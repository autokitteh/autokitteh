-- +goose Up
-- add column "deleted_at" to table: "orgs"
ALTER TABLE `orgs` ADD COLUMN `deleted_at` datetime NULL;
-- create index "idx_orgs_deleted_at" to table: "orgs"
CREATE INDEX `idx_orgs_deleted_at` ON `orgs` (`deleted_at`);

-- +goose Down
-- reverse: create index "idx_orgs_deleted_at" to table: "orgs"
DROP INDEX `idx_orgs_deleted_at`;
-- reverse: add column "deleted_at" to table: "orgs"
ALTER TABLE `orgs` DROP COLUMN `deleted_at`;
