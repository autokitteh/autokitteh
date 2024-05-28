-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ALTER COLUMN "connection_id" SET NOT NULL;

-- +goose Down
-- reverse: modify "triggers" table
ALTER TABLE "triggers" ALTER COLUMN "connection_id" DROP NOT NULL;
