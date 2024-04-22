-- +goose Up
-- modify "env_vars" table
ALTER TABLE "env_vars" ALTER COLUMN "env_id" SET NOT NULL, ALTER COLUMN "name" SET NOT NULL, ADD PRIMARY KEY ("env_id", "name");

-- +goose Down
-- reverse: modify "env_vars" table
ALTER TABLE "env_vars" DROP CONSTRAINT "env_vars_pkey", ALTER COLUMN "name" DROP NOT NULL, ALTER COLUMN "env_id" DROP NOT NULL;
