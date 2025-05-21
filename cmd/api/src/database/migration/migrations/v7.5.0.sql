-- is_generic column not actually needed.
ALTER TABLE ingest_tasks
DROP COLUMN IF EXISTS is_generic;

IF NOT EXISTS (SELECT 1 FROM parameters WHERE key = 'analysis.restrict_outbound_ntlm_default_value')
BEGIN
    INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES (
        'analysis.restrict_outbound_ntlm_default_value', 
        'Restrict Outbound NTLM Default Value', 
        'When enabled, any computer''s Restrict Outbound NTLM registry value is treated as Restricting if the registry doesn''t exist on that computer for NTLM edge processing. When disabled, treat the missing registry as Not Restricting.', '{ "enabled": false }',
        current_timestamp, current_timestamp
    );
END