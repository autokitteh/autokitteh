package dispatcher

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	wf "go.autokitteh.dev/autokitteh/internal/backend/workflows"
	cctx "go.autokitteh.dev/autokitteh/internal/context"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (d *dispatcher) getEventSessionData(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) ([]wf.SessionData, error) {
	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	z := d.Z.With(zap.String("event_id", event.ID().String()))

	if opts.Env != "" {
		z = z.With(zap.String("env", opts.Env))
	}

	if opts.DeploymentID.IsValid() {
		z = z.With(zap.String("deployment_id", opts.DeploymentID.String()))
	}

	if !event.IsValid() {
		z.Error("could not find event")
		return nil, sdkerrors.ErrNotFound
	}

	optsEnvID, err := resolveEnv(ctx, &d.Services, opts.Env)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			z.Info("env is not configured")
			return nil, nil
		}
		return nil, fmt.Errorf("env: %w", err)
	}
	if optsEnvID.IsValid() {
		z = z.With(zap.String("env_id", optsEnvID.String()))
	}

	cid := event.ConnectionID()

	conn, err := d.Services.Connections.Get(ctx, cid)
	if err != nil { // any error, including NotFound
		z.Error("get connection", zap.Error(err))
		return nil, err
	}

	iid := conn.IntegrationID()

	ts, err := d.Services.Triggers.List(ctx, sdkservices.ListTriggersFilter{ConnectionID: cid})
	if err != nil {
		return nil, fmt.Errorf("list triggers: %w", err)
	}

	if len(ts) == 0 {
		z.Info("no triggers for eventID")
		return nil, nil
	}

	var sds []wf.SessionData

	eventType := event.Type()
	for _, t := range ts {
		envID := t.EnvID()
		triggerEventType := t.EventType()

		z := z.With(zap.String("env_id", envID.String()), zap.String("event_type", triggerEventType))

		if triggerEventType != "" && eventType != triggerEventType {
			z.Debug("irrelevant event type")
			continue
		}

		if !envID.IsValid() && optsEnvID.IsValid() && envID != optsEnvID {
			z.Debug("irrelevant env", zap.String("expected", optsEnvID.String()))
			continue
		}

		relevant, additionalTriggerData, err := processSpecialTrigger(t, iid, event)
		if err != nil {
			continue
		} else if !relevant {
			z.Debug("irrelevant event for special trigger")
			continue
		}

		if relevant, err := event.Matches(t.Filter()); err != nil {
			z.Debug("filter error", zap.Error(err))
			// TODO(ENG-566): alert user their filter is bad. Integrate with alerting and monitoring.
			continue
		} else if !relevant {
			z.Debug("irrelevant event")
			continue
		}

		activeDeployments, err := d.Services.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateActive, EnvID: envID})
		if err != nil {
			z.Panic("could not fetch active deployments", zap.Error(err))
		}

		var testingDeployments []sdktypes.Deployment

		if optsEnvID.IsValid() || opts.DeploymentID.IsValid() {
			testingDeployments, err = d.Services.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateTesting, EnvID: envID})
			if err != nil {
				z.Panic("could not fetch testing deployments", zap.Error(err))
			}
		}

		if len(activeDeployments)+len(testingDeployments) == 0 {
			z.Debug("no deployments")
			continue
		}

		deployments := append(activeDeployments, testingDeployments...)

		if optsEnvID.IsValid() || opts.DeploymentID.IsValid() {
			deployments = kittehs.Filter(deployments, func(deployment sdktypes.Deployment) bool {
				return (!optsEnvID.IsValid() || optsEnvID == deployment.EnvID()) &&
					(!opts.DeploymentID.IsValid() || opts.DeploymentID == deployment.ID())
			})
		}

		cl := t.CodeLocation()
		for _, dep := range deployments {
			sds = append(sds, wf.SessionData{Deployment: dep, CodeLocation: cl, Trigger: t, AdditionalTriggerData: additionalTriggerData})
			z.Debug("relevant deployment found", zap.String("deployment_id", dep.ID().String()))
		}
	}
	return sds, nil
}

