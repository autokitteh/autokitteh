-- +goose Up
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
