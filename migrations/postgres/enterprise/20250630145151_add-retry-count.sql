-- +goose Up
-- modify "workflow_execution_requests" table
ALTER TABLE "workflow_execution_requests" ADD COLUMN "retry_count" bigint NULL DEFAULT 0;

-- +goose Down
-- reverse: modify "workflow_execution_requests" table
ALTER TABLE "workflow_execution_requests" DROP COLUMN "retry_count";
