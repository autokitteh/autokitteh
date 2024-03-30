package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (d *dispatcher) createEventRecord(ctx context.Context, eventID sdktypes.EventID, state sdktypes.EventState) error {
	record := sdktypes.NewEventRecord(eventID, state)
	if err := d.services.Events.AddEventRecord(ctx, record); err != nil {
		d.z.Panic("Failed updating event state record", zap.String("eventID", eventID.String()), zap.String("state", state.String()), zap.Error(err))
	}
	return nil
}

func (d *dispatcher) resolveEnv(ctx context.Context, env string) (envID sdktypes.EnvID, err error) {
	if env == "" {
		return sdktypes.InvalidEnvID, nil
	}

	parts := strings.Split(env, "/")
	switch len(parts) {
	case 1:
		if sdktypes.IsEnvID(parts[0]) {
			return sdktypes.ParseEnvID(parts[0])
		}
	case 2:
		var pid sdktypes.ProjectID
		if sdktypes.IsProjectID(parts[0]) {
			pid, err = sdktypes.ParseProjectID(parts[0])
		} else {
			var name sdktypes.Symbol
			if name, err = sdktypes.ParseSymbol(parts[0]); err != nil {
				return
			}

			var p sdktypes.Project
			if p, err = d.services.Projects.GetByName(context.Background(), name); p.IsValid() {
				pid = p.ID()
			}
		}

		if err != nil {
			return
		}

		if !pid.IsValid() {
			return sdktypes.InvalidEnvID, sdkerrors.ErrNotFound
		}

		var name sdktypes.Symbol
		if name, err = sdktypes.ParseSymbol(parts[1]); err != nil {
			return
		}

		var env sdktypes.Env
		if env, err = d.services.Envs.GetByName(ctx, pid, name); err != nil {
			return
		}

		if !env.IsValid() {
			err = sdkerrors.ErrNotFound
			return
		}

		return env.ID(), err
	}

	return
}

func (d *dispatcher) getEventSessionData(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) ([]sessionData, error) {
	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	z := d.z.With(zap.String("event_id", event.ID().String()))

	if !event.IsValid() {
		z.Error("could not find event")
		return nil, sdkerrors.ErrNotFound
	}

	optsEnvID, err := d.resolveEnv(ctx, opts.Env)
	if err != nil {
		return nil, fmt.Errorf("env: %w", err)
	}
	if optsEnvID.IsValid() {
		z = z.With(zap.String("env_id", optsEnvID.String()))
	}

	iid, it := event.IntegrationID(), event.IntegrationToken()

	connections, err := d.services.Connections.List(ctx, sdkservices.ListConnectionsFilter{IntegrationID: iid, IntegrationToken: it})
	if err != nil {
		z.Panic("could not fetch connections", zap.Error(err))
	}

	if len(connections) == 0 {
		z.Info("no connections for eventID", zap.String("integrationID", iid.String()), zap.String("integration token", it))
		return nil, nil
	}

	triggers := make([]sdktypes.Trigger, 0, len(connections))
	for _, c := range connections {
		m, err := d.services.Triggers.List(
			ctx, sdkservices.ListTriggersFilter{ConnectionID: c.ID()},
		)
		if err != nil {
			z.Panic("could not fetch triggers")
		}
		triggers = append(triggers, m...)
	}

	if len(triggers) == 0 {
		z.Info("no triggers for eventID")
		return nil, nil
	}

	var sds []sessionData

	eventType := event.Type()
	for _, t := range triggers {
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

		if relevant, err := event.Matches(t.Filter()); err != nil {
			z.Debug("filter error", zap.Error(err))
			// TODO(ENG-566): alert user their filter is bad. Integrate with alerting and monitoring.
			continue
		} else if !relevant {
			z.Debug("irrelevant event")
			continue
		}

		activeDeployments, err := d.services.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateActive, EnvID: envID})
		if err != nil {
			z.Panic("could not fetch active deployments", zap.Error(err))
		}

		var testingDeployments []sdktypes.Deployment

		if optsEnvID.IsValid() || opts.DeploymentID.IsValid() {
			testingDeployments, err = d.services.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateTesting, EnvID: envID})
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
			sds = append(sds, sessionData{deployment: dep, codeLocation: cl})
			z.Debug("relevant deployment found", zap.String("deployment_id", dep.ID().String()))
		}
	}
	return sds, nil
}

