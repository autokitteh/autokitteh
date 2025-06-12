-- +goose Up
-- drop index "idx_values_project_id" from table: "values"
DROP INDEX "idx_values_project_id";
-- modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "values_pkey", ADD PRIMARY KEY ("project_id", "key");

-- +goose Down
-- reverse: modify "values" table
ALTER TABLE "values" DROP CONSTRAINT "values_pkey", ADD PRIMARY KEY ("key");
-- reverse: drop index "idx_values_project_id" from table: "values"
CREATE INDEX "idx_values_project_id" ON "values" ("project_id");
