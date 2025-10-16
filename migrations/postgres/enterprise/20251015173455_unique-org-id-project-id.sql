-- +goose Up
-- drop index "idx_connections_org_id" from table: "connections"
DROP INDEX "idx_connections_org_id";
-- create index "idx_connection_org_id_project_id" to table: "connections"
CREATE UNIQUE INDEX "idx_connection_org_id_project_id" ON "connections" ("org_id", "project_id");

-- +goose Down
-- reverse: create index "idx_connection_org_id_project_id" to table: "connections"
DROP INDEX "idx_connection_org_id_project_id";
-- reverse: drop index "idx_connections_org_id" from table: "connections"
CREATE INDEX "idx_connections_org_id" ON "connections" ("org_id");
