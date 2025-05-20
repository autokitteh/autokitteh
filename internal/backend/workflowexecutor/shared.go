package workflowexecutor

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

func workflowID(sessionID sdktypes.SessionID) string { return sessionID.String() }

func (e *executor) WorkflowSessionName() string {
	return "session"
}

func (e *executor) execute(ctx context.Context, sessionID sdktypes.SessionID, args any, memo map[string]string) (string, error) {
	opts := e.cfg.SessionWorkflow.ToStartWorkflowOptions(e.WorkflowQueue(), workflowID(sessionID), fmt.Sprintf("session %v", sessionID), memo)
	r, err := e.svcs.Temporal.TemporalClient().ExecuteWorkflow(
		ctx,
		opts,
		e.WorkflowSessionName(),
		args,
	)
	if err != nil {
		return "", err
	}
	e.l.Info("Started workflow", zap.String("workflow_id", r.GetID()), zap.String("workflow_name", e.WorkflowSessionName()))

	return r.GetID(), nil
}
