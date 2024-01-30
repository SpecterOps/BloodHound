ALTER TABLE IF EXISTS audit_logs
DROP CONSTRAINT IF EXISTS audit_logs_status_check;

ALTER TABLE IF EXISTS audit_logs
ADD CONSTRAINT status_check
CHECK (status IN ('intent', 'success', 'failure'));

ALTER TABLE IF EXISTS audit_logs
ALTER COLUMN status SET DEFAULT 'intent';