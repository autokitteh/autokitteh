-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- drop "event_records" table
DROP TABLE `event_records`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: drop "event_records" table
CREATE TABLE `event_records` (
  `event_id` uuid NOT NULL,
  `seq` integer NULL,
  `state` integer NULL,
  `created_at` datetime NULL,
  PRIMARY KEY (`event_id`, `seq`),
  CONSTRAINT `fk_event_records_event` FOREIGN KEY (`event_id`) REFERENCES `events` (`event_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX `idx_event_records_state` ON `event_records` (`state`);
