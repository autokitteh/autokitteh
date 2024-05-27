-- +goose Up
-- add column "compressed_inputs" to table: "sessions"
ALTER TABLE `sessions` ADD COLUMN `compressed_inputs` blob NULL;
-- add column "compressed_complete" to table: "session_call_attempts"
ALTER TABLE `session_call_attempts` ADD COLUMN `compressed_complete` blob NULL;
-- add column "compressed_data" to table: "session_call_specs"
ALTER TABLE `session_call_specs` ADD COLUMN `compressed_data` blob NULL;
-- add column "compressed_data" to table: "events"
ALTER TABLE `events` ADD COLUMN `compressed_data` blob NULL;

-- +goose Down
-- reverse: add column "compressed_data" to table: "events"
ALTER TABLE `events` DROP COLUMN `compressed_data`;
-- reverse: add column "compressed_data" to table: "session_call_specs"
ALTER TABLE `session_call_specs` DROP COLUMN `compressed_data`;
-- reverse: add column "compressed_complete" to table: "session_call_attempts"
ALTER TABLE `session_call_attempts` DROP COLUMN `compressed_complete`;
-- reverse: add column "compressed_inputs" to table: "sessions"
ALTER TABLE `sessions` DROP COLUMN `compressed_inputs`;
