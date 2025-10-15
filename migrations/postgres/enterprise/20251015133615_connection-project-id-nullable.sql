-- +goose Up
-- modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "project_id" DROP NOT NULL, ALTER COLUMN "org_id" SET NOT NULL;

-- +goose Down
-- reverse: modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "org_id" DROP NOT NULL, ALTER COLUMN "project_id" SET NOT NULL;
