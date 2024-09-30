-- +goose Up
-- add column "memo" to table: "sessions"
ALTER TABLE `sessions` ADD COLUMN `memo` json NULL;

-- +goose Down
-- reverse: add column "memo" to table: "sessions"
ALTER TABLE `sessions` DROP COLUMN `memo`;
