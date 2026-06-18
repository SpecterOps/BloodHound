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

-- +goose Up
INSERT INTO asset_groups (name, tag, system_group, created_at, updated_at)
VALUES ('Decoy', 'decoy', true, current_timestamp, current_timestamp)
ON CONFLICT (tag) DO NOTHING;

WITH decoy_kind AS (
    INSERT INTO kind (name)
    VALUES ('Tag_Decoy')
    ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
    RETURNING id
)
INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by, analysis_enabled, glyph)
SELECT id, 4, 'Decoy', 'Accounts that should be excluded from predefined attack path queries.', current_timestamp, 'BloodHound', current_timestamp, 'BloodHound', false, 'mask'
FROM decoy_kind
ON CONFLICT DO NOTHING;

-- +goose Down
WITH decoy_tag AS (
    SELECT id, kind_id
    FROM asset_group_tags
    WHERE type = 4
        AND name = 'Decoy'
        AND created_by = 'BloodHound'
        AND updated_by = 'BloodHound'
        AND analysis_enabled = false
        AND glyph = 'mask'
), decoy_selectors AS (
    SELECT id
    FROM asset_group_tag_selectors
    WHERE asset_group_tag_id IN (SELECT id FROM decoy_tag)
), deleted_selector_nodes AS (
    DELETE FROM asset_group_tag_selector_nodes
    WHERE selector_id IN (SELECT id FROM decoy_selectors)
    RETURNING selector_id
), deleted_selector_seeds AS (
    DELETE FROM asset_group_tag_selector_seeds
    WHERE selector_id IN (SELECT id FROM decoy_selectors)
    RETURNING selector_id
), deleted_selectors AS (
    DELETE FROM asset_group_tag_selectors
    WHERE id IN (SELECT id FROM decoy_selectors)
    RETURNING id
), deleted_history AS (
    DELETE FROM asset_group_history
    WHERE asset_group_tag_id IN (SELECT id FROM decoy_tag)
    RETURNING id
), deleted_tag AS (
    DELETE FROM asset_group_tags
    WHERE id IN (SELECT id FROM decoy_tag)
    RETURNING kind_id
)
DELETE FROM kind
WHERE id IN (SELECT kind_id FROM deleted_tag)
    AND NOT EXISTS (
        SELECT 1
        FROM asset_group_tags
        WHERE asset_group_tags.kind_id = kind.id
    );

DELETE FROM asset_groups
WHERE tag = 'decoy'
    AND name = 'Decoy'
    AND system_group = true;
