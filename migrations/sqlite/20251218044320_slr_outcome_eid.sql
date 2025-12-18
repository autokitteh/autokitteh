-- +goose Up
-- add column "outcome_event_id" to table: "session_log_records"
ALTER TABLE `session_log_records` ADD COLUMN `outcome_event_id` uuid NULL;
-- create index "idx_outcome_event_id" to table: "session_log_records"
CREATE INDEX `idx_outcome_event_id` ON `session_log_records` (`outcome_event_id`) WHERE outcome_event_id is not null;

-- +goose Down
-- reverse: create index "idx_outcome_event_id" to table: "session_log_records"
DROP INDEX `idx_outcome_event_id`;
-- reverse: add column "outcome_event_id" to table: "session_log_records"
ALTER TABLE `session_log_records` DROP COLUMN `outcome_event_id`;
