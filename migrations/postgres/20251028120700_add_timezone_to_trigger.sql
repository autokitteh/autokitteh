-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "timezone" text NULL;

-- +goose Down
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "timezone";
