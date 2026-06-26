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

-- Add analysis_step column to analysis_request_switch to allow callers to specify
-- which steps of the analysis pipeline should run. The column stores an AnalysisStep
-- bitmask (ADPostProcessing=1, AzurePostProcessing=2, Tagging=4, Analysis=8).
ALTER TABLE analysis_request_switch
  ADD COLUMN IF NOT EXISTS analysis_step INTEGER NOT NULL DEFAULT 0;

-- If there currently is an analysis request in the DB set it to full analysis.
UPDATE analysis_request_switch
SET analysis_step = 15;

-- +goose Down

ALTER TABLE analysis_request_switch
  DROP COLUMN IF EXISTS analysis_step;
