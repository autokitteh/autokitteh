-- +goose Up
-- add column "timezone" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `timezone` text NULL;

-- +goose Down
-- reverse: add column "timezone" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `timezone`;
