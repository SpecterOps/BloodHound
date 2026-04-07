-- Update the 'auth_tokens' table adding created_by column
ALTER TABLE auth_tokens
  ADD COLUMN IF NOT EXISTS created_by text;

-- Add the foreign key if it doesn't already exist
DO $$
  BEGIN
    IF NOT EXISTS (
      SELECT 1
      FROM pg_constraint
      WHERE conname = 'fk_auth_tokens_created_by'
    ) THEN
      ALTER TABLE auth_tokens
        ADD CONSTRAINT fk_auth_tokens_created_by
          FOREIGN KEY (created_by)
            REFERENCES users(id);
    END IF;
  END $$;
