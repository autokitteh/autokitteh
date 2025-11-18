-- +goose Up
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "deduplication_key" text NULL;
-- create index "idx_ddk" to table: "events"
CREATE UNIQUE INDEX "idx_ddk" ON "events" ("destination_id", "deduplication_key", "event_type") WHERE (deduplication_key IS NOT NULL);

-- +goose Down
-- reverse: create index "idx_ddk" to table: "events"
DROP INDEX "idx_ddk";
-- reverse: modify "events" table
ALTER TABLE "events" DROP COLUMN "deduplication_key";
