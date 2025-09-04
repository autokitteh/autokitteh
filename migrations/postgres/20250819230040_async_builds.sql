-- +goose Up
-- modify "builds" table
ALTER TABLE "builds" ADD COLUMN "status" integer NULL;
-- create index "idx_builds_status" to table: "builds"
CREATE INDEX "idx_builds_status" ON "builds" ("status");

-- +goose Down
-- reverse: create index "idx_builds_status" to table: "builds"
DROP INDEX "idx_builds_status";
-- reverse: modify "builds" table
ALTER TABLE "builds" DROP COLUMN "status";
