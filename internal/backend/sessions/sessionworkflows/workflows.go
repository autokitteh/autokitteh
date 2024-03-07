package sessionworkflows

import (
	"context"
	"fmt"

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
	opts := cfg.Temporal.Worker
	opts.DisableRegistrationAliasing = true
	opts.OnFatalError = func(err error) { z.Error("temporal worker error", zap.Error(err)) }

	worker := worker.New(svcs.Temporal, taskQueueName, opts)

	ws := workflows{z: z, cfg: cfg, worker: worker, sessions: sessions, calls: calls, svcs: svcs}

	worker.RegisterWorkflowWithOptions(
		ws.sessionWorkflow,
		workflow.RegisterOptions{Name: sessionWorkflowName},
	)

	return &ws
}

func (ws *workflows) StartWorkers(ctx context.Context) error { return ws.worker.Start() }

func (ws *workflows) StartWorkflow(ctx context.Context, session sdktypes.Session, debug bool) error {
	sessionID := session.ID()

	wid := workflowID(sessionID)

	z := ws.z.With(zap.String("session_id", sessionID.String()), zap.String("workflow_id", wid))

	// NOTE: If we have a crash here, after CreateSession and before ExecuteWorkflow
	//       we might have a zombie session id. It should be ok as it is rare
	//       and the user will get a 501.

	memo, _ := kittehs.JoinMaps(map[string]string{
		"session_id":    sessionID.Value(),
		"deployment_id": session.DeploymentID().String(),
		"build_id":      session.BuildID().String(),
		"entrypoint":    session.EntryPoint().CanonicalString(),
		"workflow_id":   wid,
	}, session.Memo())

	swopts := client.StartWorkflowOptions{
		ID:                  workflowID(sessionID),
		TaskQueue:           taskQueueName,
		WorkflowTaskTimeout: ws.cfg.Temporal.WorkflowTaskTimeout,
		Memo:                kittehs.TransformMapValues(memo, func(s string) any { return s }),
	}

	r, err := ws.svcs.Temporal.ExecuteWorkflow(
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

func (ws *workflows) cleanupSession(ctx workflow.Context, data *sessiondata.Data) {
	z := ws.z.With(zap.String("session_id", data.SessionID.String()))

	workflow.Go(ctx, func(ctx workflow.Context) {
		// TODO: can this be done async? do we care?
		if err := workflow.ExecuteLocalActivity(
			ctx,
			ws.deactivateDrainedDeployment,
			data.Deployment.ID(),
		).Get(ctx, nil); err != nil {
			z.Error("deactivate drained deployment failed", zap.Error(err))
		}
	})
}

func (ws *workflows) getSessionDebugData(ctx workflow.Context, data *sessiondata.Data) any {
	z := ws.z.With(zap.String("session_id", data.SessionID.String()))

	goCtx := temporalclient.NewWorkflowContextAsGOContext(ctx)

	history, err := ws.sessions.GetLog(goCtx, data.SessionID)
	if err != nil {
		z.Warn("get history failed", zap.Error(err))
		return "error retreiving history"
	}

	return history
}

func (ws *workflows) sessionWorkflow(ctx workflow.Context, params *sessionWorkflowParams) (any, error) {
	z := ws.z.With(zap.String("session_id", params.SessionID.String()))

	ctx = workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: ws.cfg.Temporal.LocalScheduleToCloseTimeout,
	})

	// TODO(ENG-322): Save data in snapshot, otherwise changes between retries would
	//                blow us up due to non determinism.
	data, err := ws.getSessionData(ctx, params.SessionID)
	if err != nil {
		z.Panic("get session data", zap.Error(err))
	}

	defer ws.cleanupSession(ctx, data)

	if err := runWorkflow(ctx, z, ws, data, params.Debug); err != nil {
		return nil, err
	}

	var debug any

	if params.Debug {
		debug = ws.getSessionDebugData(ctx, data)
	}

	return debug, err
}

func (ws *workflows) deactivateDrainedDeployment(ctx context.Context, deploymentID sdktypes.DeploymentID) error {
	deactivate := false

	if err := ws.svcs.DB.Transaction(ctx, func(tx db.DB) error {
		dep, err := tx.GetDeployment(ctx, deploymentID)
		if err != nil {
			return fmt.Errorf("deployments.get: %w", err)
		}

		if dep.State() == sdktypes.DeploymentStateDraining {
			_, nRunning, err := tx.ListSessions(ctx, sdkservices.ListSessionsFilter{
				DeploymentID: deploymentID,
				StateType:    sdktypes.SessionStateTypeCreated,
				CountOnly:    true,
			})
			if err != nil {
				return fmt.Errorf("sessions.count: %w", err)
			}

			_, nCreated, err := tx.ListSessions(ctx, sdkservices.ListSessionsFilter{
				DeploymentID: deploymentID,
				StateType:    sdktypes.SessionStateTypeRunning,
				CountOnly:    true,
			})
			if err != nil {
				return fmt.Errorf("sessions.count: %w", err)
			}

			deactivate = nRunning+nCreated == 0
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

		if err := ws.svcs.Temporal.TerminateWorkflow(ctx, wid, "", reason); err != nil {
			// TODO: translate error
			return err
		}

		// TODO: we might want to create a workflow that forcibly terminates another workflow. That way
		//       we can avoid a dirty state if the terminator crashes between the temporal termination
		//       and the state update. Another way is to periodically check on all workflows and make sure
		//       that they are indeed running in termporal once in a while.
		return ws.updateSessionState(ctx, sessionID, sdktypes.NewSessionStateError(fmt.Errorf("terminated: %s", reason), nil))
	}

	if err := ws.svcs.Temporal.CancelWorkflow(ctx, wid, ""); err != nil {
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
