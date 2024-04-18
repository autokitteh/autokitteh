package sessioncalls

import (
	"context"
	"errors"
	"fmt"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/akmodules"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var localRetryPolicy = kittehs.RetryPolicy{
	MaxAttempts: 3,
}

// Pass executors for session specific calls (functions that are
// defined in runtime script, as opposed to integrations).
var executorsForSessions = make(map[string]*sdkexecutor.Executors, 16)

type CallParams struct {
	SessionID     sdktypes.SessionID
	Debug         bool
	ForceInternal bool
	CallSpec      sdktypes.SessionCallSpec

	Poller    sdktypes.Value         // TODO: need to be in Call.
	Executors *sdkexecutor.Executors // needed for session specific calls (global modules, script functions).
}

type Calls interface {
	StartWorkers(context.Context) error
	Call(ctx workflow.Context, params *CallParams) (sdktypes.SessionCallAttemptResult, error)
}

type calls struct {
	z      *zap.Logger
	config Config
	svcs   *sessionsvcs.Svcs

	// For every instance there are two worker pools created:
	// - generalWorker: used only for activities that do not need to run on
	//                  a specific instance. These activities are only used
	//                  to invoke integrations calls that are not built-in.
	// - uniqueWorker: used only for activities that need to run on a specific
	//                 instance. These activities have to run on these instances
	//                 since they must be able to access instance specific state
	//                 or available to run on the same worker as the workflow
	//                 (such as session modules, in akmodules).
	//                 This state is the workflow state as constructed by the
	//                 user script.
	generalWorker worker.Worker
	uniqueWorker  worker.Worker
}

const (
	taskQueueName    = "session_call_activities"
	callActivityName = "session_call_activity"
)

func uniqueWorkerCallTaskQueueName() string {
	return fmt.Sprintf("%s_%s", taskQueueName, fixtures.ProcessID())
}

func New(z *zap.Logger, config Config, svcs *sessionsvcs.Svcs) Calls {
	opts := config.Temporal.Worker
	opts.DisableRegistrationAliasing = true
	opts.OnFatalError = func(err error) { z.Error("temporal worker error", zap.Error(err)) }
	opts.DisableWorkflowWorker = true // these workers serve only activities.

	cs := calls{
		z:             z,
		config:        config,
		svcs:          svcs,
		generalWorker: worker.New(svcs.Temporal, taskQueueName, opts),
		uniqueWorker:  worker.New(svcs.Temporal, uniqueWorkerCallTaskQueueName(), opts),
	}

	cs.generalWorker.RegisterActivityWithOptions(
		cs.sessionCallActivity,
		activity.RegisterOptions{
			Name:                          callActivityName,
			DisableAlreadyRegisteredCheck: true,
		},
	)

	cs.uniqueWorker.RegisterActivityWithOptions(
		cs.sessionCallActivity,
		activity.RegisterOptions{
			Name:                          callActivityName,
			DisableAlreadyRegisteredCheck: true,
		},
	)

	return &cs
}

func (cs *calls) StartWorkers(ctx context.Context) error {
	if err := cs.generalWorker.Start(); err != nil {
		return err
	}

	return cs.uniqueWorker.Start()
}

func (cs *calls) createSessionCall(ctx context.Context, sessionID sdktypes.SessionID, spec sdktypes.SessionCallSpec, t time.Time) (already bool, err error) {
	err = cs.svcs.DB.CreateSessionCall(ctx, sessionID, spec, t)
	if err != nil && errors.Is(err, sdkerrors.ErrAlreadyExists) {
		err = nil
		already = true
	}

	return
}

func (cs *calls) getLastSessionCallAttemptResult(wctx workflow.Context, sessionID sdktypes.SessionID, seq uint32) (result sdktypes.SessionCallAttemptResult, err error) {
	goCtx := temporalclient.NewWorkflowContextAsGOContext(wctx)

	err = localRetryPolicy.Execute(goCtx, func(int) (err error) {
		// Do not execute in an activity: we don't want to expose the activity result to Temporal.
		result, err = cs.svcs.DB.GetSessionCallAttemptResult(goCtx, sessionID, seq, -1)
		return
	})

	return result, err
}

