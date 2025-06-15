package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	terminateSessionWorkflowName        = "terminate_session"
	delayedTerminateSessionWorkflowName = "delayed_terminate_session"
)

type StartWorkflowOptions struct{}

type Workflows interface {
	StartWorkers(context.Context) error
	StartWorkflow(ctx context.Context, session sdktypes.Session, opts StartWorkflowOptions) error
	StartChildWorkflow(wctx workflow.Context, data sessiondata.Data) (workflow.ChildWorkflowFuture, error)
	GetWorkflowLog(ctx context.Context, filter sdkservices.SessionLogRecordsFilter) (*sdkservices.GetLogResults, error)
	StopWorkflow(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool, cancelTimeout time.Duration) error
}

type sessionWorkflowParams struct {
	Data sessiondata.Data
	Opts StartWorkflowOptions
}

type workflows struct {
	l        *zap.Logger
	cfg      Config
	worker   worker.Worker
	svcs     *sessionsvcs.Svcs
	sessions sdkservices.Sessions
	calls    sessioncalls.Calls
}

func workflowID(sessionID sdktypes.SessionID) string { return sessionID.String() }

func New(
	l *zap.Logger,
	cfg Config,
	sessions sdkservices.Sessions,
	svcs *sessionsvcs.Svcs,
	calls sessioncalls.Calls,
) Workflows {
	initMetrics()
	return &workflows{l: l, cfg: cfg, sessions: sessions, calls: calls, svcs: svcs}
}

func (ws *workflows) StartWorkers(ctx context.Context) error {
	ws.worker = temporalclient.NewWorker(ws.l.Named("sessionworkflowsworker"), ws.svcs.Temporal.TemporalClient(), ws.svcs.WorkflowExecutor.WorkflowQueue(), ws.cfg.Worker)
	if ws.worker == nil {
		return nil
	}

	ws.worker.RegisterWorkflowWithOptions(
		ws.sessionWorkflow,
		workflow.RegisterOptions{Name: ws.svcs.WorkflowExecutor.WorkflowSessionName()},
	)

	// Legacy termination workflow, using explicit parameters instead of a struct.
	ws.worker.RegisterWorkflowWithOptions(
		ws.legacyTerminateSessionWorkflow,
		workflow.RegisterOptions{Name: terminateSessionWorkflowName},
	)

	ws.worker.RegisterWorkflowWithOptions(
		ws.terminateSessionWorkflow,
		workflow.RegisterOptions{Name: delayedTerminateSessionWorkflowName},
	)

	ws.registerActivities()

	return ws.worker.Start()
}

func memo(session sdktypes.Session, oid sdktypes.OrgID) map[string]string {
	memo := map[string]string{
		"process_id":      fixtures.ProcessID(),
		"session_id":      session.ID().String(),
		"session_uuid":    session.ID().UUIDValue().String(),
		"deployment_id":   session.DeploymentID().String(),
		"deployment_uuid": session.DeploymentID().UUIDValue().String(),
		"project_id":      session.ProjectID().String(),
		"project_uuid":    session.ProjectID().UUIDValue().String(),
		"org_id":          oid.String(),
		"org_uuid":        oid.UUIDValue().String(),
	}

	if session.ParentSessionID().IsValid() {
		memo["parent_session_id"] = session.ParentSessionID().String()
		memo["parent_session_uuid"] = session.ParentSessionID().UUIDValue().String()
	}

	for k, v := range session.Memo() {
		memo["session__"+k] = v
	}

	return memo
}

func (ws *workflows) StartChildWorkflow(wctx workflow.Context, data sessiondata.Data) (workflow.ChildWorkflowFuture, error) {
	childSessionID := data.Session.ID()

	l := ws.l.With(zap.Any("child_session_id", childSessionID))

	memo := memo(data.Session, data.OrgID)

	f := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(
			wctx,
			ws.cfg.SessionWorkflow.ToChildWorkflowOptions(
				ws.svcs.WorkflowExecutor.WorkflowQueue(),
				workflowID(childSessionID),
				fmt.Sprintf("session %v", childSessionID),
				enums.PARENT_CLOSE_POLICY_ABANDON,
				memo,
			),
		),
		ws.svcs.WorkflowExecutor.WorkflowSessionName(),
		sessionWorkflowParams{Data: data},
	)

	var r workflow.Execution
	if err := f.GetChildWorkflowExecution().Get(wctx, &r); err != nil {
		return nil, fmt.Errorf("child workflow execution: %w", err)
	}

	l.With(zap.String("workflow_id", r.ID), zap.String("run_id", r.RunID), zap.Any("memo", memo)).Info("initiated child session workflow")

	return f, nil
}

func (ws *workflows) StartWorkflow(ctx context.Context, session sdktypes.Session, opts StartWorkflowOptions) error {

	sessionID := session.ID()

	// l := ws.l.Sugar().With("session_id", sessionID)

	data, err := sessiondata.Get(ctx, ws.svcs, session)
	if err != nil {
		return fmt.Errorf("get session data: %w", err)
	}

	memo := memo(session, data.OrgID)

	params := sessionWorkflowParams{
		Data: *data,
		Opts: opts,
	}
	if err = ws.svcs.WorkflowExecutor.Execute(ctx, sessionID, params, memo); err != nil {
		return fmt.Errorf("execute session workflow: %w", err)
	}

	return nil
}

