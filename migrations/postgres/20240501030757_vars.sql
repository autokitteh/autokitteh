-- +goose Up
-- create "vars" table
CREATE TABLE "vars" (
  "scope_id" uuid NOT NULL,
  "name" text NOT NULL,
  "value" text NULL,
  "is_secret" boolean NULL,
  "integration_id" uuid NULL,
  PRIMARY KEY ("scope_id", "name")
);
-- create index "idx_vars_integration_id" to table: "vars"
CREATE INDEX "idx_vars_integration_id" ON "vars" ("integration_id");
-- create index "idx_vars_name" to table: "vars"
CREATE INDEX "idx_vars_name" ON "vars" ("name");
-- create index "idx_vars_scope_id" to table: "vars"
CREATE INDEX "idx_vars_scope_id" ON "vars" ("scope_id");
-- drop "env_vars" table
DROP TABLE "env_vars";

-- +goose Down
-- reverse: drop "env_vars" table
CREATE TABLE "env_vars" (
  "env_id" uuid NOT NULL,
  "name" text NOT NULL,
  "value" text NULL,
  "secret_value" text NULL,
  "is_secret" boolean NULL,
  "membership_id" text NULL,
  PRIMARY KEY ("env_id", "name"),
  CONSTRAINT "fk_env_vars_env" FOREIGN KEY ("env_id") REFERENCES "envs" ("env_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "idx_env_vars_env_id" ON "env_vars" ("env_id");
CREATE UNIQUE INDEX "idx_env_vars_membership_id" ON "env_vars" ("membership_id");
-- reverse: create index "idx_vars_scope_id" to table: "vars"
DROP INDEX "idx_vars_scope_id";
-- reverse: create index "idx_vars_name" to table: "vars"
DROP INDEX "idx_vars_name";
-- reverse: create index "idx_vars_integration_id" to table: "vars"
DROP INDEX "idx_vars_integration_id";
-- reverse: create "vars" table
DROP TABLE "vars";
