-- OpenGraph Pathfinding feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp,
    'opengraph_pathfinding',
    'OpenGraph Pathfinding',
    'Enable OpenGraph Pathfinding',
    false,
    false)
ON CONFLICT DO NOTHING;

-- Set `opengraph_search` feature flag to enabled by default
UPDATE feature_flags SET enabled = true WHERE key = 'opengraph_search';