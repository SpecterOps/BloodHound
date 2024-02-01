ALTER TABLE IF EXISTS audit_logs
  RENAME COLUMN source TO source_ip_address;

ALTER TABLE IF EXISTS audit_logs
  DROP CONSTRAINT IF EXISTS audit_logs_status_check,
  ADD CONSTRAINT status_check
  CHECK (status IN ('intent', 'success', 'failure')),
  ALTER COLUMN status SET DEFAULT 'intent',
  ALTER COLUMN source_ip_address TYPE TEXT,
  ADD COLUMN IF NOT EXISTS commit_id TEXT;

-- Add indices for scalability
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_email ON audit_logs USING btree (actor_email);
CREATE INDEX IF NOT EXISTS idx_audit_logs_source_ip_address ON audit_logs USING btree (source_ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs USING btree (status);
