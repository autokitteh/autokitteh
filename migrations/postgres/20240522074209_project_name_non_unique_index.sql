-- +goose Up
-- drop index "idx_projects_name" from table: "projects"
DROP INDEX "idx_projects_name";
-- create index "idx_projects_name" to table: "projects"
CREATE INDEX "idx_projects_name" ON "projects" ("name");

-- +goose Down
-- reverse: create index "idx_projects_name" to table: "projects"
DROP INDEX "idx_projects_name";
-- reverse: drop index "idx_projects_name" from table: "projects"
CREATE UNIQUE INDEX "idx_projects_name" ON "projects" ("name");
