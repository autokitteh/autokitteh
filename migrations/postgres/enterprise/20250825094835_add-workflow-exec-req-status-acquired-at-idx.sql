-- +goose Up
-- create index "idx_workflow_execution_request_status_acquired_at" to table: "workflow_execution_requests"
CREATE INDEX "idx_workflow_execution_request_status_acquired_at" ON "workflow_execution_requests" ("status", "acquired_at") WHERE ((status = 'pending'::text) OR (status = 'in_progress'::text));

-- +goose Down
-- reverse: create index "idx_workflow_execution_request_status_acquired_at" to table: "workflow_execution_requests"
DROP INDEX "idx_workflow_execution_request_status_acquired_at";
