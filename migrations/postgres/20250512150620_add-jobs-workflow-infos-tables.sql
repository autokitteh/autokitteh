-- +goose Up
-- create "jobs" table
CREATE TABLE "jobs" (
  "job_id" uuid NOT NULL,
  "type" text NULL,
  "status" text NULL,
  "data" jsonb NULL,
  "retry_count" smallint NULL DEFAULT 0,
  "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "start_processing_time" timestamptz NULL,
  "end_processing_time" timestamptz NULL,
  PRIMARY KEY ("job_id")
);
-- create index "idx_job_type_status" to table: "jobs"
CREATE INDEX "idx_job_type_status" ON "jobs" ("type", "status");
-- create index "idx_jobs_created_at" to table: "jobs"
CREATE INDEX "idx_jobs_created_at" ON "jobs" ("created_at");
-- create "worker_infos" table
CREATE TABLE "worker_infos" (
  "worker_id" text NOT NULL,
  "active_workflows" bigint NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("worker_id")
);

-- +goose Down
-- reverse: create "worker_infos" table
DROP TABLE "worker_infos";
-- reverse: create index "idx_jobs_created_at" to table: "jobs"
DROP INDEX "idx_jobs_created_at";
-- reverse: create index "idx_job_type_status" to table: "jobs"
DROP INDEX "idx_job_type_status";
-- reverse: create "jobs" table
DROP TABLE "jobs";
