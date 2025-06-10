-- +goose Up
-- drop index "idx_sessions_current_state_type" from table: "sessions"
DROP INDEX "idx_sessions_current_state_type";
-- drop index "idx_sessions_deployment_id" from table: "sessions"
DROP INDEX "idx_sessions_deployment_id";
-- create index "idx_active_sessions" to table: "sessions"
CREATE INDEX "idx_active_sessions" ON "sessions" ("deployment_id", "current_state_type") WHERE ((current_state_type = 1) OR (current_state_type = 2));

-- +goose Down
-- reverse: create index "idx_active_sessions" to table: "sessions"
DROP INDEX "idx_active_sessions";
-- reverse: drop index "idx_sessions_deployment_id" from table: "sessions"
CREATE INDEX "idx_sessions_deployment_id" ON "sessions" ("deployment_id");
-- reverse: drop index "idx_sessions_current_state_type" from table: "sessions"
CREATE INDEX "idx_sessions_current_state_type" ON "sessions" ("current_state_type");
