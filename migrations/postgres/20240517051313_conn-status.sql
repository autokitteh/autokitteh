-- +goose Up
-- modify "connections" table
ALTER TABLE "connections" ADD COLUMN "status_code" integer NULL, ADD COLUMN "status_message" text NULL;
-- create index "idx_connections_integration_id" to table: "connections"
CREATE INDEX "idx_connections_integration_id" ON "connections" ("integration_id");
-- create index "idx_connections_status_code" to table: "connections"
CREATE INDEX "idx_connections_status_code" ON "connections" ("status_code");

-- +goose Down
-- reverse: create index "idx_connections_status_code" to table: "connections"
DROP INDEX "idx_connections_status_code";
-- reverse: create index "idx_connections_integration_id" to table: "connections"
DROP INDEX "idx_connections_integration_id";
-- reverse: modify "connections" table
ALTER TABLE "connections" DROP COLUMN "status_message", DROP COLUMN "status_code";
