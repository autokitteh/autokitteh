-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_deployments" table
CREATE TABLE `new_deployments` (
  `deployment_id` uuid NOT NULL,
  `env_id` uuid NULL,
  `build_id` uuid NOT NULL,
  `state` integer NULL,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`deployment_id`),
  CONSTRAINT `fk_deployments_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_deployments_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "deployments" to new temporary table "new_deployments"
INSERT INTO `new_deployments` (`deployment_id`, `env_id`, `build_id`, `state`, `created_at`, `updated_at`, `deleted_at`) SELECT `deployment_id`, `env_id`, `build_id`, `state`, `created_at`, `updated_at`, `deleted_at` FROM `deployments`;
-- drop "deployments" table after copying rows
DROP TABLE `deployments`;
-- rename temporary table "new_deployments" to "deployments"
ALTER TABLE `new_deployments` RENAME TO `deployments`;
-- create index "idx_deployments_deleted_at" to table: "deployments"
CREATE INDEX `idx_deployments_deleted_at` ON `deployments` (`deleted_at`);
-- create index "idx_deployments_env_id" to table: "deployments"
CREATE INDEX `idx_deployments_env_id` ON `deployments` (`env_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_deployments_env_id" to table: "deployments"
DROP INDEX `idx_deployments_env_id`;
-- reverse: create index "idx_deployments_deleted_at" to table: "deployments"
DROP INDEX `idx_deployments_deleted_at`;
-- reverse: create "new_deployments" table
DROP TABLE `new_deployments`;
