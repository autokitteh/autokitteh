-- +goose Up
-- add column "type" to table: "session_log_records"
ALTER TABLE `session_log_records` ADD COLUMN `type` text NULL;
-- create index "idx_session_log_records_type" to table: "session_log_records"
CREATE INDEX `idx_session_log_records_type` ON `session_log_records` (`type`);

-- +goose Down
-- reverse: create index "idx_session_log_records_type" to table: "session_log_records"
DROP INDEX `idx_session_log_records_type`;
-- reverse: add column "type" to table: "session_log_records"
ALTER TABLE `session_log_records` DROP COLUMN `type`;
