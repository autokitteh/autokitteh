-- +goose Up
-- modify "org_members" table
ALTER TABLE "org_members" ADD COLUMN "roles" jsonb NULL, ADD COLUMN "updated_by" uuid NULL, ADD COLUMN "updated_at" timestamptz NULL;

UPDATE "org_members" SET "roles"='["admin"]';

-- +goose Down
-- reverse: modify "org_members" table
ALTER TABLE "org_members" DROP COLUMN "updated_at", DROP COLUMN "updated_by", DROP COLUMN "roles";
