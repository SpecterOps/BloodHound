INSERT INTO asset_groups (name, tag, system_group)
SELECT 'Owned', 'owned', true
WHERE NOT EXISTS (SELECT 1 FROM asset_groups WHERE tag='owned')
