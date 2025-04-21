package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) subscribe(wctx workflow.Context) func(context.Context, sdktypes.RunID, string, string) (string, error) {
	return func(ctx context.Context, rid sdktypes.RunID, name, filter string) (string, error) {
		_, span := w.startCallbackSpan(ctx, "subscribe")
		defer span.End()

		_, connection := kittehs.FindFirst(w.data.Connections, func(c sdktypes.Connection) bool {
			return c.Name().String() == name
		})

		_, trigger := kittehs.FindFirst(w.data.Triggers, func(t sdktypes.Trigger) bool {
			return t.Name().String() == name
		})

		var did sdktypes.EventDestinationID

		switch {
		case connection.IsValid() && trigger.IsValid():
			return "", errors.New("ambiguous source name - matching both a connection and a trigger")
		case connection.IsValid():
			did = sdktypes.NewEventDestinationID(connection.ID())
		case trigger.IsValid():
			did = sdktypes.NewEventDestinationID(trigger.ID())
		default:
			return "", fmt.Errorf("source %q not found", name)
		}

		signalID, err := w.createEventSubscription(wctx, filter, did)
		if err != nil {
			return "", err
		}

		return signalID.String(), nil
	}
}

func (w *sessionWorkflow) unsubscribe(wctx workflow.Context) func(context.Context, sdktypes.RunID, string) error {
	return func(ctx context.Context, rid sdktypes.RunID, signalID string) error {
		_, span := w.startCallbackSpan(ctx, "unsubscribe")
		defer span.End()

		uuid, err := uuid.Parse(signalID)
		if err != nil {
			return err
		}

		w.removeEventSubscription(wctx, uuid)

		return nil
	}
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
func (w *sessionWorkflow) nextEvent(wctx workflow.Context) func(context.Context, sdktypes.RunID, []string, time.Duration) (sdktypes.Value, error) {
	return func(ctx context.Context, rid sdktypes.RunID, signals []string, timeout time.Duration) (sdktypes.Value, error) {
		_, span := w.startCallbackSpan(ctx, "next_event")
		defer span.End()

		if len(signals) == 0 {
			return sdktypes.Nothing, nil
		}

		signalUUIDs, err := kittehs.TransformError(signals, uuid.Parse)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		// check if there is an event already in one of the signals
		for _, signalID := range signalUUIDs {
			event, err := w.getNextEvent(wctx, signalID)
			if err != nil {
				return sdktypes.InvalidValue, err
			}
			if event != nil {
				return sdktypes.WrapValue(event)
			}
		}

		var timeoutFuture workflow.Future
		if timeout != 0 {
			timeoutFuture = workflow.NewTimer(wctx, timeout)
		}

		for {
			// no event, wait for first signal
			signalID, err := w.waitOnFirstSignal(wctx, signalUUIDs, timeoutFuture)
			if err != nil {
				return sdktypes.InvalidValue, err
			}

			if signalID == uuid.Nil {
				return sdktypes.InvalidValue, nil
			}

			// get next event on this signal
			event, err := w.getNextEvent(wctx, signalID)
			if err != nil {
				return sdktypes.InvalidValue, err
			}

			if event != nil {
				return sdktypes.WrapValue(event)
			}
		}
	}
}
