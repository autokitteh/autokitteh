-- +goose Up
CREATE VIEW deployment_stats AS
SELECT 
    d.deployment_id,
    d.project_id,
    d.build_id,
    d.state,
    d.created_at,
    d.updated_at,
    COUNT(CASE WHEN s.current_state_type = 0 THEN 1 END) as created_count,
    COUNT(CASE WHEN s.current_state_type = 1 THEN 1 END) as running_count,
    COUNT(CASE WHEN s.current_state_type = 2 THEN 1 END) as error_count,
    COUNT(CASE WHEN s.current_state_type = 3 THEN 1 END) as completed_count,
    COUNT(CASE WHEN s.current_state_type = 4 THEN 1 END) as stopped_count
FROM deployments d
LEFT JOIN sessions s ON d.deployment_id = s.deployment_id
WHERE d.deleted_at IS NULL
GROUP BY d.deployment_id, d.project_id, d.build_id, d.state, d.created_at, d.updated_at; 

-- +goose Down
DROP VIEW IF EXISTS deployment_stats;
