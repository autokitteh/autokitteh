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

-- +goose Down
-- reverse: create index "idx_jobs_created_at" to table: "jobs"
DROP INDEX "idx_jobs_created_at";
-- reverse: create index "idx_job_type_status" to table: "jobs"
DROP INDEX "idx_job_type_status";
-- reverse: create "jobs" table
DROP TABLE "jobs";
