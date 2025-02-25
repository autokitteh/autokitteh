package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	sleepOp       = "sleep"
	startOp       = "start"
	subscribeOp   = "subscribe"
	nextEventOp   = "next_event"
	unsubscribeOp = "unsubscribe"
	signalOp      = "signal"
	nextSignalOp  = "next_signal"
)

func (w *sessionWorkflow) syscall(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if len(args) == 0 {
		return sdktypes.InvalidValue, errors.New("expecting syscall operation name as first argument")
	}

	var op string

	if err := sdktypes.DefaultValueWrapper.UnwrapInto(&op, args[0]); err != nil {
		return sdktypes.InvalidValue, err
	}

	args = args[1:]

	switch op {
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
	case signalOp:
		return w.signal(ctx, args, kwargs)
	case nextSignalOp:
		return w.nextSignal(ctx, args, kwargs)
	default:
		return sdktypes.InvalidValue, fmt.Errorf("unknown op %q", op)
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
	w.l.Info("syscalls:start", zap.Any("fn", args[0]))
	var (
		loc    string
		inputs map[string]sdktypes.Value
		memo   map[string]string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "loc", &loc, "inputs?", &inputs, "memo?", &memo); err != nil {
		return sdktypes.InvalidValue, err
	}

	cl, err := sdktypes.ParseCodeLocation(loc)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("invalid location: %w", err)
	}

	session := sdktypes.NewSession(w.data.Build.ID(), cl, inputs, memo).
		WithParentSessionID(w.data.Session.ID()).
		WithDeploymentID(w.data.Session.DeploymentID()).
		WithProjectID(w.data.Session.ProjectID())

	sessionID, err := w.ws.sessions.Start(authcontext.SetAuthnSystemUser(ctx), session)
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
		return sdktypes.InvalidValue, errors.New("ambiguous source name - matching both a connection and a trigger")
	}

	var did sdktypes.EventDestinationID

	switch {
	case connection.IsValid():
		did = sdktypes.NewEventDestinationID(connection.ID())
	case trigger.IsValid():
		did = sdktypes.NewEventDestinationID(trigger.ID())
	default:
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

const userSignalNamePrefix = "user_"

func userSignalName(name string) string               { return userSignalNamePrefix + name }
func sessionSignalName(sid sdktypes.SessionID) string { return sid.String() }

type signal struct {
	Source sdktypes.SessionID `json:"source"`
	Value  sdktypes.Value     `json:"value"`
}

func (w *sessionWorkflow) signal(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		v    sdktypes.Value
		sid  string
		name string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "session_id", &sid, "name", &name, "value?", &v); err != nil {
		return sdktypes.InvalidValue, err
	}

	wctx := sessioncontext.GetWorkflowContext(ctx)

	if _, err := sdktypes.ParseSessionID(sid); err != nil {
		return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("invalid session_id: %w", err)
	}

	if !v.IsValid() {
		v = sdktypes.Nothing
	}

	signal := signal{Source: w.data.Session.ID(), Value: v}

	if err := workflow.SignalExternalWorkflow(wctx, sid, "", userSignalName(name), &signal).Get(wctx, nil); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.Nothing, nil
}

func (w *sessionWorkflow) nextSignal(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var timeout time.Duration

	if err := sdkmodule.UnpackArgs(nil, kwargs, "timeout=?", &timeout); err != nil {
		return sdktypes.InvalidValue, err
	}

	if len(args) == 0 {
		return sdktypes.Nothing, nil
	}

	var names []string
	for _, arg := range args {
		if !arg.IsString() {
			return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("expecting signal name as string, got %q", arg)
		}

		v := arg.GetString().Value()
		name := userSignalName(v)

		if strings.HasPrefix(arg.GetString().Value(), sdktypes.SessionIDKind+"_") {
			sid, err := sdktypes.ParseSessionID(v)
			if err != nil {
				return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("invalid session_id %q: %w", arg, err)
			}

			name = sessionSignalName(sid)
		}

		names = append(names, name)
	}

	wctx := sessioncontext.GetWorkflowContext(ctx)

	selector := workflow.NewSelector(wctx)

	if timeout != 0 {
		selector.AddFuture(workflow.NewTimer(wctx, timeout), func(workflow.Future) {})
	}

	var (
		signal signal
		rxName string
	)

	for _, name := range names {
		selector.AddReceive(workflow.GetSignalChannel(wctx, name), func(c workflow.ReceiveChannel, _ bool) {
			rxName = name

			if !c.ReceiveAsync(&signal) {
				w.l.Warn("next_signal: expected but not received", zap.String("name", name))
			}
		})
	}

	// Select doesn't respond to cancellations unless we add a receive on the context done channel.
	var cancelled bool
	selector.AddReceive(wctx.Done(), func(c workflow.ReceiveChannel, _ bool) { cancelled = true })

	selector.Select(wctx)

	if cancelled {
		return sdktypes.InvalidValue, wctx.Err()
	}

	if !signal.Source.IsValid() {
		return sdktypes.Nothing, nil
	}

	if !signal.Value.IsValid() {
		signal.Value = sdktypes.Nothing
	}

	if strings.HasPrefix(rxName, userSignalNamePrefix) {
		rxName = rxName[len(userSignalNamePrefix):]
	}

	return sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
		"name":   sdktypes.NewStringValue(rxName),
		"source": sdktypes.NewStringValue(signal.Source.String()),
		"value":  signal.Value,
	}), nil
}
