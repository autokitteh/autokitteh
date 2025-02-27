-- +goose Up
-- drop index "idx_event_type_seq" from table: "events"
DROP INDEX "idx_event_type_seq";
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "org_id" uuid NULL, ADD CONSTRAINT "fk_events_org" FOREIGN KEY ("org_id") REFERENCES "orgs" ("org_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "idx_event_type_seq" to table: "events"
CREATE INDEX "idx_event_type_seq" ON "events" ("event_type", "seq");
-- create index "idx_events_org_id" to table: "events"
CREATE INDEX "idx_events_org_id" ON "events" ("org_id");
-- create index "idx_org_id_seq" to table: "events"
CREATE INDEX "idx_org_id_seq" ON "events" ("org_id", "seq");

-- +goose Down
-- reverse: create index "idx_org_id_seq" to table: "events"
DROP INDEX "idx_org_id_seq";
-- reverse: create index "idx_events_org_id" to table: "events"
DROP INDEX "idx_events_org_id";
-- reverse: create index "idx_event_type_seq" to table: "events"
DROP INDEX "idx_event_type_seq";
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_org", DROP COLUMN "org_id";
-- reverse: drop index "idx_event_type_seq" from table: "events"
CREATE INDEX "idx_event_type_seq" ON "events" ("event_type");
