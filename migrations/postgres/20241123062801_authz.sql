-- +goose Up
-- modify "builds" table
ALTER TABLE "builds" ADD COLUMN "owner_user_id" uuid NOT NULL;
-- create index "idx_builds_owner_user_id" to table: "builds"
CREATE INDEX "idx_builds_owner_user_id" ON "builds" ("owner_user_id");
-- modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "project_id" SET NOT NULL;
-- modify "deployments" table
ALTER TABLE "deployments" ALTER COLUMN "project_id" SET NOT NULL;
-- modify "projects" table
ALTER TABLE "projects" ADD COLUMN "owner_user_id" uuid NOT NULL;
-- create index "idx_projects_owner_user_id" to table: "projects"
CREATE INDEX "idx_projects_owner_user_id" ON "projects" ("owner_user_id");
-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "owner_user_id" uuid NOT NULL;
-- create index "idx_sessions_owner_user_id" to table: "sessions"
CREATE INDEX "idx_sessions_owner_user_id" ON "sessions" ("owner_user_id");
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "env_id" uuid NULL;
-- create index "idx_triggers_env_id" to table: "triggers"
CREATE INDEX "idx_triggers_env_id" ON "triggers" ("env_id");
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "disabled" boolean NULL;
-- modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "values_pkey", ADD PRIMARY KEY ("key");
-- create index "idx_values_project_id" to table: "values"
CREATE INDEX "idx_values_project_id" ON "values" ("project_id");
-- create "owneds" table
CREATE TABLE "owneds" (
  "owner_user_id" uuid NOT NULL
);
-- create index "idx_owneds_owner_user_id" to table: "owneds"
CREATE INDEX "idx_owneds_owner_user_id" ON "owneds" ("owner_user_id");
-- create "belongs_to_projects" table
CREATE TABLE "belongs_to_projects" (
  "project_id" uuid NOT NULL,
  CONSTRAINT "fk_belongs_to_projects_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
CREATE INDEX "idx_belongs_to_projects_project_id" ON "belongs_to_projects" ("project_id");
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "project_id" uuid NOT NULL, ADD
 CONSTRAINT "fk_events_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "idx_events_project_id" to table: "events"
CREATE INDEX "idx_events_project_id" ON "events" ("project_id");
-- modify "vars" table
ALTER TABLE "vars" ADD COLUMN "project_id" uuid NOT NULL, ADD
 CONSTRAINT "fk_vars_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "idx_vars_project_id" to table: "vars"
CREATE INDEX "idx_vars_project_id" ON "vars" ("project_id");

-- +goose Down
-- reverse: create index "idx_vars_project_id" to table: "vars"
DROP INDEX "idx_vars_project_id";
-- reverse: modify "vars" table
ALTER TABLE "vars" DROP CONSTRAINT "fk_vars_project", DROP COLUMN "project_id";
-- reverse: create index "idx_events_project_id" to table: "events"
DROP INDEX "idx_events_project_id";
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_project", DROP COLUMN "project_id";
-- reverse: create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
DROP INDEX "idx_belongs_to_projects_project_id";
-- reverse: create "belongs_to_projects" table
DROP TABLE "belongs_to_projects";
-- reverse: create index "idx_owneds_owner_user_id" to table: "owneds"
DROP INDEX "idx_owneds_owner_user_id";
-- reverse: create "owneds" table
DROP TABLE "owneds";
-- reverse: create index "idx_values_project_id" to table: "values"
DROP INDEX "idx_values_project_id";
-- reverse: modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "values_pkey", ADD PRIMARY KEY ("project_id", "key");
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "disabled";
-- reverse: create index "idx_triggers_env_id" to table: "triggers"
DROP INDEX "idx_triggers_env_id";
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "env_id";
-- reverse: create index "idx_sessions_owner_user_id" to table: "sessions"
DROP INDEX "idx_sessions_owner_user_id";
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "owner_user_id";
-- reverse: create index "idx_projects_owner_user_id" to table: "projects"
DROP INDEX "idx_projects_owner_user_id";
-- reverse: modify "projects" table
ALTER TABLE "projects" DROP COLUMN "owner_user_id";
-- reverse: modify "deployments" table
ALTER TABLE "deployments" ALTER COLUMN "project_id" DROP NOT NULL;
-- reverse: modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "project_id" DROP NOT NULL;
-- reverse: create index "idx_builds_owner_user_id" to table: "builds"
DROP INDEX "idx_builds_owner_user_id";
-- reverse: modify "builds" table
ALTER TABLE "builds" DROP COLUMN "owner_user_id";
