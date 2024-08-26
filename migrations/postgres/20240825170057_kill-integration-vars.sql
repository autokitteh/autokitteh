-- +goose Up
-- modify "vars" table
ALTER TABLE "vars" DROP COLUMN "integration_id";

-- +goose Down
-- reverse: modify "vars" table
ALTER TABLE "vars" ADD COLUMN "integration_id" uuid NULL;
