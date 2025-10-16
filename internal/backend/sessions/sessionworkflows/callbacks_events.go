package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/types"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
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

		l := w.l.With(zap.Any("destination_id", did))

		if err := sdktypes.VerifyEventFilter(filter); err != nil {
			l.With(zap.Error(err)).Sugar().Infof("invalid filter in workflow code: %v", err)
			return "", sdkerrors.NewInvalidArgumentError("invalid filter: %w", err)
		}

		// generate a unique signal id.
		var (
			signalID    uuid.UUID
			minSequence uint64
		)

		isActivity := activity.IsActivity(ctx)

		if isActivity {
			signalID = uuid.New()

			var err error
			if minSequence, err = w.ws.getLatestEventSequenceActivity(ctx); err != nil {
				return "", fmt.Errorf("get current sequence in activity: %w", err)
			}
		} else {
			wctx = temporalclient.WithActivityOptions(wctx, w.ws.svcs.WorkflowExecutor.WorkflowQueue(), w.ws.cfg.Activity)

			if err := workflow.SideEffect(wctx, func(wctx workflow.Context) any {
				return uuid.New()
			}).Get(&signalID); err != nil {
				return "", fmt.Errorf("generate signal ID: %w", err)
			}

			if err := workflow.ExecuteActivity(wctx, getLastEventSequenceActivityName).Get(wctx, &minSequence); err != nil {
				return "", fmt.Errorf("get current sequence: %w", err)
			}
		}

		// must be set before signal is saved, otherwise the signal might reach the workflow before
		// the map is updated.
		w.lastReadEventSeqForSignal[signalID] = minSequence

		signal := types.Signal{
			ID:            signalID,
			WorkflowID:    w.workflowExecutionID,
			DestinationID: did,
			Filter:        filter,
		}

		if isActivity {
			if err := w.ws.saveSignalActivity(ctx, &signal); err != nil {
				return "", fmt.Errorf("save signal in activity: %w", err)
			}
		} else {
			if err := workflow.ExecuteActivity(wctx, saveSignalActivityName, &signal).Get(wctx, nil); err != nil {
				return "", fmt.Errorf("save signal: %w", err)
			}
		}

		l.Sugar().Infof("created event subscription: %v", signalID)

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

		if activity.IsActivity(ctx) {
			w.removeEventSubscriptionInActivity(ctx, uuid)
		} else {
			w.removeEventSubscription(wctx, uuid)
		}

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
	return func(ctx context.Context, _ sdktypes.RunID, signals []string, timeout time.Duration) (sdktypes.Value, error) {
		if activity.IsActivity(ctx) {
			return w.nextEventInActivity(ctx, signals, timeout)
		}

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
				return sdktypes.Nothing, nil
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

func (w *sessionWorkflow) nextEventInActivity(ctx context.Context, signals []string, timeout time.Duration) (sdktypes.Value, error) {
	_, span := w.startCallbackSpan(ctx, "next_event_in_activity")
	defer span.End()

	if len(signals) == 0 {
		return sdktypes.Nothing, nil
	}

	signalUUIDs, err := kittehs.TransformError(signals, uuid.Parse)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	var tmoCh <-chan time.Time
	if timeout != 0 {
		tmoCh = time.After(timeout)
	}

	for {
		// check if there is an event already in one of the signals
		for _, signalID := range signalUUIDs {
			event, err := w.getNextEventInActivity(ctx, signalID)
			if err != nil {
				return sdktypes.InvalidValue, err
			}
			if event != nil {
				return sdktypes.WrapValue(event)
			}
		}

		select {
		case <-time.After(w.ws.cfg.NextEventInActivityPollDuration):
			w.l.Debug("next_event_in_activity: poll timeout reached")

		case <-ctx.Done():
			return sdktypes.InvalidValue, ctx.Err()

		case <-tmoCh:
			w.l.Debug("next_event_in_activity: user timeout reached")
			return sdktypes.Nothing, nil
		}
	}
}
