-- +goose Up
-- add column "webhook_sync" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `webhook_sync` numeric NULL;
-- add column "webhook_response_timeout" to table: "triggers"
ALTER TABLE `triggers` ADD COLUMN `webhook_response_timeout` integer NULL;

-- +goose Down
-- reverse: add column "webhook_response_timeout" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `webhook_response_timeout`;
-- reverse: add column "webhook_sync" to table: "triggers"
ALTER TABLE `triggers` DROP COLUMN `webhook_sync`;
