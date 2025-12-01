CREATE TABLE schemas (
  id SERIAL PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  version TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP
);

CREATE INDEX idx_schemas_deleted_at ON schemas(deleted_at);

CREATE TABLE schema_fields (
  id SERIAL PRIMARY KEY,
  schema_id INTEGER NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
  field_name TEXT NOT NULL,
  field_type TEXT NOT NULL,
  is_required BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP,
  UNIQUE(schema_id, field_name)
);

CREATE INDEX idx_schema_fields_deleted_at ON schema_fields(deleted_at);

CREATE TABLE schema_metadata (
  id SERIAL PRIMARY KEY,
  schema_id INTEGER UNIQUE NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
  description TEXT,
  author TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP
);

CREATE INDEX idx_schema_metadata_deleted_at ON schema_metadata(deleted_at);

CREATE TABLE schema_tags (
    id SERIAL PRIMARY KEY,
    schema_id INTEGER NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
    tag_key TEXT NOT NULL,
    tag_value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(schema_id, tag_key)
);

CREATE INDEX idx_schema_tags_deleted_at ON schema_tags(deleted_at);
