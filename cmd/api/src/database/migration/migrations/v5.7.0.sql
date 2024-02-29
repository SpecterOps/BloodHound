ALTER TABLE ingest_tasks
ADD COLUMN IF NOT EXISTS file_type integer DEFAULT 0;
