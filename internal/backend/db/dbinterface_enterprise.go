//go:build enterprise
// +build enterprise

package db

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type WorkflowExecutionRequest struct {
	SessionID  sdktypes.SessionID
	WorkflowID string
	Args       any
	Memo       map[string]string
	RetryCount int
}

type DB interface {
	Shared

	// WorkflowExecutionRequest
	CreateWorkflowExecutionRequest(ctx context.Context, request WorkflowExecutionRequest) error
	GetWorkflowExecutionRequests(ctx context.Context, workerID string, maxRequests int) ([]WorkflowExecutionRequest, error)
	UpdateWorkflowExecutionRequestStatus(ctx context.Context, workflowID string, status string) (bool, error)
	CountInProgressWorkflowExecutionRequests(ctx context.Context, workerID string) (int64, error)
	GetInProgressWorkflowIDs(ctx context.Context, workerID string) ([]string, error)
}
