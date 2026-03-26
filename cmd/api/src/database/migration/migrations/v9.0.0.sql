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


-- Add support_account flag to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS support_account BOOL DEFAULT FALSE;

-- Rename opengraph_collector_platform_support feature flag to openhound_support
UPDATE feature_flags
SET key         = 'openhound_support',
    name        = 'OpenHound Support',
    description = 'Enable creation and communication with OpenHound platform'
WHERE key = 'opengraph_collector_platform_support';

-- Migrate environment_id to environmentid in node properties
DO $$
BEGIN
    IF EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'node'
    ) THEN
        UPDATE node
        SET properties = properties - 'environment_id' || jsonb_build_object('environmentid', properties->>'environment_id')
        WHERE properties ? 'environment_id' AND NOT properties ? 'environmentid';
    END IF;
END $$;

-- Add index on node properties.environmentid field
DO $$
BEGIN
    IF EXISTS (
        SELECT FROM information_schema.tables
        WHERE table_schema = 'public'
        AND table_name = 'node'
    ) THEN
        CREATE INDEX IF NOT EXISTS node_environmentid_idx ON node USING btree ((properties->>'environmentid'));
    END IF;
END $$;