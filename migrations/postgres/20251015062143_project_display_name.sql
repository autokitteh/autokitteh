-- +goose Up
-- modify "projects" table
ALTER TABLE "projects" ADD COLUMN "display_name" text NULL;

-- +goose Down
-- reverse: modify "projects" table
ALTER TABLE "projects" DROP COLUMN "display_name";
