-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_sessions" table
CREATE TABLE `new_sessions` (
  `session_id` uuid NOT NULL,
  `parent_session_id` uuid NULL,
  `build_id` uuid NULL,
  `env_id` uuid NULL,
  `deployment_id` uuid NULL,
  `event_id` uuid NULL,
  `current_state_type` integer NULL,
  `entrypoint` text NULL,
  `inputs` json NULL,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`session_id`),
  CONSTRAINT `fk_sessions_env` FOREIGN KEY (`env_id`) REFERENCES `envs` (`env_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_build` FOREIGN KEY (`build_id`) REFERENCES `builds` (`build_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_parent_session` FOREIGN KEY (`parent_session_id`) REFERENCES `sessions` (`session_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_event` FOREIGN KEY (`event_id`) REFERENCES `events` (`event_id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_sessions_deployment` FOREIGN KEY (`deployment_id`) REFERENCES `deployments` (`deployment_id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "sessions" to new temporary table "new_sessions"
INSERT INTO `new_sessions` (`session_id`, `build_id`, `env_id`, `deployment_id`, `event_id`, `current_state_type`, `entrypoint`, `inputs`, `created_at`, `updated_at`, `deleted_at`) SELECT `session_id`, `build_id`, `env_id`, `deployment_id`, `event_id`, `current_state_type`, `entrypoint`, `inputs`, `created_at`, `updated_at`, `deleted_at` FROM `sessions`;
-- drop "sessions" table after copying rows
DROP TABLE `sessions`;
-- rename temporary table "new_sessions" to "sessions"
ALTER TABLE `new_sessions` RENAME TO `sessions`;
-- create index "idx_sessions_event_id" to table: "sessions"
CREATE INDEX `idx_sessions_event_id` ON `sessions` (`event_id`);
-- create index "idx_sessions_deployment_id" to table: "sessions"
CREATE INDEX `idx_sessions_deployment_id` ON `sessions` (`deployment_id`);
-- create index "idx_sessions_env_id" to table: "sessions"
CREATE INDEX `idx_sessions_env_id` ON `sessions` (`env_id`);
-- create index "idx_sessions_build_id" to table: "sessions"
CREATE INDEX `idx_sessions_build_id` ON `sessions` (`build_id`);
-- create index "idx_sessions_parent_session_id" to table: "sessions"
CREATE INDEX `idx_sessions_parent_session_id` ON `sessions` (`parent_session_id`);
-- create index "idx_sessions_deleted_at" to table: "sessions"
CREATE INDEX `idx_sessions_deleted_at` ON `sessions` (`deleted_at`);
-- create index "idx_sessions_current_state_type" to table: "sessions"
CREATE INDEX `idx_sessions_current_state_type` ON `sessions` (`current_state_type`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "idx_sessions_current_state_type" to table: "sessions"
DROP INDEX `idx_sessions_current_state_type`;
-- reverse: create index "idx_sessions_deleted_at" to table: "sessions"
DROP INDEX `idx_sessions_deleted_at`;
-- reverse: create index "idx_sessions_parent_session_id" to table: "sessions"
DROP INDEX `idx_sessions_parent_session_id`;
-- reverse: create index "idx_sessions_build_id" to table: "sessions"
DROP INDEX `idx_sessions_build_id`;
-- reverse: create index "idx_sessions_env_id" to table: "sessions"
DROP INDEX `idx_sessions_env_id`;
-- reverse: create index "idx_sessions_deployment_id" to table: "sessions"
DROP INDEX `idx_sessions_deployment_id`;
-- reverse: create index "idx_sessions_event_id" to table: "sessions"
DROP INDEX `idx_sessions_event_id`;
-- reverse: create "new_sessions" table
DROP TABLE `new_sessions`;
