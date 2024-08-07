-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_projects" table
CREATE TABLE `new_projects` (
  `project_id` uuid NOT NULL,
  `name` text NOT NULL,
  `root_url` text NULL,
  `resources` blob NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`project_id`)
);
-- copy rows from old table "projects" to new temporary table "new_projects"
INSERT INTO `new_projects` (`project_id`, `name`, `root_url`, `resources`, `deleted_at`) SELECT `project_id`, `name`, `root_url`, `resources`, `deleted_at` FROM `projects`;
-- drop "projects" table after copying rows
DROP TABLE `projects`;
-- rename temporary table "new_projects" to "projects"
ALTER TABLE `new_projects` RENAME TO `projects`;
-- create index "idx_projects_deleted_at" to table: "projects"
CREATE INDEX `idx_projects_deleted_at` ON `projects` (`deleted_at`);
-- create index "idx_projects_name" to table: "projects"
CREATE INDEX `idx_projects_name` ON `projects` (`name`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- manually written
PRAGMA foreign_keys = off;
-- reverse: create "old_projects" table
CREATE TABLE `old_projects` (
  `project_id` uuid NOT NULL,
  `name` text,
  `root_url` text NULL,
  `resources` blob NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`project_id`)
);
-- reverse: copy rows from newer "projects" table to temporary table "old_projects"
INSERT INTO `old_projects` (`project_id`, `name`, `root_url`, `resources`, `deleted_at`) SELECT `project_id`, `name`, `root_url`, `resources`, `deleted_at` FROM `projects`;
-- reverse: drop newer "projects" table after copying rows
DROP TABLE `projects`;
-- reverse: rename temporary table "old_projects" to "projects"
ALTER TABLE `old_projects` RENAME TO `projects`;
-- reverse: create index "idx_projects_deleted_at" to table: "projects"
CREATE INDEX `idx_projects_deleted_at` ON `projects` (`deleted_at`);
-- reverse: create index "idx_projects_name" to table: "projects"
CREATE INDEX `idx_projects_name` ON `projects` (`name`);
-- reverse: enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
