-- +goose Up
-- add column "org_id" to table: "events"
ALTER TABLE `events` ADD COLUMN `org_id` uuid NULL;
-- create index "idx_events_org_id" to table: "events"
CREATE INDEX `idx_events_org_id` ON `events` (`org_id`);

-- +goose Down
-- reverse: create index "idx_events_org_id" to table: "events"
DROP INDEX `idx_events_org_id`;
-- reverse: add column "org_id" to table: "events"
ALTER TABLE `events` DROP COLUMN `org_id`;
