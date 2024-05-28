-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_triggers" table
CREATE TABLE `new_triggers` (
  `trigger_id` uuid NOT NULL,
  `project_id` uuid NOT NULL,
  `connection_id` uuid NOT NULL,
  `env_id` uuid NOT NULL,
  `name` text NULL,
  `event_type` text NULL,
  `filter` text NULL,
  `code_location` text NULL,
  `data` json NULL,
  `unique_name` text NULL,
  PRIMARY KEY (`trigger_id`),
  CONSTRAINT `fk_triggers_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_triggers_connection` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`connection_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "triggers" to new temporary table "new_triggers"
INSERT INTO `new_triggers` (`trigger_id`, `project_id`, `connection_id`, `env_id`, `name`, `event_type`, `filter`, `code_location`, `data`, `unique_name`) SELECT `trigger_id`, `project_id`, `connection_id`, `env_id`, `name`, `event_type`, `filter`, `code_location`, `data`, `unique_name` FROM `triggers`;
-- drop "triggers" table after copying rows
DROP TABLE `triggers`;
-- rename temporary table "new_triggers" to "triggers"
ALTER TABLE `new_triggers` RENAME TO `triggers`;
-- create index "idx_triggers_unique_name" to table: "triggers"
CREATE UNIQUE INDEX `idx_triggers_unique_name` ON `triggers` (`unique_name`);
-- create index "idx_triggers_env_id" to table: "triggers"
CREATE INDEX `idx_triggers_env_id` ON `triggers` (`env_id`);
-- create index "idx_triggers_connection_id" to table: "triggers"
CREATE INDEX `idx_triggers_connection_id` ON `triggers` (`connection_id`);
-- create index "idx_triggers_project_id" to table: "triggers"
CREATE INDEX `idx_triggers_project_id` ON `triggers` (`project_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_triggers_project_id" to table: "triggers"
DROP INDEX `idx_triggers_project_id`;
-- reverse: create index "idx_triggers_connection_id" to table: "triggers"
DROP INDEX `idx_triggers_connection_id`;
-- reverse: create index "idx_triggers_env_id" to table: "triggers"
DROP INDEX `idx_triggers_env_id`;
-- reverse: create index "idx_triggers_unique_name" to table: "triggers"
DROP INDEX `idx_triggers_unique_name`;
-- reverse: create "new_triggers" table
DROP TABLE `new_triggers`;
