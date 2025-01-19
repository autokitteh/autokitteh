-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_session_call_attempts" table
CREATE TABLE `new_session_call_attempts` (
  `session_id` uuid NOT NULL,
  `seq` integer NULL,
  `attempt` integer NULL,
  `start` json NULL,
  `complete` json NULL,
  CONSTRAINT `fk_session_call_attempts_session` FOREIGN KEY (`session_id`) REFERENCES `sessions` (`session_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "session_call_attempts" to new temporary table "new_session_call_attempts"
INSERT INTO `new_session_call_attempts` (`session_id`, `seq`, `attempt`, `start`, `complete`) SELECT `session_id`, `seq`, `attempt`, `start`, `complete` FROM `session_call_attempts`;
-- drop "session_call_attempts" table after copying rows
DROP TABLE `session_call_attempts`;
-- rename temporary table "new_session_call_attempts" to "session_call_attempts"
ALTER TABLE `new_session_call_attempts` RENAME TO `session_call_attempts`;
-- create index "idx_session_id_seq_attempt" to table: "session_call_attempts"
CREATE UNIQUE INDEX `idx_session_id_seq_attempt` ON `session_call_attempts` (`session_id`, `seq`, `attempt`);
-- create "new_session_call_specs" table
CREATE TABLE `new_session_call_specs` (
  `session_id` uuid NOT NULL,
  `seq` integer NULL,
  `data` json NULL,
  PRIMARY KEY (`session_id`, `seq`),
  CONSTRAINT `fk_session_call_specs_session` FOREIGN KEY (`session_id`) REFERENCES `sessions` (`session_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "session_call_specs" to new temporary table "new_session_call_specs"
INSERT INTO `new_session_call_specs` (`session_id`, `seq`, `data`) SELECT `session_id`, `seq`, `data` FROM `session_call_specs`;
-- drop "session_call_specs" table after copying rows
DROP TABLE `session_call_specs`;
-- rename temporary table "new_session_call_specs" to "session_call_specs"
ALTER TABLE `new_session_call_specs` RENAME TO `session_call_specs`;
-- create "new_session_log_records" table
CREATE TABLE `new_session_log_records` (
  `session_id` uuid NOT NULL,
  `seq` integer NOT NULL,
  `data` json NULL,
  `type` text NULL,
  PRIMARY KEY (`session_id`, `seq`),
  CONSTRAINT `fk_session_log_records_session` FOREIGN KEY (`session_id`) REFERENCES `sessions` (`session_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "session_log_records" to new temporary table "new_session_log_records"
INSERT INTO `new_session_log_records` (`session_id`, `seq`, `data`, `type`) SELECT `session_id`, `seq`, `data`, `type` FROM `session_log_records`;
-- drop "session_log_records" table after copying rows
DROP TABLE `session_log_records`;
-- rename temporary table "new_session_log_records" to "session_log_records"
ALTER TABLE `new_session_log_records` RENAME TO `session_log_records`;
-- create index "idx_session_log_records_type" to table: "session_log_records"
CREATE INDEX `idx_session_log_records_type` ON `session_log_records` (`type`);
-- create "new_projects" table
CREATE TABLE `new_projects` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `org_id` uuid NULL,
  `name` text NOT NULL,
  `root_url` text NULL,
  `resources` blob NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`project_id`)
);
-- copy rows from old table "projects" to new temporary table "new_projects"
INSERT INTO `new_projects` (`created_by`, `created_at`, `project_id`, `org_id`, `name`, `root_url`, `resources`, `updated_by`, `updated_at`) SELECT `created_by`, `created_at`, `project_id`, `org_id`, `name`, `root_url`, `resources`, `updated_by`, `updated_at` FROM `projects`;
-- drop "projects" table after copying rows
DROP TABLE `projects`;
-- rename temporary table "new_projects" to "projects"
ALTER TABLE `new_projects` RENAME TO `projects`;
-- create index "idx_projects_org_id" to table: "projects"
CREATE INDEX `idx_projects_org_id` ON `projects` (`org_id`);
-- create index "idx_projects_name" to table: "projects"
CREATE INDEX `idx_projects_name` ON `projects` (`name`);
-- create "new_signals" table
CREATE TABLE `new_signals` (
  `signal_id` uuid NOT NULL,
  `destination_id` uuid NOT NULL,
  `connection_id` uuid NULL,
  `trigger_id` uuid NULL,
  `created_at` datetime NULL,
  `workflow_id` text NULL,
  `filter` text NULL,
  PRIMARY KEY (`signal_id`),
  CONSTRAINT `fk_signals_trigger` FOREIGN KEY (`trigger_id`) REFERENCES `triggers` (`trigger_id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `fk_signals_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "signals" to new temporary table "new_signals"
INSERT INTO `new_signals` (`signal_id`, `destination_id`, `connection_id`, `trigger_id`, `created_at`, `workflow_id`, `filter`) SELECT `signal_id`, `destination_id`, `connection_id`, `trigger_id`, `created_at`, `workflow_id`, `filter` FROM `signals`;
-- drop "signals" table after copying rows
DROP TABLE `signals`;
-- rename temporary table "new_signals" to "signals"
ALTER TABLE `new_signals` RENAME TO `signals`;
-- create index "idx_signals_destination_id" to table: "signals"
CREATE INDEX `idx_signals_destination_id` ON `signals` (`destination_id`);
-- create "new_triggers" table
CREATE TABLE `new_triggers` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `trigger_id` uuid NOT NULL,
  `connection_id` uuid NULL,
  `source_type` text NULL,
  `event_type` text NULL,
  `filter` text NULL,
  `code_location` text NULL,
  `name` text NULL,
  `unique_name` text NOT NULL,
  `webhook_slug` text NULL,
  `schedule` text NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`trigger_id`),
  CONSTRAINT `fk_triggers_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `fk_triggers_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "triggers" to new temporary table "new_triggers"
INSERT INTO `new_triggers` (`created_by`, `created_at`, `project_id`, `trigger_id`, `connection_id`, `source_type`, `event_type`, `filter`, `code_location`, `name`, `unique_name`, `webhook_slug`, `schedule`, `updated_by`, `updated_at`) SELECT `created_by`, `created_at`, `project_id`, `trigger_id`, `connection_id`, `source_type`, `event_type`, `filter`, `code_location`, `name`, `unique_name`, `webhook_slug`, `schedule`, `updated_by`, `updated_at` FROM `triggers`;
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
-- create "new_builds" table
CREATE TABLE `new_builds` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `build_id` uuid NOT NULL,
  `data` blob NULL,
  PRIMARY KEY (`build_id`),
  CONSTRAINT `fk_builds_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "builds" to new temporary table "new_builds"
INSERT INTO `new_builds` (`created_by`, `created_at`, `project_id`, `build_id`, `data`) SELECT `created_by`, `created_at`, `project_id`, `build_id`, `data` FROM `builds`;
-- drop "builds" table after copying rows
DROP TABLE `builds`;
-- rename temporary table "new_builds" to "builds"
ALTER TABLE `new_builds` RENAME TO `builds`;
-- create index "idx_builds_project_id" to table: "builds"
CREATE INDEX `idx_builds_project_id` ON `builds` (`project_id`);
-- create "new_connections" table
CREATE TABLE `new_connections` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `connection_id` uuid NOT NULL,
  `integration_id` uuid NULL,
  `name` text NULL,
  `status_code` integer NULL,
  `status_message` text NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`connection_id`),
  CONSTRAINT `fk_connections_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "connections" to new temporary table "new_connections"
INSERT INTO `new_connections` (`created_by`, `created_at`, `project_id`, `connection_id`, `integration_id`, `name`, `status_code`, `status_message`, `updated_by`, `updated_at`) SELECT `created_by`, `created_at`, `project_id`, `connection_id`, `integration_id`, `name`, `status_code`, `status_message`, `updated_by`, `updated_at` FROM `connections`;
-- drop "connections" table after copying rows
DROP TABLE `connections`;
-- rename temporary table "new_connections" to "connections"
ALTER TABLE `new_connections` RENAME TO `connections`;
-- create index "idx_connections_integration_id" to table: "connections"
CREATE INDEX `idx_connections_integration_id` ON `connections` (`integration_id`);
-- create index "idx_connections_project_id" to table: "connections"
CREATE INDEX `idx_connections_project_id` ON `connections` (`project_id`);
-- create index "idx_connections_status_code" to table: "connections"
CREATE INDEX `idx_connections_status_code` ON `connections` (`status_code`);
-- create "new_events" table
CREATE TABLE `new_events` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NULL,
  `event_id` uuid NOT NULL,
  `destination_id` uuid NOT NULL,
  `integration_id` uuid NULL,
  `connection_id` uuid NULL,
  `trigger_id` uuid NULL,
  `event_type` text NULL,
  `data` json NULL,
  `memo` json NULL,
  `seq` integer NULL PRIMARY KEY AUTOINCREMENT,
  CONSTRAINT `fk_events_trigger` FOREIGN KEY (`trigger_id`) REFERENCES `triggers` (`trigger_id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `fk_events_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `fk_events_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "events" to new temporary table "new_events"
INSERT INTO `new_events` (`created_by`, `created_at`, `project_id`, `event_id`, `destination_id`, `integration_id`, `connection_id`, `trigger_id`, `event_type`, `data`, `memo`, `seq`) SELECT `created_by`, `created_at`, `project_id`, `event_id`, `destination_id`, `integration_id`, `connection_id`, `trigger_id`, `event_type`, `data`, `memo`, `seq` FROM `events`;
-- drop "events" table after copying rows
DROP TABLE `events`;
-- rename temporary table "new_events" to "events"
ALTER TABLE `new_events` RENAME TO `events`;
-- create index "idx_events_trigger_id" to table: "events"
CREATE INDEX `idx_events_trigger_id` ON `events` (`trigger_id`);
-- create index "idx_events_connection_id" to table: "events"
CREATE INDEX `idx_events_connection_id` ON `events` (`connection_id`);
-- create index "idx_events_integration_id" to table: "events"
CREATE INDEX `idx_events_integration_id` ON `events` (`integration_id`);
-- create index "idx_events_destination_id" to table: "events"
CREATE INDEX `idx_events_destination_id` ON `events` (`destination_id`);
-- create index "idx_events_event_id" to table: "events"
CREATE UNIQUE INDEX `idx_events_event_id` ON `events` (`event_id`);
-- create index "idx_events_project_id" to table: "events"
CREATE INDEX `idx_events_project_id` ON `events` (`project_id`);
-- create index "idx_event_type" to table: "events"
CREATE INDEX `idx_event_type` ON `events` (`event_type`);
-- create index "idx_event_type_seq" to table: "events"
CREATE INDEX `idx_event_type_seq` ON `events` (`event_type`);
-- create "new_values" table
CREATE TABLE `new_values` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `key` text NOT NULL,
  `value` blob NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`key`),
  CONSTRAINT `fk_values_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "values" to new temporary table "new_values"
INSERT INTO `new_values` (`created_by`, `created_at`, `project_id`, `key`, `value`, `updated_by`, `updated_at`) SELECT `created_by`, `created_at`, `project_id`, `key`, `value`, `updated_by`, `updated_at` FROM `values`;
-- drop "values" table after copying rows
DROP TABLE `values`;
-- rename temporary table "new_values" to "values"
ALTER TABLE `new_values` RENAME TO `values`;
-- create index "idx_values_project_id" to table: "values"
CREATE INDEX `idx_values_project_id` ON `values` (`project_id`);
-- create "new_sessions" table
CREATE TABLE `new_sessions` (
  `created_by` uuid NULL,
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
  PRIMARY KEY (`session_id`),
  CONSTRAINT `fk_sessions_event` FOREIGN KEY (`event_id`) REFERENCES `events` (`event_id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `fk_sessions_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `fk_sessions_deployment` FOREIGN KEY (`deployment_id`) REFERENCES `deployments` (`deployment_id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `fk_sessions_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "sessions" to new temporary table "new_sessions"
INSERT INTO `new_sessions` (`created_by`, `created_at`, `project_id`, `session_id`, `build_id`, `deployment_id`, `event_id`, `current_state_type`, `entrypoint`, `inputs`, `memo`, `updated_by`, `updated_at`) SELECT `created_by`, `created_at`, `project_id`, `session_id`, `build_id`, `deployment_id`, `event_id`, `current_state_type`, `entrypoint`, `inputs`, `memo`, `updated_by`, `updated_at` FROM `sessions`;
-- drop "sessions" table after copying rows
DROP TABLE `sessions`;
-- rename temporary table "new_sessions" to "sessions"
ALTER TABLE `new_sessions` RENAME TO `sessions`;
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
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `deployment_id` uuid NOT NULL,
  `build_id` uuid NOT NULL,
  `state` integer NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`deployment_id`),
  CONSTRAINT `fk_deployments_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `fk_deployments_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "deployments" to new temporary table "new_deployments"
INSERT INTO `new_deployments` (`created_by`, `created_at`, `project_id`, `deployment_id`, `build_id`, `state`, `updated_by`, `updated_at`) SELECT `created_by`, `created_at`, `project_id`, `deployment_id`, `build_id`, `state`, `updated_by`, `updated_at` FROM `deployments`;
-- drop "deployments" table after copying rows
DROP TABLE `deployments`;
-- rename temporary table "new_deployments" to "deployments"
ALTER TABLE `new_deployments` RENAME TO `deployments`;
-- create index "idx_deployments_state" to table: "deployments"
CREATE INDEX `idx_deployments_state` ON `deployments` (`state`);
-- create index "idx_deployments_project_id" to table: "deployments"
CREATE INDEX `idx_deployments_project_id` ON `deployments` (`project_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_deployments_project_id" to table: "deployments"
DROP INDEX `idx_deployments_project_id`;
-- reverse: create index "idx_deployments_state" to table: "deployments"
DROP INDEX `idx_deployments_state`;
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
-- reverse: create "new_sessions" table
DROP TABLE `new_sessions`;
-- reverse: create index "idx_values_project_id" to table: "values"
DROP INDEX `idx_values_project_id`;
-- reverse: create "new_values" table
DROP TABLE `new_values`;
-- reverse: create index "idx_event_type_seq" to table: "events"
DROP INDEX `idx_event_type_seq`;
-- reverse: create index "idx_event_type" to table: "events"
DROP INDEX `idx_event_type`;
-- reverse: create index "idx_events_project_id" to table: "events"
DROP INDEX `idx_events_project_id`;
-- reverse: create index "idx_events_event_id" to table: "events"
DROP INDEX `idx_events_event_id`;
-- reverse: create index "idx_events_destination_id" to table: "events"
DROP INDEX `idx_events_destination_id`;
-- reverse: create index "idx_events_integration_id" to table: "events"
DROP INDEX `idx_events_integration_id`;
-- reverse: create index "idx_events_connection_id" to table: "events"
DROP INDEX `idx_events_connection_id`;
-- reverse: create index "idx_events_trigger_id" to table: "events"
DROP INDEX `idx_events_trigger_id`;
-- reverse: create "new_events" table
DROP TABLE `new_events`;
-- reverse: create index "idx_connections_status_code" to table: "connections"
DROP INDEX `idx_connections_status_code`;
-- reverse: create index "idx_connections_project_id" to table: "connections"
DROP INDEX `idx_connections_project_id`;
-- reverse: create index "idx_connections_integration_id" to table: "connections"
DROP INDEX `idx_connections_integration_id`;
-- reverse: create "new_connections" table
DROP TABLE `new_connections`;
-- reverse: create index "idx_builds_project_id" to table: "builds"
DROP INDEX `idx_builds_project_id`;
-- reverse: create "new_builds" table
DROP TABLE `new_builds`;
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
-- reverse: create index "idx_signals_destination_id" to table: "signals"
DROP INDEX `idx_signals_destination_id`;
-- reverse: create "new_signals" table
DROP TABLE `new_signals`;
-- reverse: create index "idx_projects_name" to table: "projects"
DROP INDEX `idx_projects_name`;
-- reverse: create index "idx_projects_org_id" to table: "projects"
DROP INDEX `idx_projects_org_id`;
-- reverse: create "new_projects" table
DROP TABLE `new_projects`;
-- reverse: create index "idx_session_log_records_type" to table: "session_log_records"
DROP INDEX `idx_session_log_records_type`;
-- reverse: create "new_session_log_records" table
DROP TABLE `new_session_log_records`;
-- reverse: create "new_session_call_specs" table
DROP TABLE `new_session_call_specs`;
-- reverse: create index "idx_session_id_seq_attempt" to table: "session_call_attempts"
DROP INDEX `idx_session_id_seq_attempt`;
-- reverse: create "new_session_call_attempts" table
DROP TABLE `new_session_call_attempts`;
