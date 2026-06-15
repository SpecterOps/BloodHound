-- Copyright 2026 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0
-- +goose Up

-- Delete any custom_node_kinds records whose kind_name is prefixed with 'Tag_'.
-- These names are reserved by the PZ system and shouldn't have slipped in here, but this is just in case.
DELETE FROM custom_node_kinds WHERE kind_name LIKE 'Tag_%';

-- Insert any kind_name values from custom_node_kinds that are not yet present in the
-- kind table. The WHERE NOT EXISTS filter avoids using up kind IDs (a SMALLSERIAL sequence) 
-- as would happen with ON CONFLICT DO NOTHING.
INSERT INTO kind (name)
SELECT kind_name FROM custom_node_kinds cnk
WHERE NOT EXISTS (
    SELECT 1 FROM kind WHERE kind.name = cnk.kind_name
);

-- Add the new kind_id FK column
ALTER TABLE custom_node_kinds ADD COLUMN IF NOT EXISTS kind_id INTEGER;

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


-- we have to recreate ensure_stubbed_custom_node_kind_for_ingest to use the new kind_id column instead of kind_name
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ensure_stubbed_custom_node_kind_for_ingest(node_kind_name TEXT, custom_node_kind_config JSONB) RETURNS void AS $$
DECLARE
  kind_row kind%rowtype;
  custom_node_kind_id INTEGER;
BEGIN
  -- Lock before checking existence so concurrent callers do not duplicate stubs.
  PERFORM pg_advisory_xact_lock(hashtext(node_kind_name));

  -- Try to find the existing kind
  SELECT * INTO kind_row FROM kind WHERE name = node_kind_name;

  -- If there is no kind, we should add it
  IF kind_row.id IS NULL THEN

    -- Yes, this is a select, but upsert_kind is a function that adds a kind
    SELECT * INTO kind_row FROM upsert_kind(node_kind_name);
  END IF;

  -- See if we have an existing custom node kind (note, we skip, do not update)
  SELECT id INTO custom_node_kind_id FROM custom_node_kinds WHERE kind_id = kind_row.id;
  IF custom_node_kind_id IS NULL THEN
    -- Insert based on the config
    INSERT INTO custom_node_kinds (kind_id, schema_node_kind_id, config)
    VALUES (kind_row.id, NULL, custom_node_kind_config);
  END IF;
END $$
LANGUAGE plpgsql;
-- +goose StatementEnd


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

-- reset the function to the previous version that relies on kind_name instead of kind_id
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ensure_stubbed_custom_node_kind_for_ingest(node_kind_name TEXT, custom_node_kind_config JSONB) RETURNS void AS $$
DECLARE
  kind_row kind%rowtype;
  custom_node_kind_id INTEGER;
BEGIN
  -- Lock before checking existence so concurrent callers do not duplicate stubs.
  PERFORM pg_advisory_xact_lock(hashtext(node_kind_name));

  -- Try to find the existing kind
  SELECT * INTO kind_row FROM kind WHERE name = node_kind_name;

  -- If there is no kind, we should add it
  IF kind_row.id IS NULL THEN

    -- Yes, this is a select, but upsert_kind is a function that adds a kind
    SELECT * INTO kind_row FROM upsert_kind(node_kind_name);
  END IF;

  -- See if we have an existing custom node kind (note, we skip, do not update)
  SELECT id INTO custom_node_kind_id FROM custom_node_kinds WHERE kind_name = node_kind_name;
  IF custom_node_kind_id IS NULL THEN
    -- Insert based on the config 
    INSERT INTO custom_node_kinds (kind_name, schema_node_kind_id, config)
    VALUES (node_kind_name, NULL, custom_node_kind_config);
  END IF;
END $$
LANGUAGE plpgsql;
-- +goose StatementEnd