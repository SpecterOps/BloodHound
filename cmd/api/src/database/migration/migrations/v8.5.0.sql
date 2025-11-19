-- OpenGraph Search feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp, 
    'opengraph_search',
    'OpenGraph Search',
    'Enable OpenGraph Search', 
    false,
    false)
ON CONFLICT DO NOTHING;