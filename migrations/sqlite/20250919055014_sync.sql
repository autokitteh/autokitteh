-- +goose Up
-- add column "sync" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `is_sync` numeric NULL;

-- +goose Down
-- reverse: add column "sync" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `is_sync`;
