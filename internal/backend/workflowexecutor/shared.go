package workflowexecutor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func workflowID(sessionID sdktypes.SessionID) string { return sessionID.String() }

func (e *executor) WorkflowSessionName() string {
	return "session"
}

func (e *executor) execute(ctx context.Context, sessionID sdktypes.SessionID, workflowID string, args any, memo map[string]string) error {
	opts := e.cfg.SessionWorkflow.ToStartWorkflowOptions(e.WorkflowQueue(), workflowID, fmt.Sprintf("session %v", sessionID), memo)
	r, err := e.svcs.Temporal.TemporalClient().ExecuteWorkflow(
		ctx,
		opts,
		e.WorkflowSessionName(),
		args,
	)
	if err != nil {
		return err
	}
	e.l.Info("Started workflow", zap.String("workflow_id", r.GetID()), zap.String("workflow_name", e.WorkflowSessionName()), zap.String("session_id", sessionID.String()))

	return nil
}
