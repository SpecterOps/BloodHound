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

-- Add analysis_enabled flag to asset_group_tags
ALTER TABLE asset_group_tags ADD COLUMN IF NOT EXISTS analysis_enabled BOOL;

-- Set analysis_enabled to true for tier zero and false for other tiers
UPDATE asset_group_tags SET analysis_enabled = position = 1 WHERE type = 1 AND analysis_enabled IS NULL;

-- Add EULA custom text
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('eula.custom_text', 'EULA Custom Text', 'This configuration parameter overrides the EULA agreement text with provided text.', '{"custom_text": ""}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Add Auth Session TTL Hours
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('auth.session_ttl_hours', 'Auth Session TTL Hours', 'This configuration parameter determines the length of time in hours a logged in session stays active before expiration.', '{"hours": 8}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Add NTLM default value processing
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ( 'analysis.restrict_outbound_ntlm_default_value','Restrict Outbound NTLM Default Value','When enabled, any computer''s Restrict Outbound NTLM registry value is treated as Restricting if the registry doesn''t exist on that computer for NTLM edge processing. When disabled, treat the missing registry as Not Restricting.', '{ "enabled": false }', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
