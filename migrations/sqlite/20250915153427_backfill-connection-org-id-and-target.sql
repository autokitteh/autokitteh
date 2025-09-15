-- +goose Up
-- Set all current connections to be project scoped and update org id
UPDATE connections 
SET org_id = orgs.org_id, scope = 'project'
FROM projects 
JOIN orgs USING(org_id)
WHERE connections.project_id = projects.project_id AND projects.org_id IS NOT NULL;

-- +goose Down
