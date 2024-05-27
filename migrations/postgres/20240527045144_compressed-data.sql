-- +goose Up
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "compressed_data" bytea NULL;
-- modify "session_call_attempts" table
ALTER TABLE "session_call_attempts" ADD COLUMN "compressed_complete" bytea NULL;
-- modify "session_call_specs" table
ALTER TABLE "session_call_specs" ADD COLUMN "compressed_data" bytea NULL;
-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "compressed_inputs" bytea NULL;

-- +goose Down
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "compressed_inputs";
-- reverse: modify "session_call_specs" table
ALTER TABLE "session_call_specs" DROP COLUMN "compressed_data";
-- reverse: modify "session_call_attempts" table
ALTER TABLE "session_call_attempts" DROP COLUMN "compressed_complete";
-- reverse: modify "events" table
ALTER TABLE "events" DROP COLUMN "compressed_data";
