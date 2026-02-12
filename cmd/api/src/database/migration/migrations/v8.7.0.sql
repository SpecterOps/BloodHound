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

-- OpenGraph Extension Management feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'opengraph_extension_management',
        'OpenGraph Extension Management',
        'Enable OpenGraph Extension Management',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Scheduled Analysis Configuration feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'scheduled_analysis_configuration',
        'Scheduled Analysis Configuration',
        'Enable Scheduled Analysis Configuration form in the UI',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Remove pathfinding feature flag. We are keying off of opengraph_extension_management instead
DELETE FROM feature_flags WHERE key = 'opengraph_pathfinding';

-- upsert_kind is a stop-gap function used in open graph schema node,
-- relationship and source kind creation.
-- This function is needed to stave off kind table id exhaustion while maintaining
-- an acceptable level of performance. An exception will be raised if
-- the function fails to insert the kind after 5 attempts.
--
-- Underlying Issues: The kind table's id column is a SMALLINT, Postgres will
-- increase a SERIAL PK on conflict even if DO NOTHING or DO UPDATE is used and
-- a table lock on the Kind table greatly decreases performance.
CREATE OR REPLACE FUNCTION upsert_kind(node_kind_name TEXT) RETURNS kind AS $$
DECLARE
    kind_row kind%rowtype;
BEGIN
    -- Try to find existing kind based on name
    SELECT * INTO kind_row FROM kind WHERE name = node_kind_name;
    IF kind_row IS NOT NULL THEN
        RETURN kind_row;
    END IF;
    -- Insert with retry, handles the race condition where two transactions try to add the same kind at the same time
    FOR i IN 1..5 LOOP
        BEGIN
            INSERT INTO kind (name)
            VALUES (node_kind_name)
            RETURNING * INTO kind_row;
            RETURN kind_row;
        EXCEPTION
            WHEN unique_violation THEN
                -- Check if the insert conflict was for same kind
                SELECT * INTO kind_row FROM kind WHERE name = node_kind_name;
                IF kind_row IS NOT NULL THEN
                    RETURN kind_row;
                END IF;
        END;
    END LOOP;
    -- failed to insert kind after 5 retries
    RAISE EXCEPTION 'failed to insert kind % after 5 retries', node_kind_name;
END;
$$ LANGUAGE plpgsql;