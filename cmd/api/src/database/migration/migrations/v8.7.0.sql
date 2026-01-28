-- Add ETAC configuration parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES (
           'auth.environment_targeted_access_control',
           'Environment Targeted Access Control Configuration',
           'This configuration parameter is used to enable and disable features for environment targeted access controls',
           '{"enabled": false}',
           current_timestamp,
           current_timestamp
       )
ON CONFLICT DO NOTHING;