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

-- Mark existing source_kinds table old
ALTER TABLE source_kinds RENAME CONSTRAINT source_kinds_pkey TO source_kinds_pkey_old;
ALTER TABLE source_kinds RENAME CONSTRAINT source_kinds_kind_id_key to source_kinds_kind_id_key_old;
ALTER TABLE source_kinds RENAME CONSTRAINT source_kinds_kind_id_fkey TO source_kinds_kind_id_fkey_old;
ALTER TABLE source_kinds RENAME TO source_kinds_old;
ALTER SEQUENCE source_kinds_id_seq RENAME TO source_kinds_old_id_seq;

-- Swap source_kind id to integer
CREATE TABLE IF NOT EXISTS source_kinds (
  id SERIAL,
  active BOOLEAN DEFAULT true NOT NULL,
  kind_id INTEGER NOT NULL REFERENCES kind(id),
  PRIMARY KEY(id),
  UNIQUE(kind_id)
);

-- Copy old source kinds to new source kinds starting with schema tied source kinds, then the rest
INSERT INTO source_kinds (id, active, kind_id)
  SELECT se.source_kind_id,  sk.active, sk.kind_id FROM schema_environments se JOIN source_kinds_old sk ON se.source_kind_id = sk.id;
SELECT setval('source_kinds_id_seq', COALESCE((SELECT MAX(id) FROM source_kinds), 1), true);
INSERT INTO source_kinds (active, kind_id)
  SELECT active, kind_id FROM source_kinds_old WHERE id NOT IN (SELECT source_kind_id FROM schema_environments);

-- Drop old source kinds
ALTER TABLE schema_environments DROP CONSTRAINT schema_environments_source_kind_id_fkey;
DROP TABLE source_kinds_old;

-- Add source_kind_id constraint back to schema_environments
ALTER TABLE schema_environments ADD CONSTRAINT schema_environments_source_kind_id_fkey FOREIGN KEY (source_kind_id) REFERENCES source_kinds(id) ON DELETE RESTRICT;
