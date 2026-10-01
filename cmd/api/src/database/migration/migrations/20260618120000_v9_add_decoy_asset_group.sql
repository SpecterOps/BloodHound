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
LOCK TABLE asset_groups, asset_group_tags, kind IN SHARE ROW EXCLUSIVE MODE;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM asset_groups WHERE name = 'Decoy') THEN
        RAISE EXCEPTION USING
            MESSAGE = 'Decoy migration blocked: asset_groups.name "Decoy" is already in use.',
            HINT = 'Rename the existing asset group before retrying this migration.';
    END IF;

    IF EXISTS (SELECT 1 FROM asset_groups WHERE tag = 'decoy') THEN
        RAISE EXCEPTION USING
            MESSAGE = 'Decoy migration blocked: asset_groups.tag "decoy" is already in use.',
            HINT = 'Rename the existing asset group tag before retrying this migration.';
    END IF;

    IF EXISTS (SELECT 1 FROM asset_group_tags WHERE type = 4) THEN
        RAISE EXCEPTION USING
            MESSAGE = 'Decoy migration blocked: asset_group_tags.type 4 is already in use.',
            HINT = 'Resolve the existing type 4 asset group tag before retrying this migration.';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM asset_group_tags
        WHERE name = 'Decoy'
            AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION USING
            MESSAGE = 'Decoy migration blocked: active asset_group_tags.name "Decoy" is already in use.',
            HINT = 'Rename the existing asset group tag before retrying this migration.';
    END IF;

    IF EXISTS (SELECT 1 FROM asset_group_tags WHERE glyph = 'mask') THEN
        RAISE EXCEPTION USING
            MESSAGE = 'Decoy migration blocked: asset_group_tags.glyph "mask" is already in use.',
            HINT = 'Choose a different glyph for the existing asset group tag before retrying this migration.';
    END IF;

    IF EXISTS (SELECT 1 FROM kind WHERE name = 'Tag_Decoy') THEN
        RAISE EXCEPTION USING
            MESSAGE = 'Decoy migration blocked: kind.name "Tag_Decoy" is already in use.',
            HINT = 'Resolve the existing graph kind before retrying this migration.';
    END IF;
END
$$;
-- +goose StatementEnd

INSERT INTO asset_groups (name, tag, system_group, created_at, updated_at)
VALUES ('Decoy', 'decoy', true, current_timestamp, current_timestamp);

WITH decoy_kind AS (
    INSERT INTO kind (name)
    VALUES ('Tag_Decoy')
    RETURNING id
)
INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by, glyph)
SELECT id, 4, 'Decoy', 'Accounts that should be excluded from predefined attack path queries.', current_timestamp, 'BloodHound', current_timestamp, 'BloodHound', 'mask'
FROM decoy_kind;

-- +goose Down
WITH decoy_tag AS (
    SELECT asset_group_tags.id, asset_group_tags.kind_id
    FROM asset_group_tags
    JOIN kind ON kind.id = asset_group_tags.kind_id
    WHERE asset_group_tags.type = 4
        AND asset_group_tags.name = 'Decoy'
        AND asset_group_tags.created_by = 'BloodHound'
        AND kind.name = 'Tag_Decoy'
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
    RETURNING id, kind_id
)
DELETE FROM kind
WHERE id IN (SELECT kind_id FROM deleted_tag)
    AND NOT EXISTS (
        SELECT 1
        FROM asset_group_tags
        WHERE asset_group_tags.kind_id = kind.id
            AND asset_group_tags.id NOT IN (SELECT id FROM deleted_tag)
    );

DELETE FROM asset_groups
WHERE tag = 'decoy'
    AND name = 'Decoy'
    AND system_group = true;
