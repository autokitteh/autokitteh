-- +goose Up
-- modify "connections" table'
ALTER TABLE "connections" ADD COLUMN "org_id" uuid NULL, ALTER COLUMN "name" SET NOT NULL, ALTER COLUMN "project_id" DROP NOT NULL;

-- Backfill org id to connections
UPDATE connections c
SET org_id = o.org_id
FROM projects p
JOIN orgs o USING(org_id)
WHERE c.project_id = p.project_id AND p.org_id IS NOT NULL;

-- Set org id not null 
ALTER TABLE "connections" ALTER COLUMN "org_id" SET NOT NULL;

-- create index "idx_connection_org_id_name" to table: "connections"
CREATE UNIQUE INDEX "idx_connection_org_id_name" ON "connections" ("org_id", "name") WHERE ((project_id IS NULL) AND (deleted_at IS NULL));
-- create index "idx_connection_org_id_project_id_name" to table: "connections"
CREATE UNIQUE INDEX "idx_connection_org_id_project_id_name" ON "connections" ("org_id", "project_id", "name") WHERE ((project_id IS NOT NULL) AND (deleted_at IS NULL));


-- +goose Down
-- reverse: create index "idx_connections_org_id" to table: "connections"
DROP INDEX "idx_connection_org_id_name";
DROP INDEX "idx_connection_org_id_project_id_name";

-- reverse: modify "connections" table
ALTER TABLE "connections" DROP COLUMN "org_id", ALTER COLUMN "name" DROP NOT NULL, ALTER COLUMN "project_id" SET NOT NULL;
