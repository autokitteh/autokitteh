-- +goose Up
-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "is_durable" boolean NULL;

UPDATE "triggers" SET is_durable=TRUE;

-- +goose Down
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "is_durable";
