ALTER TABLE IF EXISTS audit_logs
DROP CONSTRAINT IF EXISTS audit_logs_status_check,
ADD CONSTRAINT status_check
CHECK (status IN ('intent', 'success', 'failure')),
ALTER COLUMN status SET DEFAULT 'intent',
ADD COLUMN IF NOT EXISTS commit_id TEXT;