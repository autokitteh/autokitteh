package dispatcher

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (d *dispatcher) createEventRecord(ctx context.Context, eventID sdktypes.EventID, state sdktypes.EventState) error {
	record, err := sdktypes.StrictEventRecordFromProto(&sdktypes.EventRecordPB{EventId: eventID.String(), State: state.ToProto()})
	if err != nil {
		d.z.Error("Failed updating event state record", zap.String("eventID", eventID.String()), zap.String("state", state.String()), zap.Error(err))
		return err
	}
	if err = d.services.Events.AddEventRecord(ctx, record); err != nil {
		d.z.Panic("Failed updating event state record", zap.String("eventID", eventID.String()), zap.String("state", state.String()), zap.Error(err))
	}
	return nil
}

func (d *dispatcher) getEventSessionData(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) ([]sessionData, error) {
	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	z := d.z.With(zap.String("event_id", sdktypes.GetEventID(event).String()))

	if event == nil {
		z.Error("could not find event")
		return nil, sdkerrors.ErrNotFound
	}

	iid := sdktypes.GetEventIntegrationID(event)

	it := sdktypes.GetEventIntegrationToken(event)

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
			ctx, sdkservices.ListTriggersFilter{ConnectionID: sdktypes.GetConnectionID(c)},
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

	eventType := sdktypes.GetEventType(event)
	for _, t := range triggers {
		envID := sdktypes.GetTriggerEnvID(t)
		triggerEventType := sdktypes.GetTriggerEventType(t)

		z := z.With(zap.String("env_id", envID.String()), zap.String("event_type", triggerEventType))

		if triggerEventType != eventType {
			z.Debug("irrelevant event type")
			continue
		}

		if envID != nil && opts.EnvID != nil && envID.String() != opts.EnvID.String() {
			z.Debug("irrelevant env", zap.String("expected", opts.EnvID.String()))
			continue
		}

		activeDeployments, err := d.services.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateActive, EnvID: envID})
		if err != nil {
			z.Panic("could not fetch active deployments", zap.Error(err))
		}

		var testingDeployments []sdktypes.Deployment

		if opts.EnvID != nil || opts.DeploymentID != nil {
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

		if opts.EnvID != nil || opts.DeploymentID != nil {
			deployments = kittehs.Filter(deployments, func(deployment sdktypes.Deployment) bool {
				return (opts.EnvID == nil || opts.EnvID.String() == sdktypes.GetDeploymentEnvID(deployment).String()) &&
					(opts.DeploymentID == nil || opts.DeploymentID.String() == sdktypes.GetDeploymentID(deployment).String())
			})
		}

		cl := sdktypes.GetTriggerCodeLocation(t)
		for _, dep := range deployments {
			did := sdktypes.GetDeploymentID(dep)
			sds = append(sds, sessionData{deploymentID: did, codeLocation: cl})
			z.Debug("relevant deployment found", zap.String("deployment_id", did.String()))
		}
	}
	return sds, nil
}

func (d *dispatcher) signalWorkflows(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) error {
	z := d.z.With(zap.String("event_id", sdktypes.GetEventID(event).String()))

	if event == nil {
		z.Error("could not find event")
		return sdkerrors.ErrNotFound
	}

	eid := sdktypes.GetEventID(event)

	iid := sdktypes.GetEventIntegrationID(event)

	it := sdktypes.GetEventIntegrationToken(event)

	connections, err := d.services.Connections.List(ctx, sdkservices.ListConnectionsFilter{IntegrationID: iid, IntegrationToken: it})
	if err != nil {
		z.Panic("could not fetch connections", zap.Error(err))
	}

	if len(connections) == 0 {
		z.Info("no connections for eventID", zap.String("integrationID", iid.String()), zap.String("integration token", it))
		return nil
	}

	connection := connections[0]

	signals, err := d.db.ListSignalsWaitingOnConnection(ctx, sdktypes.GetConnectionID(connection), sdktypes.GetEventType(event))
	if err != nil {
		z.Error("could not fetch signals", zap.Error(err))
		return err
	}

	z.Debug("found signals", zap.Int("count", len(signals)))
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

	if event == nil {
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
	deploymentID sdktypes.DeploymentID
	codeLocation sdktypes.CodeLocation
}

func (d *dispatcher) startSessions(ctx workflow.Context, event sdktypes.Event, sessionsData []sessionData) {
	// DO NOT PASS Memo. It is not intended for automation use, just auditing.
	inputs := sdktypes.EventToValues(event)

	for _, sd := range sessionsData {
		session := sdktypes.NewSession(sd.deploymentID, nil, sdktypes.GetEventID(event), sd.codeLocation, inputs, nil)

		goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

		// TODO(ENG-197): change to local activity.
		sessionID, err := d.services.Sessions.Start(goCtx, session)
		if err != nil {
			// Panic in order to make the workflow retry.
			d.z.Panic("could not start session")
		}
		d.z.Info("started session", zap.String("session_id", sessionID.String()))
	}
}
