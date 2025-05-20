-- +goose Up
-- create "workflow_execution_requests" table
CREATE TABLE "workflow_execution_requests" (
  "session_id" text NOT NULL,
  "aqcuired_at" timestamptz NULL,
  "aqcuired_by" text NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("session_id")
);

-- +goose Down
-- reverse: create "workflow_execution_requests" table
DROP TABLE "workflow_execution_requests";
