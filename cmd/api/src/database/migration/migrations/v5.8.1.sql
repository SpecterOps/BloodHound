-- Copyright 2024 Specter Ops, Inc.
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

ALTER TABLE file_upload_jobs
  ADD COLUMN IF NOT EXISTS total_files integer DEFAULT 0;

ALTER TABLE file_upload_jobs
  ADD COLUMN IF NOT EXISTS failed_files integer DEFAULT 0;

ALTER TABLE ad_data_quality_stats
  ADD COLUMN IF NOT EXISTS issuancepolicies BIGINT DEFAULT 0;

ALTER TABLE ad_data_quality_aggregations
  ADD COLUMN IF NOT EXISTS issuancepolicies BIGINT DEFAULT 0;

ALTER TABLE azure_data_quality_stats
ADD COLUMN IF NOT EXISTS automation_accounts BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS container_registries BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS function_apps BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS logic_apps BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS managed_clusters BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS vm_scale_sets BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS web_apps BIGINT DEFAULT 0;

ALTER TABLE azure_data_quality_aggregations
ADD COLUMN IF NOT EXISTS automation_accounts BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS container_registries BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS function_apps BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS logic_apps BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS managed_clusters BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS vm_scale_sets BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS web_apps BIGINT DEFAULT 0;
