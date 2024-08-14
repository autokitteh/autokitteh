-- +goose Up
-- modify "projects" table
ALTER TABLE "projects" ALTER COLUMN "name" SET NOT NULL;

-- +goose Down
-- reverse: modify "projects" table
ALTER TABLE "projects" ALTER COLUMN "name" DROP NOT NULL;
