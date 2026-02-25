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
ALTER TABLE IF EXISTS schema_environments
    DROP CONSTRAINT IF EXISTS schema_environments_environment_kind_id_source_kind_id_key;

DO $$
    BEGIN
        IF NOT EXISTS (
                      SELECT 1
                      FROM pg_constraint
                      WHERE conname = 'schema_environments_environment_kind_id_key'
        ) THEN
            ALTER TABLE schema_environments
                ADD CONSTRAINT schema_environments_environment_kind_id_key UNIQUE (environment_kind_id);
        END IF;
    END$$;
