-- Copyright 2025 Specter Ops, Inc.
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

-- is_generic column not actually needed.
ALTER TABLE ingest_tasks
DROP COLUMN IF EXISTS is_generic;

-- create explore_table_view feature flag, disable it, and make it non user-updatable.
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'explore_table_view',
        'Explore Table View',
        'Adds a layout option to the Explore page that will display all nodes in a table view. It also will automatically display the table when a cypher query returned only nodes.',
        false,
        false)
ON CONFLICT DO NOTHING;
