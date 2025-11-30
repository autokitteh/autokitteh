-- +goose Up
-- migrate "authType" variable name to "auth_type" in vars table
UPDATE "vars" SET "name" = 'auth_type' WHERE "name" = 'authType';

-- +goose Down
-- reverse: migrate "authType" variable name to "auth_type" in vars table
UPDATE "vars" SET "name" = 'authType' WHERE "name" = 'auth_type';
