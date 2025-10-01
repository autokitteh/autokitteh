-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
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
  `is_durable` numeric NULL,
  `is_sync` numeric NULL,
  `name` text NULL,
  `unique_name` text NOT NULL,
  `webhook_slug` text NULL,
  `schedule` text NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`trigger_id`),
  CONSTRAINT `fk_triggers_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "triggers" to new temporary table "new_triggers"
INSERT INTO `new_triggers` (`created_by`, `created_at`, `project_id`, `trigger_id`, `connection_id`, `source_type`, `event_type`, `filter`, `code_location`, `is_durable`, `is_sync`, `name`, `unique_name`, `webhook_slug`, `schedule`, `updated_by`, `updated_at`, `deleted_at`) SELECT `created_by`, `created_at`, `project_id`, `trigger_id`, `connection_id`, `source_type`, `event_type`, `filter`, `code_location`, `is_durable`, `is_sync`, `name`, `unique_name`, `webhook_slug`, `schedule`, `updated_by`, `updated_at`, `deleted_at` FROM `triggers`;
-- drop "triggers" table after copying rows
DROP TABLE `triggers`;
-- rename temporary table "new_triggers" to "triggers"
ALTER TABLE `new_triggers` RENAME TO `triggers`;
-- create index "idx_triggers_deleted_at" to table: "triggers"
CREATE INDEX `idx_triggers_deleted_at` ON `triggers` (`deleted_at`);
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
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
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
-- reverse: create index "idx_triggers_deleted_at" to table: "triggers"
DROP INDEX `idx_triggers_deleted_at`;
-- reverse: create "new_triggers" table
DROP TABLE `new_triggers`;
