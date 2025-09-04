-- +goose Up
-- add column "is_durable" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `is_durable` numeric NULL;

-- +goose Down
-- reverse: add column "is_durable" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `is_durable`;
