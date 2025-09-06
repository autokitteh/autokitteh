-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "is_durable" boolean NULL;

UPDATE "triggers" SET is_durable=TRUE;

-- +goose Down
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "is_durable";
