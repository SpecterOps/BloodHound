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

-- OIDC Provider
CREATE TABLE IF NOT EXISTS oidc_providers
(
  id         BIGSERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  client_id  TEXT NOT NULL,
  issuer     TEXT NOT NULL,

  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

  UNIQUE (name)
);


-- Add Scheduled Analysis Configs
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
  VALUES ('analysis.scheduled',
        'Scheduled Analysis',
        'This configuration parameter allows setting a schedule for analysis. When enabled, analysis will only run when the scheduled time arrives',
        '{"enabled": false, "rrule": ""}',
        current_timestamp,current_timestamp) ON CONFLICT DO NOTHING;

-- Add last analysis time to datapipe status so we can track scheduled analysis time properly
ALTER TABLE datapipe_status
ADD COLUMN IF NOT EXISTS "last_analysis_run_at" TIMESTAMP with time zone;
