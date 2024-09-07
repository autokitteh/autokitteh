-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_events" table
CREATE TABLE `new_events` (
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
  CONSTRAINT `fk_events_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "events" to new temporary table "new_events"
INSERT INTO `new_events` (`event_id`, `integration_id`, `connection_id`, `event_type`, `data`, `memo`, `created_at`, `seq`, `deleted_at`) SELECT `event_id`, `integration_id`, `connection_id`, `event_type`, `data`, `memo`, `created_at`, `seq`, `deleted_at` FROM `events`;
-- drop "events" table after copying rows
DROP TABLE `events`;
-- rename temporary table "new_events" to "events"
ALTER TABLE `new_events` RENAME TO `events`;
-- create index "idx_event_type" to table: "events"
CREATE INDEX `idx_event_type` ON `events` (`event_type`);
-- create index "idx_event_type_seq" to table: "events"
CREATE INDEX `idx_event_type_seq` ON `events` (`event_type`);
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
-- create index "idx_events_deleted_at" to table: "events"
CREATE INDEX `idx_events_deleted_at` ON `events` (`deleted_at`);
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
  CONSTRAINT `fk_signals_trigger` FOREIGN KEY (`trigger_id`) REFERENCES `triggers` (`trigger_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_signals_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "signals" to new temporary table "new_signals"
INSERT INTO `new_signals` (`signal_id`, `connection_id`, `created_at`, `workflow_id`, `filter`) SELECT `signal_id`, `connection_id`, `created_at`, `workflow_id`, `filter` FROM `signals`;
-- drop "signals" table after copying rows
DROP TABLE `signals`;
-- rename temporary table "new_signals" to "signals"
ALTER TABLE `new_signals` RENAME TO `signals`;
-- create index "idx_signals_destination_id" to table: "signals"
CREATE INDEX `idx_signals_destination_id` ON `signals` (`destination_id`);
-- create "new_triggers" table
CREATE TABLE `new_triggers` (
  `trigger_id` uuid NOT NULL,
  `project_id` uuid NOT NULL,
  `source_type` text NULL,
  `connection_id` uuid NULL,
  `env_id` uuid NOT NULL,
  `name` text NULL,
  `event_type` text NULL,
  `filter` text NULL,
  `code_location` text NULL,
  `unique_name` text NOT NULL,
  `webhook_slug` text NULL,
  `schedule` text NULL,
  PRIMARY KEY (`trigger_id`),
  CONSTRAINT `fk_triggers_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "triggers" to new temporary table "new_triggers"
INSERT INTO `new_triggers` (`trigger_id`, `project_id`, `connection_id`, `env_id`, `name`, `event_type`, `filter`, `code_location`, `unique_name`) SELECT `trigger_id`, `project_id`, `connection_id`, `env_id`, `name`, `event_type`, `filter`, `code_location`, `unique_name` FROM `triggers`;
-- drop "triggers" table after copying rows
DROP TABLE `triggers`;
-- rename temporary table "new_triggers" to "triggers"
ALTER TABLE `new_triggers` RENAME TO `triggers`;
-- create index "idx_triggers_source_type" to table: "triggers"
CREATE INDEX `idx_triggers_source_type` ON `triggers` (`source_type`);
-- create index "idx_triggers_project_id" to table: "triggers"
CREATE INDEX `idx_triggers_project_id` ON `triggers` (`project_id`);
-- create index "idx_triggers_webhook_slug" to table: "triggers"
CREATE INDEX `idx_triggers_webhook_slug` ON `triggers` (`webhook_slug`);
-- create index "idx_triggers_unique_name" to table: "triggers"
CREATE UNIQUE INDEX `idx_triggers_unique_name` ON `triggers` (`unique_name`);
-- create index "idx_triggers_env_id" to table: "triggers"
CREATE INDEX `idx_triggers_env_id` ON `triggers` (`env_id`);
-- create index "idx_triggers_connection_id" to table: "triggers"
CREATE INDEX `idx_triggers_connection_id` ON `triggers` (`connection_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_triggers_connection_id" to table: "triggers"
DROP INDEX `idx_triggers_connection_id`;
-- reverse: create index "idx_triggers_env_id" to table: "triggers"
DROP INDEX `idx_triggers_env_id`;
-- reverse: create index "idx_triggers_unique_name" to table: "triggers"
DROP INDEX `idx_triggers_unique_name`;
-- reverse: create index "idx_triggers_webhook_slug" to table: "triggers"
DROP INDEX `idx_triggers_webhook_slug`;
-- reverse: create index "idx_triggers_project_id" to table: "triggers"
DROP INDEX `idx_triggers_project_id`;
-- reverse: create index "idx_triggers_source_type" to table: "triggers"
DROP INDEX `idx_triggers_source_type`;
-- reverse: create "new_triggers" table
DROP TABLE `new_triggers`;
-- reverse: create index "idx_signals_destination_id" to table: "signals"
DROP INDEX `idx_signals_destination_id`;
-- reverse: create "new_signals" table
DROP TABLE `new_signals`;
-- reverse: create index "idx_events_deleted_at" to table: "events"
DROP INDEX `idx_events_deleted_at`;
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
-- reverse: create index "idx_event_type_seq" to table: "events"
DROP INDEX `idx_event_type_seq`;
-- reverse: create index "idx_event_type" to table: "events"
DROP INDEX `idx_event_type`;
-- reverse: create "new_events" table
DROP TABLE `new_events`;
