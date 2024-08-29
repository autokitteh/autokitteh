-- +goose Up
-- add column "links" to table: "connections"
ALTER TABLE `connections` ADD COLUMN `links` json NULL;

-- +goose Down
-- reverse: add column "links" to table: "connections"
ALTER TABLE `connections` DROP COLUMN `links`;
