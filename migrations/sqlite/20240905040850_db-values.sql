-- +goose Up
-- create "values" table
CREATE TABLE `values` (
  `env_id` uuid NOT NULL,
  `key` text NOT NULL,
  `value` blob NULL,
  PRIMARY KEY (`env_id`, `key`)
);

-- +goose Down
-- reverse: create "values" table
DROP TABLE `values`;
