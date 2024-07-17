package sessionworkflows

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessiondata"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

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

func New(z *zap.Logger, cfg Config, sessions sdkservices.Sessions, svcs *sessionsvcs.Svcs, calls sessioncalls.Calls) Workflows {
	return &workflows{z: z, cfg: cfg, sessions: sessions, calls: calls, svcs: svcs}
}

func (ws *workflows) StartWorkers(ctx context.Context) error {
	opts := ws.cfg.Temporal.Worker
	opts.DisableRegistrationAliasing = true
	opts.OnFatalError = func(err error) { ws.z.Error("temporal worker error", zap.Error(err)) }
	opts.DeadlockDetectionTimeout = time.Second * 10

	ws.worker = worker.New(ws.svcs.TemporalClient(), taskQueueName, opts)

	ws.worker.RegisterWorkflowWithOptions(
		ws.sessionWorkflow,
		workflow.RegisterOptions{Name: sessionWorkflowName},
	)

	return ws.worker.Start()
}

func (ws *workflows) StartWorkflow(ctx context.Context, session sdktypes.Session, debug bool) error {
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

func (ws *workflows) getSessionData(ctx workflow.Context, sessionID sdktypes.SessionID) (*sessiondata.Data, error) {
	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	// This cannot run through activity as it would expose potentialy sensitive data to temporal.
	data, err := sessiondata.Get(goCtx, ws.z, ws.svcs, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session data: %w", err)
	}

	return data, nil
}

func (ws *workflows) cleanupSession(data *sessiondata.Data) {
	z := ws.z.With(zap.String("session_id", data.SessionID.String()))

	if depID := data.Session.DeploymentID(); depID.IsValid() {
		// We cannot rely on workflow context here as it might have been canceled.
		go func() {
			ctx, cancel := withLimitedTimeout(context.Background())
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

	history, err := ws.sessions.ListSessionLogRecords(ctx, sdkservices.ListSessionLogRecordsFilter{SessionID: data.SessionID})
	if err != nil {
		z.Warn("get history failed", zap.Error(err))
		return "error retreiving history"
	}

	return struct {
		Prints  []string
		History []sdktypes.SessionLogRecord
	}{
		Prints:  prints,
		History: history.Records,
	}
}

func (ws *workflows) sessionWorkflow(wctx workflow.Context, params *sessionWorkflowParams) (debug any, _ error) {
	z := ws.z.With(zap.String("session_id", params.SessionID.String()))

	wctx = workflow.WithLocalActivityOptions(wctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: ws.cfg.Temporal.LocalScheduleToCloseTimeout,
	})

	// TODO(ENG-322): Save data in snapshot, otherwise changes between retries would
	//                blow us up due to non determinism.

	var prints []string

	data, err := ws.getSessionData(wctx, params.SessionID)
	if err == nil {
		prints, err = runWorkflow(wctx, z, ws, data, params.Debug)
	}

	if err != nil {
		z := z.With(zap.Error(err))

		if errors.Is(err, workflow.ErrCanceled) || errors.Is(wctx.Err(), workflow.ErrCanceled) {
			z.Debug("workflow canceled")
			ws.stopped(params.SessionID)
		} else {
			ws.errored(params.SessionID, err, prints)

			if _, ok := sdktypes.FromError(err); ok {
				// User level error, no need to indicate the workflow as errored.
				err = nil

				z.Debug("program error")
			} else {
				z.Error("workflow error")
			}
		}
	}

	if data != nil {
		ws.cleanupSession(data)

		if params.Debug {
			debug = ws.getSessionDebugData(data, prints)
		}
	}

	return debug, err
}

func (ws *workflows) updateFinalSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error {
	return ws.svcs.DB.Transaction(ctx, func(tx db.DB) error {
		if err := tx.UpdateSessionState(ctx, sessionID, state); err != nil {
			return err
		}
		//TODO: Decide how to handle latest record
		// this assume this is for sure the latest record
		record := sdktypes.NewStateSessionLogRecord(math.MaxInt32, state)
		return tx.SaveSessionLogRecord(ctx, sessionID, record)
	})
}

func (ws *workflows) stopped(sessionID sdktypes.SessionID) {
	ctx, cancel := withLimitedTimeout(context.Background())
	defer cancel()

	reason := "<unknown>"

	if result, err := ws.svcs.DB.ListSessionLogRecords(ctx, sdkservices.ListSessionLogRecordsFilter{SessionID: sessionID}); err == nil {
		for _, rec := range result.Records {
			if r, ok := rec.GetStopRequest(); ok {
				reason = r
				break
			}
		}
	}

	if err := ws.updateFinalSessionState(ctx, sessionID, sdktypes.NewSessionStateStopped(reason)); err != nil {
		ws.z.Error("update session", zap.Error(err), zap.String("session_id", sessionID.String()))
	}
}

func (ws *workflows) errored(sessionID sdktypes.SessionID, err error, prints []string) {
	ctx, cancel := withLimitedTimeout(context.Background())
	defer cancel()

	if err := ws.updateFinalSessionState(ctx, sessionID, sdktypes.NewSessionStateError(err, prints)); err != nil {
		ws.z.Error("update session", zap.Error(err), zap.String("session_id", sessionID.String()))
	}

}

func (ws *workflows) deactivateDrainedDeployment(ctx context.Context, deploymentID sdktypes.DeploymentID) error {
	deactivate := false

	if err := ws.svcs.DB.Transaction(ctx, func(tx db.DB) error {
		dep, err := tx.GetDeployment(ctx, deploymentID)
		if err != nil {
			return fmt.Errorf("deployments.get: %w", err)
		}

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

	if err := ws.svcs.DB.Transaction(ctx, func(tx db.DB) error {
		result, err := ws.svcs.DB.ListSessionLogRecords(ctx,
			sdkservices.ListSessionLogRecordsFilter{
				SessionID: sessionID,
				PaginationRequest: sdktypes.PaginationRequest{
					PageSize:  1,
					Ascending: false,
				}})

		if err != nil {
			return err
		}

		seq := int32(1)
		if len(result.Records) == 1 {
			seq = int32(result.Records[0].Seq())
		}

		r := sdktypes.NewStopRequestSessionLogRecord(seq, reason)
		return ws.svcs.DB.SaveSessionLogRecord(ctx, sessionID, r)
	}); err != nil {
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
