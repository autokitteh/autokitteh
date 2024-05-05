package dispatcher

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
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
	Triggers     sdkservices.Triggers
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
		return sdktypes.InvalidEventID, fmt.Errorf("save event: %w", err)
	}

	z := d.z.With(zap.String("event_id", eventID.String()))

	z.Debug("event saved")

	if event.Type() == "cron_trigger" {
		err = d.startSchedulerWorkflow(ctx, eventID, opts, event.Data()["schedule"].GetString().Value())
	} else {
		err = d.startWorkflow(ctx, eventID, opts)
	}

	if err != nil {
		z.Error("startWorkflow failed, orphaned event", zap.Error(err))

		return sdktypes.InvalidEventID, fmt.Errorf("event %v saved, but startWorkflow failed: %w", eventID, err)
	}

	return eventID, nil
}

func (d *dispatcher) Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	event, err := d.services.Events.Get(ctx, eventID)
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
	w := worker.New(d.temporal, taskQueueName, worker.Options{})
	w.RegisterWorkflow(d.eventsWorkflow)

	if err := w.Start(); err != nil {
		return fmt.Errorf("worker start: %w", err)
	}
	return nil
}

func (d *dispatcher) startWorkflow(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) error {
	options := client.StartWorkflowOptions{
		ID:        eventID.String(),
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

// use legacy temporal scheduler for now
func (d *dispatcher) startSchedulerWorkflow(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions, schedule string) error {
	if schedule == "" {
		return fmt.Errorf("failed starting scheduler workflow: no schedule")
	}

	options := client.StartWorkflowOptions{
		ID:           eventID.String(),
		TaskQueue:    taskQueueName,
		CronSchedule: schedule,
	}
	input := eventsWorkflowInput{
		EventID: eventID,
		Options: opts,
	}
	_, err := d.temporal.ExecuteWorkflow(ctx, options, d.eventsWorkflow, input)
	if err != nil {
		d.z.Error("Failed starting scheduler workflow", zap.String("eventID", eventID.String()), zap.Error(err))
		return fmt.Errorf("failed starting scheduler workflow: %w", err)
	}
	d.z.Info("Started scheduler workflow", zap.String("eventID", eventID.String()), zap.String("schedule", schedule))

	return nil
}

