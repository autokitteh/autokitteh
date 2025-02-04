-- +goose Up
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "session_id" uuid NULL;
-- create index "idx_events_session_id" to table: "events"
CREATE INDEX "idx_events_session_id" ON "events" ("session_id");
-- modify "sessions" table
ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_event";
-- modify "signals" table
ALTER TABLE "signals" ADD COLUMN "session_id" uuid NULL;

-- +goose Down
-- reverse: modify "signals" table
ALTER TABLE "signals" DROP COLUMN "session_id";
-- reverse: modify "sessions" table
ALTER TABLE "sessions" ADD
 CONSTRAINT "fk_sessions_event" FOREIGN KEY ("event_id") REFERENCES "events" ("event_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: create index "idx_events_session_id" to table: "events"
DROP INDEX "idx_events_session_id";
-- reverse: modify "events" table
ALTER TABLE "events" DROP COLUMN "session_id";
