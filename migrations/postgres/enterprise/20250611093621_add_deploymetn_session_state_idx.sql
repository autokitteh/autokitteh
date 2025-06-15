-- +goose Up
-- create index "idx_active_sessions" to table: "sessions"
CREATE INDEX "idx_active_sessions" ON "sessions" ("deployment_id", "current_state_type") WHERE ((current_state_type = 1) OR (current_state_type = 2));

-- +goose Down
-- reverse: create index "idx_active_sessions" to table: "sessions"
DROP INDEX "idx_active_sessions";
