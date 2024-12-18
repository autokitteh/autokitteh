package sessioncalls

import (
	"context"
	"errors"
	"fmt"
	"sync"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type CallParams struct {
	SessionID sdktypes.SessionID
	CallSpec  sdktypes.SessionCallSpec

	Executors *sdkexecutor.Executors // needed for session specific calls (global modules, script functions).
}

type Calls interface {
	StartWorkers(context.Context) error
	Call(ctx workflow.Context, params *CallParams) (sdktypes.SessionCallAttemptResult, error)
}

type calls struct {
	l      *zap.Logger
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
	//                 (such as session modules, in modules).
	//                 This state is the workflow state as constructed by the
	//                 user script.
	generalWorker worker.Worker
	uniqueWorker  worker.Worker

	// Pass executors for session specific calls (functions that are
	// defined in runtime script, as opposed to integrations).
	executorsForSessionsMu sync.RWMutex
	executorsForSessions   map[sdktypes.SessionID]*sdkexecutor.Executors
}

const (
	generalTaskQueueName = "session_call_activities"
)

func uniqueWorkerCallTaskQueueName() string {
	return fmt.Sprintf("%s_%s", generalTaskQueueName, fixtures.ProcessID())
}

func New(l *zap.Logger, config Config, svcs *sessionsvcs.Svcs) Calls {
	return &calls{
		l:                    l,
		config:               config,
		svcs:                 svcs,
		executorsForSessions: make(map[sdktypes.SessionID]*sdkexecutor.Executors, 16),
	}
}

func (cs *calls) StartWorkers(ctx context.Context) error {
	cs.generalWorker = temporalclient.NewWorker(cs.l.Named("sessionscallsworker"), cs.svcs.Temporal(), generalTaskQueueName, cs.config.GeneralWorker)
	cs.uniqueWorker = temporalclient.NewWorker(cs.l.Named("sessionscallsworker"), cs.svcs.Temporal(), uniqueWorkerCallTaskQueueName(), cs.config.UniqueWorker)

	cs.registerActivities()

	if w := cs.generalWorker; w != nil {
		if err := w.Start(); err != nil {
			return err
		}
	}

	if w := cs.uniqueWorker; w != nil {
		if err := w.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (cs *calls) Call(wctx workflow.Context, params *CallParams) (sdktypes.SessionCallAttemptResult, error) {
	spec := params.CallSpec
	sid := params.SessionID

	fnv, _, _ := spec.Data()
	fnvf := fnv.GetFunction()

	seq := spec.Seq()

	l := cs.l.With(zap.Any("session_id", sid), zap.Any("seq", seq), zap.Any("v", fnv))

	wctx = workflow.WithActivityOptions(wctx, cs.config.activityConfig().ToOptions(generalTaskQueueName))

	fut := workflow.ExecuteActivity(
		wctx,
		createSessionCallActivityName,
		sid,
		spec,
	)
	if err := fut.Get(wctx, nil); err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("create_call: %w", err)
	}

	var attempt uint32

	fut = workflow.ExecuteActivity(
		wctx,
		createSessionCallAttemptActivityName,
		sid,
		seq,
	)
	if err := fut.Get(wctx, &attempt); err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("create_session_call_attempt: %w", err)
	}

	var result sdktypes.SessionCallAttemptResult

	if fnvf.HasFlag(sdktypes.PureFunctionFlag) {
		goCtx := temporalclient.NewWorkflowContextAsGOContext(wctx)

		if fnvf.HasFlag(sdktypes.PrivilegedFunctionFlag) {
			goCtx = sessioncontext.WithWorkflowContext(goCtx, wctx)
		}

		var err error
		if result, err = cs.executeCall(goCtx, params.CallSpec, params.Executors); err != nil {
			l.With(zap.Error(err)).Sugar().Infof("pure call failed: %v", err)
			return sdktypes.NewSessionCallAttemptResult(sdktypes.InvalidValue, fmt.Errorf("internal call: %w", err)), nil
		}
	} else {
		xid := fnvf.ExecutorID()

		// either a non-integration (like a run) or a session module.
		unique := !xid.IsIntegrationID() || modules.IsAKModuleExecutorID(xid)

		l := l.With(zap.Bool("unique", unique))

		if unique {
			l.Info("unique call")

			// store the executors in global scope so when the unique worker activity is executed
			// is could get the session specific executors.

			cs.executorsForSessionsMu.Lock()
			cs.executorsForSessions[params.SessionID] = params.Executors
			cs.executorsForSessionsMu.Unlock()

			defer func() {
				cs.executorsForSessionsMu.Lock()
				delete(cs.executorsForSessions, sid)
				cs.executorsForSessionsMu.Unlock()
			}()
		}

		for retry := true; retry; {
			aopts := cs.config.activityConfig().ToOptions(generalTaskQueueName)

			if unique {
				// When a local execution is required, it is scheduled on a unique worker. These local executions
				// need access to the state that is built by the workflow.

				uniqueTaskQueueName := uniqueWorkerCallTaskQueueName()
				if uniqueTaskQueueName == "" {
					return sdktypes.InvalidSessionCallAttemptResult, errors.New("local session worker for activities is not registered")
				}

				cfg := cs.config.uniqueActivityConfig()
				if cfg.ScheduleToStartTimeout == 0 {
					l.Warn("call activity schedule-to-start timeout is 0 and local exec is needed")
				}

				aopts = cfg.ToOptions(uniqueTaskQueueName)
			}

			actx := workflow.WithActivityOptions(wctx, aopts)

			var ret callActivityOutputs

			future := workflow.ExecuteActivity(
				actx,
				callActivityName,
				&callActivityInputs{
					SessionID: sid,
					CallSpec:  params.CallSpec,
					Unique:    unique,
				},
			)
			if err := future.Get(wctx, &ret); err != nil {
				var terr *temporal.TimeoutError
				if ok := errors.As(err, &terr); ok && terr.TimeoutType() == enumspb.TIMEOUT_TYPE_SCHEDULE_TO_START {
					// An activity that was scheduled on a unique worker (ie the workflow worker) did not get to be started.
					// This might happen in cases of scale-down events or a recovery from a crashed worker.
					// In this case we just reshecule it again.
					l.Warn("call activity schedule to start timeout, retrying")
					continue
				}

				return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("call activity error: %w", err)
			}

			if retry = ret.Retry; retry {
				// An activity scheduled on a unique worker started to run after a crash before the associated workflow
				// got started. Since in this case the required session executors where not registered, we just
				// retry the activity and give a chance to the workflow register the session workers.
				l.Warn("call activity retrying explicitly")
				continue
			}

			result = ret.Result
		}
	}

	l.Debug("call returned", zap.Uint32("attempts", attempt+1))

	fut = workflow.ExecuteActivity(
		wctx,
		completeSessionCallAttemptActivityName,
		sid,
		seq,
		attempt,
		sdktypes.NewSessionCallAttemptComplete(true, 0, result),
	)
	if err := fut.Get(wctx, nil); err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("complete_session_call_attempt: %w", err)
	}

	return result, nil
}
