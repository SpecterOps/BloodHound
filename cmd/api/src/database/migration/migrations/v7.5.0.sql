-- is_generic column not actually needed.
ALTER TABLE ingest_tasks
DROP COLUMN IF EXISTS is_generic;