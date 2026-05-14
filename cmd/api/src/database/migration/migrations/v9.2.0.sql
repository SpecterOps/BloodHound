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
-- where in the analysis pipeline to begin. The default value of 1 corresponds to
-- AnalysisStepPostProcessing (== AnalysisStepAll), which starts from the beginning.
ALTER TABLE analysis_request_switch
  ADD COLUMN IF NOT EXISTS analysis_step integer NOT NULL DEFAULT 1;
