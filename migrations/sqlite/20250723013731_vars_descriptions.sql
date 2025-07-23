-- +goose Up
-- add column "description" to table: "vars"
ALTER TABLE `vars` ADD COLUMN `description` text NULL;

-- +goose Down
-- reverse: add column "description" to table: "vars"
ALTER TABLE `vars` DROP COLUMN `description`;
