-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_vars" table
CREATE TABLE `new_vars` (
  `var_id` uuid NOT NULL,
  `name` text NOT NULL,
  `value` text NULL,
  `is_secret` numeric NULL,
  `integration_id` uuid NULL,
  PRIMARY KEY (`var_id`, `name`)
);
-- copy rows from old table "vars" to new temporary table "new_vars"
INSERT INTO `new_vars` (`name`, `value`, `is_secret`, `integration_id`) SELECT `name`, `value`, `is_secret`, `integration_id` FROM `vars`;
-- drop "vars" table after copying rows
DROP TABLE `vars`;
-- rename temporary table "new_vars" to "vars"
ALTER TABLE `new_vars` RENAME TO `vars`;
-- create index "idx_vars_integration_id" to table: "vars"
CREATE INDEX `idx_vars_integration_id` ON `vars` (`integration_id`);
-- create index "idx_vars_name" to table: "vars"
CREATE INDEX `idx_vars_name` ON `vars` (`name`);
-- create index "idx_vars_var_id" to table: "vars"
CREATE INDEX `idx_vars_var_id` ON `vars` (`var_id`);
-- create "users" table
CREATE TABLE `users` (
  `user_id` uuid NOT NULL,
  `provider` text NOT NULL,
  `email` text NOT NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`user_id`)
);
-- create index "idx_provider_email_name_idx" to table: "users"
CREATE UNIQUE INDEX `idx_provider_email_name_idx` ON `users` (`email`, `provider`, `name`);
-- create "ownerships" table
CREATE TABLE `ownerships` (
  `entity_id` uuid NOT NULL,
  `entity_type` text NOT NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`entity_id`),
  CONSTRAINT `fk_ownerships_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "ownerships" table
DROP TABLE `ownerships`;
-- reverse: create index "idx_provider_email_name_idx" to table: "users"
DROP INDEX `idx_provider_email_name_idx`;
-- reverse: create "users" table
DROP TABLE `users`;
-- reverse: create index "idx_vars_var_id" to table: "vars"
DROP INDEX `idx_vars_var_id`;
-- reverse: create index "idx_vars_name" to table: "vars"
DROP INDEX `idx_vars_name`;
-- reverse: create index "idx_vars_integration_id" to table: "vars"
DROP INDEX `idx_vars_integration_id`;
-- reverse: create "new_vars" table
DROP TABLE `new_vars`;
