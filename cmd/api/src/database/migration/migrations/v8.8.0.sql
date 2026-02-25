-- Drop the compound unique constraint on schema_environments (environment_kind_id, source_kind_id)
-- and add a unique constraint on just environment_kind_id
ALTER TABLE IF EXISTS schema_environments
    DROP CONSTRAINT IF EXISTS schema_environments_environment_kind_id_source_kind_id_key;

DO $$
    BEGIN
        IF NOT EXISTS (
                      SELECT 1
                      FROM pg_constraint
                      WHERE conname = 'schema_environments_environment_kind_id_key'
        ) THEN
            ALTER TABLE schema_environments
                ADD CONSTRAINT schema_environments_environment_kind_id_key UNIQUE (environment_kind_id);
        END IF;
    END$$;
