-- +goose Up
-- create "deployment_session_stats" table
CREATE TABLE "deployment_session_stats" (
  "deployment_id" uuid NOT NULL,
  "session_state" bigint NOT NULL,
  "count" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("deployment_id", "session_state")
);

-- +goose Down
-- reverse: create "deployment_session_stats" table
DROP TABLE "deployment_session_stats";
