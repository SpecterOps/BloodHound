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

INSERT INTO parameters (key, name, description, value, created_at, updated_at)
SELECT
    'analysis.restrict_outbound_ntlm_default_value', 
    'Restrict Outbound NTLM Default Value', 
    'When enabled, any computer''s Restrict Outbound NTLM registry value is treated as Restricting if the registry doesn''t exist on that computer for NTLM edge processing. When disabled, treat the missing registry as Not Restricting.', '{ "enabled": false }',
    current_timestamp, current_timestamp
WHERE NOT EXISTS (SELECT 1 FROM parameters WHERE key = 'analysis.restrict_outbound_ntlm_default_value');
