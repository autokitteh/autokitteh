package sessioncalls

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/backend/internal/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/backend/internal/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type CallParams struct {
	SessionID     sdktypes.SessionID
	Debug         bool
	ForceInternal bool
	CallSpec      sdktypes.SessionCallSpec

	Poller    sdktypes.Value         // TODO: need to be in Call.
	Executors *sdkexecutor.Executors // HACK: needed for session specific calls (builtins, script functions).
}

type Calls interface {
	StartWorkers(context.Context) error
	Call(ctx workflow.Context, params *CallParams) (sdktypes.SessionCallAttemptResult, error)
}

type calls struct {
	z      *zap.Logger
	config Config
	worker worker.Worker
	svcs   *sessionsvcs.Svcs
}

const (
	taskQueueName    = "session_call_activities"
	callActivityName = "session_call_activity"
)

func New(z *zap.Logger, config Config, svcs *sessionsvcs.Svcs) Calls {
	opts := config.Temporal.Worker
	opts.DisableRegistrationAliasing = true
	opts.OnFatalError = func(err error) { z.Error("temporal worker error", zap.Error(err)) }

	worker := worker.New(svcs.Temporal, taskQueueName, opts)

	cs := calls{z: z, config: config, worker: worker, svcs: svcs}

	worker.RegisterActivityWithOptions(
		cs.sessionCallActivity,
		activity.RegisterOptions{
			Name:                          callActivityName,
			DisableAlreadyRegisteredCheck: true,
		},
	)

	return &cs
}

func (cs *calls) StartWorkers(ctx context.Context) error { return cs.worker.Start() }

func (cs *calls) Call(ctx workflow.Context, params *CallParams) (sdktypes.SessionCallAttemptResult, error) {
	fnv, _, _ := params.CallSpec.Data()

	seq := params.CallSpec.Seq()

	z := cs.z.With(zap.String("session_id", params.SessionID.String()), zap.Uint32("seq", seq), zap.Any("v", fnv))

	// TODO: If replaying, make sure arguments are the same?
	if err := workflow.ExecuteLocalActivity(
		ctx, cs.svcs.DB.CreateSessionCall, params.SessionID, params.CallSpec,
	).Get(ctx, nil); err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("db.create_call: %w", err)
	}

	var attempt uint32

	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	if params.ForceInternal || fnv.GetFunction().HasFlag(sdktypes.PureFunctionFlag) {
		z.Debug("function call outside activity")

		var err error

		if fnv.GetFunction().HasFlag(sdktypes.PrivilidgedFunctionFlag) {
			goCtx = sessioncontext.WithWorkflowContext(goCtx, ctx)
		}

		if _, attempt, err = cs.executeCall(goCtx, params.SessionID, seq, params.Poller, params.Executors); err != nil {
			return sdktypes.NewSessionCallAttemptResult(sdktypes.InvalidValue, fmt.Errorf("internal call: %w", err)), nil
		}
	} else {
		executorsForSessions[params.SessionID.String()] = params.Executors // HACK

		actx := workflow.WithActivityOptions(
			ctx,
			workflow.ActivityOptions{
				TaskQueue:              taskQueueName,
				ActivityID:             fmt.Sprintf("session_call_%s_%d", params.SessionID.Value(), seq),
				ScheduleToCloseTimeout: cs.config.Temporal.ActivityScheduleToCloseTimeout,
				HeartbeatTimeout:       cs.config.Temporal.ActivityHeartbeatTimeout,
			},
		)

		ret := callActivityOutputs{Retry: true}

		for ret.Retry {
			if err := workflow.ExecuteActivity(
				actx,
				callActivityName,
				&callActivityInputs{
					SessionID: params.SessionID,
					Seq:       seq,
					Debug:     params.Debug,
					Poller:    params.Poller,
				},
			).Get(ctx, &ret); err != nil {
				return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("call activity error: %w", err)
			}

			if ret.Retry {
				z.Warn("call activity retrying explicitly")
			}
		}

		attempt = ret.Attempt
	}

	// Do not execute in an activity: we don't want to expose the activity result to Temporal.
	result, err := cs.svcs.DB.GetSessionCallAttemptResult(goCtx, params.SessionID, seq, int64(attempt))
	if err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("db.get_session_call_attempt: %w", err)
	}

	z.Debug("call returned", zap.Uint32("attempts", attempt+1), zap.Any("result", result))

	return result, nil
}
