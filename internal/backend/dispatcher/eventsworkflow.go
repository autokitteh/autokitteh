package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	eventInputsSymbolValue   = sdktypes.NewSymbolValue(sdktypes.NewSymbol("event"))
	triggerInputsSymbolValue = sdktypes.NewSymbolValue(sdktypes.NewSymbol("trigger"))
	dataSymbolValue          = sdktypes.NewSymbolValue(sdktypes.NewSymbol("data"))
)

type sessionData struct {
	sdktypes.Deployment
	sdktypes.CodeLocation
	sdktypes.Trigger
}

func (d *Dispatcher) getEventSessionData(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) ([]sessionData, error) {
	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	z := d.L.With(zap.String("event_id", event.ID().String()))

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

	optsEnvID, err := d.resolveEnv(ctx, opts.Env)
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

	var ts []sdktypes.Trigger
	dstid := event.DestinationID()

	if cid := dstid.ToConnectionID(); cid.IsValid() {
		if ts, err = d.Triggers.List(ctx, sdkservices.ListTriggersFilter{ConnectionID: cid}); err != nil {
			return nil, fmt.Errorf("list triggers: %w", err)
		}
	} else if tid := dstid.ToTriggerID(); tid.IsValid() {
		t, err := d.Triggers.Get(ctx, tid)
		if err != nil {
			return nil, fmt.Errorf("get trigger: %w", err)
		}

		ts = append(ts, t)
	}

	if len(ts) == 0 {
		z.Info("no triggers for event")
		return nil, nil
	}

	var sds []sessionData

	deploymentsForEnv := make(map[sdktypes.EnvID][]sdktypes.Deployment)

	eventType := event.Type()
	for _, t := range ts {
		envID := t.EnvID()
		triggerEventType := t.EventType()

		z := z.With(zap.String("env_id", envID.String()), zap.String("event_type", triggerEventType))

		if !t.CodeLocation().IsValid() {
			z.Info("no entry point, ignoring trigger")
			continue
		}

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

		deployments, found := deploymentsForEnv[envID]
		if !found {
			activeDeployments, err := d.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateActive, EnvID: envID})
			if err != nil {
				z.Panic("could not fetch active deployments", zap.Error(err))
			}

			var testingDeployments []sdktypes.Deployment

			if optsEnvID.IsValid() || opts.DeploymentID.IsValid() {
				testingDeployments, err = d.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateTesting, EnvID: envID})
				if err != nil {
					z.Panic("could not fetch testing deployments", zap.Error(err))
				}
			}

			if len(activeDeployments)+len(testingDeployments) != 0 {
				deployments = append(activeDeployments, testingDeployments...)

				if optsEnvID.IsValid() || opts.DeploymentID.IsValid() {
					deployments = kittehs.Filter(deployments, func(deployment sdktypes.Deployment) bool {
						return (!optsEnvID.IsValid() || optsEnvID == deployment.EnvID()) &&
							(!opts.DeploymentID.IsValid() || opts.DeploymentID == deployment.ID())
					})
				}
			}

			deploymentsForEnv[envID] = deployments

		}

		if len(deployments) == 0 {
			z.Info("no deployments for env")
			continue
		}

		cl := t.CodeLocation()
		for _, dep := range deployments {
			sds = append(sds, sessionData{Deployment: dep, CodeLocation: cl, Trigger: t})
			z.Debug("relevant deployment found", zap.String("deployment_id", dep.ID().String()))
		}
	}
	return sds, nil
}

