-- +goose Up
-- modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "memo" jsonb NULL;

-- +goose Down
-- reverse: modify "sessions" table
ALTER TABLE "sessions" DROP COLUMN "memo";
