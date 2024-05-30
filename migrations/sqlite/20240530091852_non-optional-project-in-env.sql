-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_envs" table
CREATE TABLE `new_envs` (
  `env_id` uuid NOT NULL,
  `project_id` uuid NOT NULL,
  `name` text NULL,
  `deleted_at` datetime NULL,
  `membership_id` text NULL,
  PRIMARY KEY (`env_id`),
  CONSTRAINT `fk_envs_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "envs" to new temporary table "new_envs"
INSERT INTO `new_envs` (`env_id`, `project_id`, `name`, `deleted_at`, `membership_id`) SELECT `env_id`, `project_id`, `name`, `deleted_at`, `membership_id` FROM `envs`;
-- drop "envs" table after copying rows
DROP TABLE `envs`;
-- rename temporary table "new_envs" to "envs"
ALTER TABLE `new_envs` RENAME TO `envs`;
-- create index "idx_envs_project_id" to table: "envs"
CREATE INDEX `idx_envs_project_id` ON `envs` (`project_id`);
-- create index "idx_envs_membership_id" to table: "envs"
CREATE UNIQUE INDEX `idx_envs_membership_id` ON `envs` (`membership_id`);
-- create index "idx_envs_deleted_at" to table: "envs"
CREATE INDEX `idx_envs_deleted_at` ON `envs` (`deleted_at`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_envs_deleted_at" to table: "envs"
DROP INDEX `idx_envs_deleted_at`;
-- reverse: create index "idx_envs_membership_id" to table: "envs"
DROP INDEX `idx_envs_membership_id`;
-- reverse: create index "idx_envs_project_id" to table: "envs"
DROP INDEX `idx_envs_project_id`;
-- reverse: create "new_envs" table
DROP TABLE `new_envs`;
