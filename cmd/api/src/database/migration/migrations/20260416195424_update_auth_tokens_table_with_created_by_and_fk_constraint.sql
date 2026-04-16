-- Update the 'auth_tokens' table adding created_by column
-- As of current, we don't have a way to backfill the data, so we are leaving this field optional for now
-- +goose Up
ALTER TABLE auth_tokens
  ADD COLUMN IF NOT EXISTS created_by text;

-- Create fk_auth_tokens_created_by to auth_tokens referencing users.id if it doesn't exist
-- +goose StatementBegin
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
-- +goose StatementEnd

-- +goose Down
ALTER TABLE auth_tokens
  DROP CONSTRAINT IF EXISTS fk_auth_tokens_created_by;

ALTER TABLE auth_tokens
  DROP COLUMN IF EXISTS created_by;
