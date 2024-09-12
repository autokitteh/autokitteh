package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const workflowDeadlockTimeout = time.Second * 10

type Workflows interface {
	StartWorkers(context.Context) error
	StartWorkflow(ctx context.Context, session sdktypes.Session, debug bool) error
	StopWorkflow(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool) error
}

type sessionWorkflowParams struct {
	SessionID sdktypes.SessionID
	Debug     bool
}

type workflows struct {
	z        *zap.Logger
	cfg      Config
	worker   worker.Worker
	svcs     *sessionsvcs.Svcs
	sessions sdkservices.Sessions
	calls    sessioncalls.Calls
}

const (
	taskQueueName       = "session_workflows"
	sessionWorkflowName = "session_workflow"
)

func workflowID(sessionID sdktypes.SessionID) string { return sessionID.String() }

func New(z *zap.Logger,
	cfg Config,
	sessions sdkservices.Sessions,
	svcs *sessionsvcs.Svcs,
	calls sessioncalls.Calls,
	telemetry *telemetry.Telemetry,
) Workflows {
	initMetrics(telemetry)
	return &workflows{z: z, cfg: cfg, sessions: sessions, calls: calls, svcs: svcs}
}

func (ws *workflows) StartWorkers(ctx context.Context) error {
	opts := ws.cfg.Temporal.Worker
	opts.DisableRegistrationAliasing = true
	opts.MaxConcurrentWorkflowTaskExecutionSize = 150
	opts.OnFatalError = func(err error) { ws.z.Error("temporal worker error", zap.Error(err)) }
	opts.DeadlockDetectionTimeout = workflowDeadlockTimeout

	ws.worker = worker.New(ws.svcs.TemporalClient(), taskQueueName, opts)

	ws.worker.RegisterWorkflowWithOptions(
		ws.sessionWorkflow,
		workflow.RegisterOptions{Name: sessionWorkflowName},
	)

	return ws.worker.Start()
}

func (ws *workflows) StartWorkflow(ctx context.Context, session sdktypes.Session, debug bool) error {
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.SessionWorkflow)

	sessionID := session.ID()

	wid := workflowID(sessionID)

	z := ws.z.With(zap.String("session_id", sessionID.String()), zap.String("workflow_id", wid))

	// NOTE: If we have a crash here, after CreateSession and before ExecuteWorkflow
	//       we might have a zombie session id. It should be ok as it is rare
	//       and the user will get a 501.

	memo := map[string]string{
		"session_id":    sessionID.Value().String(),
		"deployment_id": session.DeploymentID().String(),
		"entrypoint":    session.EntryPoint().CanonicalString(),
		"workflow_id":   wid,
	}
	maps.Copy(memo, session.Memo())

	swopts := client.StartWorkflowOptions{
		ID:                  workflowID(sessionID),
		TaskQueue:           taskQueueName,
		WorkflowTaskTimeout: ws.cfg.Temporal.WorkflowTaskTimeout,
		Memo:                kittehs.TransformMapValues(memo, func(s string) any { return s }),
	}

	r, err := ws.svcs.TemporalClient().ExecuteWorkflow(
		ctx,
		swopts,
		sessionWorkflowName,
		&sessionWorkflowParams{SessionID: sessionID, Debug: debug},
	)
	if err != nil {
		return fmt.Errorf("execute session workflow: %w", err)
	}

	z.Info("executed session workflow", zap.String("workflow_run_id", r.GetRunID()))

	return nil
}

func (ws *workflows) getSessionData(wctx workflow.Context, sessionID sdktypes.SessionID) (*sessiondata.Data, error) {
	ctx := temporalclient.NewWorkflowContextAsGOContext(wctx)
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.SessionWorkflow)

	// This cannot run through activity as it would expose potentialy sensitive data to temporal.
	data, err := sessiondata.Get(ctx, ws.z, ws.svcs, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session data: %w", err)
	}

	return data, nil
}

