-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- drop "ownerships" table
DROP TABLE `ownerships`;
-- create "new_users" table
CREATE TABLE `new_users` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `user_id` uuid NOT NULL,
  `email` text NOT NULL,
  `display_name` text NULL,
  `status` integer NULL,
  `default_org_id` uuid NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`user_id`)
);
-- copy rows from old table "users" to new temporary table "new_users"
INSERT INTO `new_users` (`created_by`, `created_at`, `user_id`, `email`, `display_name`, `status`, `default_org_id`, `updated_by`, `updated_at`) SELECT `created_by`, `created_at`, `user_id`, `email`, `display_name`, `status`, `default_org_id`, `updated_by`, `updated_at` FROM `users`;
-- drop "users" table after copying rows
DROP TABLE `users`;
-- rename temporary table "new_users" to "users"
ALTER TABLE `new_users` RENAME TO `users`;
-- create index "idx_users_status" to table: "users"
CREATE INDEX `idx_users_status` ON `users` (`status`);
-- create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX `idx_users_email` ON `users` (`email`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_users_email" to table: "users"
DROP INDEX `idx_users_email`;
-- reverse: create index "idx_users_status" to table: "users"
DROP INDEX `idx_users_status`;
-- reverse: create "new_users" table
DROP TABLE `new_users`;
-- reverse: drop "ownerships" table
CREATE TABLE `ownerships` (
  `entity_id` uuid NOT NULL,
  `entity_type` text NOT NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`entity_id`)
);
