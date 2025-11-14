-- +goose Up
-- add column "published" to table: "store_values"
ALTER TABLE `store_values` ADD COLUMN `published` numeric NULL;

-- +goose Down
-- reverse: add column "published" to table: "store_values"
ALTER TABLE `store_values` DROP COLUMN `published`;
