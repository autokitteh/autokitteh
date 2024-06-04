-- +goose Up
-- modify "envs" table
ALTER TABLE "envs" ALTER COLUMN "project_id" SET NOT NULL;

-- +goose Down
-- reverse: modify "envs" table
ALTER TABLE "envs" ALTER COLUMN "project_id" DROP NOT NULL;
