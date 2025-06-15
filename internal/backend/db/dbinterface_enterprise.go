//go:build enterprise
// +build enterprise

package db

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type WorkflowExecutionRequest struct {
	SessionID sdktypes.SessionID
	Args      any
	Memo      map[string]string
}

type DB interface {
	Shared
	// WorkerInfo
	GetWorkerInfo(ctx context.Context, id string) (scheme.WorkerInfo, error)
	IncActiveWorkflows(ctx context.Context, workerID string) (int, error)
	DecActiveWorkflows(ctx context.Context, workerID string) (int, error)

	// WorkflowExecutionRequest
	CreateWorkflowExecutionRequest(ctx context.Context, request WorkflowExecutionRequest) error
	GetWorkflowExecutionRequests(ctx context.Context, workerID string, maxRequests int) ([]WorkflowExecutionRequest, error)
	DeleteWorkflowExecutionRequest(ctx context.Context, sessionID sdktypes.SessionID) error
}
