-- +goose Up
-- modify "triggers" table
ALTER TABLE "triggers" ALTER COLUMN "connection_id" DROP NOT NULL, DROP COLUMN "data", ALTER COLUMN "unique_name" SET NOT NULL, ADD COLUMN "source_type" text NULL, ADD COLUMN "webhook_slug" text NULL, ADD COLUMN "schedule" text NULL;
-- create index "idx_triggers_source_type" to table: "triggers"
CREATE INDEX "idx_triggers_source_type" ON "triggers" ("source_type");
-- create index "idx_triggers_webhook_slug" to table: "triggers"
CREATE INDEX "idx_triggers_webhook_slug" ON "triggers" ("webhook_slug");
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "destination_id" uuid, ADD COLUMN "trigger_id" uuid NULL, ADD
 CONSTRAINT "fk_events_trigger" FOREIGN KEY ("trigger_id") REFERENCES "triggers" ("trigger_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- copy connection_id to new trigger_id column
UPDATE "events" SET destination_id=connection_id;
-- modify "events" table
ALTER TABLE "events" ALTER COLUMN "destination_id" SET NOT NULL;
-- create index "idx_events_destination_id" to table: "events"
CREATE INDEX "idx_events_destination_id" ON "events" ("destination_id");
-- create index "idx_events_trigger_id" to table: "events"
CREATE INDEX "idx_events_trigger_id" ON "events" ("trigger_id");
-- drop index "idx_connection_id_event_type" from table: "signals"
DROP INDEX "idx_connection_id_event_type";
-- modify "signals" table
ALTER TABLE "signals" ALTER COLUMN "signal_id" TYPE uuid using signal_id::uuid, ALTER COLUMN "connection_id" DROP NOT NULL, ADD COLUMN "destination_id" uuid, ADD COLUMN "trigger_id" uuid NULL, ADD
 CONSTRAINT "fk_signals_trigger" FOREIGN KEY ("trigger_id") REFERENCES "triggers" ("trigger_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- copy connection_id to destination_id on signals
UPDATE "signals" SET destination_id=connection_id;
-- modify "signals" table
ALTER TABLE "signals" ALTER COLUMN "destination_id" SET NOT NULL;
-- create index "idx_signals_destination_id" to table: "signals"
CREATE INDEX "idx_signals_destination_id" ON "signals" ("destination_id");

-- +goose Down
-- reverse: create index "idx_signals_destination_id" to table: "signals"
DROP INDEX "idx_signals_destination_id";
-- reverse: modify "signals" table
ALTER TABLE "signals" DROP CONSTRAINT "fk_signals_trigger", DROP COLUMN "trigger_id", DROP COLUMN "destination_id", ALTER COLUMN "connection_id" SET NOT NULL, ALTER COLUMN "signal_id" TYPE text;
-- reverse: drop index "idx_connection_id_event_type" from table: "signals"
CREATE INDEX "idx_connection_id_event_type" ON "signals" ("connection_id");
-- reverse: create index "idx_events_trigger_id" to table: "events"
DROP INDEX "idx_events_trigger_id";
-- reverse: create index "idx_events_destination_id" to table: "events"
DROP INDEX "idx_events_destination_id";
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_trigger", DROP COLUMN "trigger_id", DROP COLUMN "destination_id";
-- reverse: create index "idx_triggers_webhook_slug" to table: "triggers"
DROP INDEX "idx_triggers_webhook_slug";
-- reverse: create index "idx_triggers_source_type" to table: "triggers"
DROP INDEX "idx_triggers_source_type";
-- reverse: modify "triggers" table
ALTER TABLE "triggers" DROP COLUMN "schedule", DROP COLUMN "webhook_slug", DROP COLUMN "source_type", ALTER COLUMN "unique_name" DROP NOT NULL, ADD COLUMN "data" jsonb NULL, ALTER COLUMN "connection_id" SET NOT NULL;