func (ws *workflows) sessionWorkflow(wctx workflow.Context, params sessionWorkflowParams) error {
	wi := workflow.GetInfo(wctx)
	session := params.Data.Session
	parentSessionID := session.ParentSessionID()
	sid := session.ID()
	isReplaying := workflow.IsReplaying(wctx)

	l := ws.l.With(
		zap.String("session_id", sid.String()),
		zap.Bool("replay", isReplaying),
		zap.String("workflow_id", wi.WorkflowExecution.ID),
		zap.String("run_id", wi.WorkflowExecution.RunID),
		zap.Int32("attempt", wi.Attempt),
		zap.String("parent_session_id", parentSessionID.String()),
	)

	eventTime := struct {
		createdAt time.Time
		valid     bool
	}{
		createdAt: time.Time{},
		valid:     false,
	}

	if eventWrapped, ok := session.Inputs()["event"]; ok {
		if eventUnwrapped, err := sdktypes.UnwrapValue(eventWrapped); err == nil {
			if eventMap, ok := eventUnwrapped.(map[string]any); ok {
				if createdAt, ok := eventMap["created_at"].(time.Time); ok {
					eventTime.createdAt = createdAt
					eventTime.valid = true
				}
			}
		}
	}

	wctx = temporalclient.WithActivityOptions(wctx, ws.svcs.WorkflowExecutor.WorkflowQueue(), ws.cfg.Activity)

	// metrics is using context for the data contained, not for cancellations, as it stores in memory.
	// if it will be stuck anyway for some reason, the workflow deadlock timeout would kick in.
	metricsCtx := context.Background()

	// TODO(ENG-322): Save data in snapshot, otherwise changes between retries would
	//                blow us up due to non determinism.

	if isReplaying {
		if params.Data.Session.State().IsFinal() {
			// HACK: If we somehow get a signal to this workflow after it completed,
			//       in certain delicate timing, the workflow rekicks in replay.
			//       Though replay would work just fine, and result in a no-op,
			//       it's nice to not have to run through its entirely. Temporal
			//       just seems to ignore it...
			l.Info("already completed")
			sessionStaleReplaysCounter.Add(metricsCtx, 1)
			return nil
		}
	}

	l.Info("session workflow started")

	sessionsCreatedCounter.Add(
		metricsCtx,
		1,
		metric.WithAttributes(attribute.Bool("replay", isReplaying)),
	)

	startTime := time.Now() // we want actual start time for metrics.
	prints, retVal, err := runWorkflow(wctx, l, ws, params)
	duration := time.Since(startTime)

	// from this point on we should not be doing anything really that should be cancelled.
	// the original wctx might have been cancelled, so we work with a disconnected context
	// to avoid any belated non-timely cancellations.
	dwctx, done := workflow.NewDisconnectedContext(wctx)
	defer done()

	sessionDurationHistogram.Record(metricsCtx, duration.Milliseconds(),
		metric.WithAttributes(attribute.Bool("replay", isReplaying), attribute.Bool("success", err == nil)))

	l = l.With(zap.Duration("duration", duration))

	if eventTime.valid {
		invocationDelay := time.Since(eventTime.createdAt)
		l = l.With(zap.Duration("invocation_delay", invocationDelay))

		if !isReplaying {
			sessionInvocationDelayHistogram.Record(metricsCtx, invocationDelay.Milliseconds())
		}
	}

	workflowErr := err

	if err != nil {
		l := l.With(zap.Error(err))

		if wctxErr := wctx.Err(); errors.Is(err, workflow.ErrCanceled) || errors.Is(wctxErr, workflow.ErrCanceled) {
			sessionsStoppedCounter.Add(metricsCtx, 1)

			l.With(zap.Any("ctx_err", wctxErr)).Info("session workflow canceled")

			ws.stopped(dwctx, sid)
		} else {
			ws.errored(dwctx, sid, err, kittehs.Transform(prints, func(p sdkservices.SessionPrint) string { return p.Value.GetString().Value() }))

			if _, ok := sdktypes.FromError(err); ok {
				// User level error (convertible to ProgramError).
				l.Info("session workflow program error")
				sessionsProgramErrorsCounter.Add(metricsCtx, 1)

				// No need to indicate the workflow as errored.
				workflowErr = nil
			} else {
				l.Sugar().Errorf("session workflow error: %v", err)
				if sdkerrors.IsRetryableError(err) {
					sessionsRetryErrorsCounter.Add(metricsCtx, 1)
					l.Panic("panic session to retry")
				}
				sessionsErroredCounter.Add(metricsCtx, 1)
			}
		}
	} else {
		sessionsCompletedCounter.Add(metricsCtx, 1)
		l.Info("session workflow completed with no errors")
	}

	// Signal parent session on completion.
	if parentSessionID.IsValid() {
		payload := map[string]sdktypes.Value{
			"completed": sdktypes.NewBooleanValue(err == nil),
		}

		if err == nil {
			payload["value"] = retVal
		}

		if err := workflow.SignalExternalWorkflow(
			dwctx,
			parentSessionID.String(),
			"",
			sessionSignalName(session.ID()),
			sdktypes.NewDictValueFromStringMap(payload),
		).Get(dwctx, nil); err != nil {
			l.With(zap.Error(err)).Sugar().Errorf("signal parent session: %v", err)
		}
	}

	_ = workflow.ExecuteActivity(wctx, deactivateDrainedDeploymentActivityName, session.DeploymentID()).Get(wctx, nil)
	_ = workflow.ExecuteActivity(wctx, notifyWorkflowEndedActivity, session.ID()).Get(wctx, nil)

	return workflowErr
}

