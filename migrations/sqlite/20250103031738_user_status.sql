-- +goose Up
-- add column "status" to table: "users"
ALTER TABLE `users` ADD COLUMN `status` integer NULL;
-- create index "idx_users_status" to table: "users"
CREATE INDEX `idx_users_status` ON `users` (`status`);

-- +goose Down
-- reverse: create index "idx_users_status" to table: "users"
DROP INDEX `idx_users_status`;
-- reverse: add column "status" to table: "users"
ALTER TABLE `users` DROP COLUMN `status`;
