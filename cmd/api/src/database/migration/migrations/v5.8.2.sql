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

CREATE INDEX IF NOT EXISTS idx_ad_data_quality_aggregations_created_at ON ad_data_quality_aggregations USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_ad_data_quality_aggregations_updated_at ON ad_data_quality_aggregations USING btree (updated_at);

CREATE INDEX IF NOT EXISTS idx_ad_asset_groups_updated_at ON asset_groups USING btree (updated_at);
CREATE INDEX IF NOT EXISTS idx_ad_asset_groups_created_at ON asset_groups USING btree (created_at);

CREATE INDEX IF NOT EXISTS idx_azure_data_quality_aggregations_created_at ON azure_data_quality_aggregations USING btree (created_at);

CREATE INDEX IF NOT EXISTS idx_azure_data_quality_stats_created_at ON azure_data_quality_stats USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_azure_data_quality_stats_updated_at ON azure_data_quality_stats USING btree (updated_at);

CREATE INDEX IF NOT EXISTS idx_file_upload_jobs_status ON file_upload_jobs USING btree (status);
CREATE INDEX IF NOT EXISTS idx_file_upload_jobs_created_at ON file_upload_jobs USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_file_upload_jobs_updated_at ON file_upload_jobs USING btree (updated_at);
CREATE INDEX IF NOT EXISTS idx_file_upload_jobs_start_time ON file_upload_jobs USING btree (start_time);
CREATE INDEX IF NOT EXISTS idx_file_upload_jobs_end_time ON file_upload_jobs USING btree (end_time);

CREATE INDEX IF NOT EXISTS idx_ingest_tasks_task_id ON ingest_tasks USING btree (task_id);

CREATE INDEX IF NOT EXISTS idx_users_eula_accepted ON users USING btree (eula_accepted);
