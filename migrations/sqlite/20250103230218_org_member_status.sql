-- +goose Up
-- add column "status" to table: "org_members"
ALTER TABLE `org_members` ADD COLUMN `status` integer NULL;
-- create index "idx_org_members_status" to table: "org_members"
CREATE INDEX `idx_org_members_status` ON `org_members` (`status`);

-- +goose Down
-- reverse: create index "idx_org_members_status" to table: "org_members"
DROP INDEX `idx_org_members_status`;
-- reverse: add column "status" to table: "org_members"
ALTER TABLE `org_members` DROP COLUMN `status`;
