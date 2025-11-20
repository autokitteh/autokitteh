-- +goose Up
-- modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "project_id" DROP NOT NULL, ADD COLUMN "org_id" uuid NOT NULL;
-- create index "idx_connection_org_id_project_id" to table: "connections"
CREATE UNIQUE INDEX "idx_connection_org_id_project_id" ON "connections" ("org_id", "project_id");

-- +goose Down
-- reverse: create index "idx_connection_org_id_project_id" to table: "connections"
DROP INDEX "idx_connection_org_id_project_id";
-- reverse: modify "connections" table
ALTER TABLE "connections" DROP COLUMN "org_id", ALTER COLUMN "project_id" SET NOT NULL;
