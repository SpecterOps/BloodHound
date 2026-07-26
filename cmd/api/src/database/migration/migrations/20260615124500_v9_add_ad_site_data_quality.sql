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
ALTER TABLE ad_data_quality_aggregations ADD COLUMN IF NOT EXISTS sites BIGINT DEFAULT 0;
ALTER TABLE ad_data_quality_aggregations ADD COLUMN IF NOT EXISTS siteservers BIGINT DEFAULT 0;
ALTER TABLE ad_data_quality_aggregations ADD COLUMN IF NOT EXISTS sitesubnets BIGINT DEFAULT 0;

ALTER TABLE ad_data_quality_stats ADD COLUMN IF NOT EXISTS sites BIGINT DEFAULT 0;
ALTER TABLE ad_data_quality_stats ADD COLUMN IF NOT EXISTS siteservers BIGINT DEFAULT 0;
ALTER TABLE ad_data_quality_stats ADD COLUMN IF NOT EXISTS sitesubnets BIGINT DEFAULT 0;

WITH src_data AS (
    SELECT *
    FROM (VALUES
        (
            'Sites',
            true,
            true,
            E'MATCH (n:Site) \nRETURN n;',
            E'Control over an Active Directory site may allow users to compromise all assets associated with the site through the application of Group Policy Objects. Since AD Sites contain at least a Domain Controller as a Site Server, this results in the potential compromise of at least one domain in the forest. Therefore, Active Directory Site objects are Tier Zero.'
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
    WHERE name = 'Sites'
        AND is_default = true
        AND created_by = 'BloodHound'
);

DELETE FROM asset_group_tag_selectors
WHERE name = 'Sites'
    AND is_default = true
    AND created_by = 'BloodHound';

ALTER TABLE ad_data_quality_stats DROP COLUMN IF EXISTS sitesubnets;
ALTER TABLE ad_data_quality_stats DROP COLUMN IF EXISTS siteservers;
ALTER TABLE ad_data_quality_stats DROP COLUMN IF EXISTS sites;

ALTER TABLE ad_data_quality_aggregations DROP COLUMN IF EXISTS sitesubnets;
ALTER TABLE ad_data_quality_aggregations DROP COLUMN IF EXISTS siteservers;
ALTER TABLE ad_data_quality_aggregations DROP COLUMN IF EXISTS sites;
