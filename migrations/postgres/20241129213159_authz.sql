-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "env_id" uuid NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;
-- create index "idx_triggers_env_id" to table: "triggers"
CREATE INDEX "idx_triggers_env_id" ON "triggers" ("env_id");
-- modify "connections" table
ALTER TABLE "connections" ALTER COLUMN "project_id" SET NOT NULL, ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;
-- modify "deployments" table
ALTER TABLE "deployments" ALTER COLUMN "project_id" SET NOT NULL, ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "updated_by" uuid NULL;
-- modify "builds" table
ALTER TABLE "builds" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "owner_id" uuid NOT NULL, ADD COLUMN "owner_user_id" uuid NOT NULL, ADD COLUMN "owner_org_id" uuid NOT NULL;
-- create index "idx_builds_owner_id" to table: "builds"
CREATE INDEX "idx_builds_owner_id" ON "builds" ("owner_id");
-- create index "idx_builds_owner_org_id" to table: "builds"
CREATE INDEX "idx_builds_owner_org_id" ON "builds" ("owner_org_id");
-- create index "idx_builds_owner_user_id" to table: "builds"
CREATE INDEX "idx_builds_owner_user_id" ON "builds" ("owner_user_id");
-- modify "projects" table
ALTER TABLE "projects" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "owner_id" uuid NOT NULL, ADD COLUMN "owner_user_id" uuid NOT NULL, ADD COLUMN "owner_org_id" uuid NOT NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;
-- create index "idx_projects_owner_id" to table: "projects"
CREATE INDEX "idx_projects_owner_id" ON "projects" ("owner_id");
-- create index "idx_projects_owner_org_id" to table: "projects"
CREATE INDEX "idx_projects_owner_org_id" ON "projects" ("owner_org_id");
-- create index "idx_projects_owner_user_id" to table: "projects"
CREATE INDEX "idx_projects_owner_user_id" ON "projects" ("owner_user_id");
-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "owner_id" uuid NOT NULL, ADD COLUMN "owner_user_id" uuid NOT NULL, ADD COLUMN "owner_org_id" uuid NOT NULL, ADD COLUMN "updated_by" uuid NULL;
-- create index "idx_sessions_owner_id" to table: "sessions"
CREATE INDEX "idx_sessions_owner_id" ON "sessions" ("owner_id");
-- create index "idx_sessions_owner_org_id" to table: "sessions"
CREATE INDEX "idx_sessions_owner_org_id" ON "sessions" ("owner_org_id");
-- create index "idx_sessions_owner_user_id" to table: "sessions"
CREATE INDEX "idx_sessions_owner_user_id" ON "sessions" ("owner_user_id");
-- create "orgs" table
CREATE TABLE "orgs" (
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "org_id" uuid NOT NULL,
  "name" text NOT NULL,
  "updated_by" uuid NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("org_id")
);
-- create index "idx_orgs_name" to table: "orgs"
CREATE UNIQUE INDEX "idx_orgs_name" ON "orgs" ("name");
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "disabled" boolean NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;
-- modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "values_pkey", ADD COLUMN "created_by" uuid NOT NULL, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "updated_by" uuid NULL, ADD PRIMARY KEY ("key");
-- create index "idx_values_project_id" to table: "values"
CREATE INDEX "idx_values_project_id" ON "values" ("project_id");
-- create "owneds" table
CREATE TABLE "owneds" (
  "owner_id" uuid NOT NULL,
  "owner_user_id" uuid NOT NULL,
  "owner_org_id" uuid NOT NULL
);
-- create index "idx_owneds_owner_id" to table: "owneds"
CREATE INDEX "idx_owneds_owner_id" ON "owneds" ("owner_id");
-- create index "idx_owneds_owner_org_id" to table: "owneds"
CREATE INDEX "idx_owneds_owner_org_id" ON "owneds" ("owner_org_id");
-- create index "idx_owneds_owner_user_id" to table: "owneds"
CREATE INDEX "idx_owneds_owner_user_id" ON "owneds" ("owner_user_id");
-- create "bases" table
CREATE TABLE "bases" (
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NULL
);
-- create "org_members" table
CREATE TABLE "org_members" (
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "org_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("org_id", "user_id")
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
-- reverse: create index "idx_events_project_id" to table: "events"
DROP INDEX "idx_events_project_id";
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_project", DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "project_id", DROP COLUMN "created_by";
-- reverse: create index "idx_belongs_to_projects_project_id" to table: "belongs_to_projects"
DROP INDEX "idx_belongs_to_projects_project_id";
-- reverse: create "belongs_to_projects" table
DROP TABLE "belongs_to_projects";
-- reverse: create "org_members" table
DROP TABLE "org_members";
-- reverse: create "bases" table
DROP TABLE "bases";
-- reverse: create index "idx_owneds_owner_user_id" to table: "owneds"
DROP INDEX "idx_owneds_owner_user_id";
-- reverse: create index "idx_owneds_owner_org_id" to table: "owneds"
DROP INDEX "idx_owneds_owner_org_id";
-- reverse: create index "idx_owneds_owner_id" to table: "owneds"
DROP INDEX "idx_owneds_owner_id";
-- reverse: create "owneds" table
DROP TABLE "owneds";
-- reverse: create index "idx_values_project_id" to table: "values"
DROP INDEX "idx_values_project_id";
-- reverse: modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "values_pkey", DROP COLUMN "updated_by", DROP COLUMN "created_at", DROP COLUMN "created_by", ADD PRIMARY KEY ("project_id", "key");
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "disabled", DROP COLUMN "created_at", DROP COLUMN "created_by";
-- reverse: create index "idx_orgs_name" to table: "orgs"
DROP INDEX "idx_orgs_name";
-- reverse: create "orgs" table
DROP TABLE "orgs";
-- reverse: create index "idx_sessions_owner_user_id" to table: "sessions"
DROP INDEX "idx_sessions_owner_user_id";
-- reverse: create index "idx_sessions_owner_org_id" to table: "sessions"
DROP INDEX "idx_sessions_owner_org_id";
-- reverse: create index "idx_sessions_owner_id" to table: "sessions"
DROP INDEX "idx_sessions_owner_id";
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "updated_by", DROP COLUMN "owner_org_id", DROP COLUMN "owner_user_id", DROP COLUMN "owner_id", DROP COLUMN "created_by";
-- reverse: create index "idx_projects_owner_user_id" to table: "projects"
DROP INDEX "idx_projects_owner_user_id";
-- reverse: create index "idx_projects_owner_org_id" to table: "projects"
DROP INDEX "idx_projects_owner_org_id";
-- reverse: create index "idx_projects_owner_id" to table: "projects"
DROP INDEX "idx_projects_owner_id";
-- reverse: modify "projects" table
ALTER TABLE "projects" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "owner_org_id", DROP COLUMN "owner_user_id", DROP COLUMN "owner_id", DROP COLUMN "created_at", DROP COLUMN "created_by";
-- reverse: create index "idx_builds_owner_user_id" to table: "builds"
DROP INDEX "idx_builds_owner_user_id";
-- reverse: create index "idx_builds_owner_org_id" to table: "builds"
DROP INDEX "idx_builds_owner_org_id";
-- reverse: create index "idx_builds_owner_id" to table: "builds"
DROP INDEX "idx_builds_owner_id";
-- reverse: modify "builds" table
ALTER TABLE "builds" DROP COLUMN "owner_org_id", DROP COLUMN "owner_user_id", DROP COLUMN "owner_id", DROP COLUMN "created_by";
-- reverse: modify "deployments" table
ALTER TABLE "deployments" DROP COLUMN "updated_by", DROP COLUMN "created_by", ALTER COLUMN "project_id" DROP NOT NULL;
-- reverse: modify "connections" table
ALTER TABLE "connections" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "created_at", DROP COLUMN "created_by", ALTER COLUMN "project_id" DROP NOT NULL;
-- reverse: create index "idx_triggers_env_id" to table: "triggers"
DROP INDEX "idx_triggers_env_id";
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "env_id", DROP COLUMN "created_at", DROP COLUMN "created_by";
