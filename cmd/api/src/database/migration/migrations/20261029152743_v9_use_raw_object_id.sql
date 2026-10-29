-- +goose Up
UPDATE feature_flags
SET enabled = TRUE
WHERE
  key = 'use_raw_object_id'
  -- this exists only on the new installation
  AND NOT EXISTS (SELECT 1 FROM installations);

-- +goose Down
-- Intentionaly preserve the flag's current state. Disabling it after
-- ingestion may cause object ID collisions and incorrect relationship resolution
