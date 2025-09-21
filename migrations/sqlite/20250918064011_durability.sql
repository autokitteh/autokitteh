-- +goose Up
-- add column "is_durable" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `is_durable` numeric NOT NULL DEFAULT false;
-- add column "is_durable" to table: "sessions"
ALTER TABLE `sessions` ADD COLUMN `is_durable` numeric NOT NULL DEFAULT false;

-- +goose Down
-- reverse: add column "is_durable" to table: "sessions"
ALTER TABLE `sessions` DROP COLUMN `is_durable`;
-- reverse: add column "is_durable" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `is_durable`;
