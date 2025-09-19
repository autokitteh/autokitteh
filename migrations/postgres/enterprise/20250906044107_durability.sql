-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "is_durable" boolean NULL DEFAULT false;
UPDATE "triggers" SET is_durable = true;
ALTER TABLE "triggers" ALTER COLUMN "is_durable" SET NOT NULL;

-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "is_durable" boolean NULL DEFAULT false;
UPDATE "sessions" SET is_durable = true;
ALTER TABLE "sessions" ALTER COLUMN "is_durable" SET NOT NULL;

-- +goose Down
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "is_durable";

-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "is_durable";
