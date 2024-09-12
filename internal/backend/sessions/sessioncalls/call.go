package sessioncalls

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (cs *calls) invoke(ctx context.Context, callv sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value, executors *sdkexecutor.Executors, timeout time.Duration) (sdktypes.SessionCallAttemptResult, error) {
	xid := callv.GetFunction().ExecutorID()

	var caller sdkexecutor.Caller

	f := callv.GetFunction()

	if iid := xid.ToIntegrationID(); iid.IsValid() && !f.HasFlag(sdktypes.PureFunctionFlag) && !modules.IsAKModuleExecutorID(xid) {
		// The executor is an integration, and not a pure function or a session module that must run in a session workflow,
		// so it can run in any worker and using a stateless integration.

		var err error
		if caller, err = cs.svcs.Integrations.Attach(ctx, iid); err != nil {
			return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("get integration: %w", err)
		}
	} else {
		// Function required to run in the same worker as the workflow.
		// In these cases, it's not registered in the integrations
		// service, but must exist in the executors map, since the activity
		// must have been executed in the same worker as the originating workflow.

		if executors == nil {
			return sdktypes.InvalidSessionCallAttemptResult, errors.New("no registered executors")
		}

		caller = executors.GetCaller(xid)
	}

	if caller == nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("executor not found: %q", xid)
	}

	if timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()

		if !activity.IsActivity(ctx) {
			if wctx := sessioncontext.GetWorkflowContext(ctx); wctx != nil {
				wctx, cancel := workflow.WithCancel(wctx)
				ctx = sessioncontext.WithWorkflowContext(ctx, wctx)
				go func() {
					<-ctx.Done()
					cancel()
				}()
			}
		}
	}

	v, err := caller.Call(ctx, callv, args, kwargs)
	if err != nil {
		if errors.Is(err, workflow.ErrCanceled) && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			err = context.DeadlineExceeded
		}

		if errors.Is(err, context.DeadlineExceeded) {
			v = sdktypes.InvalidValue
			err = sdktypes.NewProgramError(
				fixtures.TimeoutError,
				nil,
				map[string]string{
					"duration": timeout.String(),
				},
			).ToError()
		}
	}

	return sdktypes.NewSessionCallAttemptResult(v, err), nil
}

// This is executed either in an activity (for regular calls) or directly in a workflow (for internal calls).
func (cs *calls) executeCall(ctx context.Context, sessionID sdktypes.SessionID, seq uint32, executors *sdkexecutor.Executors) (debug any, attempt uint32, _ error) {
	z := cs.z.With(zap.Uint32("seq", seq))

	call, err := cs.svcs.DB.GetSessionCallSpec(ctx, sessionID, seq)
	if err != nil {
		return nil, 0, fmt.Errorf("db.get_call: %w", err)
	}

	callv, args, kwargs, opts, err := parseCallSpec(call)
	if err != nil {
		return nil, 0, err
	}

	z = z.With(zap.Any("function", callv))

	if attempt, err = cs.svcs.DB.StartSessionCallAttempt(ctx, sessionID, seq); err != nil {
		return nil, 0, err
	}

	z.Debug("calling")

	var result sdktypes.SessionCallAttemptResult

	func() {
		defer func() {
			if reason := recover(); reason != nil {
				result = sdktypes.NewSessionCallAttemptResult(sdktypes.InvalidValue, fmt.Errorf("panic: %v", reason))
				return
			}
		}()

		if result, err = cs.invoke(ctx, callv, args, kwargs, executors, opts.Timeout); err != nil {
			z.Panic("call integration", zap.Error(err))
		}
	}()

	if err := result.GetError(); err != nil {
		z.Debug("call returned an error", zap.Error(err))
	} else {
		z.Debug("call returned a value")
	}

	if opts.Catch {
		result = sdktypes.NewSessionCallAttemptResult(result.ToValueTuple(), nil)
	}

	// TODO: this is an error in db access inside the activity. If this fails we might be in a troubling state.
	//       need to at least manually retry this.
	if err := cs.svcs.DB.CompleteSessionCallAttempt(
		ctx, sessionID, seq, attempt,
		sdktypes.NewSessionCallAttemptComplete(true, 0, result),
	); err != nil {
		return nil, 0, err
	}

	return call, attempt, nil
}
