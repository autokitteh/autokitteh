-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- drop "users" table
DROP TABLE `users`;
-- create "new_ownerships" table
CREATE TABLE `new_ownerships` (
  `entity_id` uuid NOT NULL,
  `entity_type` text NOT NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`entity_id`)
);
-- copy rows from old table "ownerships" to new temporary table "new_ownerships"
INSERT INTO `new_ownerships` (`entity_id`, `entity_type`, `user_id`) SELECT `entity_id`, `entity_type`, `user_id` FROM `ownerships`;
-- drop "ownerships" table after copying rows
DROP TABLE `ownerships`;
-- rename temporary table "new_ownerships" to "ownerships"
ALTER TABLE `new_ownerships` RENAME TO `ownerships`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "new_ownerships" table
DROP TABLE `new_ownerships`;
-- reverse: drop "users" table
CREATE TABLE `users` (
  `user_id` uuid NOT NULL,
  `provider` text NOT NULL,
  `email` text NOT NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`user_id`)
);
CREATE UNIQUE INDEX `idx_provider_email_name_idx` ON `users` (`email`, `provider`, `name`);
