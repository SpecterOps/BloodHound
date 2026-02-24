-- Remigrate old custom AGI selectors to PZ selectors for any instances without PZ feature flag enabled
DO $$
  BEGIN
		IF
      (SELECT enabled FROM feature_flags WHERE key  = 'tier_management_engine') = false
    THEN
       -- Delete custom selectors
       DELETE FROM asset_group_tag_selectors WHERE is_default = false AND asset_group_tag_id IN ((SELECT id FROM asset_group_tags WHERE position = 1), (SELECT id FROM asset_group_tags WHERE type = 3));

       -- Re-Migrate existing Tier Zero selectors
       WITH inserted_selector AS (
         INSERT INTO asset_group_tag_selectors 
            (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, disabled_by, disabled_at, allow_disable, auto_certify)
         SELECT 
            (SELECT id FROM asset_group_tags WHERE position = 1), 
            current_timestamp, 
            'BloodHound', 
            current_timestamp, 
            'BloodHound', 
            s.name, 
            s.selector, 
            false, 
            CASE WHEN is_disabled THEN 'BloodHound' END AS disabled_by,
            CASE WHEN is_disabled THEN current_timestamp END as disabled_at,
            true, 
            2
         FROM asset_group_selectors s 
            JOIN asset_groups ag ON ag.id = s.asset_group_id
            -- Computes is_disabled once per row to be reused for disabled_by and disabled_at. 
            -- The lateral subquery returns exactly one row, so the cross join does not multiply rows.
            CROSS JOIN LATERAL ( 
                SELECT EXISTS (
					SELECT 1
	                FROM node n
    	            JOIN kind k on k.id = ANY(n.kind_ids::integer[])
        	        WHERE properties->>'objectid' = s.selector
            	    AND (k.name = 'OU' OR k.name = 'Container')
				) as is_disabled
            ) disable_check
         WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
            RETURNING id, description
         )
       INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

      -- Re-Migrate existing Owned selectors
      WITH inserted_selector AS (
        INSERT INTO asset_group_tag_selectors 
            (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, disabled_by, disabled_at, allow_disable, auto_certify)
        SELECT 
            (SELECT id FROM asset_group_tags WHERE type = 3), 
            current_timestamp, 
            'BloodHound', 
            current_timestamp, 
            'BloodHound', 
            s.name, 
            s.selector, 
            false,
            CASE WHEN is_disabled THEN 'BloodHound' END AS disabled_by,
            CASE WHEN is_disabled THEN current_timestamp END as disabled_at,
            true, 
            0
        FROM asset_group_selectors s 
            JOIN asset_groups ag ON ag.id = s.asset_group_id
            -- Computes is_disabled once per row to be reused for disabled_by and disabled_at. 
            -- The lateral subquery returns exactly one row, so the cross join does not multiply rows.
            CROSS JOIN LATERAL ( 
                SELECT EXISTS (
					SELECT 1
	                FROM node n
    	            JOIN kind k on k.id = ANY(n.kind_ids::integer[])
        	        WHERE properties->>'objectid' = s.selector
            	    AND (k.name = 'OU' OR k.name = 'Container')
				) as is_disabled
            ) disable_check
        WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
          RETURNING id, description
        )
      INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;
    END IF;
  END;
$$;

-- Before we add unique constraint, rename any duplicates with `_X` to prevent constraint failing
WITH duplicate_selectors AS (
  SELECT id, name, asset_group_tag_id, ROW_NUMBER() OVER (PARTITION BY name, asset_group_tag_id ORDER BY id) AS rowNumber
  FROM asset_group_tag_selectors
)
UPDATE asset_group_tag_selectors agts
SET name = agts.name || '_' || rowNumber FROM duplicate_selectors
WHERE agts.id = duplicate_selectors.id AND duplicate_selectors.rowNumber > 1;
