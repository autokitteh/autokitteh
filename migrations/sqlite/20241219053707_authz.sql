-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_builds" table
CREATE TABLE `new_builds` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `build_id` uuid NOT NULL,
  `data` blob NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`build_id`),
  CONSTRAINT `fk_builds_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "builds" to new temporary table "new_builds"
INSERT INTO `new_builds` (`created_at`, `project_id`, `build_id`, `data`, `deleted_at`) SELECT `created_at`, `project_id`, `build_id`, `data`, `deleted_at` FROM `builds`;
-- drop "builds" table after copying rows
DROP TABLE `builds`;
-- rename temporary table "new_builds" to "builds"
ALTER TABLE `new_builds` RENAME TO `builds`;
-- create index "idx_builds_deleted_at" to table: "builds"
CREATE INDEX `idx_builds_deleted_at` ON `builds` (`deleted_at`);
-- create index "idx_builds_project_id" to table: "builds"
CREATE INDEX `idx_builds_project_id` ON `builds` (`project_id`);
-- create "new_connections" table
CREATE TABLE `new_connections` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `connection_id` uuid NOT NULL,
  `integration_id` uuid NULL,
  `name` text NULL,
  `status_code` integer NULL,
  `status_message` text NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
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
-- create index "idx_connections_integration_id" to table: "connections"
CREATE INDEX `idx_connections_integration_id` ON `connections` (`integration_id`);
-- create index "idx_connections_project_id" to table: "connections"
CREATE INDEX `idx_connections_project_id` ON `connections` (`project_id`);
-- create index "idx_connections_deleted_at" to table: "connections"
CREATE INDEX `idx_connections_deleted_at` ON `connections` (`deleted_at`);
-- create index "idx_connections_status_code" to table: "connections"
CREATE INDEX `idx_connections_status_code` ON `connections` (`status_code`);
-- add column "created_by" to table: "projects"
ALTER TABLE `projects` ADD COLUMN `created_by` uuid NOT NULL;
-- add column "created_at" to table: "projects"
ALTER TABLE `projects` ADD COLUMN `created_at` datetime NULL;
-- add column "org_id" to table: "projects"
ALTER TABLE `projects` ADD COLUMN `org_id` uuid NOT NULL;
-- add column "updated_by" to table: "projects"
ALTER TABLE `projects` ADD COLUMN `updated_by` uuid NULL;
-- add column "updated_at" to table: "projects"
ALTER TABLE `projects` ADD COLUMN `updated_at` datetime NULL;
-- create index "idx_projects_org_id" to table: "projects"
CREATE INDEX `idx_projects_org_id` ON `projects` (`org_id`);
-- create "new_vars" table
CREATE TABLE `new_vars` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `var_id` uuid NOT NULL,
  `name` text NOT NULL,
  `value` text NULL,
  `is_secret` numeric NULL,
  `integration_id` uuid NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`var_id`, `name`),
  CONSTRAINT `fk_vars_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "vars" to new temporary table "new_vars"
