-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- add column "owner_user_id" to table: "builds"
ALTER TABLE `builds` ADD COLUMN `owner_user_id` uuid NOT NULL;
-- create index "idx_builds_owner_user_id" to table: "builds"
CREATE INDEX `idx_builds_owner_user_id` ON `builds` (`owner_user_id`);
-- create "new_connections" table
CREATE TABLE `new_connections` (
  `project_id` uuid NOT NULL,
  `connection_id` uuid NOT NULL,
  `integration_id` uuid NULL,
  `name` text NULL,
  `status_code` integer NULL,
  `status_message` text NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`connection_id`),
  CONSTRAINT `fk_connections_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "connections" to new temporary table "new_connections"
INSERT INTO `new_connections` (`project_id`, `connection_id`, `integration_id`, `name`, `status_code`, `status_message`, `deleted_at`) SELECT `project_id`, `connection_id`, `integration_id`, `name`, `status_code`, `status_message`, `deleted_at` FROM `connections`;
-- drop "connections" table after copying rows
DROP TABLE `connections`;
-- rename temporary table "new_connections" to "connections"
ALTER TABLE `new_connections` RENAME TO `connections`;
-- create index "idx_connections_deleted_at" to table: "connections"
CREATE INDEX `idx_connections_deleted_at` ON `connections` (`deleted_at`);
-- create index "idx_connections_status_code" to table: "connections"
CREATE INDEX `idx_connections_status_code` ON `connections` (`status_code`);
-- create index "idx_connections_integration_id" to table: "connections"
CREATE INDEX `idx_connections_integration_id` ON `connections` (`integration_id`);
-- create index "idx_connections_project_id" to table: "connections"
CREATE INDEX `idx_connections_project_id` ON `connections` (`project_id`);
-- add column "owner_user_id" to table: "projects"
ALTER TABLE `projects` ADD COLUMN `owner_user_id` uuid NOT NULL;
-- create index "idx_projects_owner_user_id" to table: "projects"
CREATE INDEX `idx_projects_owner_user_id` ON `projects` (`owner_user_id`);
-- create "new_vars" table
CREATE TABLE `new_vars` (
  `project_id` uuid NOT NULL,
  `var_id` uuid NOT NULL,
  `name` text NOT NULL,
  `value` text NULL,
  `is_secret` numeric NULL,
  `integration_id` uuid NULL,
  PRIMARY KEY (`var_id`, `name`),
  CONSTRAINT `fk_vars_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "vars" to new temporary table "new_vars"
