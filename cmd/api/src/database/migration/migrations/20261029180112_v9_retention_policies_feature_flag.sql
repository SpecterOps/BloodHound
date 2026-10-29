-- +goose Up
INSERT INTO feature_flags (
    created_at, updated_at, key, name, description, enabled, user_updatable
)
VALUES (
    current_timestamp,
    current_timestamp,
    'retention_policies',
    'Retention Policies',
    'Enables retention policy management.',
    false,
    false
)
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM feature_flags WHERE key = 'retention_policies';
