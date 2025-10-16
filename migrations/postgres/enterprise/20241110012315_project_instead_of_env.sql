-- +goose Up
-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "project_id" uuid NULL;
-- create index "idx_sessions_project_id" to table: "sessions"
CREATE INDEX "idx_sessions_project_id" ON "sessions" ("project_id");
-- modify "triggers" table
ALTER TABLE "triggers" ALTER COLUMN "env_id" DROP NOT NULL;
-- modify "deployments" table
ALTER TABLE "deployments" ADD COLUMN "project_id" uuid NULL, ADD
 CONSTRAINT "fk_deployments_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "idx_deployments_project_id" to table: "deployments"
CREATE INDEX "idx_deployments_project_id" ON "deployments" ("project_id");

-- +goose Down
-- reverse: create index "idx_deployments_project_id" to table: "deployments"
DROP INDEX "idx_deployments_project_id";
-- reverse: modify "deployments" table
ALTER TABLE "deployments" DROP CONSTRAINT "fk_deployments_project", DROP COLUMN "project_id";
-- reverse: modify "triggers" table
ALTER TABLE "triggers" ALTER COLUMN "env_id" SET NOT NULL;
-- reverse: create index "idx_sessions_project_id" to table: "sessions"
DROP INDEX "idx_sessions_project_id";
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "project_id";
