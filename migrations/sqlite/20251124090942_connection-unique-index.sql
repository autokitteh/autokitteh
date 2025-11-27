-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_connections" table
CREATE TABLE `new_connections` (
  `created_by` uuid NULL,
  `created_at` datetime NULL,
  `project_id` uuid NULL,
  `org_id` uuid NOT NULL,
  `connection_id` uuid NOT NULL,
  `integration_id` uuid NULL,
  `name` text NOT NULL,
  `status_code` integer NULL,
  `status_message` text NULL,
  `updated_by` uuid NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`connection_id`),
  CONSTRAINT `fk_connections_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`project_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "connections" to new temporary table "new_connections"
INSERT INTO `new_connections` (`created_by`, `created_at`, `project_id`, `connection_id`, `integration_id`, `name`, `status_code`, `status_message`, `updated_by`, `updated_at`, `deleted_at`) SELECT `created_by`, `created_at`, `project_id`, `connection_id`, `integration_id`, `name`, `status_code`, `status_message`, `updated_by`, `updated_at`, `deleted_at` FROM `connections`;
-- drop "connections" table after copying rows
DROP TABLE `connections`;
-- rename temporary table "new_connections" to "connections"
ALTER TABLE `new_connections` RENAME TO `connections`;
-- create index "idx_connections_deleted_at" to table: "connections"
CREATE INDEX `idx_connections_deleted_at` ON `connections` (`deleted_at`);
-- create index "idx_connections_status_code" to table: "connections"
CREATE INDEX `idx_connections_status_code` ON `connections` (`status_code`);
-- create index "idx_connections_integration_id" to table: "connections"
CREATE INDEX `idx_connections_integration_id` ON `connections` (`integration_id`);
-- create index "idx_connection_org_id_name" to table: "connections"
CREATE UNIQUE INDEX `idx_connection_org_id_name` ON `connections` (`org_id`, `name`) WHERE project_id is null and deleted_at is null;
-- create index "idx_connections_project_id" to table: "connections"
CREATE INDEX `idx_connections_project_id` ON `connections` (`project_id`);
-- create index "idx_connection_org_id_project_id_name" to table: "connections"
CREATE UNIQUE INDEX `idx_connection_org_id_project_id_name` ON `connections` (`org_id`, `project_id`, `name`) WHERE project_id is not null and deleted_at is null;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_connection_org_id_project_id_name" to table: "connections"
DROP INDEX `idx_connection_org_id_project_id_name`;
-- reverse: create index "idx_connections_project_id" to table: "connections"
DROP INDEX `idx_connections_project_id`;
-- reverse: create index "idx_connection_org_id_name" to table: "connections"
DROP INDEX `idx_connection_org_id_name`;
-- reverse: create index "idx_connections_integration_id" to table: "connections"
DROP INDEX `idx_connections_integration_id`;
-- reverse: create index "idx_connections_status_code" to table: "connections"
DROP INDEX `idx_connections_status_code`;
-- reverse: create index "idx_connections_deleted_at" to table: "connections"
DROP INDEX `idx_connections_deleted_at`;
-- reverse: create "new_connections" table
DROP TABLE `new_connections`;