// workflow is stopped, so this assumes the given wctx is a disconnected context.
func (ws *workflows) stopped(wctx workflow.Context, sessionID sdktypes.SessionID) {
	var reason string

	if err := workflow.ExecuteActivity(wctx, getSessionStopReasonActivityName, sessionID).Get(wctx, &reason); err != nil {
		// error here is always a temporal error since the local activity above would never return an error.
		// it is just nice like that.
		// in any case, getting the reason is not critical, so we just log it and move on.
		ws.l.Sugar().With("session_id", sessionID.String(), "err", err).Errorf("get stop reason error: %v", err)
		reason = "<unknown>"
	}

	_ = ws.updateSessionState(wctx, sessionID, sdktypes.NewSessionStateStopped(reason))
}

func (ws *workflows) errored(wctx workflow.Context, sessionID sdktypes.SessionID, err error, prints []string) {
	_ = ws.updateSessionState(wctx, sessionID, sdktypes.NewSessionStateError(err, prints))
}

func (ws *workflows) StopWorkflow(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool, forceDelay time.Duration) error {
	s, err := ws.svcs.DB.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	if s.State().IsFinal() {
		return fmt.Errorf("%w: session already in final state", sdkerrors.ErrConflict)
	}

	// Always log termination request.
	loggedReason := reason
	if force {
		loggedReason = "[forced] " + reason
	}

	if err := ws.svcs.DB.AddSessionStopRequest(ctx, sessionID, loggedReason); err != nil {
		return err
	}

	wid := workflowID(sessionID)

	// Always first try to terminate politely.
	//
	// Since this cancellation is polite, it is not guaranteed to be successful.
	// We should not be waiting for the cancellation to actually go through, just ask
	// temporal nicely to do it, unlike the forceful termination done below.
	//
	// When the workflow is actually stopped, the session state will be updated to stopped by
	// the workflow itself.
	if err := ws.svcs.Temporal.TemporalClient().CancelWorkflow(ctx, wid, ""); err != nil {
		var notFound *serviceerror.NotFound
		if errors.As(err, &notFound) {
			// workflow might have already ended?
			s, err := ws.svcs.DB.GetSession(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("get session: %w", err)
			}

			if s.State().IsFinal() {
				return nil
			}

			// nope! workflow not found although we still think it's there - we might have lost it?
			ws.l.Error("workflow lost", zap.String("workflow_id", wid))
			if err := ws.svcs.DB.UpdateSessionState(ctx, sessionID, sdktypes.NewSessionStateError(errors.New("workflow lost"), nil)); err != nil {
				return fmt.Errorf("update session state: %w", err)
			}

			return nil
		}

		return err
	}

	if !force {
		// This is not a forced stop, we do not need to wait for the workflow to actually stop.
		return nil
	}

	// run the termination in a separate workflow to avoid having the workflow terminated but not updated in
	// the db if the caller croaks.
	r, err := ws.svcs.Temporal.TemporalClient().ExecuteWorkflow(
		ctx,
		ws.cfg.TerminationWorkflow.ToStartWorkflowOptions(
			ws.svcs.WorkflowExecutor.WorkflowQueue(),
			"terminate_"+wid,
			fmt.Sprintf("stop %v", sessionID),
			map[string]string{
				"process_id":   fixtures.ProcessID(),
				"session_id":   sessionID.String(),
				"session_uuid": sessionID.UUIDValue().String(),
			},
		),
		delayedTerminateSessionWorkflowName,
		&terminateSessionWorkflowParams{
			SessionID: sessionID,
			Reason:    reason,
			Delay:     forceDelay,
		},
	)
	if err != nil {
		return fmt.Errorf("execute terminate workflow: %w", err)
	}

	if forceDelay == 0 {
		// wait for the deed to be done, as the termination workflow itself should not block on anything,
		// this should be quick.
		return r.Get(ctx, nil)
	}

	return nil
}

func (ws *workflows) updateSessionState(wctx workflow.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error {
	ws.l.Sugar().With("session_id", sessionID, "state", state.Type()).Infof("updating session state to %v", state.Type())
	return workflow.ExecuteActivity(wctx, updateSessionStateActivityName, sessionID, state).Get(wctx, nil)
}
