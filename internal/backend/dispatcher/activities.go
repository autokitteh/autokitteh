package dispatcher

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/types"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	getEventSessionDataActivityName = "get_event_session_data"
	startSessionActivityName        = "start_session"
	listWaitingSignalsActivityName  = "list_waiting_signals"
	signalWorkflowActivityName      = "signal_workflow"
)

func (d *Dispatcher) registerActivities(w worker.Worker) {
	w.RegisterActivityWithOptions(
		d.getEventSessionDataActivity,
		activity.RegisterOptions{Name: getEventSessionDataActivityName},
	)

	w.RegisterActivityWithOptions(
		d.startSessionActivity,
		activity.RegisterOptions{Name: startSessionActivityName},
	)

	w.RegisterActivityWithOptions(
		d.listWaitingSignalsActivity,
		activity.RegisterOptions{Name: listWaitingSignalsActivityName},
	)

	w.RegisterActivityWithOptions(
		d.signalWorkflowActivity,
		activity.RegisterOptions{Name: signalWorkflowActivityName},
	)
}

type sessionData struct {
	Deployment   sdktypes.Deployment
	CodeLocation sdktypes.CodeLocation
	Trigger      sdktypes.Trigger
	Connection   sdktypes.Connection
}

func (d *Dispatcher) listWaitingSignalsActivity(ctx context.Context, dstid sdktypes.EventDestinationID) ([]*types.Signal, error) {
	sigs, err := d.svcs.DB.ListWaitingSignals(ctx, dstid)
	return sigs, temporalclient.TranslateError(err, "list waiting signals for %v", dstid)
}

func (d *Dispatcher) startSessionActivity(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.Dispatcher)
	ctx = akCtx.WithOwnershipOf(ctx, d.svcs.DB.GetOwnership, session.EnvID().UUIDValue())

	sid, err := d.svcs.Sessions.Start(ctx, session)
	return sid, temporalclient.TranslateError(err, "start session %v", session.ID())
}

func (d *Dispatcher) getEventSessionDataActivity(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) ([]sessionData, error) {
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.EventWorkflow)
	ctx = akCtx.WithOwnershipOf(ctx, d.svcs.DB.GetOwnership, event.DestinationID().UUIDValue())

	eid := event.ID()
	dstid := event.DestinationID()

	sl := d.sl.With("event_id", eid, "destination_id", dstid)

	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	if opts.Env != "" {
		sl = sl.With("env", opts.Env)
	}

	if opts.DeploymentID.IsValid() {
		sl = sl.With("deployment_id", opts.DeploymentID)
	}

	optsEnvID, err := d.resolveEnv(ctx, opts.Env)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			sl.Infof("env %q not found", opts.Env)
			return nil, nil
		}
		return nil, fmt.Errorf("env: %w", err)
	}
	if optsEnvID.IsValid() {
		sl = sl.With("env_id", optsEnvID)
	}

	var ts []sdktypes.Trigger

	if cid := dstid.ToConnectionID(); cid.IsValid() {
		if ts, err = d.svcs.Triggers.List(ctx, sdkservices.ListTriggersFilter{ConnectionID: cid}); err != nil {
			return nil, temporalclient.TranslateError(err, "list triggers for %v", cid)
		}

		sl.Infof("found %d triggers for connection %v", len(ts), cid)
	} else if tid := dstid.ToTriggerID(); tid.IsValid() {
		t, err := d.svcs.Triggers.Get(ctx, tid)
		if err != nil {
			return nil, temporalclient.TranslateError(err, "get trigger %v", tid)
		}

		ts = append(ts, t)
	}

	if len(ts) == 0 {
		sl.Infof("no triggers for event %v", eid)
		return nil, nil
	}

	var sds []sessionData

	deploymentsForEnv := make(map[sdktypes.EnvID][]sdktypes.Deployment)

	eventType := event.Type()
	for _, t := range ts {
		envID := t.EnvID()
		triggerEventType := t.EventType()

		sl := sl.With("trigger_id", t.ID())

		if !t.CodeLocation().IsValid() {
			sl.Info("no entry point, ignoring trigger")
			continue
		}

		if triggerEventType != "" && eventType != triggerEventType {
			sl.Infof("irrelevant event type %v != required %v", triggerEventType, eventType)
			continue
		}

		if !envID.IsValid() && optsEnvID.IsValid() && envID != optsEnvID {
			sl.Infof("irrelevant env %v != required %v", envID, optsEnvID)
			continue
		}

		if relevant, err := event.Matches(t.Filter()); err != nil {
			sl.With("err", err).Infof("filter error: %v", err)
			// TODO(ENG-566): alert user their filter is bad. Integrate with alerting and monitoring.
			continue
		} else if !relevant {
			sl.Info("irrelevant event")
			continue
		}

		deployments, found := deploymentsForEnv[envID]
		if !found {
			activeDeployments, err := d.svcs.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateActive, EnvID: envID})
			if err != nil {
				return nil, temporalclient.TranslateError(err, "list active deployments for %v", envID)
			}

			var testingDeployments []sdktypes.Deployment

			if optsEnvID.IsValid() || opts.DeploymentID.IsValid() {
				testingDeployments, err = d.svcs.Deployments.List(ctx, sdkservices.ListDeploymentsFilter{State: sdktypes.DeploymentStateTesting, EnvID: envID})
				if err != nil {
					return nil, temporalclient.TranslateError(err, "list testing deployments for %v", envID)
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
			sl.Info("no deployments for env")
			continue
		}

		var c sdktypes.Connection
		if cid := t.ConnectionID(); cid.IsValid() { // only if this trigger has conneciton defined
			c, err = d.svcs.Connections.Get(ctx, t.ConnectionID())
			if err != nil {
				sl.With("err", err).Errorf("could not fetch connection %v: %v", t.ConnectionID(), err)
				// fallthrough - not critical, informational only.
			}
		}

		cl := t.CodeLocation()
		for _, dep := range deployments {
			sds = append(sds, sessionData{Deployment: dep, CodeLocation: cl, Trigger: t, Connection: c})
			sl.Infof("relevant deployment %v found for %v", dep.ID(), eid)
		}
	}

	return sds, nil
}

func (d *Dispatcher) signalWorkflowActivity(ctx context.Context, wid string, sigid uuid.UUID, eid sdktypes.EventID) error {
	sl := d.sl.With("workflow_id", wid, "signal_id", sigid, "event_id", eid)

	if err := d.svcs.LazyTemporalClient().SignalWorkflow(ctx, wid, "", sigid.String(), eid); err != nil {
		var nferr *serviceerror.NotFound
		if errors.As(err, &nferr) {
			sl.Warnf("workflow %v not found for %v - removing signal", wid, sigid)

			if err := d.svcs.DB.RemoveSignal(ctx, sigid); err != nil {
				sl.With("err", err).Error("db remove signal %v: %w", sigid, err)
				return temporalclient.TranslateError(err, "remove signal %v", sigid)
			}

			// might be some race condition after workflow was done, not really an error.
			return nil
		}

		sl.With("err", err).Errorf("signal workflow %v for %v: %v", wid, sigid, err)

		return temporalclient.TranslateError(err, "signal workflow %v for %v", wid, sigid)
	}

	return nil
}
