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
WITH src_data AS (
    SELECT *
    FROM (VALUES
        (
            'Domain Controller Site Servers',
            true,
            false,
            E'MATCH (n:SiteServer)-[:ServerIs]->(:Computer)-[:DCFor]->(:Domain)\nRETURN n;',
            E'Active Directory Site Server objects that reference domain controllers are Tier Zero because control over the Site Server object may impact a domain controller and therefore the domain.'
        )
    ) AS s (name, enabled, allow_disable, cypher, description)
),
inserted_selectors AS (
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
    )
    SELECT
        (SELECT id FROM asset_group_tags WHERE type = 1 AND position = 1 LIMIT 1),
        current_timestamp,
        'BloodHound',
        current_timestamp,
        'BloodHound',
        CASE WHEN NOT d.enabled THEN current_timestamp ELSE NULL END,
        CASE WHEN NOT d.enabled THEN 'BloodHound' ELSE NULL END,
        d.name,
        d.description,
        true,
        d.allow_disable,
        2
    FROM src_data d
    WHERE NOT EXISTS (SELECT 1 FROM asset_group_tag_selectors WHERE name = d.name)
    RETURNING id, name
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
    s.id,
    2,
    d.cypher
FROM inserted_selectors s
JOIN src_data d ON d.name = s.name;

-- +goose Down
DELETE FROM asset_group_tag_selector_seeds
WHERE selector_id IN (
    SELECT id
    FROM asset_group_tag_selectors
    WHERE name = 'Domain Controller Site Servers'
        AND is_default = true
        AND created_by = 'BloodHound'
);

DELETE FROM asset_group_tag_selectors
WHERE name = 'Domain Controller Site Servers'
    AND is_default = true
    AND created_by = 'BloodHound';
