-- Add new columns for audit_logs
ALTER TABLE audit_logs 
ADD COLUMN IF NOT EXISTS actor_email VARCHAR(330) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS source VARCHAR(40) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS status VARCHAR(15) CHECK (status IN ('success', 'failure')) DEFAULT 'success';

-- Populate actor_email for existing records by looking up the email address from the users table
UPDATE audit_logs
SET actor_email = COALESCE((SELECT email_address FROM users WHERE audit_logs.actor_id = users.id), 'unknown');
