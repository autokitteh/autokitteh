-- +goose Up
-- modify "deployments" table
ALTER TABLE "deployments" ALTER COLUMN "build_id" SET NOT NULL;

-- +goose Down
-- reverse: modify "deployments" table
ALTER TABLE "deployments" ALTER COLUMN "build_id" DROP NOT NULL;
