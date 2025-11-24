-- +goose Up
-- drop index "idx_connections_org_id" from table: "connections"
DROP INDEX "idx_connections_org_id";
-- modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "name" SET NOT NULL;
-- create index "idx_connection_org_id_name" to table: "connections"
CREATE UNIQUE INDEX "idx_connection_org_id_name" ON "connections" ("org_id", "name") WHERE ((project_id IS NULL) AND (deleted_at IS NULL));
-- create index "idx_connection_org_id_project_id_name" to table: "connections"
CREATE UNIQUE INDEX "idx_connection_org_id_project_id_name" ON "connections" ("org_id", "project_id", "name") WHERE ((project_id IS NOT NULL) AND (deleted_at IS NULL));

-- +goose Down
-- reverse: create index "idx_connection_org_id_project_id_name" to table: "connections"
DROP INDEX "idx_connection_org_id_project_id_name";
-- reverse: create index "idx_connection_org_id_name" to table: "connections"
DROP INDEX "idx_connection_org_id_name";
-- reverse: modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "name" DROP NOT NULL;
-- reverse: drop index "idx_connections_org_id" from table: "connections"
CREATE INDEX "idx_connections_org_id" ON "connections" ("org_id");
