-- +goose Up
-- drop index "idx_session_log_records_session_id" from table: "session_log_records"
DROP INDEX "idx_session_log_records_session_id";
-- modify "session_log_records" table
ALTER TABLE "session_log_records" ADD COLUMN "seq" bigint;

CREATE SEQUENCE temp_id;

ALTER TABLE "session_log_records" ADD COLUMN "temp_id" int default nextval('temp_id');

-- map between temp_id (specific row) and its index in the session_id partition
WITH cte AS (
   SELECT temp_id, ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY data->'t') AS new_seq FROM session_log_records
 )

UPDATE session_log_records
SET seq = cte.new_seq FROM cte
WHERE session_log_records.temp_id = cte.temp_id;

-- cleanup
ALTER TABLE "session_log_records" DROP COLUMN temp_id;
DROP SEQUENCE temp_id;

ALTER TABLE "session_log_records" ALTER COLUMN "seq" SET NOT NULL;

ALTER TABLE "session_log_records" ADD PRIMARY KEY ("session_id", "seq");

-- +goose Down
-- reverse: modify "session_log_records" table
ALTER TABLE "session_log_records" DROP CONSTRAINT "session_log_records_pkey", DROP COLUMN "seq";
-- reverse: drop index "idx_session_log_records_session_id" from table: "session_log_records"
CREATE INDEX "idx_session_log_records_session_id" ON "session_log_records" ("session_id");