INSERT INTO `new_vars` (`var_id`, `name`, `value`, `is_secret`, `integration_id`) SELECT `var_id`, `name`, `value`, `is_secret`, `integration_id` FROM `vars`;
-- drop "vars" table after copying rows
DROP TABLE `vars`;
-- rename temporary table "new_vars" to "vars"
ALTER TABLE `new_vars` RENAME TO `vars`;
-- create index "idx_vars_var_id" to table: "vars"
CREATE INDEX `idx_vars_var_id` ON `vars` (`var_id`);
-- create index "idx_vars_project_id" to table: "vars"
CREATE INDEX `idx_vars_project_id` ON `vars` (`project_id`);
-- create index "idx_vars_integration_id" to table: "vars"
CREATE INDEX `idx_vars_integration_id` ON `vars` (`integration_id`);
-- create index "idx_vars_name" to table: "vars"
CREATE INDEX `idx_vars_name` ON `vars` (`name`);
-- create "new_events" table
CREATE TABLE `new_events` (
  `project_id` uuid NOT NULL,
  `event_id` uuid NOT NULL,
  `destination_id` uuid NOT NULL,
  `integration_id` uuid NULL,
  `connection_id` uuid NULL,
  `trigger_id` uuid NULL,
  `event_type` text NULL,
  `data` json NULL,
  `memo` json NULL,
  `created_at` datetime NULL,
  `seq` integer NULL PRIMARY KEY AUTOINCREMENT,
  `deleted_at` datetime NULL,
  CONSTRAINT `fk_events_trigger` FOREIGN KEY (`trigger_id`) REFERENCES `triggers` (`trigger_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_events_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_events_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "events" to new temporary table "new_events"
INSERT INTO `new_events` (`event_id`, `destination_id`, `integration_id`, `connection_id`, `trigger_id`, `event_type`, `data`, `memo`, `created_at`, `seq`, `deleted_at`) SELECT `event_id`, `destination_id`, `integration_id`, `connection_id`, `trigger_id`, `event_type`, `data`, `memo`, `created_at`, `seq`, `deleted_at` FROM `events`;
-- drop "events" table after copying rows
DROP TABLE `events`;
-- rename temporary table "new_events" to "events"
ALTER TABLE `new_events` RENAME TO `events`;
-- create index "idx_event_type_seq" to table: "events"
CREATE INDEX `idx_event_type_seq` ON `events` (`event_type`);
-- create index "idx_events_integration_id" to table: "events"
CREATE INDEX `idx_events_integration_id` ON `events` (`integration_id`);
-- create index "idx_events_trigger_id" to table: "events"
CREATE INDEX `idx_events_trigger_id` ON `events` (`trigger_id`);
-- create index "idx_events_connection_id" to table: "events"
CREATE INDEX `idx_events_connection_id` ON `events` (`connection_id`);
-- create index "idx_events_destination_id" to table: "events"
CREATE INDEX `idx_events_destination_id` ON `events` (`destination_id`);
-- create index "idx_events_event_id" to table: "events"
CREATE UNIQUE INDEX `idx_events_event_id` ON `events` (`event_id`);
-- create index "idx_events_project_id" to table: "events"
CREATE INDEX `idx_events_project_id` ON `events` (`project_id`);
-- create index "idx_events_deleted_at" to table: "events"
CREATE INDEX `idx_events_deleted_at` ON `events` (`deleted_at`);
-- create index "idx_event_type" to table: "events"
CREATE INDEX `idx_event_type` ON `events` (`event_type`);
-- create "new_values" table
CREATE TABLE `new_values` (
  `project_id` uuid NOT NULL,
  `key` text NOT NULL,
  `value` blob NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`key`),
  CONSTRAINT `fk_values_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "values" to new temporary table "new_values"
INSERT INTO `new_values` (`project_id`, `key`, `value`, `updated_at`) SELECT `project_id`, `key`, `value`, `updated_at` FROM `values`;
-- drop "values" table after copying rows
DROP TABLE `values`;
-- rename temporary table "new_values" to "values"
ALTER TABLE `new_values` RENAME TO `values`;
-- create index "idx_values_project_id" to table: "values"
CREATE INDEX `idx_values_project_id` ON `values` (`project_id`);
-- add column "disabled" to table: "users"
ALTER TABLE `users` ADD COLUMN `disabled` numeric NULL;
-- add column "owner_user_id" to table: "sessions"
ALTER TABLE `sessions` ADD COLUMN `owner_user_id` uuid NOT NULL;
-- create index "idx_sessions_owner_user_id" to table: "sessions"
CREATE INDEX `idx_sessions_owner_user_id` ON `sessions` (`owner_user_id`);
-- create "new_deployments" table
CREATE TABLE `new_deployments` (
  `project_id` uuid NOT NULL,
  `deployment_id` uuid NOT NULL,
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
INSERT INTO `new_deployments` (`project_id`, `deployment_id`, `build_id`, `state`, `created_at`, `updated_at`, `deleted_at`) SELECT `project_id`, `deployment_id`, `build_id`, `state`, `created_at`, `updated_at`, `deleted_at` FROM `deployments`;
-- drop "deployments" table after copying rows
DROP TABLE `deployments`;
-- rename temporary table "new_deployments" to "deployments"
ALTER TABLE `new_deployments` RENAME TO `deployments`;
-- create index "idx_deployments_deleted_at" to table: "deployments"
CREATE INDEX `idx_deployments_deleted_at` ON `deployments` (`deleted_at`);
-- create index "idx_deployments_project_id" to table: "deployments"
CREATE INDEX `idx_deployments_project_id` ON `deployments` (`project_id`);
-- add column "env_id" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `env_id` uuid NULL;
-- create index "idx_triggers_env_id" to table: "triggers"
CREATE INDEX `idx_triggers_env_id` ON `triggers` (`env_id`);
-- create "belongs_to_projects" table
CREATE TABLE `belongs_to_projects` (
  `project_id` uuid NOT NULL,
  CONSTRAINT `fk_belongs_to_projects_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
CREATE INDEX `idx_belongs_to_projects_project_id` ON `belongs_to_projects` (`project_id`);
-- create "owneds" table
CREATE TABLE `owneds` (
  `owner_user_id` uuid NOT NULL
);
-- create index "idx_owneds_owner_user_id" to table: "owneds"
CREATE INDEX `idx_owneds_owner_user_id` ON `owneds` (`owner_user_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_owneds_owner_user_id" to table: "owneds"
DROP INDEX `idx_owneds_owner_user_id`;
-- reverse: create "owneds" table
DROP TABLE `owneds`;
-- reverse: create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
DROP INDEX `idx_belongs_to_projects_project_id`;
-- reverse: create "belongs_to_projects" table
DROP TABLE `belongs_to_projects`;
-- reverse: create index "idx_triggers_env_id" to table: "triggers"
DROP INDEX `idx_triggers_env_id`;
-- reverse: add column "env_id" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `env_id`;
-- reverse: create index "idx_deployments_project_id" to table: "deployments"
DROP INDEX `idx_deployments_project_id`;
-- reverse: create index "idx_deployments_deleted_at" to table: "deployments"
DROP INDEX `idx_deployments_deleted_at`;
-- reverse: create "new_deployments" table
DROP TABLE `new_deployments`;
-- reverse: create index "idx_sessions_owner_user_id" to table: "sessions"
DROP INDEX `idx_sessions_owner_user_id`;
-- reverse: add column "owner_user_id" to table: "sessions"
ALTER TABLE `sessions` DROP COLUMN `owner_user_id`;
-- reverse: add column "disabled" to table: "users"
ALTER TABLE `users` DROP COLUMN `disabled`;
-- reverse: create index "idx_values_project_id" to table: "values"
DROP INDEX `idx_values_project_id`;
-- reverse: create "new_values" table
DROP TABLE `new_values`;
-- reverse: create index "idx_event_type" to table: "events"
DROP INDEX `idx_event_type`;
-- reverse: create index "idx_events_deleted_at" to table: "events"
DROP INDEX `idx_events_deleted_at`;
-- reverse: create index "idx_events_project_id" to table: "events"
DROP INDEX `idx_events_project_id`;
-- reverse: create index "idx_events_event_id" to table: "events"
DROP INDEX `idx_events_event_id`;
-- reverse: create index "idx_events_destination_id" to table: "events"
DROP INDEX `idx_events_destination_id`;
-- reverse: create index "idx_events_connection_id" to table: "events"
DROP INDEX `idx_events_connection_id`;
-- reverse: create index "idx_events_trigger_id" to table: "events"
DROP INDEX `idx_events_trigger_id`;
-- reverse: create index "idx_events_integration_id" to table: "events"
DROP INDEX `idx_events_integration_id`;
-- reverse: create index "idx_event_type_seq" to table: "events"
DROP INDEX `idx_event_type_seq`;
-- reverse: create "new_events" table
DROP TABLE `new_events`;
-- reverse: create index "idx_vars_name" to table: "vars"
DROP INDEX `idx_vars_name`;
-- reverse: create index "idx_vars_integration_id" to table: "vars"
DROP INDEX `idx_vars_integration_id`;
-- reverse: create index "idx_vars_project_id" to table: "vars"
DROP INDEX `idx_vars_project_id`;
-- reverse: create index "idx_vars_var_id" to table: "vars"
DROP INDEX `idx_vars_var_id`;
-- reverse: create "new_vars" table
DROP TABLE `new_vars`;
-- reverse: create index "idx_projects_owner_user_id" to table: "projects"
DROP INDEX `idx_projects_owner_user_id`;
-- reverse: add column "owner_user_id" to table: "projects"
ALTER TABLE `projects` DROP COLUMN `owner_user_id`;
-- reverse: create index "idx_connections_project_id" to table: "connections"
DROP INDEX `idx_connections_project_id`;
-- reverse: create index "idx_connections_integration_id" to table: "connections"
DROP INDEX `idx_connections_integration_id`;
-- reverse: create index "idx_connections_status_code" to table: "connections"
DROP INDEX `idx_connections_status_code`;
-- reverse: create index "idx_connections_deleted_at" to table: "connections"
DROP INDEX `idx_connections_deleted_at`;
-- reverse: create "new_connections" table
DROP TABLE `new_connections`;
-- reverse: create index "idx_builds_owner_user_id" to table: "builds"
DROP INDEX `idx_builds_owner_user_id`;
-- reverse: add column "owner_user_id" to table: "builds"
ALTER TABLE `builds` DROP COLUMN `owner_user_id`;
