//go:build enterprise
// +build enterprise

package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) CreateWorkflowExecutionRequest(ctx context.Context, r db.WorkflowExecutionRequest) error {
	argsBytes, err := json.Marshal(r.Args)
	if err != nil {
		return fmt.Errorf("failed to marshal args: %w", err)
	}

	memoData, err := json.Marshal(r.Memo)
	if err != nil {
		return fmt.Errorf("failed to marshal memo: %w", err)
	}

	request := &scheme.WorkflowExecutionRequest{
		WorkflowID: r.WorkflowID,
		SessionID:  r.SessionID.UUIDValue(),
		Args:       argsBytes,
		Memo:       memoData,
		Status:     "pending",
	}

	return gdb.writer.WithContext(ctx).Create(request).Error
}

func (gdb *gormdb) GetWorkflowExecutionRequests(ctx context.Context, workerID string, maxRequests int) ([]db.WorkflowExecutionRequest, error) {
	//TODO: Decide on lock limit of tasks
	var requests []scheme.WorkflowExecutionRequest
	if err := gdb.writer.WithContext(ctx).Model(&scheme.WorkflowExecutionRequest{}).Raw(`
	UPDATE workflow_execution_requests
	SET acquired_at = NOW(), acquired_by = $1
	WHERE session_id IN (SELECT session_id FROM workflow_execution_requests WHERE status = 'pending' OR acquired_at < NOW() - INTERVAL '1 day' LIMIT $2 FOR UPDATE SKIP LOCKED)
	RETURNING *;
	`, workerID, maxRequests).Scan(&requests).Error; err != nil {
		return nil, err
	}

	results := make([]db.WorkflowExecutionRequest, len(requests))
	for i, request := range requests {
		var args any
		if err := json.Unmarshal(request.Args, &args); err != nil {
			return nil, fmt.Errorf("failed to unmarshal args: %w", err)
		}
		var memo map[string]string
		if err := json.Unmarshal(request.Memo, &memo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal memo: %w", err)
		}
		results[i] = db.WorkflowExecutionRequest{
			WorkflowID: request.WorkflowID,
			SessionID:  sdktypes.NewIDFromUUID[sdktypes.SessionID](request.SessionID),
			Args:       args,
			Memo:       memo,
		}
	}

	return results, nil
}

func (gdb *gormdb) CountInProgressWorkflowExecutionRequests(ctx context.Context, workerID string) (int64, error) {
	var count int64
	if err := gdb.writer.WithContext(ctx).Model(&scheme.WorkflowExecutionRequest{}).
		Where("acquired_by = ? AND status = 'in_progress'", workerID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (gdb *gormdb) UpdateRequestStatus(ctx context.Context, workflowID string, status string) error {
	return gdb.writer.WithContext(ctx).Model(&scheme.WorkflowExecutionRequest{}).
		Where("workflow_id = ?", workflowID).
		Update("status", status).Error
}
