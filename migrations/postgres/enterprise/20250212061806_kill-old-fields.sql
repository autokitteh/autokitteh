-- +goose Up
-- modify "users" table
ALTER TABLE "users" DROP COLUMN "disabled";
-- drop "ownerships" table
DROP TABLE "ownerships";

-- +goose Down
-- reverse: drop "ownerships" table
CREATE TABLE "ownerships" (
  "entity_id" uuid NOT NULL,
  "entity_type" text NOT NULL,
  "user_id" text NOT NULL,
  PRIMARY KEY ("entity_id")
);
-- reverse: modify "users" table
ALTER TABLE "users" ADD COLUMN "disabled" boolean NULL;
