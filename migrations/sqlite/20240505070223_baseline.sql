-- +goose Up
-- create "projects" table
CREATE TABLE `projects` (
  `project_id` uuid NULL,
  `name` text NULL,
  `root_url` text NULL,
  `resources` blob NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`project_id`)
);
-- create index "idx_projects_deleted_at" to table: "projects"
CREATE INDEX `idx_projects_deleted_at` ON `projects` (`deleted_at`);
-- create index "idx_projects_name" to table: "projects"
CREATE UNIQUE INDEX `idx_projects_name` ON `projects` (`name`);
-- create "connections" table
CREATE TABLE `connections` (
  `connection_id` uuid NULL,
  `integration_id` uuid NULL,
  `project_id` uuid NULL,
  `name` text NULL,
  PRIMARY KEY (`connection_id`),
  CONSTRAINT `fk_connections_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_connections_project_id" to table: "connections"
CREATE INDEX `idx_connections_project_id` ON `connections` (`project_id`);
-- create "signals" table
CREATE TABLE `signals` (
  `signal_id` text NULL,
  `connection_id` uuid NULL,
  `created_at` datetime NULL,
  `workflow_id` text NULL,
  `filter` text NULL,
  PRIMARY KEY (`signal_id`),
  CONSTRAINT `fk_signals_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_connection_id_event_type" to table: "signals"
CREATE INDEX `idx_connection_id_event_type` ON `signals` (`connection_id`);
-- create "secrets" table
CREATE TABLE `secrets` (
  `name` text NULL,
  `data` json NULL,
  PRIMARY KEY (`name`)
);
-- create "vars" table
CREATE TABLE `vars` (
  `scope_id` uuid NULL,
  `name` text NULL,
  `value` text NULL,
  `is_secret` numeric NULL,
  `integration_id` uuid NULL,
  PRIMARY KEY (`scope_id`, `name`)
);
-- create index "idx_vars_scope_id" to table: "vars"
CREATE INDEX `idx_vars_scope_id` ON `vars` (`scope_id`);
-- create index "idx_vars_integration_id" to table: "vars"
CREATE INDEX `idx_vars_integration_id` ON `vars` (`integration_id`);
-- create index "idx_vars_name" to table: "vars"
CREATE INDEX `idx_vars_name` ON `vars` (`name`);
-- create "events" table
CREATE TABLE `events` (
  `event_id` uuid NULL,
  `integration_id` text NULL,
  `connection_id` text NULL,
  `event_type` text NULL,
  `data` json NULL,
  `memo` json NULL,
  `created_at` datetime NULL,
  `seq` integer NULL PRIMARY KEY AUTOINCREMENT,
  `deleted_at` datetime NULL
);
-- create index "idx_events_deleted_at" to table: "events"
CREATE INDEX `idx_events_deleted_at` ON `events` (`deleted_at`);
-- create index "idx_event_type" to table: "events"
CREATE INDEX `idx_event_type` ON `events` (`event_type`);
-- create index "idx_event_type_seq" to table: "events"
CREATE INDEX `idx_event_type_seq` ON `events` (`event_type`);
-- create index "idx_events_connection_id" to table: "events"
CREATE INDEX `idx_events_connection_id` ON `events` (`connection_id`);
-- create index "idx_events_integration_id" to table: "events"
CREATE INDEX `idx_events_integration_id` ON `events` (`integration_id`);
-- create index "idx_events_event_id" to table: "events"
CREATE UNIQUE INDEX `idx_events_event_id` ON `events` (`event_id`);
-- create "builds" table
CREATE TABLE `builds` (
  `build_id` uuid NULL,
  `data` blob NULL,
  `created_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`build_id`)
);
-- create index "idx_builds_deleted_at" to table: "builds"
CREATE INDEX `idx_builds_deleted_at` ON `builds` (`deleted_at`);
-- create "envs" table
CREATE TABLE `envs` (
  `env_id` uuid NULL,
  `project_id` uuid NULL,
  `name` text NULL,
  `deleted_at` datetime NULL,
  `membership_id` text NULL,
  PRIMARY KEY (`env_id`),
  CONSTRAINT `fk_envs_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_envs_membership_id" to table: "envs"
CREATE UNIQUE INDEX `idx_envs_membership_id` ON `envs` (`membership_id`);
-- create index "idx_envs_deleted_at" to table: "envs"
CREATE INDEX `idx_envs_deleted_at` ON `envs` (`deleted_at`);
-- create index "idx_envs_project_id" to table: "envs"
CREATE INDEX `idx_envs_project_id` ON `envs` (`project_id`);
-- create "deployments" table
CREATE TABLE `deployments` (
  `deployment_id` uuid NULL,
  `env_id` uuid NULL,
  `build_id` uuid NULL,
  `state` integer NULL,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`deployment_id`),
  CONSTRAINT `fk_deployments_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_deployments_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_deployments_deleted_at" to table: "deployments"
CREATE INDEX `idx_deployments_deleted_at` ON `deployments` (`deleted_at`);
-- create index "idx_deployments_env_id" to table: "deployments"
CREATE INDEX `idx_deployments_env_id` ON `deployments` (`env_id`);
-- create "sessions" table
CREATE TABLE `sessions` (
  `session_id` uuid NULL,
  `build_id` uuid NULL,
  `env_id` uuid NULL,
  `deployment_id` uuid NULL,
  `event_id` uuid NULL,
  `current_state_type` integer NULL,
  `entrypoint` text NULL,
  `inputs` json NULL,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`session_id`),
  CONSTRAINT `fk_sessions_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_event` FOREIGN KEY (`event_id`) REFERENCES `events` (`event_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_deployment` FOREIGN KEY (`deployment_id`) REFERENCES `deployments` (`deployment_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_sessions_deployment_id" to table: "sessions"
CREATE INDEX `idx_sessions_deployment_id` ON `sessions` (`deployment_id`);
-- create index "idx_sessions_env_id" to table: "sessions"
CREATE INDEX `idx_sessions_env_id` ON `sessions` (`env_id`);
-- create index "idx_sessions_build_id" to table: "sessions"
CREATE INDEX `idx_sessions_build_id` ON `sessions` (`build_id`);
-- create index "idx_sessions_deleted_at" to table: "sessions"
CREATE INDEX `idx_sessions_deleted_at` ON `sessions` (`deleted_at`);
-- create index "idx_sessions_current_state_type" to table: "sessions"
CREATE INDEX `idx_sessions_current_state_type` ON `sessions` (`current_state_type`);
-- create index "idx_sessions_event_id" to table: "sessions"
CREATE INDEX `idx_sessions_event_id` ON `sessions` (`event_id`);
-- create "session_call_specs" table
CREATE TABLE `session_call_specs` (
  `session_id` uuid NULL,
  `seq` integer NULL,
  `data` json NULL,
  PRIMARY KEY (`session_id`, `seq`),
  CONSTRAINT `fk_session_call_specs_session` FOREIGN KEY (`session_id`) REFERENCES `sessions` (`session_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "event_records" table
CREATE TABLE `event_records` (
  `seq` integer NULL,
  `event_id` uuid NULL,
  `state` integer NULL,
  `created_at` datetime NULL,
  PRIMARY KEY (`seq`, `event_id`),
  CONSTRAINT `fk_event_records_event` FOREIGN KEY (`event_id`) REFERENCES `events` (`event_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_event_records_state" to table: "event_records"
CREATE INDEX `idx_event_records_state` ON `event_records` (`state`);
-- create "triggers" table
CREATE TABLE `triggers` (
  `trigger_id` uuid NULL,
  `project_id` uuid NULL,
  `connection_id` uuid NULL,
  `env_id` uuid NULL,
  `name` text NULL,
  `event_type` text NULL,
  `filter` text NULL,
  `code_location` text NULL,
  `data` json NULL,
  `unique_name` text NULL,
  PRIMARY KEY (`trigger_id`),
  CONSTRAINT `fk_triggers_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_triggers_project_id" to table: "triggers"
CREATE INDEX `idx_triggers_project_id` ON `triggers` (`project_id`);
-- create index "idx_triggers_unique_name" to table: "triggers"
CREATE UNIQUE INDEX `idx_triggers_unique_name` ON `triggers` (`unique_name`);
-- create index "idx_triggers_env_id" to table: "triggers"
CREATE INDEX `idx_triggers_env_id` ON `triggers` (`env_id`);
-- create index "idx_triggers_connection_id" to table: "triggers"
CREATE INDEX `idx_triggers_connection_id` ON `triggers` (`connection_id`);
-- create "integrations" table
CREATE TABLE `integrations` (
  `integration_id` uuid NULL,
  `unique_name` text NULL,
  `display_name` text NULL,
  `description` text NULL,
  `logo_url` text NULL,
  `user_links` json NULL,
  `connection_url` text NULL,
  `api_key` text NULL,
  `signing_key` text NULL,
  PRIMARY KEY (`integration_id`)
);
-- create index "idx_integrations_unique_name" to table: "integrations"
CREATE UNIQUE INDEX `idx_integrations_unique_name` ON `integrations` (`unique_name`);
-- create "session_call_attempts" table
CREATE TABLE `session_call_attempts` (
  `session_id` uuid NULL,
  `seq` integer NULL,
  `attempt` integer NULL,
  `start` json NULL,
  `complete` json NULL,
  CONSTRAINT `fk_session_call_attempts_session` FOREIGN KEY (`session_id`) REFERENCES `sessions` (`session_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_session_id_seq_attempt" to table: "session_call_attempts"
CREATE UNIQUE INDEX `idx_session_id_seq_attempt` ON `session_call_attempts` (`session_id`, `seq`, `attempt`);
-- create "session_log_records" table
CREATE TABLE `session_log_records` (
  `session_id` uuid NULL,
  `data` json NULL,
  CONSTRAINT `fk_session_log_records_session` FOREIGN KEY (`session_id`) REFERENCES `sessions` (`session_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_session_log_records_session_id" to table: "session_log_records"
CREATE INDEX `idx_session_log_records_session_id` ON `session_log_records` (`session_id`);

-- +goose Down
-- reverse: create index "idx_session_log_records_session_id" to table: "session_log_records"
DROP INDEX `idx_session_log_records_session_id`;
-- reverse: create "session_log_records" table
DROP TABLE `session_log_records`;
-- reverse: create index "idx_session_id_seq_attempt" to table: "session_call_attempts"
DROP INDEX `idx_session_id_seq_attempt`;
-- reverse: create "session_call_attempts" table
DROP TABLE `session_call_attempts`;
-- reverse: create index "idx_integrations_unique_name" to table: "integrations"
DROP INDEX `idx_integrations_unique_name`;
-- reverse: create "integrations" table
DROP TABLE `integrations`;
-- reverse: create index "idx_triggers_connection_id" to table: "triggers"
DROP INDEX `idx_triggers_connection_id`;
-- reverse: create index "idx_triggers_env_id" to table: "triggers"
DROP INDEX `idx_triggers_env_id`;
-- reverse: create index "idx_triggers_unique_name" to table: "triggers"
DROP INDEX `idx_triggers_unique_name`;
-- reverse: create index "idx_triggers_project_id" to table: "triggers"
DROP INDEX `idx_triggers_project_id`;
-- reverse: create "triggers" table
DROP TABLE `triggers`;
-- reverse: create index "idx_event_records_state" to table: "event_records"
DROP INDEX `idx_event_records_state`;
-- reverse: create "event_records" table
DROP TABLE `event_records`;
-- reverse: create "session_call_specs" table
DROP TABLE `session_call_specs`;
-- reverse: create index "idx_sessions_event_id" to table: "sessions"
DROP INDEX `idx_sessions_event_id`;
-- reverse: create index "idx_sessions_current_state_type" to table: "sessions"
DROP INDEX `idx_sessions_current_state_type`;
-- reverse: create index "idx_sessions_deleted_at" to table: "sessions"
DROP INDEX `idx_sessions_deleted_at`;
-- reverse: create index "idx_sessions_build_id" to table: "sessions"
DROP INDEX `idx_sessions_build_id`;
-- reverse: create index "idx_sessions_env_id" to table: "sessions"
DROP INDEX `idx_sessions_env_id`;
-- reverse: create index "idx_sessions_deployment_id" to table: "sessions"
DROP INDEX `idx_sessions_deployment_id`;
-- reverse: create "sessions" table
DROP TABLE `sessions`;
-- reverse: create index "idx_deployments_env_id" to table: "deployments"
DROP INDEX `idx_deployments_env_id`;
-- reverse: create index "idx_deployments_deleted_at" to table: "deployments"
DROP INDEX `idx_deployments_deleted_at`;
-- reverse: create "deployments" table
DROP TABLE `deployments`;
-- reverse: create index "idx_envs_project_id" to table: "envs"
DROP INDEX `idx_envs_project_id`;
-- reverse: create index "idx_envs_deleted_at" to table: "envs"
DROP INDEX `idx_envs_deleted_at`;
-- reverse: create index "idx_envs_membership_id" to table: "envs"
DROP INDEX `idx_envs_membership_id`;
-- reverse: create "envs" table
DROP TABLE `envs`;
-- reverse: create index "idx_builds_deleted_at" to table: "builds"
DROP INDEX `idx_builds_deleted_at`;
-- reverse: create "builds" table
DROP TABLE `builds`;
-- reverse: create index "idx_events_event_id" to table: "events"
DROP INDEX `idx_events_event_id`;
-- reverse: create index "idx_events_integration_id" to table: "events"
DROP INDEX `idx_events_integration_id`;
-- reverse: create index "idx_events_connection_id" to table: "events"
DROP INDEX `idx_events_connection_id`;
-- reverse: create index "idx_event_type_seq" to table: "events"
DROP INDEX `idx_event_type_seq`;
-- reverse: create index "idx_event_type" to table: "events"
DROP INDEX `idx_event_type`;
-- reverse: create index "idx_events_deleted_at" to table: "events"
DROP INDEX `idx_events_deleted_at`;
-- reverse: create "events" table
DROP TABLE `events`;
-- reverse: create index "idx_vars_name" to table: "vars"
DROP INDEX `idx_vars_name`;
-- reverse: create index "idx_vars_integration_id" to table: "vars"
DROP INDEX `idx_vars_integration_id`;
-- reverse: create index "idx_vars_scope_id" to table: "vars"
DROP INDEX `idx_vars_scope_id`;
-- reverse: create "vars" table
DROP TABLE `vars`;
-- reverse: create "secrets" table
DROP TABLE `secrets`;
-- reverse: create index "idx_connection_id_event_type" to table: "signals"
DROP INDEX `idx_connection_id_event_type`;
-- reverse: create "signals" table
DROP TABLE `signals`;
-- reverse: create index "idx_connections_project_id" to table: "connections"
DROP INDEX `idx_connections_project_id`;
-- reverse: create "connections" table
DROP TABLE `connections`;
-- reverse: create index "idx_projects_name" to table: "projects"
DROP INDEX `idx_projects_name`;
-- reverse: create index "idx_projects_deleted_at" to table: "projects"
DROP INDEX `idx_projects_deleted_at`;
-- reverse: create "projects" table
DROP TABLE `projects`;
