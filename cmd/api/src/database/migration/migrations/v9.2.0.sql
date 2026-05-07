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

-- Add analysis_step column to analysis_request_switch to allow callers to specify
-- which steps of the analysis pipeline should run. The column stores an AnalysisStep
-- bitmask (PostProcessing=1, Tagging=2, Analysis=4). The default value of 7
-- corresponds to AnalysisStepAll, which selects every step.
ALTER TABLE analysis_request_switch
  ADD COLUMN IF NOT EXISTS analysis_step integer NOT NULL DEFAULT 7;
