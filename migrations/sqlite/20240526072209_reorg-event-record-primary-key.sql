-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_event_records" table
CREATE TABLE `new_event_records` (
  `event_id` uuid NOT NULL,
  `seq` integer NULL,
  `state` integer NULL,
  `created_at` datetime NULL,
  PRIMARY KEY (`event_id`, `seq`),
  CONSTRAINT `fk_event_records_event` FOREIGN KEY (`event_id`) REFERENCES `events` (`event_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "event_records" to new temporary table "new_event_records"
INSERT INTO `new_event_records` (`event_id`, `seq`, `state`, `created_at`) SELECT `event_id`, `seq`, `state`, `created_at` FROM `event_records`;
-- drop "event_records" table after copying rows
DROP TABLE `event_records`;
-- rename temporary table "new_event_records" to "event_records"
ALTER TABLE `new_event_records` RENAME TO `event_records`;
-- create index "idx_event_records_state" to table: "event_records"
CREATE INDEX `idx_event_records_state` ON `event_records` (`state`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_event_records_state" to table: "event_records"
DROP INDEX `idx_event_records_state`;
-- reverse: create "new_event_records" table
DROP TABLE `new_event_records`;
