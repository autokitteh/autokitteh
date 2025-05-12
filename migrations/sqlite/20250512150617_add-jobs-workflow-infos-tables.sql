-- +goose Up
-- create "jobs" table
CREATE TABLE `jobs` (
  `job_id` uuid NOT NULL,
  `type` text NULL,
  `status` text NULL,
  `data` json NULL,
  `retry_count` integer NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  `start_processing_time` datetime NULL,
  `end_processing_time` datetime NULL,
  PRIMARY KEY (`job_id`)
);
-- create index "idx_jobs_created_at" to table: "jobs"
CREATE INDEX `idx_jobs_created_at` ON `jobs` (`created_at`);
-- create index "idx_job_type_status" to table: "jobs"
CREATE INDEX `idx_job_type_status` ON `jobs` (`type`, `status`);
-- create "worker_infos" table
CREATE TABLE `worker_infos` (
  `worker_id` text NOT NULL,
  `active_workflows` integer NULL,
  `updated_at` datetime NULL,
  PRIMARY KEY (`worker_id`)
);

-- +goose Down
-- reverse: create "worker_infos" table
DROP TABLE `worker_infos`;
-- reverse: create index "idx_job_type_status" to table: "jobs"
DROP INDEX `idx_job_type_status`;
-- reverse: create index "idx_jobs_created_at" to table: "jobs"
DROP INDEX `idx_jobs_created_at`;
-- reverse: create "jobs" table
DROP TABLE `jobs`;
