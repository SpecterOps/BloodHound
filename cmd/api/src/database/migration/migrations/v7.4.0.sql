-- Copyright 2025 Specter Ops, Inc.
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

-- Migration to add `Microsoft 365 groups` column to relevant tables

-- Add `groups365` column to `azure_data_quality_aggregations` table
ALTER TABLE IF EXISTS azure_data_quality_aggregations
    ADD COLUMN groups365 bigint;

-- Add `groups365` column to `azure_data_quality_stats` table
ALTER TABLE IF EXISTS azure_data_quality_stats
    ADD COLUMN groups365 bigint;