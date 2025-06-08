-- +goose Up
-- modify "orgs" table
ALTER TABLE "orgs" ADD COLUMN "deleted_at" timestamptz NULL;
-- create index "idx_orgs_deleted_at" to table: "orgs"
CREATE INDEX "idx_orgs_deleted_at" ON "orgs" ("deleted_at");

-- +goose Down
-- reverse: create index "idx_orgs_deleted_at" to table: "orgs"
DROP INDEX "idx_orgs_deleted_at";
-- reverse: modify "orgs" table
ALTER TABLE "orgs" DROP COLUMN "deleted_at";
