-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_secrets" table
CREATE TABLE `new_secrets` (
  `key` text NULL,
  `value` text NULL,
  PRIMARY KEY (`key`)
);
-- drop "secrets" table without copying rows (no columns)
DROP TABLE `secrets`;
-- rename temporary table "new_secrets" to "secrets"
ALTER TABLE `new_secrets` RENAME TO `secrets`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "new_secrets" table
DROP TABLE `new_secrets`;
