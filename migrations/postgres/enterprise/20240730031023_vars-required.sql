-- +goose Up
-- modify "vars" table
ALTER TABLE "vars" ADD COLUMN "is_required" boolean NULL;

-- +goose Down
-- reverse: modify "vars" table
ALTER TABLE "vars" DROP COLUMN "is_required";
