-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "is_sync" boolean NULL;

-- +goose Down
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "is_sync";