func (cs *calls) Call(ctx workflow.Context, params *CallParams) (sdktypes.SessionCallAttemptResult, error) {
	spec := params.CallSpec

	fnv, _, _ := spec.Data()
	fnvf := fnv.GetFunction()

	seq := spec.Seq()

	z := cs.z.With(zap.String("session_id", params.SessionID.String()), zap.Uint32("seq", seq), zap.Any("v", fnv))

	var already bool

	if err := workflow.ExecuteLocalActivity(
		ctx, cs.createSessionCall, params.SessionID, spec, workflow.Now(ctx),
	).Get(ctx, &already); err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("db.create_call: %w", err)
	}

	if already {
		hasPoller := params.Poller.IsValid()

		z.Debug("call already began", zap.Bool("has_poller", hasPoller))

		// If call has a poller, no problem running the call again and let
		// the poller function sort out if it's done. Otherwise, if no poller
		// and the call finished, we will just return the result.
		if !hasPoller {
			result, err := cs.getLastSessionCallAttemptResult(ctx, params.SessionID, seq)
			if err != nil {
				return sdktypes.InvalidSessionCallAttemptResult, err
			}

			if result.IsValid() {
				z.Debug("call already executed", zap.Any("result", result))
				return result, nil
			}
		}
	}

	var attempt uint32

	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	if params.ForceInternal || fnvf.HasFlag(sdktypes.PureFunctionFlag) {
		z.Debug("function call outside activity")

		var err error

		if fnvf.HasFlag(sdktypes.PrivilidgedFunctionFlag) {
			goCtx = sessioncontext.WithWorkflowContext(goCtx, ctx)
		}

		if _, attempt, err = cs.executeCall(goCtx, params.SessionID, seq, params.Poller, params.Executors, func() time.Time {
			return workflow.Now(ctx)
		}); err != nil {
			return sdktypes.NewSessionCallAttemptResult(sdktypes.InvalidValue, fmt.Errorf("internal call: %w", err)), nil
		}
	} else {
		xid := fnvf.ExecutorID()

		// either a non-integration (like a run) or a session module.
		local := !xid.IsIntegrationID() || akmodules.IsAKModuleExecutorID(xid)

		if local {
			executorsForSessions[params.SessionID.String()] = params.Executors
			defer func() { delete(executorsForSessions, params.SessionID.String()) }()
		}

		for retry := true; retry; {
			aopts := workflow.ActivityOptions{
				TaskQueue:              taskQueueName,
				ActivityID:             fmt.Sprintf("session_call_%s_%d", params.SessionID.Value(), seq),
				ScheduleToCloseTimeout: cs.config.Temporal.ActivityScheduleToCloseTimeout,
				HeartbeatTimeout:       cs.config.Temporal.ActivityHeartbeatTimeout,
			}

			if fnvf.HasFlag(sdktypes.ShortHeartbeatTimeout) && cs.config.Temporal.ShortActivityHeartbeatTimeout > 0 {
				aopts.HeartbeatTimeout = cs.config.Temporal.ShortActivityHeartbeatTimeout
			}

			if local {
				// When a local execution is required, it is scheduled on a unique worker. These local executions
				// need access to the state that is built by the workflow.

				if aopts.TaskQueue = uniqueWorkerCallTaskQueueName(); aopts.TaskQueue == "" {
					return sdktypes.InvalidSessionCallAttemptResult, errors.New("local session worker for activities is not registerd")
				}

				aopts.ScheduleToStartTimeout = cs.config.Temporal.ActivityScheduleToStartTimeout
				if aopts.ScheduleToStartTimeout == 0 {
					z.Warn("call activity schedule-to-start timeout is 0 and local exec is needed")
				}

				aopts.ActivityID = "local_" + aopts.ActivityID
			}

			actx := workflow.WithActivityOptions(ctx, aopts)

			var ret callActivityOutputs

			future := workflow.ExecuteActivity(
				actx,
				callActivityName,
				&callActivityInputs{
					SessionID:     params.SessionID,
					Seq:           seq,
					Debug:         params.Debug,
					Poller:        params.Poller,
					AutoHeartbeat: !fnvf.HasFlag(sdktypes.DisableAutoHeartbeatFlag),
				},
			)
			if err := future.Get(ctx, &ret); err != nil {
				var terr *temporal.TimeoutError
				if ok := errors.As(err, &terr); ok && terr.TimeoutType() == enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START {
					// An activity that was scheduled on a unique worker (ie the workflow worker) did not get to be started.
					// This might happen in cases of scale-down events or a recovery from a crashed worker.
					// In this case we just reshecule it again.
					z.Warn("call activity schedule to start timeout, retrying")
					continue
				}

				return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("call activity error: %w", err)
			}

			attempt = ret.Attempt

			if retry = ret.Retry; retry {
				// An activity scheduled on a unique worker started to run after a crash before the associated workflow
				// got started. Since in this case the required session executors where not registered, we just
				// retry the activity and give a chance to the workflow register the session workers.
				z.Warn("call activity retrying explicitly")
			}
		}
	}

	result, err := cs.getLastSessionCallAttemptResult(ctx, params.SessionID, seq)
	if err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, err
	}

	z.Debug("call returned", zap.Uint32("attempts", attempt+1), zap.Any("result", result))

	return result, nil
}
