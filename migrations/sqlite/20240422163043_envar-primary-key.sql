-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_env_vars" table
CREATE TABLE `new_env_vars` (
  `env_id` uuid NULL,
  `name` text NULL,
  `value` text NULL,
  `secret_value` text NULL,
  `is_secret` numeric NULL,
  `membership_id` text NULL,
  PRIMARY KEY (`env_id`, `name`),
  CONSTRAINT `fk_env_vars_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "env_vars" to new temporary table "new_env_vars"
INSERT INTO `new_env_vars` (`env_id`, `name`, `value`, `secret_value`, `is_secret`, `membership_id`) SELECT `env_id`, `name`, `value`, `secret_value`, `is_secret`, `membership_id` FROM `env_vars`;
-- drop "env_vars" table after copying rows
DROP TABLE `env_vars`;
-- rename temporary table "new_env_vars" to "env_vars"
ALTER TABLE `new_env_vars` RENAME TO `env_vars`;
-- create index "idx_env_vars_membership_id" to table: "env_vars"
CREATE UNIQUE INDEX `idx_env_vars_membership_id` ON `env_vars` (`membership_id`);
-- create index "idx_env_vars_env_id" to table: "env_vars"
CREATE INDEX `idx_env_vars_env_id` ON `env_vars` (`env_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_env_vars_env_id" to table: "env_vars"
DROP INDEX `idx_env_vars_env_id`;
-- reverse: create index "idx_env_vars_membership_id" to table: "env_vars"
DROP INDEX `idx_env_vars_membership_id`;
-- reverse: create "new_env_vars" table
DROP TABLE `new_env_vars`;
