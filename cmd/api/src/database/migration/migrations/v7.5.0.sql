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

 -- Add Tier Management Parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('analysis.tiering', 'Multi-Tier Analysis Configuration', 'This configuration parameter determines the limits of tiering with respect to analysis', '{"tier_limit": 1, "label_limit": 0, "multi_tier_analysis_enabled": false}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
