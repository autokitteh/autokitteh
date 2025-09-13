-- +goose Up
-- add column "is_durable" to table: "sessions"
ALTER TABLE `sessions` ADD COLUMN `is_durable` numeric NULL;

UPDATE `triggers` SET is_durable=TRUE;

-- +goose Down
-- reverse: add column "is_durable" to table: "sessions"
ALTER TABLE `sessions` DROP COLUMN `is_durable`;
