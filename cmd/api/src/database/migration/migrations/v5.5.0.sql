-- Add columns for audit_logs. remove actor_name as it will now be pulled from the users table directly.
ALTER TABLE audit_logs 
ADD COLUMN IF NOT EXISTS source VARCHAR(40) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS status VARCHAR(15) CHECK (status IN ('success', 'failure')) DEFAULT 'success',
DROP COLUMN IF EXISTS actor_name;
