-- +goose Up
-- create "workflow_execution_requests" table
CREATE TABLE "workflow_execution_requests" (
  "session_id" text NOT NULL,
  "args" bytea NULL,
  "memo" bytea NULL,
  "acquired_at" timestamptz NULL,
  "acquired_by" text NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("session_id")
);

-- +goose Down
-- reverse: create "workflow_execution_requests" table
DROP TABLE "workflow_execution_requests";
