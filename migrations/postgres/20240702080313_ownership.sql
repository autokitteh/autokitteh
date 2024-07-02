-- +goose Up
-- modify "vars" table
ALTER TABLE "vars" DROP CONSTRAINT "vars_pkey", DROP COLUMN "scope_id", ADD COLUMN "var_id" uuid NOT NULL, ADD PRIMARY KEY ("var_id", "name");
-- create index "idx_vars_var_id" to table: "vars"
CREATE INDEX "idx_vars_var_id" ON "vars" ("var_id");
-- create "users" table
CREATE TABLE "users" (
  "user_id" uuid NOT NULL,
  "provider" text NOT NULL,
  "email" text NOT NULL,
  "name" text NOT NULL,
  PRIMARY KEY ("user_id")
);
-- create index "idx_provider_email_name_idx" to table: "users"
CREATE UNIQUE INDEX "idx_provider_email_name_idx" ON "users" ("email", "provider", "name");
-- create "ownerships" table
CREATE TABLE "ownerships" (
  "entity_id" uuid NOT NULL,
  "entity_type" text NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("entity_id"),
  CONSTRAINT "fk_ownerships_user" FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- +goose Down
-- reverse: create "ownerships" table
DROP TABLE "ownerships";
-- reverse: create index "idx_provider_email_name_idx" to table: "users"
DROP INDEX "idx_provider_email_name_idx";
-- reverse: create "users" table
DROP TABLE "users";
-- reverse: create index "idx_vars_var_id" to table: "vars"
DROP INDEX "idx_vars_var_id";
-- reverse: modify "vars" table
ALTER TABLE "vars" DROP CONSTRAINT "vars_pkey", DROP COLUMN "var_id", ADD COLUMN "scope_id" uuid NOT NULL, ADD PRIMARY KEY ("scope_id", "name");
