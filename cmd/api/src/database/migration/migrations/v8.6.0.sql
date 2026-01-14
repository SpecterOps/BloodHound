-- OpenGraph Findings feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp,
    'opengraph_findings',
    'OpenGraph Findings',
    'Enable OpenGraph Findings',
    false,
    false)
ON CONFLICT DO NOTHING;