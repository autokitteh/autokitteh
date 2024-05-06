-- +goose Up
-- modify "secrets" table
ALTER TABLE "secrets" DROP CONSTRAINT "secrets_pkey", DROP COLUMN "name", DROP COLUMN "data", ADD COLUMN "key" text NOT NULL, ADD COLUMN "value" jsonb NULL, ADD PRIMARY KEY ("key");

-- +goose Down
-- reverse: modify "secrets" table
ALTER TABLE "secrets" DROP CONSTRAINT "secrets_pkey", DROP COLUMN "value", DROP COLUMN "key", ADD COLUMN "data" jsonb NULL, ADD COLUMN "name" text NOT NULL, ADD PRIMARY KEY ("name");
