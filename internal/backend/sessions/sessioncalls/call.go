package sessioncalls

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (cs *calls) checkPollPolicy(ctx context.Context, pollfn, result sdktypes.Value, executors *sdkexecutor.Executors) (retry bool, interval time.Duration, err error) {
	pollRet, err := executors.Call(ctx, pollfn, []sdktypes.Value{result}, nil)
	if err != nil {
		err = fmt.Errorf("poll function: %w", err)
		return
	}

	if pollRet.IsNothing() {
		// Nothing: no retry.
		return
	}

	if isBool := pollRet.IsBoolean(); isBool {
		// True: no retry. False: immediate retry.
		return !pollRet.GetBoolean().Value(), 0, nil
	}

	// Duration: retry with specified interval.
	if interval, err = pollRet.ToDuration(); err != nil {
		err = fmt.Errorf("poller return value must be either a boolean or convertible to duration: %w", err)
		return
	}

	retry = true
	return
}

func (cs *calls) callIntegration(ctx context.Context, callv sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value, executors *sdkexecutor.Executors) (sdktypes.SessionCallAttemptResult, error) {
	xid := callv.GetFunction().ExecutorID()
	if !xid.IsIntegrationID() {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("call function must be an integration function, got %q", xid)
	}

	intg, err := cs.svcs.Integrations.Get(ctx, xid.ToIntegrationID())
	if err != nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("get integration: %w", err)
	}

	var caller sdkexecutor.Caller = intg

	if caller == nil {
		caller = executors.GetCaller(xid)
	}

	if caller == nil {
		return sdktypes.InvalidSessionCallAttemptResult, fmt.Errorf("executor not found: %q", xid)
	}

	return sdktypes.NewSessionCallAttemptResult(caller.Call(ctx, callv, args, kwargs)), nil
}

// called from `sessions.sessionCallActivity`.
func (cs *calls) executeCall(ctx context.Context, sessionID sdktypes.SessionID, seq uint32, poller sdktypes.Value, executors *sdkexecutor.Executors) (debug any, attempt uint32, _ error) {
	z := cs.z.With(zap.Uint32("seq", seq))

	call, err := cs.svcs.DB.GetSessionCallSpec(ctx, sessionID, seq)
	if err != nil {
		return nil, 0, fmt.Errorf("db.get_call: %w", err)
	}

	var (
		interval time.Duration
		last     bool
	)

	callv, args, kwargs := call.Data()

	z = z.With(zap.Any("function", callv))

	if poller.IsValid() && callv.GetFunction().HasFlag(sdktypes.DisablePollingFunctionFlag) {
		z.Debug("poller exists, but function is not pollable")
		poller = sdktypes.InvalidValue
	}

	for !last {
		z := z.With(zap.Uint32("attempt", attempt))

		attempt, err = cs.svcs.DB.StartSessionCallAttempt(ctx, sessionID, seq)
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

			if result, err = cs.callIntegration(ctx, callv, args, kwargs, executors); err != nil {
				z.Panic("call integration", zap.Error(err))
			}
		}()

		if err := result.GetError(); err != nil {
			last = true
			z.Debug("call returned an error", zap.Error(err))
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

		// TODO: this is an error in db access inside the activity. If this fails we might be in a troubling state.
		//       need to at least manually retry this.
		if err := cs.svcs.DB.CompleteSessionCallAttempt(
			ctx, sessionID, seq, attempt,
			sdktypes.NewSessionCallAttemptComplete(last, interval, result),
		); err != nil {
			return nil, attempt, err
		}
	}

	return call, attempt, nil
}
