-- +goose Up
-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "is_durable" boolean NULL;

-- +goose Down
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "is_durable";
