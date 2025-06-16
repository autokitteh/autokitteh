-- +goose Up
-- drop index "idx_sessions_current_state_type" from table: "sessions"
DROP INDEX `idx_sessions_current_state_type`;

-- +goose Down
-- reverse: drop index "idx_sessions_current_state_type" from table: "sessions"
CREATE INDEX `idx_sessions_current_state_type` ON `sessions` (`current_state_type`);
