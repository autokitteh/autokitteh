-- +goose Up
-- Set all current connections to be project scoped and update org id
UPDATE connections c
SET org_id = o.org_id, scope = 'project'
FROM projects p
JOIN orgs o USING(org_id)
WHERE c.project_id = p.project_id AND p.org_id IS NOT NULL;

-- +goose Down