func (d *dispatcher) signalWorkflows(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) error {
	eid := event.ID()

	z := d.z.With(zap.String("event_id", eid.String()))

	if !event.IsValid() {
		z.Error("could not find event")
		return sdkerrors.ErrNotFound
	}

	iid, it := event.IntegrationID(), event.IntegrationToken()

	connections, err := d.services.Connections.List(ctx, sdkservices.ListConnectionsFilter{IntegrationID: iid, IntegrationToken: it})
	if err != nil {
		z.Panic("could not fetch connections", zap.Error(err))
	}

	if len(connections) == 0 {
		z.Info("no connections for eventID", zap.String("integrationID", iid.String()), zap.String("integration token", it))
		return nil
	}

	connection := connections[0]

	signals, err := d.db.ListSignalsWaitingOnConnection(ctx, connection.ID())
	if err != nil {
		z.Error("could not fetch signals", zap.Error(err))
		return err
	}

	z.Debug("found signal candidates", zap.Int("count", len(signals)))
	for _, signal := range signals {

		if err := d.temporal.SignalWorkflow(ctx, signal.WorkflowID, "", signal.SignalID, eid); err != nil {
			var nferr *serviceerror.NotFound
			if !errors.As(err, &nferr) {
				z.Error("could not signal workflow", zap.Error(err))
				return err
			}
			z.Debug("workflow not found, removing signal", zap.String("signal_id", signal.SignalID))
			if err := d.db.RemoveSignal(ctx, signal.SignalID); err != nil {
				return err
			}
		}
		z.Debug("signaled workflow", zap.String("workflow_id", signal.WorkflowID))
	}

	return nil
}

func (d *dispatcher) eventsWorkflow(ctx workflow.Context, input eventsWorkflowInput) (*eventsWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("started events workflow", "event_id", input.EventID)

	z := d.z.With(zap.String("event_id", input.EventID.String()))
	event, err := d.services.Events.Get(context.TODO(), input.EventID)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	if !event.IsValid() {
		z.Error("could not find event")
		return nil, sdkerrors.ErrNotFound
	}

	// Set event to processing
	if err := d.createEventRecord(context.Background(), input.EventID, sdktypes.EventStateProcessing); err != nil {
		return nil, err
	}

	// Fetch event related data
	sds, err := d.getEventSessionData(context.Background(), event, input.Options)
	if err != nil {
		_ = d.createEventRecord(context.Background(), input.EventID, sdktypes.EventStateFailed)
		logger.Error("Failed processing event", "EventID", input.EventID, "error", err)
		return nil, nil
	}

	// start sessions
	d.startSessions(ctx, event, sds)

	// execute waiting signals
	err = d.signalWorkflows(context.Background(), event, input.Options)
	if err != nil {
		return nil, err
	}
	// Set event to Completed
	if err = d.createEventRecord(context.Background(), input.EventID, sdktypes.EventStateCompleted); err != nil {
		return nil, err
	}

	return nil, nil
}

type sessionData struct {
	deployment   sdktypes.Deployment
	codeLocation sdktypes.CodeLocation
}

func (d *dispatcher) startSessions(ctx workflow.Context, event sdktypes.Event, sessionsData []sessionData) {
	// DO NOT PASS Memo. It is not intended for automation use, just auditing.
	inputs := event.ToValues()

	for _, sd := range sessionsData {
		dep := sd.deployment

		session := sdktypes.NewSession(dep.BuildID(), sd.codeLocation, inputs, nil).
			WithDeploymentID(dep.ID()).
			WithEventID(event.ID()).
			WithEnvID(dep.EnvID())

		goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

		// TODO(ENG-197): change to local activity.
		sessionID, err := d.services.Sessions.Start(goCtx, session)
		if err != nil {
			// Panic in order to make the workflow retry.
			d.z.Panic("could not start session")
		}
		d.z.Info("started session", zap.String("session_id", sessionID.String()))
	}

	return
}
