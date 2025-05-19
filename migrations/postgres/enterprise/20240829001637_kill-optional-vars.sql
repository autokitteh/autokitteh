-- +goose Up
-- modify "vars" table
ALTER TABLE "vars" DROP COLUMN "is_optional";

-- +goose Down
-- reverse: modify "vars" table
ALTER TABLE "vars" ADD COLUMN "is_optional" boolean NULL;
