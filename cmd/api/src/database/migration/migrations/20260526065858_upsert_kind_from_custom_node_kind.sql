-- +goose Up

-- Delete any custom_node_kinds records whose kind_name is prefixed with 'Tag_'.
-- These names are reserved by the PZ system and shouldn't have slipped in here, but this is just in case.
DELETE FROM custom_node_kinds WHERE kind_name LIKE 'Tag_%';

-- Upsert any kind_name values from custom_node_kinds that are not yet present in the
-- kind table. kind.id is SMALLSERIAL (max 32,767), so only genuinely missing names are inserted.
INSERT INTO kind (name)
SELECT kind_name FROM custom_node_kinds
ON CONFLICT (name) DO NOTHING;

-- Add the new kind_id FK column (SMALLINT matches kind.id SMALLSERIAL).
ALTER TABLE custom_node_kinds ADD COLUMN IF NOT EXISTS kind_id SMALLINT;

-- Back-fill the new kind_id column
UPDATE custom_node_kinds
SET kind_id = kind.id
FROM kind
WHERE kind.name = custom_node_kinds.kind_name;

-- remove any rows that could not be resolved to a kind_id. This should never happen but we'll kill them just in case so the following steps don't fail due to nulls in kind_id. 
DELETE FROM custom_node_kinds WHERE kind_id IS NULL;

-- Enforce NOT NULL and add the FK constraint
ALTER TABLE custom_node_kinds ALTER COLUMN kind_id SET NOT NULL;
ALTER TABLE custom_node_kinds
    ADD CONSTRAINT fk_custom_node_kinds_kind_id FOREIGN KEY (kind_id) REFERENCES kind (id);

-- kind_name previously enforced one unique constraint per row, but now that we've switched to kind_id we need to enforce the same uniqueness on kind_id instead
ALTER TABLE custom_node_kinds
    ADD CONSTRAINT custom_node_kinds_kind_id_key UNIQUE (kind_id);

-- drop the now-superseded kind_name column.
-- dropping the column implicitly removes the custom_node_kinds_kind_name_key unique constraint.
ALTER TABLE custom_node_kinds DROP COLUMN kind_name;

-- +goose Down

-- Re-add the kind_name column (nullable initially so it can be populated).
ALTER TABLE custom_node_kinds ADD COLUMN IF NOT EXISTS kind_name VARCHAR(256);

-- Populate kind_name from the kind table via kind_id.
UPDATE custom_node_kinds
SET kind_name = kind.name
FROM kind
WHERE kind.id = custom_node_kinds.kind_id;

-- Enforce NOT NULL and restore the original unique constraint name.
ALTER TABLE custom_node_kinds ALTER COLUMN kind_name SET NOT NULL;
ALTER TABLE custom_node_kinds
    ADD CONSTRAINT custom_node_kinds_kind_name_key UNIQUE (kind_name);

-- Drop the FK and unique constraints on kind_id, then drop the column.
ALTER TABLE custom_node_kinds DROP CONSTRAINT IF EXISTS fk_custom_node_kinds_kind_id;
ALTER TABLE custom_node_kinds DROP CONSTRAINT IF EXISTS custom_node_kinds_kind_id_key;
ALTER TABLE custom_node_kinds DROP COLUMN kind_id;

