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
-- +goose StatementBegin
DO $$
DECLARE
    -- Keep created_by as BloodHound for system-selector behavior and use updated_by to scope rollback ownership.
    migration_marker CONSTANT text := 'migration:20260615130000_v9_add_tier_zero_site_server_selector';
    selector_name CONSTANT text := 'Domain Controller Site Servers';
    selector_description CONSTANT text := E'An Active Directory Site Server object represents a computer associated with an Active Directory Site. When the referenced computer is a Domain Controller, a malicious Group Policy Object linked to the Site could compromise the Domain Controller and, consequently, its domain. Therefore, Site Server objects that reference Domain Controllers are classified as Tier Zero.';
    selector_cypher CONSTANT text := E'MATCH (n:SiteServer)-[:ServerIs]->(:Computer)-[:DCFor]->(:Domain)\nRETURN n;';
    resolved_selector_id integer;
    resolved_tier_zero_tag_id integer;
    selector_count bigint;
    tier_zero_tag_count bigint;
BEGIN
    SELECT MIN(tags.id), COUNT(*)
    INTO resolved_tier_zero_tag_id, tier_zero_tag_count
    FROM asset_group_tags tags
    WHERE tags.type = 1
        AND tags.position = 1
        AND tags.deleted_at IS NULL;

    IF tier_zero_tag_count <> 1 THEN
        RAISE EXCEPTION 'expected exactly one Tier Zero asset group tag for the % selector, found %', selector_name, tier_zero_tag_count;
    END IF;

    SELECT MIN(selectors.id), COUNT(*)
    INTO resolved_selector_id, selector_count
    FROM asset_group_tag_selectors selectors
    WHERE selectors.name = selector_name;

    IF selector_count = 0 THEN
        INSERT INTO asset_group_tag_selectors (
            asset_group_tag_id,
            created_at,
            created_by,
            updated_at,
            updated_by,
            disabled_at,
            disabled_by,
            name,
            description,
            is_default,
            allow_disable,
            auto_certify
        ) VALUES (
            resolved_tier_zero_tag_id,
            current_timestamp,
            'BloodHound',
            current_timestamp,
            migration_marker,
            NULL,
            NULL,
            selector_name,
            selector_description,
            true,
            true,
            2
        )
        RETURNING id INTO resolved_selector_id;

        INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
        VALUES (resolved_selector_id, 2, selector_cypher);
    ELSIF selector_count <> 1 OR NOT EXISTS (
        SELECT 1
        FROM asset_group_tag_selectors selectors
        WHERE selectors.id = resolved_selector_id
            AND selectors.asset_group_tag_id = resolved_tier_zero_tag_id
            AND selectors.description = selector_description
            AND selectors.is_default = true
            AND selectors.disabled_at IS NULL
            AND selectors.disabled_by IS NULL
            AND selectors.allow_disable = false
            AND selectors.auto_certify = 2
            AND EXISTS (
                SELECT 1
                FROM asset_group_tag_selector_seeds seeds
                WHERE seeds.selector_id = resolved_selector_id
                    AND seeds.type = 2
                    AND seeds.value = selector_cypher
            )
    ) THEN
        RAISE EXCEPTION 'the existing % selector or its seed does not match the expected configuration', selector_name;
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose Down
DELETE FROM asset_group_tag_selector_seeds
WHERE selector_id IN (
    SELECT id
    FROM asset_group_tag_selectors
    WHERE name = 'Domain Controller Site Servers'
        AND created_by = 'BloodHound'
        AND updated_by = 'migration:20260615130000_v9_add_tier_zero_site_server_selector'
)
    AND type = 2
    AND value = E'MATCH (n:SiteServer)-[:ServerIs]->(:Computer)-[:DCFor]->(:Domain)\nRETURN n;';

DELETE FROM asset_group_tag_selectors
WHERE name = 'Domain Controller Site Servers'
    AND created_by = 'BloodHound'
    AND updated_by = 'migration:20260615130000_v9_add_tier_zero_site_server_selector';
