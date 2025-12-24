-- +goose Up
-- drop index "idx_connection_org_id_name" from table: "connections"
DROP INDEX `idx_connection_org_id_name`;
-- drop index "idx_connections_project_id" from table: "connections"
DROP INDEX `idx_connections_project_id`;
-- drop index "idx_connection_org_id_project_id_name" from table: "connections"
DROP INDEX `idx_connection_org_id_project_id_name`;
-- create index "idx_connection_org_id_project_id" to table: "connections"
CREATE INDEX `idx_connection_org_id_project_id` ON `connections` (`org_id`, `project_id`);

-- +goose Down
-- reverse: create index "idx_connection_org_id_project_id" to table: "connections"
DROP INDEX `idx_connection_org_id_project_id`;
-- reverse: drop index "idx_connection_org_id_project_id_name" from table: "connections"
CREATE UNIQUE INDEX `idx_connection_org_id_project_id_name` ON `connections` (`org_id`, `project_id`, `name`) WHERE project_id is not null and deleted_at is null;
-- reverse: drop index "idx_connections_project_id" from table: "connections"
CREATE INDEX `idx_connections_project_id` ON `connections` (`project_id`);
-- reverse: drop index "idx_connection_org_id_name" from table: "connections"
CREATE UNIQUE INDEX `idx_connection_org_id_name` ON `connections` (`org_id`, `name`) WHERE project_id is null and deleted_at is null;
