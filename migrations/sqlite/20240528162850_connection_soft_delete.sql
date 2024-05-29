-- +goose Up
-- add column "deleted_at" to table: "connections"
ALTER TABLE `connections` ADD COLUMN `deleted_at` datetime NULL;
-- create index "idx_connections_deleted_at" to table: "connections"
CREATE INDEX `idx_connections_deleted_at` ON `connections` (`deleted_at`);

-- +goose Down
-- reverse: create index "idx_connections_deleted_at" to table: "connections"
DROP INDEX `idx_connections_deleted_at`;
-- reverse: add column "deleted_at" to table: "connections"
ALTER TABLE `connections` DROP COLUMN `deleted_at`;
