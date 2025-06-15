-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_values" table
CREATE TABLE `new_values` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NOT NULL,
  `key` text NOT NULL,
  `value` blob NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`project_id`, `key`),
  CONSTRAINT `fk_values_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "values" to new temporary table "new_values"
INSERT INTO `new_values` (`created_by`, `created_at`, `project_id`, `key`, `value`, `updated_by`, `updated_at`) SELECT `created_by`, `created_at`, `project_id`, `key`, `value`, `updated_by`, `updated_at` FROM `values`;
-- drop "values" table after copying rows
DROP TABLE `values`;
-- rename temporary table "new_values" to "values"
ALTER TABLE `new_values` RENAME TO `values`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "new_values" table
DROP TABLE `new_values`;