INSERT INTO `new_vars` (`var_id`, `name`, `value`, `is_secret`, `integration_id`) SELECT `var_id`, `name`, `value`, `is_secret`, `integration_id` FROM `vars`;
-- drop "vars" table after copying rows
DROP TABLE `vars`;
-- rename temporary table "new_vars" to "vars"
ALTER TABLE `new_vars` RENAME TO `vars`;
-- create index "idx_vars_project_id" to table: "vars"
CREATE INDEX `idx_vars_project_id` ON `vars` (`project_id`);
-- create index "idx_vars_integration_id" to table: "vars"
CREATE INDEX `idx_vars_integration_id` ON `vars` (`integration_id`);
-- create index "idx_vars_name" to table: "vars"
CREATE INDEX `idx_vars_name` ON `vars` (`name`);
-- create index "idx_vars_var_id" to table: "vars"
CREATE INDEX `idx_vars_var_id` ON `vars` (`var_id`);
-- create "new_events" table
CREATE TABLE `new_events` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `event_id` uuid NOT NULL,
  `destination_id` uuid NOT NULL,
  `integration_id` uuid NULL,
  `connection_id` uuid NULL,
  `trigger_id` uuid NULL,
  `event_type` text NULL,
  `data` json NULL,
  `memo` json NULL,
  `seq` integer NULL PRIMARY KEY AUTOINCREMENT,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  CONSTRAINT `fk_events_trigger` FOREIGN KEY (`trigger_id`) REFERENCES `triggers` (`trigger_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_events_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_events_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "events" to new temporary table "new_events"
INSERT INTO `new_events` (`created_at`, `event_id`, `destination_id`, `integration_id`, `connection_id`, `trigger_id`, `event_type`, `data`, `memo`, `seq`, `deleted_at`) SELECT `created_at`, `event_id`, `destination_id`, `integration_id`, `connection_id`, `trigger_id`, `event_type`, `data`, `memo`, `seq`, `deleted_at` FROM `events`;
-- drop "events" table after copying rows
DROP TABLE `events`;
-- rename temporary table "new_events" to "events"
ALTER TABLE `new_events` RENAME TO `events`;
-- create index "idx_event_type_seq" to table: "events"
CREATE INDEX `idx_event_type_seq` ON `events` (`event_type`);
-- create index "idx_events_project_id" to table: "events"
CREATE INDEX `idx_events_project_id` ON `events` (`project_id`);
-- create index "idx_event_type" to table: "events"
CREATE INDEX `idx_event_type` ON `events` (`event_type`);
-- create index "idx_events_destination_id" to table: "events"
CREATE INDEX `idx_events_destination_id` ON `events` (`destination_id`);
-- create index "idx_events_event_id" to table: "events"
CREATE UNIQUE INDEX `idx_events_event_id` ON `events` (`event_id`);
-- create index "idx_events_deleted_at" to table: "events"
CREATE INDEX `idx_events_deleted_at` ON `events` (`deleted_at`);
-- create index "idx_events_trigger_id" to table: "events"
CREATE INDEX `idx_events_trigger_id` ON `events` (`trigger_id`);
-- create index "idx_events_connection_id" to table: "events"
CREATE INDEX `idx_events_connection_id` ON `events` (`connection_id`);
-- create index "idx_events_integration_id" to table: "events"
CREATE INDEX `idx_events_integration_id` ON `events` (`integration_id`);
-- create "new_values" table
CREATE TABLE `new_values` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `key` text NOT NULL,
  `value` blob NULL,
  `updated_by` uuid NULL,
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
-- add column "created_by" to table: "users"
ALTER TABLE `users` ADD COLUMN `created_by` uuid NOT NULL;
-- add column "created_at" to table: "users"
ALTER TABLE `users` ADD COLUMN `created_at` datetime NULL;
-- add column "name" to table: "users"
ALTER TABLE `users` ADD COLUMN `name` text NOT NULL;
-- add column "disabled" to table: "users"
ALTER TABLE `users` ADD COLUMN `disabled` numeric NULL;
-- add column "default_org_id" to table: "users"
ALTER TABLE `users` ADD COLUMN `default_org_id` uuid NULL;
-- add column "updated_by" to table: "users"
ALTER TABLE `users` ADD COLUMN `updated_by` uuid NULL;
-- add column "updated_at" to table: "users"
ALTER TABLE `users` ADD COLUMN `updated_at` datetime NULL;
-- create index "idx_users_name" to table: "users"
CREATE UNIQUE INDEX `idx_users_name` ON `users` (`name`);
-- create "new_sessions" table
CREATE TABLE `new_sessions` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `session_id` uuid NOT NULL,
  `build_id` uuid NOT NULL,
  `deployment_id` uuid NULL,
  `event_id` uuid NULL,
  `current_state_type` integer NULL,
  `entrypoint` text NULL,
  `inputs` json NULL,
  `memo` json NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`session_id`),
  CONSTRAINT `fk_sessions_event` FOREIGN KEY (`event_id`) REFERENCES `events` (`event_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_deployment` FOREIGN KEY (`deployment_id`) REFERENCES `deployments` (`deployment_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "sessions" to new temporary table "new_sessions"
INSERT INTO `new_sessions` (`created_at`, `project_id`, `session_id`, `build_id`, `deployment_id`, `event_id`, `current_state_type`, `entrypoint`, `inputs`, `memo`, `updated_at`, `deleted_at`) SELECT `created_at`, `project_id`, `session_id`, `build_id`, `deployment_id`, `event_id`, `current_state_type`, `entrypoint`, `inputs`, `memo`, `updated_at`, `deleted_at` FROM `sessions`;
-- drop "sessions" table after copying rows
DROP TABLE `sessions`;
-- rename temporary table "new_sessions" to "sessions"
ALTER TABLE `new_sessions` RENAME TO `sessions`;
-- create index "idx_sessions_deleted_at" to table: "sessions"
CREATE INDEX `idx_sessions_deleted_at` ON `sessions` (`deleted_at`);
-- create index "idx_sessions_current_state_type" to table: "sessions"
CREATE INDEX `idx_sessions_current_state_type` ON `sessions` (`current_state_type`);
-- create index "idx_sessions_event_id" to table: "sessions"
CREATE INDEX `idx_sessions_event_id` ON `sessions` (`event_id`);
-- create index "idx_sessions_deployment_id" to table: "sessions"
CREATE INDEX `idx_sessions_deployment_id` ON `sessions` (`deployment_id`);
-- create index "idx_sessions_build_id" to table: "sessions"
CREATE INDEX `idx_sessions_build_id` ON `sessions` (`build_id`);
-- create index "idx_sessions_project_id" to table: "sessions"
CREATE INDEX `idx_sessions_project_id` ON `sessions` (`project_id`);
-- create "new_deployments" table
CREATE TABLE `new_deployments` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `deployment_id` uuid NOT NULL,
  `build_id` uuid NOT NULL,
  `state` integer NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`deployment_id`),
  CONSTRAINT `fk_deployments_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_deployments_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "deployments" to new temporary table "new_deployments"
INSERT INTO `new_deployments` (`created_at`, `project_id`, `deployment_id`, `build_id`, `state`, `updated_at`, `deleted_at`) SELECT `created_at`, `project_id`, `deployment_id`, `build_id`, `state`, `updated_at`, `deleted_at` FROM `deployments`;
-- drop "deployments" table after copying rows
DROP TABLE `deployments`;
-- rename temporary table "new_deployments" to "deployments"
ALTER TABLE `new_deployments` RENAME TO `deployments`;
-- create index "idx_deployments_project_id" to table: "deployments"
CREATE INDEX `idx_deployments_project_id` ON `deployments` (`project_id`);
-- create index "idx_deployments_deleted_at" to table: "deployments"
CREATE INDEX `idx_deployments_deleted_at` ON `deployments` (`deleted_at`);
-- add column "created_by" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `created_by` uuid NOT NULL;
-- add column "created_at" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `created_at` datetime NULL;
-- add column "updated_by" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `updated_by` uuid NULL;
-- add column "updated_at" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `updated_at` datetime NULL;
-- create "bases" table
CREATE TABLE `bases` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL
);
-- create "belongs_to_projects" table
CREATE TABLE `belongs_to_projects` (
  `project_id` uuid NOT NULL,
  CONSTRAINT `fk_belongs_to_projects_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
