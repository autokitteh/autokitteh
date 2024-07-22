package dispatcher

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	wf "go.autokitteh.dev/autokitteh/internal/backend/workflows"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type dispatcher struct {
	wf.Workflow
}

type Dispatcher interface {
	sdkservices.Dispatcher
	Start(context.Context) error
}

func New(
	z *zap.Logger,
	db db.DB,
	services wf.Services,
	tc temporalclient.Client,
) Dispatcher {
	return &dispatcher{wf.Workflow{Z: z, DB: db, Services: services, Tmprl: tc}}
}

func (d *dispatcher) Dispatch(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.Dispatcher)
	ctx = akCtx.WithOwnershipOf(ctx, d.DB.GetOwnership, event.ConnectionID().UUIDValue())

	eventID, err := d.Services.Events.Save(ctx, event)
	if err != nil {
		return sdktypes.InvalidEventID, fmt.Errorf("save event: %w", err)
	}

	z := d.Z.With(zap.String("event_id", eventID.String()))

	z.Debug("event saved")

	if err := d.startWorkflow(ctx, eventID, opts); err != nil {
		z.Error("startWorkflow failed, orphaned event", zap.Error(err))

		return sdktypes.InvalidEventID, fmt.Errorf("event %v saved, but startWorkflow failed: %w", eventID, err)
	}

	return eventID, nil
}

func (d *dispatcher) Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.Dispatcher)
	event, err := d.Services.Events.Get(ctx, eventID)
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

func (d *dispatcher) Start(context.Context) error {
	w := worker.New(d.Tmprl.Temporal(), wf.TaskQueueName, worker.Options{Identity: wf.DispatcherWorkerID})
	w.RegisterWorkflow(d.eventsWorkflow)

	if err := w.Start(); err != nil {
		return fmt.Errorf("worker start: %w", err)
	}
	return nil
}

func (d *dispatcher) startWorkflow(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) error {
	options := client.StartWorkflowOptions{
		ID:        eventID.String(),
		TaskQueue: wf.TaskQueueName,
	}
	input := eventsWorkflowInput{
		EventID: eventID,
		Options: opts,
	}
	_, err := d.Tmprl.Temporal().ExecuteWorkflow(ctx, options, d.eventsWorkflow, input)
	if err != nil {
		d.Z.Error("Failed starting workflow", zap.String("eventID", eventID.String()), zap.Error(err))
		return fmt.Errorf("failed starting workflow: %w", err)
	}
	d.Z.Info("Started workflow", zap.String("eventID", eventID.String()))

	return nil
}
