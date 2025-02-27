-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_events" table
CREATE TABLE `new_events` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NULL,
  `org_id` uuid NULL,
  `event_id` uuid NOT NULL,
  `destination_id` uuid NOT NULL,
  `integration_id` uuid NULL,
  `connection_id` uuid NULL,
  `trigger_id` uuid NULL,
  `event_type` text NULL,
  `data` json NULL,
  `memo` json NULL,
  `seq` integer NULL PRIMARY KEY AUTOINCREMENT,
  CONSTRAINT `fk_events_org` FOREIGN KEY (`org_id`) REFERENCES `orgs` (`org_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_events_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_events_trigger` FOREIGN KEY (`trigger_id`) REFERENCES `triggers` (`trigger_id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `fk_events_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- copy rows from old table "events" to new temporary table "new_events"
INSERT INTO `new_events` (`created_by`, `created_at`, `project_id`, `event_id`, `destination_id`, `integration_id`, `connection_id`, `trigger_id`, `event_type`, `data`, `memo`, `seq`) SELECT `created_by`, `created_at`, `project_id`, `event_id`, `destination_id`, `integration_id`, `connection_id`, `trigger_id`, `event_type`, `data`, `memo`, `seq` FROM `events`;
-- drop "events" table after copying rows
DROP TABLE `events`;
-- rename temporary table "new_events" to "events"
ALTER TABLE `new_events` RENAME TO `events`;
-- create index "idx_events_event_id" to table: "events"
CREATE UNIQUE INDEX `idx_events_event_id` ON `events` (`event_id`);
-- create index "idx_events_created_at" to table: "events"
CREATE INDEX `idx_events_created_at` ON `events` (`created_at`);
-- create index "idx_event_type" to table: "events"
CREATE INDEX `idx_event_type` ON `events` (`event_type`);
-- create index "idx_events_connection_id" to table: "events"
CREATE INDEX `idx_events_connection_id` ON `events` (`connection_id`);
-- create index "idx_events_destination_id" to table: "events"
CREATE INDEX `idx_events_destination_id` ON `events` (`destination_id`);
-- create index "idx_events_project_id" to table: "events"
CREATE INDEX `idx_events_project_id` ON `events` (`project_id`);
-- create index "idx_event_type_seq" to table: "events"
CREATE INDEX `idx_event_type_seq` ON `events` (`event_type`, `seq`);
-- create index "idx_events_trigger_id" to table: "events"
CREATE INDEX `idx_events_trigger_id` ON `events` (`trigger_id`);
-- create index "idx_events_integration_id" to table: "events"
CREATE INDEX `idx_events_integration_id` ON `events` (`integration_id`);
-- create index "idx_org_id_seq" to table: "events"
CREATE INDEX `idx_org_id_seq` ON `events` (`org_id`, `seq`);
-- create index "idx_events_org_id" to table: "events"
CREATE INDEX `idx_events_org_id` ON `events` (`org_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_events_org_id" to table: "events"
DROP INDEX `idx_events_org_id`;
-- reverse: create index "idx_org_id_seq" to table: "events"
DROP INDEX `idx_org_id_seq`;
-- reverse: create index "idx_events_integration_id" to table: "events"
DROP INDEX `idx_events_integration_id`;
-- reverse: create index "idx_events_trigger_id" to table: "events"
DROP INDEX `idx_events_trigger_id`;
-- reverse: create index "idx_event_type_seq" to table: "events"
DROP INDEX `idx_event_type_seq`;
-- reverse: create index "idx_events_project_id" to table: "events"
DROP INDEX `idx_events_project_id`;
-- reverse: create index "idx_events_destination_id" to table: "events"
DROP INDEX `idx_events_destination_id`;
-- reverse: create index "idx_events_connection_id" to table: "events"
DROP INDEX `idx_events_connection_id`;
-- reverse: create index "idx_event_type" to table: "events"
DROP INDEX `idx_event_type`;
-- reverse: create index "idx_events_created_at" to table: "events"
DROP INDEX `idx_events_created_at`;
-- reverse: create index "idx_events_event_id" to table: "events"
DROP INDEX `idx_events_event_id`;
-- reverse: create "new_events" table
DROP TABLE `new_events`;
