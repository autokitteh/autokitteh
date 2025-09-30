-- +goose Up
-- add column "org_id" to table: "connections"
ALTER TABLE `connections` ADD COLUMN `org_id` uuid NULL;
-- create index "idx_connections_org_id" to table: "connections"
CREATE INDEX `idx_connections_org_id` ON `connections` (`org_id`);

-- +goose Down
-- reverse: create index "idx_connections_org_id" to table: "connections"
DROP INDEX `idx_connections_org_id`;
-- reverse: add column "org_id" to table: "connections"
ALTER TABLE `connections` DROP COLUMN `org_id`;
