-- +goose Up
-- add column "display_name" to table: "projects"
ALTER TABLE `projects` ADD COLUMN `display_name` text NULL;

-- +goose Down
-- reverse: add column "display_name" to table: "projects"
ALTER TABLE `projects` DROP COLUMN `display_name`;
