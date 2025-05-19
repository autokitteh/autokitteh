-- +goose Up
-- modify "deployments" table
ALTER TABLE "deployments" DROP COLUMN "env_id";
-- modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "env_id";
-- modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "env_id";
-- drop "envs" table
DROP TABLE "envs";

-- +goose Down
-- reverse: drop "envs" table
CREATE TABLE "envs" (
  "env_id" uuid NOT NULL,
  "project_id" uuid NOT NULL,
  "name" text NULL,
  "deleted_at" timestamptz NULL,
  "membership_id" text NULL,
  PRIMARY KEY ("env_id"),
  CONSTRAINT "fk_envs_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "idx_envs_deleted_at" ON "envs" ("deleted_at");
CREATE UNIQUE INDEX "idx_envs_membership_id" ON "envs" ("membership_id");
CREATE INDEX "idx_envs_project_id" ON "envs" ("project_id");
-- reverse: modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "env_id" uuid NULL;
-- reverse: modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "env_id" uuid NULL;
-- reverse: modify "deployments" table
ALTER TABLE "deployments" ADD COLUMN "env_id" uuid NULL;
