-- +goose Up
-- create "locks" table
CREATE TABLE "locks" (
  "id" text NOT NULL,
  "count" bigint NOT NULL,
  PRIMARY KEY ("id")
);

-- +goose Down
-- reverse: create "locks" table
DROP TABLE "locks";
