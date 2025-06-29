-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- drop "values" table
DROP TABLE `values`;
-- create "store_values" table
CREATE TABLE `store_values` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `key` text NOT NULL,
  `value` blob NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`project_id`, `key`),
  CONSTRAINT `fk_store_values_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "store_values" table
DROP TABLE `store_values`;
-- reverse: drop "values" table
CREATE TABLE `values` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `key` text NOT NULL,
  `value` blob NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`key`),
  CONSTRAINT `fk_values_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX `idx_values_project_id` ON `values` (`project_id`);
