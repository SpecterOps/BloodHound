-- Prepend all found duplicate emails on user table with user id in preparation for unique constraint
UPDATE users SET email_address = id || '-' || lower(email_address) where lower(email_address) in (SELECT distinct(lower(email_address)) FROM users GROUP BY lower(email_address) HAVING count(lower(email_address)) > 1);

-- Add unique constraint on user emails
ALTER TABLE IF EXISTS users
  DROP CONSTRAINT IF EXISTS users_email_address_key;
ALTER TABLE IF EXISTS users
  ADD CONSTRAINT users_email_address_key UNIQUE (email_address);
