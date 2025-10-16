-- +goose Up
-- modify "orgs" table
ALTER TABLE "orgs" ADD COLUMN "name" text NULL;
-- create index "idx_orgs_name" to table: "orgs"
CREATE INDEX "idx_orgs_name" ON "orgs" ("name");

-- +goose Down
-- reverse: create index "idx_orgs_name" to table: "orgs"
DROP INDEX "idx_orgs_name";
-- reverse: modify "orgs" table
ALTER TABLE "orgs" DROP COLUMN "name";
