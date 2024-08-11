package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	pollOp        = "poll"
	fakeOp        = "fake"
	sleepOp       = "sleep"
	startOp       = "start"
	subscribeOp   = "subscribe"
	nextEventOp   = "next_event"
	unsubscribeOp = "unsubscribe"
)

func (w *sessionWorkflow) syscall(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if len(args) == 0 {
		return sdktypes.InvalidValue, fmt.Errorf("expecting syscall operation name as first argument")
	}

	var op string

	if err := sdktypes.DefaultValueWrapper.UnwrapInto(&op, args[0]); err != nil {
		return sdktypes.InvalidValue, err
	}

	args = args[1:]

	switch op {
	case pollOp:
		return w.setPoller(args, kwargs)
	case fakeOp:
		return w.fake(args, kwargs)
	case sleepOp:
		return w.sleep(ctx, args, kwargs)
	case startOp:
		return w.start(ctx, args, kwargs)
	case subscribeOp:
		return w.subscribe(ctx, args, kwargs)
	case nextEventOp:
		return w.nextEvent(ctx, args, kwargs)
	case unsubscribeOp:
		return w.unsubscribe(ctx, args, kwargs)
	default:
		return sdktypes.InvalidValue, fmt.Errorf("unknown op %q", op)
	}
}

func (w *sessionWorkflow) fake(args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var orig, fake sdktypes.Value

	if err := sdkmodule.UnpackArgs(args, kwargs, "orig", &orig, "fake", &fake); err != nil {
		return sdktypes.InvalidValue, err
	}

	// TODO: better make sure `orig` is an action call and `fake` is a runner call.
	if !orig.IsFunction() {
		return sdktypes.InvalidValue, fmt.Errorf("orig must be a function value")
	}

	del := !fake.IsValid() || fake.IsNothing()

	if !del && !fake.IsFunction() {
		return sdktypes.InvalidValue, fmt.Errorf("fake must be a function value")
	}

	id := orig.GetFunction().UniqueID()

	prev := w.fakers[id]
	if !prev.IsValid() {
		prev = sdktypes.Nothing
	}

	if del {
		delete(w.fakers, id)
	} else {
		w.fakers[id] = fake
	}

	return prev, nil
}

func (w *sessionWorkflow) setPoller(args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var fn sdktypes.Value

	if err := sdkmodule.UnpackArgs(args, kwargs, "fn", &fn); err != nil {
		return sdktypes.InvalidValue, err
	}

	if fn.IsValid() && !fn.IsFunction() && !fn.IsNothing() {
		return sdktypes.InvalidValue, fmt.Errorf("value must be either a function or nothing")
	}

	prev := w.poller

	if !fn.IsValid() {
		fn = sdktypes.Nothing
	}

	w.poller = fn

	return prev, nil
}

func (w *sessionWorkflow) getPoller() sdktypes.Value {
	if w.poller.IsNothing() {
		return sdktypes.InvalidValue
	}

	return w.poller
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
		WithParentSessionID(w.data.SessionID).
		WithDeploymentID(w.data.Session.DeploymentID()).
		WithEnvID(w.data.Env.ID())

	sessionID, err := w.ws.sessions.Start(ctx, session)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewStringValue(sessionID.String()), nil
}

func (w *sessionWorkflow) subscribe(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		connectionName string
		filter         string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "connection_name", &connectionName, "filter", &filter); err != nil {
		return sdktypes.InvalidValue, err
	}

	signalID, err := w.createEventSubscription(ctx, connectionName, filter)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(signalID)
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

	signals := kittehs.Transform(args, func(v sdktypes.Value) string { return v.GetString().Value() })

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

	var f workflow.Future
	if timeout != 0 {
		f = workflow.NewTimer(wctx, timeout)
	}

	for {
		// no event, wait for first signal
		signalID, err := w.waitOnFirstSignal(wctx, signals, f)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		if signalID == "" {
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

	w.removeEventSubscription(ctx, signalID)

	return sdktypes.Nothing, nil
}
