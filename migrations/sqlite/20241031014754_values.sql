-- +goose Up
-- create "values" table
CREATE TABLE `values` (
  `project_id` uuid NOT NULL,
  `key` text NOT NULL,
  `value` blob NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`project_id`, `key`),
  CONSTRAINT `fk_values_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- +goose Down
-- reverse: create "values" table
DROP TABLE `values`;
