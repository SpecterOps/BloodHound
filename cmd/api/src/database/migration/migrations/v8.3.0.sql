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

CREATE INDEX IF NOT EXISTS idx_agt_history_actor ON asset_group_history USING btree (actor);
CREATE INDEX IF NOT EXISTS idx_agt_history_action ON asset_group_history USING btree (action);
CREATE INDEX IF NOT EXISTS idx_agt_history_target ON asset_group_history USING btree (target);
CREATE INDEX IF NOT EXISTS idx_agt_history_email ON asset_group_history USING btree (email);
CREATE INDEX IF NOT EXISTS idx_agt_history_env_id ON asset_group_history USING btree (environment_id);
CREATE INDEX IF NOT EXISTS idx_agt_history_created_at ON asset_group_history USING btree (created_at);

DO $$
BEGIN
		IF
      (SELECT enabled FROM feature_flags WHERE key  = 'tier_management_engine') = false
    THEN
       -- Delete custom selectors
       DELETE FROM asset_group_tag_selectors WHERE is_default = false AND asset_group_tag_id IN ((SELECT id FROM asset_group_tags WHERE position = 1), (SELECT id FROM asset_group_tags WHERE type = 3));

       -- Re-Migrate existing Tier Zero selectors
       WITH inserted_selector AS (
         INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
         SELECT (SELECT id FROM asset_group_tags WHERE position = 1), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, 2
         FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
         WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
         RETURNING id, description
         )
       INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

      -- Re-Migrate existing Owned selectors
      WITH inserted_selector AS (
        INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
        SELECT (SELECT id FROM asset_group_tags WHERE type = 3), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, 0
        FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
        WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
          RETURNING id, description
          )
      INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

END IF;
END;
$$;