func (d *dispatcher) signalWorkflows(ctx context.Context, event sdktypes.Event) error {
	eid := event.ID()

	z := d.Z.With(zap.String("event_id", eid.String()))

	if !event.IsValid() {
		z.Error("could not find event")
		return sdkerrors.ErrNotFound
	}

	cid := event.ConnectionID()

	// REVIEW: could we rely on invalid connectionID instead of fetching connection and checking it?
	// if not we need to fetch connection with ignoreNotFound and check if it's valid
	if !cid.IsValid() {
		z.Info("no connections for event id", zap.String("connection_id", cid.String()))
		return nil
	}
	conn, err := d.Services.Connections.Get(ctx, cid)
	if err != nil {
		z.Error("could not fetch connections", zap.Error(err))
	}

	signals, err := d.DB.ListSignalsWaitingOnConnection(ctx, conn.ID())
	if err != nil {
		z.Error("could not fetch signals", zap.Error(err))
		return err
	}

	z.Debug("found signal candidates", zap.Int("count", len(signals)))
	for _, signal := range signals {
		match, err := event.Matches(signal.Filter)
		l := z.With(zap.String("signal_id", signal.SignalID),
			zap.String("filter", signal.Filter),
			zap.String("event_id", event.ID().String()))

		if err != nil {
			l.Error("inavlid signal filter", zap.Error(err))

			if err := d.DB.RemoveSignal(ctx, signal.SignalID); err != nil {
				l.Error("failed removing signal with invalid filter", zap.Error(err))
				continue
			}
			l.Debug("signal removed")
			continue
		}

		if !match {
			l.Debug("signal filter not matching event, skipping")
			continue
		}

		if err := d.Tmprl.Temporal().SignalWorkflow(ctx, signal.WorkflowID, "", signal.SignalID, eid); err != nil {
			var nferr *serviceerror.NotFound
			if !errors.As(err, &nferr) {
				l.Error("could not signal workflow", zap.Error(err))
				return err
			}
			l.Debug("workflow not found, removing signal")
			if err := d.DB.RemoveSignal(ctx, signal.SignalID); err != nil {
				return err
			}
		}
		l.Debug("signaled workflow", zap.String("workflow_id", signal.WorkflowID))
	}

	return nil
}

func (d *dispatcher) eventsWorkflow(wctx workflow.Context, input eventsWorkflowInput) (*eventsWorkflowOutput, error) {
	logger := workflow.GetLogger(wctx)

	logger.Info("started events workflow", "event_id", input.EventID)
	z := d.Z.With(zap.String("event_id", input.EventID.String()))

	ctx := cctx.WithRequestOrginator(context.Background(), cctx.EventWorkflow)

	event, err := d.Services.Events.Get(ctx, input.EventID)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	if !event.IsValid() {
		z.Error("could not find event")
		return nil, sdkerrors.ErrNotFound
	}

	// Set event to processing
	d.CreateEventRecord(ctx, input.EventID, sdktypes.EventStateProcessing)

	// Fetch event related data
	sds, err := d.getEventSessionData(ctx, event, input.Options)
	if err != nil {
		logger.Error("Failed processing event", "EventID", input.EventID, "error", err)
		d.CreateEventRecord(ctx, input.EventID, sdktypes.EventStateFailed)
		return nil, nil
	}

	// start sessions
	if err := d.StartSessions(wctx, event, sds); err != nil {
		return nil, err
	}

	// execute waiting signals
	err = d.signalWorkflows(ctx, event)
	if err != nil {
		return nil, err
	}
	// Set event to Completed
	d.CreateEventRecord(ctx, input.EventID, sdktypes.EventStateCompleted)
	return nil, nil
}
