-- +goose Up
-- create index "idx_deployments_state" to table: "deployments"
CREATE INDEX "idx_deployments_state" ON "deployments" ("state");

-- +goose Down
-- reverse: create index "idx_deployments_state" to table: "deployments"
DROP INDEX "idx_deployments_state";
