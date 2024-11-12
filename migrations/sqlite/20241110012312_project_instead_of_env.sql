-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_deployments" table
CREATE TABLE `new_deployments` (
  `deployment_id` uuid NOT NULL,
  `env_id` uuid NULL,
  `project_id` uuid NULL,
  `build_id` uuid NOT NULL,
  `state` integer NULL,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`deployment_id`),
  CONSTRAINT `fk_deployments_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_deployments_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_deployments_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "deployments" to new temporary table "new_deployments"
INSERT INTO `new_deployments` (`deployment_id`, `env_id`, `build_id`, `state`, `created_at`, `updated_at`, `deleted_at`) SELECT `deployment_id`, `env_id`, `build_id`, `state`, `created_at`, `updated_at`, `deleted_at` FROM `deployments`;
-- drop "deployments" table after copying rows
DROP TABLE `deployments`;
-- rename temporary table "new_deployments" to "deployments"
ALTER TABLE `new_deployments` RENAME TO `deployments`;
-- create index "idx_deployments_deleted_at" to table: "deployments"
CREATE INDEX `idx_deployments_deleted_at` ON `deployments` (`deleted_at`);
-- create index "idx_deployments_project_id" to table: "deployments"
CREATE INDEX `idx_deployments_project_id` ON `deployments` (`project_id`);
-- create index "idx_deployments_env_id" to table: "deployments"
CREATE INDEX `idx_deployments_env_id` ON `deployments` (`env_id`);
-- add column "project_id" to table: "sessions"
ALTER TABLE `sessions` ADD COLUMN `project_id` uuid NULL;
-- create index "idx_sessions_project_id" to table: "sessions"
CREATE INDEX `idx_sessions_project_id` ON `sessions` (`project_id`);
-- create "new_triggers" table
CREATE TABLE `new_triggers` (
  `trigger_id` uuid NOT NULL,
  `project_id` uuid NOT NULL,
  `connection_id` uuid NULL,
  `env_id` uuid NULL,
  `source_type` text NULL,
  `event_type` text NULL,
  `filter` text NULL,
  `code_location` text NULL,
  `name` text NULL,
  `unique_name` text NOT NULL,
  `webhook_slug` text NULL,
  `schedule` text NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`trigger_id`),
  CONSTRAINT `fk_triggers_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "triggers" to new temporary table "new_triggers"
INSERT INTO `new_triggers` (`trigger_id`, `project_id`, `connection_id`, `env_id`, `source_type`, `event_type`, `filter`, `code_location`, `name`, `unique_name`, `webhook_slug`, `schedule`, `deleted_at`) SELECT `trigger_id`, `project_id`, `connection_id`, `env_id`, `source_type`, `event_type`, `filter`, `code_location`, `name`, `unique_name`, `webhook_slug`, `schedule`, `deleted_at` FROM `triggers`;
-- drop "triggers" table after copying rows
DROP TABLE `triggers`;
-- rename temporary table "new_triggers" to "triggers"
ALTER TABLE `new_triggers` RENAME TO `triggers`;
-- create index "idx_triggers_source_type" to table: "triggers"
CREATE INDEX `idx_triggers_source_type` ON `triggers` (`source_type`);
-- create index "idx_triggers_env_id" to table: "triggers"
CREATE INDEX `idx_triggers_env_id` ON `triggers` (`env_id`);
-- create index "idx_triggers_connection_id" to table: "triggers"
CREATE INDEX `idx_triggers_connection_id` ON `triggers` (`connection_id`);
-- create index "idx_triggers_project_id" to table: "triggers"
CREATE INDEX `idx_triggers_project_id` ON `triggers` (`project_id`);
-- create index "idx_triggers_deleted_at" to table: "triggers"
CREATE INDEX `idx_triggers_deleted_at` ON `triggers` (`deleted_at`);
-- create index "idx_triggers_webhook_slug" to table: "triggers"
CREATE INDEX `idx_triggers_webhook_slug` ON `triggers` (`webhook_slug`);
-- create index "idx_triggers_unique_name" to table: "triggers"
CREATE UNIQUE INDEX `idx_triggers_unique_name` ON `triggers` (`unique_name`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_triggers_unique_name" to table: "triggers"
DROP INDEX `idx_triggers_unique_name`;
-- reverse: create index "idx_triggers_webhook_slug" to table: "triggers"
DROP INDEX `idx_triggers_webhook_slug`;
-- reverse: create index "idx_triggers_deleted_at" to table: "triggers"
DROP INDEX `idx_triggers_deleted_at`;
-- reverse: create index "idx_triggers_project_id" to table: "triggers"
DROP INDEX `idx_triggers_project_id`;
-- reverse: create index "idx_triggers_connection_id" to table: "triggers"
DROP INDEX `idx_triggers_connection_id`;
-- reverse: create index "idx_triggers_env_id" to table: "triggers"
DROP INDEX `idx_triggers_env_id`;
-- reverse: create index "idx_triggers_source_type" to table: "triggers"
DROP INDEX `idx_triggers_source_type`;
-- reverse: create "new_triggers" table
DROP TABLE `new_triggers`;
-- reverse: create index "idx_sessions_project_id" to table: "sessions"
DROP INDEX `idx_sessions_project_id`;
-- reverse: add column "project_id" to table: "sessions"
ALTER TABLE `sessions` DROP COLUMN `project_id`;
-- reverse: create index "idx_deployments_env_id" to table: "deployments"
DROP INDEX `idx_deployments_env_id`;
-- reverse: create index "idx_deployments_project_id" to table: "deployments"
DROP INDEX `idx_deployments_project_id`;
-- reverse: create index "idx_deployments_deleted_at" to table: "deployments"
DROP INDEX `idx_deployments_deleted_at`;
-- reverse: create "new_deployments" table
DROP TABLE `new_deployments`;
