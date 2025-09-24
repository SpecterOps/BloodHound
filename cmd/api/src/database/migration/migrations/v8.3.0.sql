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
            RENAME COLUMN environment TO environment_id;
    EXCEPTION
        WHEN undefined_column THEN
    END;
$$;

-- This migration changes the auto_certify column type from a boolean to an integer type
-- Then it converts the previous boolean values into enum-like integer values of 0, 1, or 2
DO $$
	BEGIN
		IF (
            SELECT data_type
            FROM information_schema.columns
            WHERE table_name = 'asset_group_tag_selectors' AND column_name = 'auto_certify'
        ) = 'boolean' THEN
		    ALTER TABLE asset_group_tag_selectors ADD COLUMN auto_certify_int INTEGER NOT NULL DEFAULT 0;
		
		    -- 0 means disabled 
		    -- 1 is enabled for all objects (seeds, children, parents) 
		    -- 2 is enabled for seeds only
		    UPDATE asset_group_tag_selectors selectors
		    SET auto_certify_int = CASE
		        WHEN selectors.is_default THEN 2                      
		        WHEN selectors.auto_certify = TRUE THEN 1             
		        WHEN EXISTS (                                         
		            SELECT *
		            FROM asset_group_tag_selector_seeds seeds
		            WHERE seeds.type = 1 AND seeds.selector_id = selectors.id
		        ) THEN 2
		        ELSE 0
		    END
		    FROM asset_group_tags tags
		    WHERE selectors.asset_group_tag_id = tags.id;
		
		    ALTER TABLE asset_group_tag_selectors DROP COLUMN auto_certify;
		    ALTER TABLE asset_group_tag_selectors RENAME COLUMN auto_certify_int TO auto_certify;
		END IF;
	END;
$$;
