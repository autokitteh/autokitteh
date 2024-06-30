-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_builds" table
CREATE TABLE `new_builds` (
  `build_id` uuid NOT NULL,
  `project_id` uuid NULL,
  `data` blob NULL,
  `created_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`build_id`),
  CONSTRAINT `fk_builds_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "builds" to new temporary table "new_builds"
INSERT INTO `new_builds` (`build_id`, `data`, `created_at`, `deleted_at`) SELECT `build_id`, `data`, `created_at`, `deleted_at` FROM `builds`;
-- drop "builds" table after copying rows
DROP TABLE `builds`;
-- rename temporary table "new_builds" to "builds"
ALTER TABLE `new_builds` RENAME TO `builds`;
-- create index "idx_builds_project_id" to table: "builds"
CREATE INDEX `idx_builds_project_id` ON `builds` (`project_id`);
-- create index "idx_builds_deleted_at" to table: "builds"
CREATE INDEX `idx_builds_deleted_at` ON `builds` (`deleted_at`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_builds_deleted_at" to table: "builds"
DROP INDEX `idx_builds_deleted_at`;
-- reverse: create index "idx_builds_project_id" to table: "builds"
DROP INDEX `idx_builds_project_id`;
-- reverse: create "new_builds" table
DROP TABLE `new_builds`;
