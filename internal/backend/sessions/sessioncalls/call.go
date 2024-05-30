package sessioncalls

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/akmodules"
	akmodule "go.autokitteh.dev/autokitteh/internal/backend/akmodules/ak"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (cs *calls) checkPolicyFn(ctx context.Context, fn sdktypes.Value, args []sdktypes.Value, executors *sdkexecutor.Executors) (retry bool, interval time.Duration, err error) {
	var res sdktypes.SessionCallAttemptResult
	if res, err = cs.invoke(ctx, fn, args, nil, executors, 0); err != nil {
		err = fmt.Errorf("policy function invoke: %w", err)
		return
	} else if err = res.GetError(); err != nil {
		err = fmt.Errorf("policy function error: %w", err)
		return
	}

	ret := res.GetValue()

	if ret.IsNothing() {
		// Nothing: no retry.
		return
	}

	if ret.IsBoolean() {
		// True: immediate retry. False: no retry.
		// TODO: invert for poll?
		retry = ret.GetBoolean().Value()
		return
	}

	// Duration: retry with specified interval.
	if interval, err = ret.ToDuration(); err != nil {
		err = fmt.Errorf("policy return value must be either nothing, a boolean, or convertible to duration: %w", err)
		return
	}

	retry = true
	return
}

func (cs *calls) checkPollPolicy(ctx context.Context, pollfn, result sdktypes.Value, executors *sdkexecutor.Executors) (bool, time.Duration, error) {
	return cs.checkPolicyFn(ctx, pollfn, []sdktypes.Value{result}, executors)
}

func (cs *calls) checkRetryPolicy(ctx context.Context, retryfn sdktypes.Value, err error, executors *sdkexecutor.Executors, attempt uint32) (bool, time.Duration, error) {
	return cs.checkPolicyFn(ctx, retryfn, []sdktypes.Value{sdktypes.WrapError(err).Value(), sdktypes.NewIntegerValue(attempt)}, executors)
}

func (cs *calls) invoke(ctx context.Context, callv sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value, executors *sdkexecutor.Executors, timeout time.Duration) (sdktypes.SessionCallAttemptResult, error) {
	xid := callv.GetFunction().ExecutorID()

	var caller sdkexecutor.Caller

	f := callv.GetFunction()

	if iid := xid.ToIntegrationID(); iid.IsValid() && !f.HasFlag(sdktypes.PureFunctionFlag) && !akmodules.IsAKModuleExecutorID(xid) {
		// The executor is an integration, and not a pure function or a session module that must run in a session workflow,
		// so it can run in any worker and using a stateless integration.

		var err error
		if caller, err = cs.svcs.Integrations.GetByID(ctx, iid); err != nil {
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
				akmodule.TimeoutError,
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
func (cs *calls) executeCall(
	ctx context.Context,
	params *executeParams,
	executors *sdkexecutor.Executors,
) (debug any, attempt uint32, _ error) {
	seq, sid := params.Seq, params.SessionID

	z := cs.z.With(zap.Uint32("seq", seq))

	call, err := cs.svcs.DB.GetSessionCallSpec(ctx, sid, seq)
	if err != nil {
		return nil, 0, fmt.Errorf("db.get_call: %w", err)
	}

	callv, args, kwargs, opts, err := parseCallSpec(call)
	if err != nil {
		return nil, 0, err
	}

	z = z.With(zap.Any("function", callv))

	var poller, retryer sdktypes.Value

	if opts.Poll.IsValid() {
		poller = opts.Poll
	}

	if opts.Retry.IsValid() {
		retryer = opts.Retry
	}

	if poller.IsValid() && callv.GetFunction().HasFlag(sdktypes.DisablePollingFunctionFlag) {
		z.Debug("poller exists, but function is not pollable")
		poller = sdktypes.InvalidValue
	}

	if retryer.IsValid() && callv.GetFunction().HasFlag(sdktypes.DisableRetryingFunctionFlag) {
		z.Debug("poller exists, but function is not pollable")
		retryer = sdktypes.InvalidValue
	}

	for interval, last := time.Duration(0), false; !last; {
		z := z.With(zap.Uint32("attempt", attempt))

		attempt, err = cs.svcs.DB.StartSessionCallAttempt(ctx, sid, seq)
		if err != nil {
			return nil, attempt, err
		}

		if interval != 0 {
			z.Debug("waiting", zap.Duration("interval", interval))
			select {
			case <-time.After(interval):
				// nop
			case <-ctx.Done():
				return nil, attempt, ctx.Err()
			}
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
			last = !retryer.IsValid()
			z.Debug("call returned an error", zap.Error(err))

			if !last {
				var retry bool

				if retry, interval, err = cs.checkRetryPolicy(ctx, retryer, err, executors, attempt); err != nil {
					result = sdktypes.NewSessionCallAttemptResult(sdktypes.InvalidValue, fmt.Errorf("retry function: %w", err))
					last = true
				} else {
					last = !retry
				}

				z.Debug("retryer results", zap.Bool("last", last), zap.Duration("t", interval))
			}
		} else {
			z.Debug("call returned a value")

			if !poller.IsValid() {
				last = true
			} else {
				var retry bool

				if retry, interval, err = cs.checkPollPolicy(ctx, poller, result.GetValue(), executors); err != nil {
					result = sdktypes.NewSessionCallAttemptResult(sdktypes.InvalidValue, fmt.Errorf("poll function: %w", err))
					last = true
				} else {
					last = !retry
				}

				z.Debug("poll results", zap.Bool("last", last), zap.Duration("t", interval))
			}
		}

		if opts.Catch {
			result = sdktypes.NewSessionCallAttemptResult(result.ToValueTuple(), nil)
		}

		// TODO: this is an error in db access inside the activity. If this fails we might be in a troubling state.
		//       need to at least manually retry this.
		if err := cs.svcs.DB.CompleteSessionCallAttempt(
			ctx, sid, seq, attempt,
			sdktypes.NewSessionCallAttemptComplete(last, interval, result),
		); err != nil {
			return nil, attempt, err
		}
	}

	return call, attempt, nil
}
