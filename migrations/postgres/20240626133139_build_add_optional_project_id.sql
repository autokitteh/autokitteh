-- +goose Up
-- modify "builds" table
ALTER TABLE "builds" ADD COLUMN "project_id" uuid NULL, ADD
 CONSTRAINT "fk_builds_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "idx_builds_project_id" to table: "builds"
CREATE INDEX "idx_builds_project_id" ON "builds" ("project_id");

-- +goose Down
-- reverse: create index "idx_builds_project_id" to table: "builds"
DROP INDEX "idx_builds_project_id";
-- reverse: modify "builds" table
ALTER TABLE "builds" DROP CONSTRAINT "fk_builds_project", DROP COLUMN "project_id";
