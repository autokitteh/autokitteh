-- +goose Up
-- add column "deduplication_key" to table: "events"
ALTER TABLE `events` ADD COLUMN `deduplication_key` text NULL;
-- create index "idx_events_deduplication_key" to table: "events"
CREATE UNIQUE INDEX `idx_events_deduplication_key` ON `events` (`deduplication_key`);

-- +goose Down
-- reverse: create index "idx_events_deduplication_key" to table: "events"
DROP INDEX `idx_events_deduplication_key`;
-- reverse: add column "deduplication_key" to table: "events"
ALTER TABLE `events` DROP COLUMN `deduplication_key`;
