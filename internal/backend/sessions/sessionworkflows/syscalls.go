package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	sleepOp       = "sleep"
	startOp       = "start"
	subscribeOp   = "subscribe"
	nextEventOp   = "next_event"
	unsubscribeOp = "unsubscribe"
)

func (w *sessionWorkflow) syscalls() *sdkservices.RunSyscalls {
	call := func(ctx context.Context, name string, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
		v := w.akModule.Values()[name]

		wctx := sessioncontext.GetWorkflowContext(ctx)
		if wctx == nil {
			return sdktypes.InvalidValue, errors.New("missing workflow context")
		}

		return w.call(wctx, v, args, kwargs)
	}

	return &sdkservices.RunSyscalls{
		Sleep: func(ctx context.Context, t time.Duration) error {
			_, err := call(ctx, sleepOp, []sdktypes.Value{sdktypes.NewDurationValue(t)}, nil)
			return err
		},
		Start: func(ctx context.Context, entrypoint sdktypes.CodeLocation, inputs map[string]sdktypes.Value, memo map[string]string) (string, error) {
			v, err := call(
				ctx,
				startOp,
				nil,
				map[string]sdktypes.Value{
					"loc":    sdktypes.NewStringValue(entrypoint.String()),
					"inputs": sdktypes.NewDictValueFromStringMap(inputs),
					"memo":   sdktypes.NewDictValueFromStringMap(kittehs.TransformMapValues(memo, sdktypes.NewStringValue)),
				},
			)
			if err != nil {
				return "", err
			}

			return v.GetString().Value(), nil
		},
		Subscribe: func(ctx context.Context, name, filter string) (uuid.UUID, error) {
			v, err := call(
				ctx,
				subscribeOp,
				nil,
				map[string]sdktypes.Value{
					"name":   sdktypes.NewStringValue(name),
					"filter": sdktypes.NewStringValue(filter),
				},
			)
			if err != nil {
				return uuid.Nil, err
			}

			return uuid.Parse(v.GetString().Value())
		},
		Unsubscribe: func(ctx context.Context, signalID uuid.UUID) error {
			_, err := call(
				ctx,
				unsubscribeOp,
				[]sdktypes.Value{sdktypes.NewStringValue(signalID.String())},
				nil,
			)

			return err
		},
		NextEvent: func(ctx context.Context, subscriptions []uuid.UUID, t time.Duration) (sdktypes.Value, error) {
			return call(
				ctx,
				nextEventOp,
				kittehs.Transform(subscriptions, func(id uuid.UUID) sdktypes.Value { return sdktypes.NewStringValue(id.String()) }),
				map[string]sdktypes.Value{
					"timeout": sdktypes.NewDurationValue(t),
				},
			)
		},
	}
}

func (w *sessionWorkflow) sleep(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var duration time.Duration

	if err := sdkmodule.UnpackArgs(args, kwargs, "duration", &duration); err != nil {
		return sdktypes.InvalidValue, err
	}

	wctx := sessioncontext.GetWorkflowContext(ctx)

	if err := workflow.Sleep(wctx, duration); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.Nothing, nil
}

func (w *sessionWorkflow) start(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		loc    string
		inputs map[string]sdktypes.Value
		memo   map[string]string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "loc", &loc, "inputs?", &inputs, "memo=?", &memo); err != nil {
		return sdktypes.InvalidValue, err
	}

	cl, err := sdktypes.ParseCodeLocation(loc)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("invalid location: %w", err)
	}

	session := sdktypes.NewSession(w.data.Build.ID(), cl, inputs, memo).
		WithParentSessionID(w.data.Session.ID()).
		WithDeploymentID(w.data.Session.DeploymentID()).
		WithEnvID(w.data.Env.ID())

	sessionID, err := w.ws.sessions.Start(ctx, session)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewStringValue(sessionID.String()), nil
}

func (w *sessionWorkflow) subscribe(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var name, filter string

	if err := sdkmodule.UnpackArgs(args, kwargs, "name", &name, "filter?", &filter); err != nil {
		return sdktypes.InvalidValue, err
	}

	_, connection := kittehs.FindFirst(w.data.Connections, func(c sdktypes.Connection) bool {
		return c.Name().String() == name
	})

	_, trigger := kittehs.FindFirst(w.data.Triggers, func(t sdktypes.Trigger) bool {
		return t.Name().String() == name
	})

	if connection.IsValid() && trigger.IsValid() {
		return sdktypes.InvalidValue, errors.New("ambigous source name - matching both a connection and a trigger")
	}

	var did sdktypes.EventDestinationID

	if connection.IsValid() {
		did = sdktypes.NewEventDestinationID(connection.ID())
	} else if trigger.IsValid() {
		did = sdktypes.NewEventDestinationID(trigger.ID())
	} else {
		return sdktypes.InvalidValue, fmt.Errorf("source %q not found", name)
	}

	signalID, err := w.createEventSubscription(sessioncontext.GetWorkflowContext(ctx), filter, did)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewStringValue(signalID.String()), nil
}

/*
Algorithm:
Inputs: signal_id or signals which is a list of signal ids
If none or both provided, raise an error
If signal_id provided, convert to signals list
For each signal in signals list:

	If event exists, return it - this will increment this event's sequence counting

If no event found, wait on first signal in signals list
Get the next event on this signal (has to exists since we got a signal on it)
return this event
*/
func (w *sessionWorkflow) nextEvent(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var timeout time.Duration
	if err := sdkmodule.UnpackArgs(nil, kwargs, "timeout=?", &timeout); err != nil {
		return sdktypes.InvalidValue, err
	}

	if len(args) == 0 {
		return sdktypes.InvalidValue, errors.New("expecting at least one signal")
	}

	if timeout < 0 {
		return sdktypes.InvalidValue, errors.New("timeout must be a non-negative value")
	}

	signals, err := kittehs.TransformError(args, func(v sdktypes.Value) (uuid.UUID, error) {
		return uuid.Parse(v.GetString().Value())
	})
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// check if there is an event already in one of the signals
	for _, signalID := range signals {
		event, err := w.getNextEvent(ctx, signalID)
		if err != nil {
			return sdktypes.InvalidValue, err
		}
		if event != nil {
			return sdktypes.WrapValue(event)
		}
	}

	wctx := sessioncontext.GetWorkflowContext(ctx)

	var timeoutFuture workflow.Future
	if timeout != 0 {
		timeoutFuture = workflow.NewTimer(wctx, timeout)
	}

	for {
		// no event, wait for first signal
		signalID, err := w.waitOnFirstSignal(wctx, signals, timeoutFuture)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		if signalID == uuid.Nil {
			return sdktypes.Nothing, nil
		}

		// get next event on this signal
		event, err := w.getNextEvent(ctx, signalID)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		if event != nil {
			return sdktypes.WrapValue(event)
		}
	}
}

func (w *sessionWorkflow) unsubscribe(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var signalID string

	if err := sdkmodule.UnpackArgs(args, kwargs, "signal_id", &signalID); err != nil {
		return sdktypes.InvalidValue, err
	}

	uuid, err := uuid.Parse(signalID)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	w.removeEventSubscription(ctx, uuid)

	return sdktypes.Nothing, nil
}
