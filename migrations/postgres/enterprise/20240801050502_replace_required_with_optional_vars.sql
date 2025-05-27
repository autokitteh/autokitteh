-- +goose Up
-- modify "vars" table
ALTER TABLE "vars" DROP COLUMN "is_required", ADD COLUMN "is_optional" boolean NULL;

-- +goose Down
-- reverse: modify "vars" table
ALTER TABLE "vars" DROP COLUMN "is_optional", ADD COLUMN "is_required" boolean NULL;
