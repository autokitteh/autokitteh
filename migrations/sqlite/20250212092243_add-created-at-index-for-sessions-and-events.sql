-- +goose Up
-- create index "idx_events_created_at" to table: "events"
CREATE INDEX `idx_events_created_at` ON `events` (`created_at`);
-- create index "idx_sessions_created_at" to table: "sessions"
CREATE INDEX `idx_sessions_created_at` ON `sessions` (`created_at`);

-- +goose Down
-- reverse: create index "idx_sessions_created_at" to table: "sessions"
DROP INDEX `idx_sessions_created_at`;
-- reverse: create index "idx_events_created_at" to table: "events"
DROP INDEX `idx_events_created_at`;
