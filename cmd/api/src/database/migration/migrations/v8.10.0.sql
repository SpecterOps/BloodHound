-- Add support_account flag to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS support_account BOOL DEFAULT FALSE;
