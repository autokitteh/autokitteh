-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "is_durable" boolean NULL;

-- +goose Down
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "is_durable";
