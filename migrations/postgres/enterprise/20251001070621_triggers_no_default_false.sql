-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ALTER COLUMN "is_durable" DROP NOT NULL, ALTER COLUMN "is_durable" DROP DEFAULT;

-- +goose Down
-- reverse: modify "triggers" table
ALTER TABLE "triggers" ALTER COLUMN "is_durable" SET NOT NULL, ALTER COLUMN "is_durable" SET DEFAULT false;
