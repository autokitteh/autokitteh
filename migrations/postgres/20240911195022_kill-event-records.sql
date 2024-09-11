-- +goose Up
-- drop "event_records" table
DROP TABLE "event_records";

-- +goose Down
-- reverse: drop "event_records" table
CREATE TABLE "event_records" (
  "event_id" uuid NOT NULL,
  "seq" bigint NOT NULL,
  "state" integer NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("event_id", "seq"),
  CONSTRAINT "fk_event_records_event" FOREIGN KEY ("event_id") REFERENCES "events" ("event_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "idx_event_records_state" ON "event_records" ("state");