CREATE INDEX `idx_belongs_to_projects_project_id` ON `belongs_to_projects` (`project_id`);
-- create "orgs" table
CREATE TABLE `orgs` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `org_id` uuid NOT NULL,
  `name` text NOT NULL,
  `display_name` text NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`org_id`)
);
-- create index "idx_orgs_name" to table: "orgs"
CREATE UNIQUE INDEX `idx_orgs_name` ON `orgs` (`name`);
-- create "org_members" table
CREATE TABLE `org_members` (
  `created_by` uuid NOT NULL,
  `created_at` datetime NULL,
  `org_id` uuid NOT NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`org_id`, `user_id`)
);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "org_members" table
DROP TABLE `org_members`;
-- reverse: create index "idx_orgs_name" to table: "orgs"
DROP INDEX `idx_orgs_name`;
-- reverse: create "orgs" table
DROP TABLE `orgs`;
-- reverse: create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
DROP INDEX `idx_belongs_to_projects_project_id`;
-- reverse: create "belongs_to_projects" table
DROP TABLE `belongs_to_projects`;
-- reverse: create "bases" table
DROP TABLE `bases`;
-- reverse: add column "updated_at" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `updated_at`;
-- reverse: add column "updated_by" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `updated_by`;
-- reverse: add column "created_at" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `created_at`;
-- reverse: add column "created_by" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `created_by`;
-- reverse: create index "idx_deployments_deleted_at" to table: "deployments"
DROP INDEX `idx_deployments_deleted_at`;
-- reverse: create index "idx_deployments_project_id" to table: "deployments"
DROP INDEX `idx_deployments_project_id`;
-- reverse: create "new_deployments" table
DROP TABLE `new_deployments`;
-- reverse: create index "idx_sessions_project_id" to table: "sessions"
DROP INDEX `idx_sessions_project_id`;
-- reverse: create index "idx_sessions_build_id" to table: "sessions"
DROP INDEX `idx_sessions_build_id`;
-- reverse: create index "idx_sessions_deployment_id" to table: "sessions"
DROP INDEX `idx_sessions_deployment_id`;
-- reverse: create index "idx_sessions_event_id" to table: "sessions"
DROP INDEX `idx_sessions_event_id`;
-- reverse: create index "idx_sessions_current_state_type" to table: "sessions"
DROP INDEX `idx_sessions_current_state_type`;
-- reverse: create index "idx_sessions_deleted_at" to table: "sessions"
DROP INDEX `idx_sessions_deleted_at`;
-- reverse: create "new_sessions" table
DROP TABLE `new_sessions`;
-- reverse: create index "idx_users_name" to table: "users"
DROP INDEX `idx_users_name`;
-- reverse: add column "updated_at" to table: "users"
ALTER TABLE `users` DROP COLUMN `updated_at`;
-- reverse: add column "updated_by" to table: "users"
ALTER TABLE `users` DROP COLUMN `updated_by`;
-- reverse: add column "default_org_id" to table: "users"
ALTER TABLE `users` DROP COLUMN `default_org_id`;
-- reverse: add column "disabled" to table: "users"
ALTER TABLE `users` DROP COLUMN `disabled`;
-- reverse: add column "name" to table: "users"
ALTER TABLE `users` DROP COLUMN `name`;
-- reverse: add column "created_at" to table: "users"
ALTER TABLE `users` DROP COLUMN `created_at`;
-- reverse: add column "created_by" to table: "users"
ALTER TABLE `users` DROP COLUMN `created_by`;
-- reverse: create index "idx_values_project_id" to table: "values"
DROP INDEX `idx_values_project_id`;
-- reverse: create "new_values" table
DROP TABLE `new_values`;
-- reverse: create index "idx_events_integration_id" to table: "events"
DROP INDEX `idx_events_integration_id`;
-- reverse: create index "idx_events_connection_id" to table: "events"
DROP INDEX `idx_events_connection_id`;
-- reverse: create index "idx_events_trigger_id" to table: "events"
DROP INDEX `idx_events_trigger_id`;
-- reverse: create index "idx_events_deleted_at" to table: "events"
DROP INDEX `idx_events_deleted_at`;
-- reverse: create index "idx_events_event_id" to table: "events"
DROP INDEX `idx_events_event_id`;
-- reverse: create index "idx_events_destination_id" to table: "events"
DROP INDEX `idx_events_destination_id`;
-- reverse: create index "idx_event_type" to table: "events"
DROP INDEX `idx_event_type`;
-- reverse: create index "idx_events_project_id" to table: "events"
DROP INDEX `idx_events_project_id`;
-- reverse: create index "idx_event_type_seq" to table: "events"
DROP INDEX `idx_event_type_seq`;
-- reverse: create "new_events" table
DROP TABLE `new_events`;
-- reverse: create index "idx_vars_var_id" to table: "vars"
DROP INDEX `idx_vars_var_id`;
-- reverse: create index "idx_vars_name" to table: "vars"
DROP INDEX `idx_vars_name`;
-- reverse: create index "idx_vars_integration_id" to table: "vars"
DROP INDEX `idx_vars_integration_id`;
-- reverse: create index "idx_vars_project_id" to table: "vars"
DROP INDEX `idx_vars_project_id`;
-- reverse: create "new_vars" table
DROP TABLE `new_vars`;
-- reverse: create index "idx_projects_org_id" to table: "projects"
DROP INDEX `idx_projects_org_id`;
-- reverse: add column "updated_at" to table: "projects"
ALTER TABLE `projects` DROP COLUMN `updated_at`;
-- reverse: add column "updated_by" to table: "projects"
ALTER TABLE `projects` DROP COLUMN `updated_by`;
-- reverse: add column "org_id" to table: "projects"
ALTER TABLE `projects` DROP COLUMN `org_id`;
-- reverse: add column "created_at" to table: "projects"
ALTER TABLE `projects` DROP COLUMN `created_at`;
-- reverse: add column "created_by" to table: "projects"
ALTER TABLE `projects` DROP COLUMN `created_by`;
-- reverse: create index "idx_connections_status_code" to table: "connections"
DROP INDEX `idx_connections_status_code`;
-- reverse: create index "idx_connections_deleted_at" to table: "connections"
DROP INDEX `idx_connections_deleted_at`;
-- reverse: create index "idx_connections_project_id" to table: "connections"
DROP INDEX `idx_connections_project_id`;
-- reverse: create index "idx_connections_integration_id" to table: "connections"
DROP INDEX `idx_connections_integration_id`;
-- reverse: create "new_connections" table
DROP TABLE `new_connections`;
-- reverse: create index "idx_builds_project_id" to table: "builds"
DROP INDEX `idx_builds_project_id`;
-- reverse: create index "idx_builds_deleted_at" to table: "builds"
DROP INDEX `idx_builds_deleted_at`;
-- reverse: create "new_builds" table
DROP TABLE `new_builds`;
