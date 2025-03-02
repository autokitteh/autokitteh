-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;

DELETE from "session_log_records" where session_id in (select session_id from sessions where deleted_at is not NULL);
DELETE from "session_call_specs" where session_id in (select session_id from sessions where deleted_at is not NULL);
DELETE from "session_call_attempts" where session_id in (select session_id from sessions where deleted_at is not NULL);
DELETE from "sessions" where deleted_at is not NULL;
DELETE from events where deleted_at is not NULL;

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
  CONSTRAINT `fk_events_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `fk_events_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_events_trigger` FOREIGN KEY (`trigger_id`) REFERENCES `triggers` (`trigger_id`) ON UPDATE NO ACTION ON DELETE SET NULL
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
  CONSTRAINT `fk_sessions_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_deployment` FOREIGN KEY (`deployment_id`) REFERENCES `deployments` (`deployment_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
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
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
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
