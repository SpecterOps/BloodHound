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


-- Set all_environments to true for existing users
UPDATE users SET all_environments = true;
-- Rename environment to environment_id to prepare for data partitioning, if the column does not exist then we throw away the error for idempotence
DO
$$
    BEGIN
        ALTER TABLE environment_access_control
            RENAME COLUMN environment_id TO environment_id;
    EXCEPTION
        WHEN undefined_column THEN
    END;
$$;
