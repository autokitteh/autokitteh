-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ADD COLUMN "deleted_at" timestamptz NULL;
-- create index "idx_triggers_deleted_at" to table: "triggers"
CREATE INDEX "idx_triggers_deleted_at" ON "triggers" ("deleted_at");

-- +goose Down
-- reverse: create index "idx_triggers_deleted_at" to table: "triggers"
DROP INDEX "idx_triggers_deleted_at";
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "deleted_at";
