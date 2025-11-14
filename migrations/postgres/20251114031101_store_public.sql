-- +goose Up
-- modify "store_values" table
ALTER TABLE "store_values" ADD COLUMN "published" boolean NULL;

-- +goose Down
-- reverse: modify "store_values" table
ALTER TABLE "store_values" DROP COLUMN "published";
