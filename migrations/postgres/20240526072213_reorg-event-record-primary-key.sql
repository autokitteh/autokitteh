-- +goose Up
-- modify "event_records" table
ALTER TABLE "event_records" DROP CONSTRAINT "event_records_pkey", ADD PRIMARY KEY ("event_id", "seq");

-- +goose Down
-- reverse: modify "event_records" table
ALTER TABLE "event_records" DROP CONSTRAINT "event_records_pkey", ADD PRIMARY KEY ("seq", "event_id");
