-- +goose Up
-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "parent_session_id" uuid NULL, ADD
 CONSTRAINT "fk_sessions_parent_session" FOREIGN KEY ("parent_session_id") REFERENCES "sessions" ("session_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "idx_sessions_parent_session_id" to table: "sessions"
CREATE INDEX "idx_sessions_parent_session_id" ON "sessions" ("parent_session_id");

-- +goose Down
-- reverse: create index "idx_sessions_parent_session_id" to table: "sessions"
DROP INDEX "idx_sessions_parent_session_id";
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_parent_session", DROP COLUMN "parent_session_id";
