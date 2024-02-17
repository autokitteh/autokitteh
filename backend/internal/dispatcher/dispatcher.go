package dispatcher

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/db"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Services struct {
	fx.In

	Connections  sdkservices.Connections
	Deployments  sdkservices.Deployments
	Events       sdkservices.Events
	Integrations sdkservices.Integrations
	Projects     sdkservices.Projects
	Mappings     sdkservices.Mappings
	Sessions     sdkservices.Sessions
	Envs         sdkservices.Envs
}

type dispatcher struct {
	z        *zap.Logger
	db       db.DB
	services Services
	temporal client.Client
}

type Dispatcher interface {
	sdkservices.Dispatcher
	Start(context.Context) error
}

func New(
	z *zap.Logger,
	db db.DB,
	services Services,
	c client.Client,
) Dispatcher {
	return &dispatcher{db: db, z: z, services: services, temporal: c}
}

func (d *dispatcher) Dispatch(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	eventID, err := d.services.Events.Save(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("save event: %w", err)
	}

	z := d.z.With(zap.String("event_id", eventID.String()))

	z.Debug("event saved")

	if err := d.startWorkflow(ctx, eventID, opts); err != nil {
		z.Error("startWorkflow failed, orphaned event", zap.Error(err))

		return nil, fmt.Errorf("event %v saved, but startWorkflow failed: %w", eventID, err)
	}

	return eventID, nil
}

func (d *dispatcher) Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	event, err := d.services.Events.Get(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if event == nil {
		return nil, sdkerrors.ErrNotFound
	}

	event = kittehs.Must1(event.Update(func(event *sdktypes.EventPB) {
		if event.Memo == nil {
			event.Memo = make(map[string]string)
		}

		// TODO: should probably be a first class field in event.
		event.Memo["redispatch_of"] = eventID.String()
	}))

	return d.Dispatch(ctx, event, opts)
}

func (d *dispatcher) Start(context.Context) error {
	w := worker.New(d.temporal, taskQueueName, worker.Options{})
	w.RegisterWorkflow(d.eventsWorkflow)

	if err := w.Start(); err != nil {
		return fmt.Errorf("worker start: %w", err)
	}
	return nil
}

func (d *dispatcher) startWorkflow(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) error {
	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s_%s", workflowName, eventID.Value()),
		TaskQueue: taskQueueName,
	}
	input := eventsWorkflowInput{
		EventID: eventID,
		Options: opts,
	}
	_, err := d.temporal.ExecuteWorkflow(ctx, options, d.eventsWorkflow, input)
	if err != nil {
		d.z.Error("Failed starting workflow", zap.String("eventID", eventID.String()), zap.Error(err))
		return fmt.Errorf("failed starting workflow: %w", err)
	}
	d.z.Info("Started workflow", zap.String("eventID", eventID.String()))

	return nil
}
