package dispatcher

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	taskQueueName = "events"
	workflowName  = "event"
)

type Svcs struct {
	fx.In

	db.DB
	temporalclient.LazyTemporalClient
	sdkservices.Events
	sdkservices.Triggers
	sdkservices.Deployments
	sdkservices.Sessions
	sdkservices.Projects
	sdkservices.Envs
}

type Dispatcher struct {
	sl   *zap.SugaredLogger
	cfg  *Config
	svcs Svcs
}

var _ sdkservices.Dispatcher = (*Dispatcher)(nil)

func New(l *zap.Logger, cfg *Config, svcs Svcs) *Dispatcher {
	return &Dispatcher{sl: l.Sugar(), cfg: cfg, svcs: svcs}
}

func (d *Dispatcher) Dispatch(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.Dispatcher)
	ctx = akCtx.WithOwnershipOf(ctx, d.svcs.DB.GetOwnership, event.DestinationID().UUIDValue())

	eid, err := d.svcs.Events.Save(ctx, event)
	if err != nil {
		return sdktypes.InvalidEventID, fmt.Errorf("save event: %w", err)
	}

	event = event.WithID(eid)

	sl := d.sl.With("event_id", eid)

	sl.Infof("event %v saved", eid)

	r, err := d.svcs.LazyTemporalClient().ExecuteWorkflow(
		ctx,
		d.cfg.Workflow.ToStartWorkflowOptions(
			taskQueueName,
			eid.String(),
			fmt.Sprintf("event %v", eid),
			nil,
		),
		workflowName,
		eventsWorkflowInput{Event: event, Options: opts},
	)
	if err != nil {
		return sdktypes.InvalidEventID, fmt.Errorf("failed starting workflow: %w", err)
	}

	sl.With("run", r).Infof("started dispatcher workflow %v for %v", r, eid)

	return eid, nil
}

func (d *Dispatcher) Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	sl := d.sl.With("event_id", eventID)

	ctx = akCtx.WithRequestOrginator(ctx, akCtx.Dispatcher)
	event, err := d.svcs.Events.Get(ctx, eventID)
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

	newEventID, err := d.Dispatch(ctx, event, opts)
	if err != nil {
		sl.With("err", err).Errorf("failed redispatching event %v: %v", eventID, err)
		return sdktypes.InvalidEventID, fmt.Errorf("failed redispatching event %v: %w", eventID, err)
	}

	sl.With("new_event_id", newEventID).Infof("redispatched event %v as %v", eventID, newEventID)

	return newEventID, err
}

func (d *Dispatcher) Start(context.Context) error {
	w := temporalclient.NewWorker(d.sl.Desugar(), d.svcs.LazyTemporalClient(), taskQueueName, temporalclient.WorkerConfig{})
	if w == nil {
		return nil
	}

	w.RegisterWorkflowWithOptions(
		d.eventsWorkflow,
		workflow.RegisterOptions{Name: workflowName},
	)

	d.registerActivities(w)

	if err := w.Start(); err != nil {
		return fmt.Errorf("worker start: %w", err)
	}

	return nil
}
