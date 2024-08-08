-- Copyright 2024 Specter Ops, Inc.
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

UPDATE asset_groups
SET tag = REGEXP_REPLACE(tag, '\s', '', 'g');

ALTER TABLE ingest_tasks
ADD COLUMN IF NOT EXISTS file_type integer DEFAULT 0;

-- Add db wipe permission
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('db', 'Wipe', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- grant admin dp wipe permission
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'db'  and permissions.name = 'Wipe')) ON CONFLICT DO NOTHING;

-- Add clear graph db FF
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'clear_graph_data', 'Clear Graph Data', 'Enables the ability to delete all nodes and edges from the graph database.', true, false) ON CONFLICT DO NOTHING;
