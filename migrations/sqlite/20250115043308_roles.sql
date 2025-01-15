-- +goose Up
-- add column "roles" to table: "org_members"
ALTER TABLE `org_members` ADD COLUMN `roles` json NULL;
-- add column "updated_by" to table: "org_members"
ALTER TABLE `org_members` ADD COLUMN `updated_by` uuid NULL;
-- add column "updated_at" to table: "org_members"
ALTER TABLE `org_members` ADD COLUMN `updated_at` datetime NULL;

-- +goose Down
-- reverse: add column "updated_at" to table: "org_members"
ALTER TABLE `org_members` DROP COLUMN `updated_at`;
-- reverse: add column "updated_by" to table: "org_members"
ALTER TABLE `org_members` DROP COLUMN `updated_by`;
-- reverse: add column "roles" to table: "org_members"
ALTER TABLE `org_members` DROP COLUMN `roles`;
