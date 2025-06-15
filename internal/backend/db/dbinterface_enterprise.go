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
}

type DB interface {
	Shared

	// WorkflowExecutionRequest
	CreateWorkflowExecutionRequest(ctx context.Context, request WorkflowExecutionRequest) error
	GetWorkflowExecutionRequests(ctx context.Context, workerID string, maxRequests int) ([]WorkflowExecutionRequest, error)
	UpdateRequestStatus(ctx context.Context, workflowID string, status string) error
	CountInProgressWorkflowExecutionRequests(ctx context.Context, workerID string) (int64, error)
}
