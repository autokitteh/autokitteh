-- +goose Up
-- create "workflow_execution_requests" table
CREATE TABLE "workflow_execution_requests" (
  "session_id" text NULL,
  "workflow_id" text NOT NULL,
  "args" bytea NULL,
  "memo" bytea NULL,
  "acquired_at" timestamptz NULL,
  "acquired_by" text NULL,
  "status" text NULL DEFAULT 'pending',
  "created_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("workflow_id")
);
-- create index "idx_acquired_by_status" to table: "workflow_execution_requests"
CREATE INDEX "idx_acquired_by_status" ON "workflow_execution_requests" ("acquired_by", "status") WHERE (status = 'in_progress'::text);

-- +goose Down
-- reverse: create index "idx_acquired_by_status" to table: "workflow_execution_requests"
DROP INDEX "idx_acquired_by_status";
-- reverse: create "workflow_execution_requests" table
DROP TABLE "workflow_execution_requests";
