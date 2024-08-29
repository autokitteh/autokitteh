-- +goose Up
-- modify "connections" table
ALTER TABLE "connections" ADD COLUMN "links" jsonb NULL;

-- +goose Down
-- reverse: modify "connections" table
ALTER TABLE "connections" DROP COLUMN "links";
