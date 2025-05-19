-- +goose Up
-- modify "ownerships" table
ALTER TABLE "ownerships" DROP CONSTRAINT "fk_ownerships_user", ALTER COLUMN "user_id" TYPE text;
-- drop "users" table
DROP TABLE "users";

-- +goose Down
-- reverse: drop "users" table
CREATE TABLE "users" (
  "user_id" uuid NOT NULL,
  "provider" text NOT NULL,
  "email" text NOT NULL,
  "name" text NOT NULL,
  PRIMARY KEY ("user_id")
);
CREATE UNIQUE INDEX "idx_provider_email_name_idx" ON "users" ("email", "provider", "name");
-- reverse: modify "ownerships" table
ALTER TABLE "ownerships" ALTER COLUMN "user_id" TYPE uuid, ADD
 CONSTRAINT "fk_ownerships_user" FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
