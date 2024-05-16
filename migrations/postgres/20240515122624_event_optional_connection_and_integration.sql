-- +goose Up
-- modify "events" table
ALTER TABLE "events" ALTER COLUMN "integration_id" DROP NOT NULL, ALTER COLUMN "connection_id" DROP NOT NULL;

-- +goose Down
-- reverse: modify "events" table
ALTER TABLE "events" ALTER COLUMN "connection_id" SET NOT NULL, ALTER COLUMN "integration_id" SET NOT NULL;