func (ws *workflows) cleanupSession(ctx context.Context, data *sessiondata.Data) {
	z := ws.z.With(zap.String("session_id", data.SessionID.String()))

	if depID := data.Session.DeploymentID(); depID.IsValid() {
		// We cannot rely on workflow context here as it might have been canceled.
		go func() {
			ctx, cancel := withLimitedTimeout(ctx)
			defer cancel()

			if err := ws.deactivateDrainedDeployment(ctx, depID); err != nil {
				z.Error("deactivate drained deployment failed", zap.Error(err))
			}
		}()
	}
}

func (ws *workflows) getSessionDebugData(data *sessiondata.Data, prints []string) any {
	z := ws.z.With(zap.String("session_id", data.SessionID.String()))

	// We use background as the workflow might have been canceled.
	ctx, cancel := withLimitedTimeout(context.Background())
	defer cancel()
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.SessionWorkflow)

	history, err := ws.sessions.GetLog(ctx, sdkservices.ListSessionLogRecordsFilter{SessionID: data.SessionID})
	if err != nil {
		z.Warn("get history failed", zap.Error(err))
		return "error retreiving history"
	}

	return struct {
		Prints  []string
		History sdktypes.SessionLog
	}{
		Prints:  prints,
		History: history.Log,
	}
}

func (ws *workflows) sessionWorkflow(wctx workflow.Context, params *sessionWorkflowParams) (debug any, _ error) {
	sessionID := params.SessionID.String()
	wi := workflow.GetInfo(wctx)
	l := ws.z.With(
		zap.String("session_id", sessionID),
		zap.Bool("replay", workflow.IsReplaying(wctx)),
		zap.String("workflow_id", wi.WorkflowExecution.ID),
		zap.String("run_id", wi.WorkflowExecution.RunID),
	)

	wctx = workflow.WithLocalActivityOptions(wctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: ws.cfg.Temporal.LocalScheduleToCloseTimeout,
	})

	ctx := akCtx.WithRequestOrginator(context.Background(), akCtx.SessionWorkflow)

	// TODO(ENG-322): Save data in snapshot, otherwise changes between retries would
	//                blow us up due to non determinism.

	var prints []string
	var duration time.Duration
	var startTime time.Time
	fields := []zap.Field{}

	data, err := ws.getSessionData(wctx, params.SessionID)
	if err == nil {

		if workflow.IsReplaying(wctx) {
			if data.Session.State().IsFinal() {
				// HACK: If we somehow get a signal to this workflow after it completed,
				//       in certain delicate timing, the workflow rekicks in replay.
				//       Though replay would work just fine, and result in a no-op,
				//       it's nice to not have to run through its entirely. Temporal
				//       just seems to ignore it...
				l.Info("already completed")
				return nil, nil
			}
			fields = append(fields, zap.Int32("attempt", wi.Attempt))
		} else {
			sessionsCreatedCounter.Add(ctx, 1)
		}

		startTime = data.Session.CreatedAt()
		l.Info("session workflow: started", fields...)

		prints, err = runWorkflow(wctx, l, ws, data, params.Debug)

		sessionDurationHistogram.Record(
			ctx,
			time.Since(startTime).Milliseconds(),
			metric.WithAttributes(
				attribute.Bool("replay", workflow.IsReplaying(wctx)),
				attribute.Bool("success", err == nil),
			),
		)

		l = l.With(zap.Duration("duration", duration))
	}

	if err != nil {
		l := l.With(zap.Error(err))

		if errors.Is(err, workflow.ErrCanceled) || errors.Is(wctx.Err(), workflow.ErrCanceled) {
			sessionsStoppedCounter.Add(ctx, 1)

			l.Info("session workflow: canceled")
			ws.stopped(ctx, params.SessionID)
		} else {
			sessionsErroredCounter.Add(ctx, 1)
			ws.errored(ctx, params.SessionID, err, prints)

			if _, ok := sdktypes.FromError(err); ok {
				// User level error, no need to indicate the workflow as errored.
				err = nil

				l.Info("session workflow: program error")
			} else {
				l.Error("session workflow: error")
			}
		}
	} else {
		sessionsCompletedCounter.Add(ctx, 1)
		l.Info("session workflow: completed")
	}

	if data != nil {
		ws.cleanupSession(ctx, data)

		if params.Debug {
			debug = ws.getSessionDebugData(data, prints)
		}
	}

	return debug, err
}

