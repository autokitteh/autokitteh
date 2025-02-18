-- +goose Up
-- drop index "idx_sessions_created_at" from table: "sessions"
DROP INDEX `idx_sessions_created_at`;

-- +goose Down
-- reverse: drop index "idx_sessions_created_at" from table: "sessions"
CREATE INDEX `idx_sessions_created_at` ON `sessions` (`created_at`);
