-- +goose Up
-- add column "status_code" to table: "connections"
ALTER TABLE `connections` ADD COLUMN `status_code` integer NULL;
-- add column "status_message" to table: "connections"
ALTER TABLE `connections` ADD COLUMN `status_message` text NULL;
-- create index "idx_connections_status_code" to table: "connections"
CREATE INDEX `idx_connections_status_code` ON `connections` (`status_code`);
-- create index "idx_connections_integration_id" to table: "connections"
CREATE INDEX `idx_connections_integration_id` ON `connections` (`integration_id`);

-- +goose Down
-- reverse: create index "idx_connections_integration_id" to table: "connections"
DROP INDEX `idx_connections_integration_id`;
-- reverse: create index "idx_connections_status_code" to table: "connections"
DROP INDEX `idx_connections_status_code`;
-- reverse: add column "status_message" to table: "connections"
ALTER TABLE `connections` DROP COLUMN `status_message`;
-- reverse: add column "status_code" to table: "connections"
ALTER TABLE `connections` DROP COLUMN `status_code`;