func (ws *workflows) stopped(ctx context.Context, sessionID sdktypes.SessionID) {
	ctx, cancel := withLimitedTimeout(ctx)
	defer cancel()

	reason := "<unknown>"

	if log, err := ws.svcs.DB.GetSessionLog(ctx, sdkservices.ListSessionLogRecordsFilter{SessionID: sessionID}); err == nil {
		for _, rec := range log.Log.Records() {
			if r, ok := rec.GetStopRequest(); ok {
				reason = r
				break
			}
		}
	}

	_ = ws.updateSessionState(ctx, sessionID, sdktypes.NewSessionStateStopped(reason))
}

func (ws *workflows) errored(ctx context.Context, sessionID sdktypes.SessionID, err error, prints []string) {
	ctx, cancel := withLimitedTimeout(ctx)
	defer cancel()

	_ = ws.updateSessionState(ctx, sessionID, sdktypes.NewSessionStateError(err, prints))
}

func (ws *workflows) deactivateDrainedDeployment(ctx context.Context, deploymentID sdktypes.DeploymentID) error {
	deactivate := false

	if err := ws.svcs.DB.Transaction(ctx, func(tx db.DB) error {
		dep, err := tx.GetDeployment(ctx, deploymentID)
		if err != nil {
			return fmt.Errorf("deployments.get: %w", err)
		}

		// TODO [ENG-1238]: use single query for this and move this to the db layer?
		if dep.State() == sdktypes.DeploymentStateDraining {
			resultRunning, err := tx.ListSessions(ctx, sdkservices.ListSessionsFilter{
				DeploymentID: deploymentID,
				StateType:    sdktypes.SessionStateTypeCreated,
				CountOnly:    true,
			})
			if err != nil {
				return fmt.Errorf("sessions.count: %w", err)
			}

			resultCreated, err := tx.ListSessions(ctx, sdkservices.ListSessionsFilter{
				DeploymentID: deploymentID,
				StateType:    sdktypes.SessionStateTypeRunning,
				CountOnly:    true,
			})
			if err != nil {
				return fmt.Errorf("sessions.count: %w", err)
			}

			deactivate = resultRunning.TotalCount+resultCreated.TotalCount == 0
		}
		return nil
	}); err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	if deactivate {
		if err := ws.svcs.Deployments.Deactivate(ctx, deploymentID); err != nil {
			return fmt.Errorf("deployment.deactivate(%v): %w", deploymentID, err)
		}
	}

	return nil
}

func (ws *workflows) StopWorkflow(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool) error {
	wid := workflowID(sessionID)
	ctx = akCtx.WithRequestOrginator(ctx, akCtx.SessionWorkflow)

	if force {
		// TODO(ENG-206): Is there a race condition here with update session?

		if err := ws.svcs.TemporalClient().TerminateWorkflow(ctx, wid, "", reason); err != nil {
			// TODO: translate error
			return fmt.Errorf("temporal: %w", err)
		}

		// TODO: we might want to create a workflow that forcibly terminates another workflow. That way
		//       we can avoid a dirty state if the terminator crashes between the temporal termination
		//       and the state update. Another way is to periodically check on all workflows and make sure
		//       that they are indeed running in termporal once in a while.

		return ws.updateSessionState(ctx, sessionID, sdktypes.NewSessionStateStopped(reason))
	}

	// In case of non-forceful termination, we log the request politely. This will also
	// let the workflow know what the reason is.
	if err := ws.svcs.DB.AddSessionStopRequest(ctx, sessionID, reason); err != nil {
		return err
	}

	if err := ws.svcs.TemporalClient().CancelWorkflow(ctx, wid, ""); err != nil {
		// TODO: translate errors.
		return err
	}

	return nil
}

func (ws *workflows) updateSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error {
	if err := ws.svcs.DB.UpdateSessionState(ctx, sessionID, state); err != nil {
		ws.z.With(zap.String("session_id", sessionID.String())).Error("update session", zap.Error(err))
		return err
	}

	return nil
}
