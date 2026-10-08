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
INSERT INTO custom_node_kinds (kind_id, schema_node_kind_id, config)
SELECT id, NULL, '{"icon": {"name": "layer-group", "color": "#ffb7ce", "type": "font-awesome"}}'
FROM kind
WHERE name = 'Zone'
ON CONFLICT (kind_id) DO NOTHING;

-- +goose Down
DELETE FROM custom_node_kinds
WHERE kind_id = (SELECT id FROM kind WHERE name = 'Zone');
