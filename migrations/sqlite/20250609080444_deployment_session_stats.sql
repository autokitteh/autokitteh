-- +goose Up
CREATE VIEW deployment_session_stats AS
SELECT 
    deployment_id,
    COUNT(CASE WHEN current_state_type = 1 THEN 1 END) AS created_count,
    COUNT(CASE WHEN current_state_type = 2 THEN 1 END) AS running_count,
    COUNT(CASE WHEN current_state_type = 3 THEN 1 END) AS error_count,
    COUNT(CASE WHEN current_state_type = 4 THEN 1 END) AS completed_count,
    COUNT(CASE WHEN current_state_type = 5 THEN 1 END) AS stopped_count
FROM sessions
GROUP BY deployment_id;

-- +goose Down
DROP VIEW IF EXISTS deployment_session_stats;
