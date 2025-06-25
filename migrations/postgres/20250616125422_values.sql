-- +goose Up
-- create "store_values" table
CREATE TABLE "store_values" (
  "created_by" uuid NULL,
  "created_at" timestamptz NULL,
  "project_id" uuid NOT NULL,
  "key" text NOT NULL,
  "value" bytea NULL,
  "updated_by" uuid NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("project_id", "key"),
  CONSTRAINT "fk_store_values_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- drop "values" table
DROP TABLE "values";

-- +goose Down
-- reverse: drop "values" table
CREATE TABLE "values" (
  "project_id" uuid NOT NULL,
  "key" text NOT NULL,
  "value" bytea NULL,
  "updated_at" timestamptz NULL,
  "created_by" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_by" uuid NULL,
  PRIMARY KEY ("key"),
  CONSTRAINT "fk_values_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "idx_values_project_id" ON "values" ("project_id");
-- reverse: create "store_values" table
DROP TABLE "store_values";
