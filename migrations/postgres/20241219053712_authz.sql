-- +goose Up
-- modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "project_id" SET NOT NULL, ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;
-- modify "deployments" table
ALTER TABLE "deployments" ALTER COLUMN "project_id" SET NOT NULL, ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "updated_by" uuid NULL;
-- modify "builds" table
ALTER TABLE "builds" ALTER COLUMN "project_id" SET NOT NULL, ADD COLUMN "created_by" uuid NOT NULL;
-- modify "projects" table
ALTER TABLE "projects" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "org_id" uuid NOT NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;
-- create index "idx_projects_org_id" to table: "projects"
CREATE INDEX "idx_projects_org_id" ON "projects" ("org_id");
-- create "orgs" table
CREATE TABLE "orgs" (
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "org_id" uuid NOT NULL,
  "name" text NOT NULL,
  "display_name" text NULL,
  "updated_by" uuid NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("org_id")
);
-- create index "idx_orgs_name" to table: "orgs"
CREATE UNIQUE INDEX "idx_orgs_name" ON "orgs" ("name");
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "name" text NOT NULL, ADD COLUMN "disabled" boolean NULL, ADD COLUMN "default_org_id" uuid NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;
-- create index "idx_users_name" to table: "users"
CREATE UNIQUE INDEX "idx_users_name" ON "users" ("name");
-- modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "values_pkey", ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "updated_by" uuid NULL, ADD PRIMARY KEY ("key");
-- create index "idx_values_project_id" to table: "values"
CREATE INDEX "idx_values_project_id" ON "values" ("project_id");
-- create "org_members" table
CREATE TABLE "org_members" (
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "org_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("org_id", "user_id")
);
-- create "bases" table
CREATE TABLE "bases" (
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NULL
);
-- create "belongs_to_projects" table
CREATE TABLE "belongs_to_projects" (
  "project_id" uuid NOT NULL,
  CONSTRAINT "fk_belongs_to_projects_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
CREATE INDEX "idx_belongs_to_projects_project_id" ON "belongs_to_projects" ("project_id");
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "project_id" uuid NOT NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL, ADD
 CONSTRAINT "fk_events_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "idx_events_project_id" to table: "events"
CREATE INDEX "idx_events_project_id" ON "events" ("project_id");
-- modify "sessions" table
ALTER TABLE "sessions" ALTER COLUMN "build_id" SET NOT NULL, ALTER COLUMN "project_id" SET NOT NULL, ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "updated_by" uuid NULL, ADD
 CONSTRAINT "fk_sessions_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "vars" table
ALTER TABLE "vars" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "project_id" uuid NOT NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL, ADD
 CONSTRAINT "fk_vars_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "idx_vars_project_id" to table: "vars"
CREATE INDEX "idx_vars_project_id" ON "vars" ("project_id");

-- +goose Down
-- reverse: create index "idx_vars_project_id" to table: "vars"
DROP INDEX "idx_vars_project_id";
-- reverse: modify "vars" table
ALTER TABLE "vars" DROP CONSTRAINT "fk_vars_project", DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "project_id", DROP COLUMN "created_at", DROP COLUMN "created_by";
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_project", DROP COLUMN "updated_by", DROP COLUMN "created_by", ALTER COLUMN "project_id" DROP NOT NULL, ALTER COLUMN "build_id" DROP NOT NULL;
-- reverse: create index "idx_events_project_id" to table: "events"
DROP INDEX "idx_events_project_id";
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_project", DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "project_id", DROP COLUMN "created_by";
-- reverse: create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
DROP INDEX "idx_belongs_to_projects_project_id";
-- reverse: create "belongs_to_projects" table
DROP TABLE "belongs_to_projects";
-- reverse: create "bases" table
DROP TABLE "bases";
-- reverse: create "org_members" table
DROP TABLE "org_members";
-- reverse: create index "idx_values_project_id" to table: "values"
DROP INDEX "idx_values_project_id";
-- reverse: modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "values_pkey", DROP COLUMN "updated_by", DROP COLUMN "created_at", DROP COLUMN "created_by", ADD PRIMARY KEY ("project_id", "key");
-- reverse: create index "idx_users_name" to table: "users"
DROP INDEX "idx_users_name";
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "default_org_id", DROP COLUMN "disabled", DROP COLUMN "name", DROP COLUMN "created_at", DROP COLUMN "created_by";
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "created_at", DROP COLUMN "created_by";
-- reverse: create index "idx_orgs_name" to table: "orgs"
DROP INDEX "idx_orgs_name";
-- reverse: create "orgs" table
DROP TABLE "orgs";
-- reverse: create index "idx_projects_org_id" to table: "projects"
DROP INDEX "idx_projects_org_id";
-- reverse: modify "projects" table
ALTER TABLE "projects" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "org_id", DROP COLUMN "created_at", DROP COLUMN "created_by";
-- reverse: modify "builds" table
ALTER TABLE "builds" DROP COLUMN "created_by", ALTER COLUMN "project_id" DROP NOT NULL;
-- reverse: modify "deployments" table
ALTER TABLE "deployments" DROP COLUMN "updated_by", DROP COLUMN "created_by", ALTER COLUMN "project_id" DROP NOT NULL;
-- reverse: modify "connections" table
ALTER TABLE "connections" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "created_at", DROP COLUMN "created_by", ALTER COLUMN "project_id" DROP NOT NULL;
