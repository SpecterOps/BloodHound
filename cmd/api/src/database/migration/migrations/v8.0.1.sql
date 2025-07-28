ALTER TABLE ingest_jobs ADD COLUMN IF NOT EXISTS task_info jsonb NOT NULL DEFAULT '{"completed_tasks": []}'::jsonb;
ALTER TABLE ingest_tasks ADD COLUMN IF NOT EXISTS provided_file_name text NOT NULL DEFAULT 'UnknownFileName';
ALTER TABLE client_ingest_tasks ADD COLUMN IF NOT EXISTS provided_file_name text DEFAULT 'UnknownFileName';
