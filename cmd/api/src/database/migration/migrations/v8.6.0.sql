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

-- OpenGraph Schema Extension Management feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'opengraph_extension_management',
        'OpenGraph Schema Extension Management',
        'Enable OpenGraph Schema Extension Management',
        false,
        false)
ON CONFLICT DO NOTHING;

ALTER TABLE IF EXISTS schema_environments
    DROP CONSTRAINT IF EXISTS schema_environments_source_kind_id_fkey;

ALTER TABLE IF EXISTS schema_environments
    ADD CONSTRAINT schema_environments_source_kind_id_fkey
    FOREIGN KEY (source_kind_id) REFERENCES source_kinds(id);


-- OpenGraph Findings feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp,
    'opengraph_findings',
    'OpenGraph Findings',
    'Enable OpenGraph Findings',
    false,
    false)
ON CONFLICT DO NOTHING;