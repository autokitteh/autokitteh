package sessionworkflows

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Returns "", nil on timeout.
func (w *sessionWorkflow) waitOnFirstSignal(wctx workflow.Context, signals []uuid.UUID, f workflow.Future) (uuid.UUID, error) {
	selector := workflow.NewSelector(wctx)

	if f != nil {
		selector.AddFuture(f, func(workflow.Future) {})
	}

	var signalID uuid.UUID

	for _, signal := range signals {
		selector.AddReceive(workflow.GetSignalChannel(wctx, signal.String()), func(c workflow.ReceiveChannel, _ bool) {
			// clear all pending signals.
			for c.ReceiveAsync(nil) {
				// nop
			}

			signalID = signal
		})
	}

	var cancelled bool

	// Select doesn't respond to cancellations unless we add a receive on the context done channel.
	selector.AddReceive(wctx.Done(), func(c workflow.ReceiveChannel, _ bool) { cancelled = true })

	// this will wait for first signal or timeout.
	selector.Select(wctx)

	if cancelled {
		return uuid.Nil, wctx.Err()
	}

	return signalID, nil
}

func (w *sessionWorkflow) getNextEventInternal(signalID uuid.UUID, get func(uuid.UUID, uint64) (sdktypes.Event, error)) (map[string]sdktypes.Value, error) {
	l := w.l.With(zap.Any("signal_id", signalID))

	minSequenceNumber, ok := w.lastReadEventSeqForSignal[signalID]
	if !ok {
		return nil, fmt.Errorf("no such subscription %q", signalID)
	}

	event, err := get(signalID, minSequenceNumber)
	if err != nil {
		return nil, fmt.Errorf("get signal event %v: %w", signalID, err)
	}

	if !event.IsValid() {
		return nil, nil
	}

	w.lastReadEventSeqForSignal[signalID] = event.Seq()

	l.With(zap.Any("event_id", event.ID())).Sugar().Infof("got event: %v", event.ID())

	return event.Data(), nil
}

func (w *sessionWorkflow) getNextEvent(wctx workflow.Context, signalID uuid.UUID) (map[string]sdktypes.Value, error) {
	return w.getNextEventInternal(signalID, func(signalID uuid.UUID, minSequenceNumber uint64) (sdktypes.Event, error) {
		wctx = temporalclient.WithActivityOptions(wctx, w.ws.svcs.WorkflowExecutor.WorkflowQueue(), w.ws.cfg.Activity)

		var event sdktypes.Event

		fut := workflow.ExecuteActivity(
			wctx,
			getSignalEventActivityName,
			signalID,
			minSequenceNumber,
		)

		if err := fut.Get(wctx, &event); err != nil {
			// was the context cancelled?
			if wctx.Err() != nil {
				return sdktypes.InvalidEvent, err
			}

			return sdktypes.InvalidEvent, fmt.Errorf("get signal event %v: %w", signalID, err)
		}

		return event, nil
	})
}

func (w *sessionWorkflow) getNextEventInActivity(ctx context.Context, signalID uuid.UUID) (map[string]sdktypes.Value, error) {
	return w.getNextEventInternal(signalID, func(signalID uuid.UUID, minSequenceNumber uint64) (sdktypes.Event, error) {
		return w.ws.getSignalEventActivity(ctx, signalID, minSequenceNumber)
	})
}

func (w *sessionWorkflow) removeEventSubscription(wctx workflow.Context, signalID uuid.UUID) {
	l := w.l.With(zap.Any("signal_id", signalID))

	wctx = temporalclient.WithActivityOptions(wctx, w.ws.svcs.WorkflowExecutor.WorkflowQueue(), w.ws.cfg.Activity)

	if err := workflow.ExecuteActivity(wctx, removeSignalActivityName, signalID).Get(wctx, nil); err != nil {
		// it is not a critical error, we can just log it. no need to panic.
		l.With(zap.Error(err)).Sugar().Errorf("remove signal: %v", err)
	}

	delete(w.lastReadEventSeqForSignal, signalID)
}

func (w *sessionWorkflow) removeEventSubscriptionInActivity(ctx context.Context, signalID uuid.UUID) {
	l := w.l.With(zap.Any("signal_id", signalID))

	if err := w.ws.removeSignalActivity(ctx, signalID); err != nil {
		// it is not a critical error, we can just log it. no need to panic.
		l.With(zap.Error(err)).Sugar().Errorf("remove signal in activity: %v", err)
	}

	delete(w.lastReadEventSeqForSignal, signalID)
}
