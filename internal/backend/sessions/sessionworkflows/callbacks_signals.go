package sessionworkflows

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const userSignalNamePrefix = "user_"

func userSignalName(name string) string               { return userSignalNamePrefix + name }
func sessionSignalName(sid sdktypes.SessionID) string { return sid.String() }

func (w *sessionWorkflow) signal(wctx workflow.Context) func(context.Context, sdktypes.RunID, sdktypes.SessionID, string, sdktypes.Value) error {
	return func(ctx context.Context, _ sdktypes.RunID, sid sdktypes.SessionID, name string, v sdktypes.Value) error {
		if activity.IsActivity(ctx) {
			return errForbiddenInActivity
		}

		_, span := w.startCallbackSpan(ctx, "signal")
		defer span.End()

		span.SetAttributes(attribute.String("name", name))

		if !v.IsValid() {
			v = sdktypes.Nothing
		}

		var f workflow.Future

		childFuture, ok := w.children[sid]
		if ok {
			f = childFuture.SignalChildWorkflow(wctx, userSignalName(name), v)
		} else {
			f = workflow.SignalExternalWorkflow(wctx, sid.String(), "", userSignalName(name), v)
		}

		if err := f.Get(wctx, nil); err != nil {
			return err
		}

		return nil
	}
}

func (w *sessionWorkflow) nextSignal(wctx workflow.Context) func(context.Context, sdktypes.RunID, []string, time.Duration) (*sdkservices.RunSignal, error) {
	return func(ctx context.Context, _ sdktypes.RunID, names []string, timeout time.Duration) (*sdkservices.RunSignal, error) {
		if activity.IsActivity(ctx) {
			return nil, errForbiddenInActivity
		}

		_, span := w.startCallbackSpan(ctx, "next_signal")
		defer span.End()

		span.SetAttributes(attribute.StringSlice("names", names), attribute.Int64("timeout", int64(timeout)))

		if len(names) == 0 {
			return nil, nil
		}

		for i, name := range names {
			if strings.HasPrefix(name, sdktypes.SessionIDKind+"_") {
				sid, err := sdktypes.ParseSessionID(name)
				if err != nil {
					return nil, sdkerrors.NewInvalidArgumentError("invalid session id %q: %w", name, err)
				}

				names[i] = sessionSignalName(sid)
			} else {
				names[i] = userSignalName(name)
			}
		}

		selector := workflow.NewSelector(wctx)

		if timeout != 0 {
			selector.AddFuture(workflow.NewTimer(wctx, timeout), func(workflow.Future) {})
		}

		var signal *sdkservices.RunSignal

		for _, name := range names {
			selector.AddReceive(workflow.GetSignalChannel(wctx, name), func(c workflow.ReceiveChannel, _ bool) {
				var v sdktypes.Value

				if !c.ReceiveAsync(&v) {
					w.l.Warn("next_signal: expected but not received", zap.String("name", name))
				}

				signal = &sdkservices.RunSignal{
					Payload: v,
					Name:    strings.TrimPrefix(name, userSignalNamePrefix),
				}
			})
		}

		// Select doesn't respond to cancellations unless we add a receive on the context done channel.
		var cancelled bool
		selector.AddReceive(wctx.Done(), func(c workflow.ReceiveChannel, _ bool) { cancelled = true })

		selector.Select(wctx)

		if cancelled {
			return nil, wctx.Err()
		}

		if signal == nil {
			return nil, nil
		}

		if !signal.Payload.IsValid() {
			signal.Payload = sdktypes.Nothing
		}

		return signal, nil
	}
}
