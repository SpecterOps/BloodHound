-- Copyright 2026 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
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
    defunct_kind_ids SMALLINT[];
BEGIN
    SELECT array_agg(id) INTO defunct_kind_ids
    FROM kind
    WHERE name IN ('ContainsIdentity', 'PropagatesACEsTo', 'GPOAppliesTo', 'CanApplyGPO');

    IF defunct_kind_ids IS NULL THEN
        RETURN;
    END IF;

    -- Remove rows that reference the defunct kinds without ON DELETE CASCADE
    DELETE FROM asset_group_history
    WHERE asset_group_tag_id IN (SELECT id FROM asset_group_tags WHERE kind_id = ANY (defunct_kind_ids));
    DELETE FROM asset_group_tags WHERE kind_id = ANY (defunct_kind_ids);
    DELETE FROM schema_environments_principal_kinds WHERE principal_kind = ANY (defunct_kind_ids);
    DELETE FROM schema_environments
    WHERE environment_kind_id = ANY (defunct_kind_ids) OR source_kind_id = ANY (defunct_kind_ids);
    DELETE FROM schema_findings WHERE kind_id = ANY (defunct_kind_ids) OR environment_id = ANY (defunct_kind_ids);
    DELETE FROM custom_node_kinds WHERE kind_id = ANY (defunct_kind_ids);
    DELETE FROM data_quality_stats WHERE schema_environment_kind_id = ANY (defunct_kind_ids);
    DELETE FROM data_quality_aggregations WHERE schema_environment_kind_id = ANY (defunct_kind_ids);
    DELETE FROM schema_kind_info WHERE kind_id = ANY (defunct_kind_ids);
    DELETE FROM schema_relationship_kinds WHERE kind_id = ANY (defunct_kind_ids);

    -- The graph edge table is created by DAWGS and has no foreign key to kind
    IF to_regclass('public.edge') IS NOT NULL THEN
        DELETE FROM edge WHERE kind_id = ANY (defunct_kind_ids);
    END IF;

    DELETE FROM kind WHERE id = ANY (defunct_kind_ids);
END $$;
-- +goose StatementEnd

-- +goose Down
-- Intentionally empty. The defunct AD edge kinds should not be restored.
