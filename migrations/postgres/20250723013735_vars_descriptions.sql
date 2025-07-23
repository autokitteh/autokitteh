-- +goose Up
-- modify "vars" table
ALTER TABLE "vars" ADD COLUMN "description" text NULL;

-- +goose Down
-- reverse: modify "vars" table
ALTER TABLE "vars" DROP COLUMN "description";
