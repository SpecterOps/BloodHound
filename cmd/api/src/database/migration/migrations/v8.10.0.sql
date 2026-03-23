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
-- Drop the compound unique constraint on schema_environments (environment_kind_id, source_kind_id)
-- and add a unique constraint on just environment_kind_id

-- Rename environment_targeted_access_control feature flag to 
UPDATE feature_flags
SET key         = 'openhound_support',
    name        = 'OpenHound Support',
    description = 'Enable creation and communication with OpenHound platform'
WHERE key = 'opengraph_collector_platform_support';
