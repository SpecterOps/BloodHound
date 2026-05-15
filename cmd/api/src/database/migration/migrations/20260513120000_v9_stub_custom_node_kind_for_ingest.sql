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
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ensure_stubbed_custom_node_kind_for_ingest(node_kind_name TEXT, custom_node_kind_config JSONB) RETURNS void AS $$
DECLARE
  kind_row kind%rowtype;
  custom_node_kind_id INTEGER;
BEGIN
  -- Lock before checking existence so only the caller that first observes a new
  -- kind creates its custom_node_kinds stub.
  PERFORM pg_advisory_xact_lock(hashtext(node_kind_name));

  SELECT * INTO kind_row FROM kind WHERE name = node_kind_name;
  IF kind_row.id IS NOT NULL THEN
    RETURN;
  END IF;

  SELECT * INTO kind_row FROM upsert_kind(node_kind_name);

  SELECT id INTO custom_node_kind_id FROM custom_node_kinds WHERE kind_name = node_kind_name;
  IF custom_node_kind_id IS NULL THEN
    INSERT INTO custom_node_kinds (kind_name, schema_node_kind_id, config)
    VALUES (node_kind_name, NULL, custom_node_kind_config);
  END IF;
END $$
LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS ensure_stubbed_custom_node_kind_for_ingest(TEXT, JSONB);
