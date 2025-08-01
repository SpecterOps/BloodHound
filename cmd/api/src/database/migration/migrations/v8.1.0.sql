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

-- File Ingest Details
ALTER TABLE ingest_jobs ADD COLUMN IF NOT EXISTS task_info jsonb NOT NULL DEFAULT '{"completed_tasks": []}'::jsonb;
ALTER TABLE ingest_tasks ADD COLUMN IF NOT EXISTS provided_file_name text NOT NULL DEFAULT '';
