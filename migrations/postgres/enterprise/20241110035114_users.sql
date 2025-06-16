-- +goose Up
-- create "users" table
CREATE TABLE "users" (
  "user_id" uuid NOT NULL,
  "email" text NOT NULL,
  "display_name" text NULL,
  PRIMARY KEY ("user_id")
);
-- create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email");

-- +goose Down
-- reverse: create index "idx_users_email" to table: "users"
DROP INDEX "idx_users_email";
-- reverse: create "users" table
DROP TABLE "users";
