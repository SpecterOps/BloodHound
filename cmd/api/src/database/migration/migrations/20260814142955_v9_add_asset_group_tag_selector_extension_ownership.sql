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
ALTER TABLE asset_group_tag_selectors
    ADD COLUMN IF NOT EXISTS rule_key text,
    ADD COLUMN IF NOT EXISTS extension_id integer;

-- idempotent-ly add foreign key constraint to extension_id column
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'asset_group_tag_selectors_extension_id_fkey'
    ) THEN
        ALTER TABLE asset_group_tag_selectors
            ADD CONSTRAINT asset_group_tag_selectors_extension_id_fkey
            FOREIGN KEY (extension_id)
            REFERENCES schema_extensions(id)
            ON DELETE CASCADE;
    END IF;
END $$;
-- +goose StatementEnd

-- require rule_key and extension_id to be populated together
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'asset_group_tag_selectors_rule_key_extension_id_check'
    ) THEN
        ALTER TABLE asset_group_tag_selectors
            ADD CONSTRAINT asset_group_tag_selectors_rule_key_extension_id_check
            CHECK (
                (rule_key IS NULL AND extension_id IS NULL)
                OR (rule_key IS NOT NULL AND extension_id IS NOT NULL)
            );
    END IF;
END $$;
-- +goose StatementEnd

-- index for looking up selectors by rule_key and ensuring uniqueness of rule_key per extension
CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_group_tag_selectors_rule_key
    ON asset_group_tag_selectors (rule_key, extension_id)
    WHERE rule_key IS NOT NULL;

-- index for looking up selectors by extension_id
CREATE INDEX IF NOT EXISTS idx_asset_group_tag_selectors_extension_id
    ON asset_group_tag_selectors (extension_id);

-- +goose Down
DROP INDEX IF EXISTS idx_asset_group_tag_selectors_extension_id;
DROP INDEX IF EXISTS idx_asset_group_tag_selectors_rule_key;

ALTER TABLE asset_group_tag_selectors
    DROP CONSTRAINT IF EXISTS asset_group_tag_selectors_extension_id_fkey,
    DROP CONSTRAINT IF EXISTS asset_group_tag_selectors_rule_key_extension_id_check,
    DROP COLUMN IF EXISTS extension_id,
    DROP COLUMN IF EXISTS rule_key;
