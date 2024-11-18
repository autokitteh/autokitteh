-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- drop "envs" table
DROP TABLE `envs`;
-- create "new_sessions" table
CREATE TABLE `new_sessions` (
  `session_id` uuid NOT NULL,
  `build_id` uuid NULL,
  `project_id` uuid NULL,
  `deployment_id` uuid NULL,
  `event_id` uuid NULL,
  `current_state_type` integer NULL,
  `entrypoint` text NULL,
  `inputs` json NULL,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `memo` json NULL,
  PRIMARY KEY (`session_id`),
  CONSTRAINT `fk_sessions_event` FOREIGN KEY (`event_id`) REFERENCES `events` (`event_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_deployment` FOREIGN KEY (`deployment_id`) REFERENCES `deployments` (`deployment_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "sessions" to new temporary table "new_sessions"
INSERT INTO `new_sessions` (`session_id`, `build_id`, `project_id`, `deployment_id`, `event_id`, `current_state_type`, `entrypoint`, `inputs`, `created_at`, `updated_at`, `deleted_at`, `memo`) SELECT `session_id`, `build_id`, `project_id`, `deployment_id`, `event_id`, `current_state_type`, `entrypoint`, `inputs`, `created_at`, `updated_at`, `deleted_at`, `memo` FROM `sessions`;
-- drop "sessions" table after copying rows
DROP TABLE `sessions`;
-- rename temporary table "new_sessions" to "sessions"
ALTER TABLE `new_sessions` RENAME TO `sessions`;
-- create index "idx_sessions_event_id" to table: "sessions"
CREATE INDEX `idx_sessions_event_id` ON `sessions` (`event_id`);
-- create index "idx_sessions_deployment_id" to table: "sessions"
CREATE INDEX `idx_sessions_deployment_id` ON `sessions` (`deployment_id`);
-- create index "idx_sessions_project_id" to table: "sessions"
CREATE INDEX `idx_sessions_project_id` ON `sessions` (`project_id`);
-- create index "idx_sessions_build_id" to table: "sessions"
CREATE INDEX `idx_sessions_build_id` ON `sessions` (`build_id`);
-- create index "idx_sessions_deleted_at" to table: "sessions"
CREATE INDEX `idx_sessions_deleted_at` ON `sessions` (`deleted_at`);
-- create index "idx_sessions_current_state_type" to table: "sessions"
CREATE INDEX `idx_sessions_current_state_type` ON `sessions` (`current_state_type`);
-- create "new_deployments" table
CREATE TABLE `new_deployments` (
  `deployment_id` uuid NOT NULL,
  `project_id` uuid NULL,
  `build_id` uuid NOT NULL,
  `state` integer NULL,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`deployment_id`),
  CONSTRAINT `fk_deployments_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_deployments_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "deployments" to new temporary table "new_deployments"
INSERT INTO `new_deployments` (`deployment_id`, `project_id`, `build_id`, `state`, `created_at`, `updated_at`, `deleted_at`) SELECT `deployment_id`, `project_id`, `build_id`, `state`, `created_at`, `updated_at`, `deleted_at` FROM `deployments`;
-- drop "deployments" table after copying rows
DROP TABLE `deployments`;
-- rename temporary table "new_deployments" to "deployments"
ALTER TABLE `new_deployments` RENAME TO `deployments`;
-- create index "idx_deployments_deleted_at" to table: "deployments"
CREATE INDEX `idx_deployments_deleted_at` ON `deployments` (`deleted_at`);
-- create index "idx_deployments_project_id" to table: "deployments"
CREATE INDEX `idx_deployments_project_id` ON `deployments` (`project_id`);
-- create "new_triggers" table
CREATE TABLE `new_triggers` (
  `trigger_id` uuid NOT NULL,
  `project_id` uuid NOT NULL,
  `connection_id` uuid NULL,
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
  CONSTRAINT `fk_triggers_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "triggers" to new temporary table "new_triggers"
INSERT INTO `new_triggers` (`trigger_id`, `project_id`, `connection_id`, `source_type`, `event_type`, `filter`, `code_location`, `name`, `unique_name`, `webhook_slug`, `schedule`, `deleted_at`) SELECT `trigger_id`, `project_id`, `connection_id`, `source_type`, `event_type`, `filter`, `code_location`, `name`, `unique_name`, `webhook_slug`, `schedule`, `deleted_at` FROM `triggers`;
-- drop "triggers" table after copying rows
DROP TABLE `triggers`;
-- rename temporary table "new_triggers" to "triggers"
ALTER TABLE `new_triggers` RENAME TO `triggers`;
-- create index "idx_triggers_webhook_slug" to table: "triggers"
CREATE INDEX `idx_triggers_webhook_slug` ON `triggers` (`webhook_slug`);
-- create index "idx_triggers_unique_name" to table: "triggers"
CREATE UNIQUE INDEX `idx_triggers_unique_name` ON `triggers` (`unique_name`);
-- create index "idx_triggers_source_type" to table: "triggers"
CREATE INDEX `idx_triggers_source_type` ON `triggers` (`source_type`);
-- create index "idx_triggers_connection_id" to table: "triggers"
CREATE INDEX `idx_triggers_connection_id` ON `triggers` (`connection_id`);
-- create index "idx_triggers_project_id" to table: "triggers"
CREATE INDEX `idx_triggers_project_id` ON `triggers` (`project_id`);
-- create index "idx_triggers_deleted_at" to table: "triggers"
CREATE INDEX `idx_triggers_deleted_at` ON `triggers` (`deleted_at`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_triggers_deleted_at" to table: "triggers"
DROP INDEX `idx_triggers_deleted_at`;
-- reverse: create index "idx_triggers_project_id" to table: "triggers"
DROP INDEX `idx_triggers_project_id`;
-- reverse: create index "idx_triggers_connection_id" to table: "triggers"
DROP INDEX `idx_triggers_connection_id`;
-- reverse: create index "idx_triggers_source_type" to table: "triggers"
DROP INDEX `idx_triggers_source_type`;
-- reverse: create index "idx_triggers_unique_name" to table: "triggers"
DROP INDEX `idx_triggers_unique_name`;
-- reverse: create index "idx_triggers_webhook_slug" to table: "triggers"
DROP INDEX `idx_triggers_webhook_slug`;
-- reverse: create "new_triggers" table
DROP TABLE `new_triggers`;
-- reverse: create index "idx_deployments_project_id" to table: "deployments"
DROP INDEX `idx_deployments_project_id`;
-- reverse: create index "idx_deployments_deleted_at" to table: "deployments"
DROP INDEX `idx_deployments_deleted_at`;
-- reverse: create "new_deployments" table
DROP TABLE `new_deployments`;
-- reverse: create index "idx_sessions_current_state_type" to table: "sessions"
DROP INDEX `idx_sessions_current_state_type`;
-- reverse: create index "idx_sessions_deleted_at" to table: "sessions"
DROP INDEX `idx_sessions_deleted_at`;
-- reverse: create index "idx_sessions_build_id" to table: "sessions"
DROP INDEX `idx_sessions_build_id`;
-- reverse: create index "idx_sessions_project_id" to table: "sessions"
DROP INDEX `idx_sessions_project_id`;
-- reverse: create index "idx_sessions_deployment_id" to table: "sessions"
DROP INDEX `idx_sessions_deployment_id`;
-- reverse: create index "idx_sessions_event_id" to table: "sessions"
DROP INDEX `idx_sessions_event_id`;
-- reverse: create "new_sessions" table
DROP TABLE `new_sessions`;
-- reverse: drop "envs" table
CREATE TABLE `envs` (
  `env_id` uuid NOT NULL,
  `project_id` uuid NOT NULL,
  `name` text NULL,
  `deleted_at` datetime NULL,
  `membership_id` text NULL,
  PRIMARY KEY (`env_id`),
  CONSTRAINT `fk_envs_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE UNIQUE INDEX `idx_envs_membership_id` ON `envs` (`membership_id`);
CREATE INDEX `idx_envs_deleted_at` ON `envs` (`deleted_at`);
CREATE INDEX `idx_envs_project_id` ON `envs` (`project_id`);
