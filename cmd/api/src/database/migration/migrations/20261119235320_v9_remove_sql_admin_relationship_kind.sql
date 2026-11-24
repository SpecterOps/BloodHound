-- Copyright 2026 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
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
DELETE FROM schema_relationship_kinds AS relationship_kind
USING kind, schema_extensions AS extension
WHERE relationship_kind.kind_id = kind.id
  AND relationship_kind.schema_extension_id = extension.id
  AND kind.name = 'SQLAdmin'
  AND extension.name = 'AD'
  AND extension.is_builtin = TRUE;

-- +goose Down
INSERT INTO schema_relationship_kinds (
    schema_extension_id, kind_id, description, is_traversable, created_at, updated_at
)
SELECT extension.id, kind.id, '', TRUE, now(), now()
FROM kind
CROSS JOIN schema_extensions AS extension
WHERE kind.name = 'SQLAdmin'
  AND extension.name = 'AD'
  AND extension.is_builtin = TRUE
ON CONFLICT (kind_id) DO NOTHING;