func (d *Dispatcher) signalWorkflows(ctx context.Context, event sdktypes.Event) error {
	eid := event.ID()

	z := d.L.With(zap.String("event_id", eid.String()))

	if !event.IsValid() {
		z.Error("could not find event")
		return sdkerrors.ErrNotFound
	}

	signals, err := d.DB.ListWaitingSignals(ctx, event.DestinationID())
	if err != nil {
		z.Error("could not fetch signals", zap.Error(err))
		return err
	}

	z.Debug(fmt.Sprintf("found %d signal candidates", len(signals)))

	for _, signal := range signals {
		match, err := event.Matches(signal.Filter)
		l := z.With(zap.String("signal_id", signal.SignalID.String()),
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

		if err := d.Client.Temporal().SignalWorkflow(ctx, signal.WorkflowID, "", signal.SignalID.String(), eid); err != nil {
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

func (d *Dispatcher) eventsWorkflow(wctx workflow.Context, input eventsWorkflowInput) (*eventsWorkflowOutput, error) {
	logger := workflow.GetLogger(wctx)

	logger.Info("started events workflow", "event_id", input.EventID)
	z := d.L.With(zap.String("event_id", input.EventID.String()))

	ctx := akCtx.WithRequestOrginator(context.Background(), akCtx.EventWorkflow)

	event, err := d.Events.Get(ctx, input.EventID)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	if !event.IsValid() {
		z.Error("could not find event")
		return nil, sdkerrors.ErrNotFound
	}

	// Set event to processing
	d.createEventRecord(ctx, input.EventID, sdktypes.EventStateProcessing)

	// Fetch event related data
	sds, err := d.getEventSessionData(ctx, event, input.Options)
	if err != nil {
		logger.Error("Failed processing event", "EventID", input.EventID, "error", err)
		d.createEventRecord(ctx, input.EventID, sdktypes.EventStateFailed)
		return nil, nil
	}

	sessionsData, err := createSessionsForWorkflow(event, sds)
	if err != nil {
		return nil, fmt.Errorf("create sessions: %w", err)
	}

	// start sessions
	if err := d.startSessions(wctx, event, sessionsData); err != nil {
		return nil, err
	}

	// execute waiting signals
	err = d.signalWorkflows(ctx, event)
	if err != nil {
		return nil, err
	}
	// Set event to Completed
	d.createEventRecord(ctx, input.EventID, sdktypes.EventStateCompleted)
	return nil, nil
}

func (d *Dispatcher) createEventRecord(ctx context.Context, eventID sdktypes.EventID, state sdktypes.EventState) {
	record := sdktypes.NewEventRecord(eventID, state)
	if err := d.Events.AddEventRecord(ctx, record); err != nil {
		d.L.Panic("Failed setting event state", zap.String("eventID", eventID.String()), zap.String("state", state.String()), zap.Error(err))
	}
}

func (d *Dispatcher) startSessions(wctx workflow.Context, event sdktypes.Event, sessions []sdktypes.Session) error {
	ctx := temporalclient.NewWorkflowContextAsGOContext(wctx)
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.SessionWorkflow)
	ctx = akCtx.WithOwnershipOf(ctx, d.DB.GetOwnership, event.ID().UUIDValue())

	for _, session := range sessions {
		// TODO(ENG-197): change to local activity.
		sessionID, err := d.Sessions.Start(ctx, session)
		if err != nil {
			d.L.Panic("could not start session") // Panic in order to make the workflow retry.
		}
		d.L.Info("started session", zap.String("session_id", sessionID.String()))
	}
	return nil
}

// used by both dispatcher and scheduler
func createSessionsForWorkflow(event sdktypes.Event, sessionsData []sessionData) ([]sdktypes.Session, error) {
	// DO NOT PASS Memo. It is not intended for automation use, just auditing.
	eventInputs := event.ToValues()
	eventStruct, err := sdktypes.NewStructValue(eventInputsSymbolValue, eventInputs)
	if err != nil {
		return nil, fmt.Errorf("start session: event: %w", err)
	}

	inputs := map[string]sdktypes.Value{
		"event": eventStruct,
		"data":  eventInputs["data"],
	}

	sessions := make([]sdktypes.Session, len(sessionsData))
	for i, sd := range sessionsData {
		if t := sd.Trigger; t.IsValid() {
			inputs = maps.Clone(inputs)
			triggerInputs := t.ToValues()

			if inputs["trigger"], err = sdktypes.NewStructValue(triggerInputsSymbolValue, triggerInputs); err != nil {
				return nil, fmt.Errorf("trigger: %w", err)
			}

			fs := inputs["data"].GetStruct().Fields()
			maps.Copy(fs, triggerInputs["data"].GetStruct().Fields())
			if inputs["data"], err = sdktypes.NewStructValue(dataSymbolValue, fs); err != nil {
				return nil, fmt.Errorf("data: %w", err)
			}

		}

		dep := sd.Deployment

		sessions[i] = sdktypes.NewSession(dep.BuildID(), sd.CodeLocation, inputs, nil).
			WithDeploymentID(dep.ID()).
			WithEventID(event.ID()).
			WithEnvID(dep.EnvID())
	}
	return sessions, nil
}
