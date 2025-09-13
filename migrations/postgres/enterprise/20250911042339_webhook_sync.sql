-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "webhook_sync" boolean NULL, ADD COLUMN "webhook_response_timeout" bigint NULL;

-- +goose Down
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "webhook_response_timeout", DROP COLUMN "webhook_sync";
