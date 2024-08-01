-- +goose Up
-- add column "is_required" to table: "vars"
ALTER TABLE `vars` ADD COLUMN `is_required` numeric NULL;

-- +goose Down
-- reverse: add column "is_required" to table: "vars"
ALTER TABLE `vars` DROP COLUMN `is_required`;
