-- +goose Up
-- add column "name" to table: "orgs"
ALTER TABLE `orgs` ADD COLUMN `name` text NULL;
-- create index "idx_orgs_name" to table: "orgs"
CREATE UNIQUE INDEX `idx_orgs_name` ON `orgs` (`name`);

-- +goose Down
-- reverse: create index "idx_orgs_name" to table: "orgs"
DROP INDEX `idx_orgs_name`;
-- reverse: add column "name" to table: "orgs"
ALTER TABLE `orgs` DROP COLUMN `name`;
