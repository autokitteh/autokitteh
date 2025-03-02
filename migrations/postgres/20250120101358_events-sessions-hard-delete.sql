-- +goose Up
-- remove deleted sessions and events
DELETE from "session_log_records" where session_id in (select session_id from sessions where deleted_at is not NULL);
DELETE from "session_call_specs" where session_id in (select session_id from sessions where deleted_at is not NULL);
DELETE from "session_call_attempts" where session_id in (select session_id from sessions where deleted_at is not NULL);
DELETE from "sessions" where deleted_at is not NULL;
DELETE from events where deleted_at is not NULL;

-- modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_connection", DROP CONSTRAINT "fk_events_trigger", DROP COLUMN "deleted_at", ADD
 CONSTRAINT "fk_events_connection" FOREIGN KEY ("connection_id") REFERENCES "connections" ("connection_id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
 CONSTRAINT "fk_events_trigger" FOREIGN KEY ("trigger_id") REFERENCES "triggers" ("trigger_id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "sessions" table
ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_event", DROP COLUMN "deleted_at", ADD
 CONSTRAINT "fk_sessions_event" FOREIGN KEY ("event_id") REFERENCES "events" ("event_id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "session_call_attempts" table
ALTER TABLE "session_call_attempts" DROP CONSTRAINT "fk_session_call_attempts_session", ADD
 CONSTRAINT "fk_session_call_attempts_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "session_call_specs" table
ALTER TABLE "session_call_specs" DROP CONSTRAINT "fk_session_call_specs_session", ADD
 CONSTRAINT "fk_session_call_specs_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "session_log_records" table
ALTER TABLE "session_log_records" DROP CONSTRAINT "fk_session_log_records_session", ADD
 CONSTRAINT "fk_session_log_records_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "session_log_records" table
ALTER TABLE "session_log_records" DROP CONSTRAINT "fk_session_log_records_session", ADD
 CONSTRAINT "fk_session_log_records_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "session_call_specs" table
ALTER TABLE "session_call_specs" DROP CONSTRAINT "fk_session_call_specs_session", ADD
 CONSTRAINT "fk_session_call_specs_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "session_call_attempts" table
ALTER TABLE "session_call_attempts" DROP CONSTRAINT "fk_session_call_attempts_session", ADD
 CONSTRAINT "fk_session_call_attempts_session" FOREIGN KEY ("session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_event", ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_sessions_event" FOREIGN KEY ("event_id") REFERENCES "events" ("event_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_trigger", DROP CONSTRAINT "fk_events_connection", ADD COLUMN "deleted_at" timestamptz NULL, ADD
 CONSTRAINT "fk_events_trigger" FOREIGN KEY ("trigger_id") REFERENCES "triggers" ("trigger_id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_events_connection" FOREIGN KEY ("connection_id") REFERENCES "connections" ("connection_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
