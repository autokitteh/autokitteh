package sessionworkflows

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/types"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) createEventSubscription(wctx workflow.Context, filter string, did sdktypes.EventDestinationID) (uuid.UUID, error) {
	wctx = temporalclient.WithActivityOptions(wctx, taskQueueName, w.ws.cfg.Activity)

	l := w.l.With(zap.Any("destination_id", did))

	if err := sdktypes.VerifyEventFilter(filter); err != nil {
		l.With(zap.Error(err)).Sugar().Infof("invalid filter in workflow code: %v", err)
		return uuid.Nil, sdkerrors.NewInvalidArgumentError("invalid filter: %w", err)
	}

	// generate a unique signal id.
	var signalID uuid.UUID
	if err := workflow.SideEffect(wctx, func(wctx workflow.Context) any {
		return uuid.New()
	}).Get(&signalID); err != nil {
		return uuid.Nil, fmt.Errorf("generate signal ID: %w", err)
	}

	var minSequence uint64
	if err := workflow.ExecuteActivity(wctx, getLastEventSequenceActivityName).Get(wctx, &minSequence); err != nil {
		return uuid.Nil, fmt.Errorf("get current sequence: %w", err)
	}

	// must be set before signal is saved, otherwise the signal might reach the workflow before
	// the map is updated.
	w.lastReadEventSeqForSignal[signalID] = minSequence

	workflowID := workflow.GetInfo(wctx).WorkflowExecution.ID

	signal := types.Signal{
		ID:            signalID,
		WorkflowID:    workflowID,
		DestinationID: did,
		Filter:        filter,
	}

	if err := workflow.ExecuteActivity(wctx, saveSignalActivityName, &signal).Get(wctx, nil); err != nil {
		return uuid.Nil, fmt.Errorf("save signal: %w", err)
	}

	l.Sugar().Infof("created event subscription: %v", signalID)

	return signalID, nil
}

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

func (w *sessionWorkflow) getNextEvent(ctx context.Context, signalID uuid.UUID) (map[string]sdktypes.Value, error) {
	l := w.l.With(zap.Any("signal_id", signalID))

	wctx := sessioncontext.GetWorkflowContext(ctx)
	wctx = temporalclient.WithActivityOptions(wctx, taskQueueName, w.ws.cfg.Activity)

	minSequenceNumber, ok := w.lastReadEventSeqForSignal[signalID]
	if !ok {
		return nil, fmt.Errorf("no such subscription %q", signalID)
	}

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
			return nil, err
		}

		return nil, fmt.Errorf("get signal event %v: %w", signalID, err)
	}

	if !event.IsValid() {
		return nil, nil
	}

	w.lastReadEventSeqForSignal[signalID] = event.Seq()

	l.With(zap.Any("event_id", event.ID())).Sugar().Infof("got event: %v", event.ID())

	return event.Data(), nil
}

func (w *sessionWorkflow) removeEventSubscription(ctx context.Context, signalID uuid.UUID) {
	l := w.l.With(zap.Any("signal_id", signalID))

	wctx := sessioncontext.GetWorkflowContext(ctx)
	wctx = temporalclient.WithActivityOptions(wctx, taskQueueName, w.ws.cfg.Activity)

	if err := workflow.ExecuteActivity(wctx, removeSignalActivityName, signalID).Get(wctx, nil); err != nil {
		// it is not a critical error, we can just log it. no need to panic.
		l.With(zap.Error(err)).Sugar().Errorf("remove signal: %v", err)
	}

	delete(w.lastReadEventSeqForSignal, signalID)
}
