package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/backend/internal/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
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
		return nil, fmt.Errorf("expecting syscall operation name as first argument")
	}

	var op string

	if err := sdkvalues.DefaultValueWrapper.UnwrapInto(&op, args[0]); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("unknown op %q", op)
	}
}

func (w *sessionWorkflow) fake(args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var orig, fake sdktypes.Value

	if err := sdkmodule.UnpackArgs(args, kwargs, "orig", &orig, "fake", &fake); err != nil {
		return nil, err
	}

	// TODO: better make sure `orig` is an action call and `fake` is a runner call.
	if !sdktypes.IsFunctionValue(orig) {
		return nil, fmt.Errorf("orig must be a function value")
	}

	del := fake == nil || sdktypes.IsNothingValue(fake)

	if !del && !sdktypes.IsFunctionValue(fake) {
		return nil, fmt.Errorf("fake must be a function value")
	}

	id := sdktypes.GetFunctionValueUniqueID(orig)

	prev := w.fakers[id]
	if prev == nil {
		prev = sdktypes.NewNothingValue()
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
		return nil, err
	}

	if fn != nil && !sdktypes.IsFunctionValue(fn) && !sdktypes.IsNothingValue(fn) {
		return nil, fmt.Errorf("value must be either a function or nothing")
	}

	prev := w.poller

	if fn == nil {
		fn = sdktypes.NewNothingValue()
	}

	w.poller = fn

	return prev, nil
}

func (w *sessionWorkflow) getPoller() sdktypes.Value {
	if sdktypes.IsNothingValue(w.poller) {
		return nil
	}

	return w.poller
}

func (w *sessionWorkflow) sleep(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var duration time.Duration

	if err := sdkmodule.UnpackArgs(args, kwargs, "duration", &duration); err != nil {
		return nil, err
	}

	wctx := sessioncontext.GetWorkflowContext(ctx)

	if err := workflow.Sleep(wctx, duration); err != nil {
		return nil, err
	}

	return sdktypes.NewNothingValue(), nil
}

func (w *sessionWorkflow) start(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		loc  string
		data map[string]sdktypes.Value
		memo map[string]string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "loc", &loc, "data?", &data, "memo?", &memo); err != nil {
		return nil, err
	}

	cl, err := sdktypes.ParseCodeLocation(loc)
	if err != nil {
		return nil, fmt.Errorf("invalid location: %w", err)
	}

	session := sdktypes.NewSession(
		sdktypes.GetDeploymentID(w.data.Deployment),
		w.data.SessionID,
		nil,
		cl,
		data,
		memo,
	)

	sessionID, err := w.ws.sessions.Start(ctx, session)
	if err != nil {
		return nil, err
	}

	return sdktypes.NewStringValue(sessionID.String()), nil
}

func (w *sessionWorkflow) subscribe(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		connectionName string
		eventName      string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "connection_name", &connectionName, "event_name", &eventName); err != nil {
		return nil, err
	}

	signalID, err := w.createEventSubscription(ctx, connectionName, eventName)
	if err != nil {
		return nil, err
	}

	return sdkvalues.DefaultValueWrapper.Wrap(signalID)
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
	if len(kwargs) > 0 {
		return nil, errors.New("unexpected keyword arguments")
	}

	if len(args) == 0 {
		return nil, errors.New("expecting at least one signal")
	}

	signals := kittehs.Transform(args, func(v sdktypes.Value) string { return sdktypes.GetStringValue(v) })

	// check if there is an event already in one of the signals
	for _, signalID := range signals {
		event, err := w.getNextEvent(ctx, signalID)
		if err != nil {
			return nil, err
		}
		if event != nil {
			return sdkvalues.DefaultValueWrapper.Wrap(event)
		}
	}

	// no event, wait for first signal
	signalID, err := w.waitOnFirstSignal(ctx, signals)
	if err != nil {
		return nil, err
	}

	// get next event on this signal
	event, err := w.getNextEvent(ctx, signalID)
	if err != nil {
		return nil, err
	}

	// should never happen, since we got a signal on this signalID
	// meaning there is an event waiting for us
	if event == nil {
		return nil, fmt.Errorf("no event received")
	}

	return sdkvalues.DefaultValueWrapper.Wrap(event)
}

func (w *sessionWorkflow) unsubscribe(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var signalID string

	if err := sdkmodule.UnpackArgs(args, kwargs, "signal_id", &signalID); err != nil {
		return nil, err
	}

	w.removeEventSubscription(ctx, signalID)

	return sdktypes.NewNothingValue(), nil
}
