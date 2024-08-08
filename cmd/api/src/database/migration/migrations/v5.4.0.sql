-- Copyright 2023 Specter Ops, Inc.
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

-- Data Quality Stats for new ADCS node types
ALTER TABLE ad_data_quality_stats
ADD COLUMN IF NOT EXISTS aiacas BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS rootcas BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS enterprisecas BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS ntauthstores BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS certtemplates BIGINT DEFAULT 0;

ALTER TABLE ad_data_quality_aggregations
ADD COLUMN IF NOT EXISTS aiacas BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS rootcas BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS enterprisecas BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS ntauthstores BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS certtemplates BIGINT DEFAULT 0;

DELETE FROM
    saved_queries
WHERE
    user_id = '00000000-0000-0000-0000-000000000000';

INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'adcs', 'Enable collection and processing of Active Directory Certificate Services Data', 'Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.', false, true) ON CONFLICT DO NOTHING;
