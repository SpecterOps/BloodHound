INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (
           current_timestamp,
           current_timestamp,
           'finished_jobs_log_v2',
           'Finished Jobs Log Update',
           'An updated Finished Jobs Log with filtering and more info.',
           false,
           false
       )
ON CONFLICT DO NOTHING;
