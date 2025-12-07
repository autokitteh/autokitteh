-- +goose Up
-- migrate "authType" variable name to "auth_type" in vars table
-- Delete any authType entries where auth_type already exists for the same var_id (keep the newer auth_type)
DELETE FROM `vars` WHERE `name` = 'authType' AND EXISTS (
  SELECT 1 FROM `vars` v2 WHERE v2.`name` = 'auth_type' AND v2.`var_id` = `vars`.`var_id`
);
-- Update remaining authType to auth_type
UPDATE `vars` SET `name` = 'auth_type' WHERE `name` = 'authType';

-- +goose Down
-- reverse: migrate "authType" variable name to "auth_type" in vars table
UPDATE `vars` SET `name` = 'authType' WHERE `name` = 'auth_type';
