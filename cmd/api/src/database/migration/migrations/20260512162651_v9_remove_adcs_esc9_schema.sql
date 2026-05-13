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
DELETE FROM schema_findings_subtypes
WHERE schema_finding_id IN (
    SELECT id
    FROM schema_findings
    WHERE name IN ('T0ADCSESC9a', 'T0ADCSESC9b')
);

DELETE FROM schema_findings
WHERE name IN ('T0ADCSESC9a', 'T0ADCSESC9b');

DELETE FROM schema_relationship_kinds
USING kind
WHERE schema_relationship_kinds.kind_id = kind.id
AND kind.name IN ('ADCSESC9a', 'ADCSESC9b');

-- +goose Down
INSERT INTO schema_relationship_kinds (
    schema_extension_id,
    kind_id,
    description,
    is_traversable,
    created_at,
    updated_at
)
SELECT
    schema_extensions.id,
    kind.id,
    '',
    true,
    current_timestamp,
    current_timestamp
FROM kind
JOIN schema_extensions
    ON schema_extensions.name = 'AD'
WHERE kind.name IN ('ADCSESC9a', 'ADCSESC9b')
ON CONFLICT (kind_id) DO UPDATE SET
    schema_extension_id = EXCLUDED.schema_extension_id,
    description = EXCLUDED.description,
    is_traversable = EXCLUDED.is_traversable,
    updated_at = current_timestamp;

INSERT INTO schema_findings (
    type,
    schema_extension_id,
    environment_id,
    kind_id,
    name,
    display_name,
    created_at
)
SELECT
    1,
    schema_extensions.id,
    schema_environments.id,
    kind.id,
    finding_definitions.finding_name,
    finding_definitions.display_name,
    current_timestamp
FROM (
    VALUES
        ('ADCSESC9a', 'T0ADCSESC9a', 'Tier Zero ADCS ESC9a'),
        ('ADCSESC9b', 'T0ADCSESC9b', 'Tier Zero ADCS ESC9b')
) AS finding_definitions(kind_name, finding_name, display_name)
JOIN schema_extensions
    ON schema_extensions.name = 'AD'
JOIN LATERAL (
    SELECT id
    FROM schema_environments
    WHERE schema_environments.schema_extension_id = schema_extensions.id
    LIMIT 1
) schema_environments
    ON true
JOIN kind
    ON kind.name = finding_definitions.kind_name
ON CONFLICT (name) DO UPDATE SET
    type = EXCLUDED.type,
    schema_extension_id = EXCLUDED.schema_extension_id,
    environment_id = EXCLUDED.environment_id,
    kind_id = EXCLUDED.kind_id,
    display_name = EXCLUDED.display_name;
