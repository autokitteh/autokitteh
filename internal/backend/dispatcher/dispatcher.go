package dispatcher

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/externalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
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

	DB       db.DB
	Temporal temporalclient.Client

	Connections    sdkservices.Connections
	Deployments    sdkservices.Deployments
	Events         sdkservices.Events
	Projects       sdkservices.Projects
	Sessions       sdkservices.Sessions
	Triggers       sdkservices.Triggers
	Tokens         authtokens.Tokens
	ExternalClient externalclient.ExternalClient
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

func (d *Dispatcher) DispatchExternal(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (*sdkservices.DispatchResponse, error) {
	pid, err := d.svcs.DB.GetProjectIDOf(ctx, event.DestinationID())
	if err != nil {
		return nil, fmt.Errorf("get project id of destination %v: %w", event.DestinationID(), err)
	}

	orgID, err := d.svcs.DB.GetOrgIDOf(ctx, pid)
	if err != nil {
		return nil, fmt.Errorf("get org id of project %v: %w", pid, err)
	}

	d.sl.Info("external dispatch found pid", pid.String(), "and orgID", orgID.String(), "for eventID", event.ID().String(), "and destinationID", event.DestinationID().String())

	cli, err := d.svcs.ExternalClient.NewOrgImpersonator(orgID)
	if err != nil {
		return nil, fmt.Errorf("create internal token: %w", err)
	}

	return cli.Dispatcher().Dispatch(ctx, event, opts)
}

func (d *Dispatcher) Dispatch(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (*sdkservices.DispatchResponse, error) {
	event = event.WithCreatedAt(time.Now())
	eid, err := d.svcs.Events.Save(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("save event: %w", err)
	}
	event = event.WithID(eid)

	sl := d.sl.With("event_id", eid)

	sl.Infof("event saved: %v", eid)

	if err := authz.CheckContext(ctx, event.ID(), "dispatch", authz.WithData("event", event), authz.WithData("opts", opts)); err != nil {
		return nil, err
	}

	memo := map[string]string{
		"event_id":         eid.String(),
		"event_uuid":       eid.UUIDValue().String(),
		"destination_id":   event.DestinationID().String(),
		"destination_uuid": event.DestinationID().UUIDValue().String(),
		"seq":              strconv.FormatUint(event.Seq(), 10),
		"process_id":       fixtures.ProcessID(),
	}

	r, err := d.svcs.Temporal.TemporalClient().ExecuteWorkflow(
		ctx,
		d.cfg.Workflow.ToStartWorkflowOptions(
			taskQueueName,
			eid.String(),
			fmt.Sprintf("event %v", eid),
			memo,
		),
		workflowName,
		eventsWorkflowInput{Event: event, Options: opts},
	)
	if err != nil {
		return nil, fmt.Errorf("failed starting workflow: %w", err)
	}

	sl.Desugar().Info("started dispatcher workflow for event: "+eid.String(),
		zap.Any("workflow_id", r.GetID()),
		zap.Any("run_id", r.GetRunID()))

	if opts == nil || !opts.Wait {
		return &sdkservices.DispatchResponse{EventID: eid}, nil
	}

	var out eventsWorkflowOutput
	if err := r.Get(ctx, &out); err != nil {
		return nil, fmt.Errorf("waiting for workflow completion: %w", err)
	}

	return &sdkservices.DispatchResponse{EventID: eid, SessionIDs: out.Started}, nil
}

func (d *Dispatcher) Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) (*sdkservices.DispatchResponse, error) {
	sl := d.sl.With("event_id", eventID)

	event, err := d.svcs.Events.Get(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if !event.IsValid() {
		return nil, sdkerrors.ErrNotFound
	}

	if err := authz.CheckContext(ctx, eventID, "redispatch", authz.WithData("event", event), authz.WithData("opts", opts)); err != nil {
		return nil, err
	}

	memo := event.Memo()
	if memo == nil {
		memo = make(map[string]string)
	}
	memo["redispatch_of"] = eventID.String()
	event = event.WithMemo(memo)

	resp, err := d.Dispatch(authcontext.SetAuthnSystemUser(ctx), event, opts)
	if err != nil {
		sl.With("err", err).Errorf("failed redispatching event %v: %v", eventID, err)
		return nil, fmt.Errorf("failed redispatching event %v: %w", eventID, err)
	}

	sl.With("response", resp).Infof("redispatched event %v as %v", eventID, resp.EventID)

	return resp, err
}

func (d *Dispatcher) Start(context.Context) error {
	w := temporalclient.NewWorker(d.sl.Desugar(), d.svcs.Temporal.TemporalClient(), taskQueueName, d.cfg.Worker)
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
