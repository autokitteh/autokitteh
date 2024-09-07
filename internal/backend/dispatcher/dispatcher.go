package dispatcher

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const taskQueueName = "events"

type Dispatcher struct {
	fx.In

	L *zap.Logger

	db.DB
	temporalclient.Client
	sdkservices.Events
	sdkservices.Triggers
	sdkservices.Deployments
	sdkservices.Sessions
	sdkservices.Projects
	sdkservices.Envs
}

func New(d Dispatcher) *Dispatcher {
	return &d
}

func (d *Dispatcher) Dispatch(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.Dispatcher)
	ctx = akCtx.WithOwnershipOf(ctx, d.DB.GetOwnership, event.DestinationID().UUIDValue())

	eventID, err := d.Events.Save(ctx, event)
	if err != nil {
		return sdktypes.InvalidEventID, fmt.Errorf("save event: %w", err)
	}

	z := d.L.With(zap.String("event_id", eventID.String()))

	z.Debug("event saved")

	if err := d.startWorkflow(ctx, eventID, opts); err != nil {
		z.Error("startWorkflow failed, orphaned event", zap.Error(err))

		return sdktypes.InvalidEventID, fmt.Errorf("event %v saved, but startWorkflow failed: %w", eventID, err)
	}

	return eventID, nil
}

func (d *Dispatcher) Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.Dispatcher)
	event, err := d.Events.Get(ctx, eventID)
	if err != nil {
		return sdktypes.InvalidEventID, err
	}

	if !event.IsValid() {
		return sdktypes.InvalidEventID, sdkerrors.ErrNotFound
	}

	memo := event.Memo()
	if memo == nil {
		memo = make(map[string]string)
	}
	memo["redispatch_of"] = eventID.String()
	event = event.WithMemo(memo)

	return d.Dispatch(ctx, event, opts)
}

func (d *Dispatcher) Start(context.Context) error {
	w := worker.New(d.Client.Temporal(), taskQueueName, worker.Options{Identity: "Dispatcher"})
	w.RegisterWorkflow(d.eventsWorkflow)

	if err := w.Start(); err != nil {
		return fmt.Errorf("worker start: %w", err)
	}
	return nil
}

func (d *Dispatcher) startWorkflow(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) error {
	options := client.StartWorkflowOptions{
		ID:        eventID.String(),
		TaskQueue: taskQueueName,
	}
	input := eventsWorkflowInput{
		EventID: eventID,
		Options: opts,
	}
	_, err := d.Client.Temporal().ExecuteWorkflow(ctx, options, d.eventsWorkflow, input)
	if err != nil {
		d.L.Error("Failed starting workflow", zap.String("eventID", eventID.String()), zap.Error(err))
		return fmt.Errorf("failed starting workflow: %w", err)
	}
	d.L.Info("Started workflow", zap.String("eventID", eventID.String()))

	return nil
}
