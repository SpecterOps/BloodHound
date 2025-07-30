-- Targeted Access Control Feature Flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'targeted_access_control',
        'Targeted Access Control',
        'Enable power users and admins to set targeted access controls on users',
        false,
        false)
ON CONFLICT DO NOTHING;
