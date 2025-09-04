package sessionworkflows

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/types"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	updateSessionStateActivityName          = "update_session_state"
	terminateWorkflowActivityName           = "terminate_workflow"
	saveSignalActivityName                  = "save_signal"
	getLastEventSequenceActivityName        = "get_last_event_sequence"
	getSessionStopReasonActivityName        = "get_session_stop_reason"
	getSignalEventActivityName              = "get_signal_event"
	removeSignalActivityName                = "remove_signal"
	deactivateDrainedDeploymentActivityName = "deactivate_drained_deployment"
	getDeploymentStateActivityName          = "get_deployment_state"
	createSessionActivityName               = "create_session"
	getProjectIDAndActiveBuildID            = "get_project_id_and_active_build_id"
	listStoreValuesActivityName             = "list_store_values"
	mutateStoreValueActivityName            = "mutate_store_value"
	notifyWorkflowEndedActivity             = "notify_workflow_ended"
	startChildSessionActivityName           = "start_child_session"
)

func (ws *workflows) registerActivities() {
	// Utils Worker activities

	// We need to register the terminate workflow activity on the utils worker,
	// since it is used to terminate workflows and should not be registered on the sessions worker.
	ws.utilsWorker.RegisterActivityWithOptions(
		ws.updateSessionStateActivity,
		activity.RegisterOptions{Name: updateSessionStateActivityName},
	)

	ws.utilsWorker.RegisterActivityWithOptions(
		ws.terminateWorkflowActivity,
		activity.RegisterOptions{Name: terminateWorkflowActivityName},
	)

	// Session Worker activities
	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.updateSessionStateActivity,
		activity.RegisterOptions{Name: updateSessionStateActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.saveSignalActivity,
		activity.RegisterOptions{Name: saveSignalActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.getLatestEventSequenceActivity,
		activity.RegisterOptions{Name: getLastEventSequenceActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.getSessionStopReasonActivity,
		activity.RegisterOptions{Name: getSessionStopReasonActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.getSignalEventActivity,
		activity.RegisterOptions{Name: getSignalEventActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.removeSignalActivity,
		activity.RegisterOptions{Name: removeSignalActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.deactivateDrainedDeploymentActivity,
		activity.RegisterOptions{Name: deactivateDrainedDeploymentActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.getDeploymentStateActivity,
		activity.RegisterOptions{Name: getDeploymentStateActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.createSessionActivity,
		activity.RegisterOptions{Name: createSessionActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.listStoreValuesActivity,
		activity.RegisterOptions{Name: listStoreValuesActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.mutateStoreValueActivity,
		activity.RegisterOptions{Name: mutateStoreValueActivityName},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.getProjectIDAndActiveBuildID,
		activity.RegisterOptions{Name: getProjectIDAndActiveBuildID},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.notifyWorkflowEndedActivity,
		activity.RegisterOptions{Name: notifyWorkflowEndedActivity},
	)

	ws.sessionsWorker.RegisterActivityWithOptions(
		ws.startChildSessionActivity,
		activity.RegisterOptions{Name: startChildSessionActivityName},
	)
}

type getProjectIDAndActiveBuildIDParams struct {
	OrgID   sdktypes.OrgID
	Project sdktypes.Symbol
}

type getProjectIDAndActiveBuildIDResponse struct {
	BuildID   sdktypes.BuildID
	ProjectID sdktypes.ProjectID
}

func (ws *workflows) getProjectIDAndActiveBuildID(ctx context.Context, params getProjectIDAndActiveBuildIDParams) (*getProjectIDAndActiveBuildIDResponse, error) {
	p, err := ws.svcs.Projects.GetByName(authcontext.SetAuthnSystemUser(ctx), params.OrgID, params.Project)
	if err != nil {
		return nil, temporalclient.TranslateError(err, "get project %v", params.Project)
	}

	ds, err := ws.svcs.Deployments.List(
		authcontext.SetAuthnSystemUser(ctx),
		sdkservices.ListDeploymentsFilter{
			OrgID:     p.OrgID(),
			ProjectID: p.ID(),
			State:     sdktypes.DeploymentStateActive,
			Limit:     1,
		},
	)
	if err != nil {
		return nil, temporalclient.TranslateError(err, "list deployments for project %v", p.ID())
	}

	if len(ds) == 0 {
		return nil, temporalclient.TranslateError(sdkerrors.ErrNotFound, "no active deployment for project")
	}

	d := ds[0]

	return &getProjectIDAndActiveBuildIDResponse{
		BuildID:   d.BuildID(),
		ProjectID: p.ID(),
	}, nil
}

func (ws *workflows) listStoreValuesActivity(ctx context.Context, pid sdktypes.ProjectID) ([]string, error) {
	return ws.svcs.Store.List(authcontext.SetAuthnSystemUser(ctx), pid)
}

func (ws *workflows) mutateStoreValueActivity(ctx context.Context, pid sdktypes.ProjectID, key, op string, operands []sdktypes.Value) (sdktypes.Value, error) {
	return ws.svcs.Store.Mutate(authcontext.SetAuthnSystemUser(ctx), pid, key, op, operands...)
}

func (ws *workflows) createSessionActivity(ctx context.Context, session sdktypes.Session) error {
	return temporalclient.TranslateError(ws.svcs.DB.CreateSession(ctx, session), "%v: create session", session.ID())
}

func (ws *workflows) updateSessionStateActivity(ctx context.Context, sid sdktypes.SessionID, state sdktypes.SessionState) error {
	return temporalclient.TranslateError(ws.svcs.DB.UpdateSessionState(ctx, sid, state), "%v: update session state", sid)
}

func (ws *workflows) getDeploymentStateActivity(ctx context.Context, did sdktypes.DeploymentID) (sdktypes.DeploymentState, error) {
	d, err := ws.svcs.Deployments.Get(authcontext.SetAuthnSystemUser(ctx), did)
	if err != nil {
		return sdktypes.DeploymentStateUnspecified, temporalclient.TranslateError(err, "%v: get deployment state", did)
	}

	return d.State(), nil
}

func (ws *workflows) removeSignalActivity(ctx context.Context, sigid uuid.UUID) error {
	return temporalclient.TranslateError(ws.svcs.DB.RemoveSignal(ctx, sigid), "%v: remove signal", sigid)
}

func (ws *workflows) getLatestEventSequenceActivity(ctx context.Context) (uint64, error) {
	seq, err := ws.svcs.DB.GetLatestEventSequence(ctx)
	err = temporalclient.TranslateError(err, "get latest event sequence")
	return seq, err
}

func (ws *workflows) deactivateDrainedDeploymentActivity(ctx context.Context, did sdktypes.DeploymentID) error {
	sl := ws.l.Sugar().With("deployment_id", did)

	drained, err := ws.svcs.DB.DeactivateDrainedDeployment(ctx, did)
	if err != nil {
		return temporalclient.TranslateError(err, "deactivate drained deployments")
	}

	if drained {
		sl.Infof("deactivated drained deployment")
	}

	return nil
}

func (ws *workflows) getSignalEventActivity(ctx context.Context, sigid uuid.UUID, minSeq uint64) (sdktypes.Event, error) {
	sl := ws.l.Sugar().With("signal_id", sigid, "seq", minSeq)

	signal, err := ws.svcs.DB.GetSignal(ctx, sigid)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			return sdktypes.InvalidEvent, nil
		}

		return sdktypes.InvalidEvent, temporalclient.TranslateError(err, "get signal %v", sigid)
	}

	filter := sdkservices.ListEventsFilter{
		DestinationID:     signal.DestinationID,
		Limit:             1,
		MinSequenceNumber: minSeq + 1,
		Order:             sdkservices.ListOrderAscending,
	}

	for {
		evs, err := ws.svcs.DB.ListEvents(ctx, filter)
		if err != nil {
			return sdktypes.InvalidEvent, temporalclient.TranslateError(err, "list events for %v minSeq: %v", signal.DestinationID, minSeq)
		}

		if len(evs) == 0 {
			sl.Debug("no events found")
			return sdktypes.InvalidEvent, nil
		}

		eid := evs[0].ID()

		event, err := ws.svcs.DB.GetEventByID(ctx, eid)
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotFound) {
				sl.With("event_id", eid).Warnf("event %v not found", eid)
				continue
			}

			return sdktypes.InvalidEvent, temporalclient.TranslateError(err, "get event %v", eid)
		}

		filter.MinSequenceNumber = event.Seq() + 1

		match, err := event.Matches(signal.Filter)
		if err != nil {
			// TODO(ENG-566): inform user.
			sl.Info("invalid signal filter", zap.Error(err), zap.String("filter", signal.Filter))
			continue
		}

		if match {
			return event, nil
		}
	}
}

func (ws *workflows) getSessionStopReasonActivity(ctx context.Context, sid sdktypes.SessionID) (string, error) {
	log, err := ws.svcs.DB.GetSessionLog(ctx, sdkservices.SessionLogRecordsFilter{SessionID: sid})
	if err != nil {
		return "", temporalclient.TranslateError(err, "get session log for %v", sid)
	}

	for _, rec := range log.Records {
		if r, ok := rec.GetStopRequest(); ok {
			return r, nil
		}
	}

	return "<unknown>", nil
}

func (ws *workflows) saveSignalActivity(ctx context.Context, signal *types.Signal) error {
	if err := ws.svcs.DB.SaveSignal(ctx, signal); err != nil {
		if errors.Is(err, sdkerrors.ErrAlreadyExists) {
			// ignore error: since siganlID is unique - this means we got replayed/retried here and the signal was already saved prior.
			ws.l.Sugar().With("signal_id", signal.ID).Warnf("signal %v already saved", signal.ID)
			return nil
		}
		return temporalclient.TranslateError(err, "save signal %v", signal.ID)
	}

	return nil
}

type terminateSessionWorkflowParams struct {
	SessionID sdktypes.SessionID
	Reason    string
	Delay     time.Duration
}

func (ws *workflows) legacyTerminateSessionWorkflow(wctx workflow.Context, sid sdktypes.SessionID, reason string) error {
	return ws.terminateSessionWorkflow(wctx, terminateSessionWorkflowParams{SessionID: sid, Reason: reason})
}

func (ws *workflows) terminateSessionWorkflow(wctx workflow.Context, params terminateSessionWorkflowParams) error {
	sid, reason := params.SessionID, params.Reason

	sl := ws.l.Sugar().With("session_id", sid)

	if t := params.Delay; t > 0 {
		sl.With("delay", t).Infof("waiting %v before terminatiion", t)

		if err := workflow.Sleep(wctx, t); err != nil {
			sl.With("err", err).Errorf("sleep error: %v", err)
			return temporalclient.TranslateError(err, "sleep")
		}
	}

	sl.Infof("terminating session workflow %s", sid)

	wctx = workflow.WithActivityOptions(wctx, ws.cfg.Activity.ToOptions(utilsWorkerQueue))
	// this is fine if it runs multiple times and should be short.
	if err := workflow.ExecuteActivity(wctx, terminateWorkflowActivityName, sid, reason).Get(wctx, nil); err != nil {
		sl.With("err", err).Errorf("terminate workflow %v activity: %v", sid, err)
		return temporalclient.TranslateError(err, "terminate workflow %v", sid)
	}

	// the terminated workflow should not be active at this point. in this case there should be no concurrent
	// updates with the below.

	if err := ws.updateSessionState(wctx, sid, sdktypes.NewSessionStateStopped(reason)); err != nil {
		sl.With("err", err).Errorf("update session %v state error: %w", sid, err)
	}

	sl.Infof("terminated session workflow %s", sid)

	return nil
}

func (ws *workflows) terminateWorkflowActivity(ctx context.Context, sid sdktypes.SessionID, reason string) error {
	err := ws.svcs.Temporal.TemporalClient().TerminateWorkflow(ctx, workflowID(sid), "", reason)
	if err != nil {
		// might happen multiple times for some reason, give it a chance to update the state later on.
		var notFound *serviceerror.NotFound
		if errors.As(err, &notFound) {
			err = nil
		}
	}

	return err
}

func (ws *workflows) notifyWorkflowEndedActivity(ctx context.Context, sid sdktypes.SessionID) error {
	return ws.svcs.WorkflowExecutor.NotifyDone(ctx, workflowID(sid))
}

func (ws *workflows) startChildSessionActivity(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
	return ws.sessions.Start(ctx, session)
}
