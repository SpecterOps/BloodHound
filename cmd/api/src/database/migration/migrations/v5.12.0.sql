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

CREATE TABLE IF NOT EXISTS datapipe_status (
        singleton BOOL PRIMARY KEY DEFAULT true,
        status TEXT NOT NULL,
        updated_at TIMESTAMP with time zone NOT NULL,
        last_complete_analysis_at TIMESTAMP with time zone,
        CONSTRAINT singleton_uni CHECK (singleton)
);

INSERT INTO
    datapipe_status (status, updated_at)
VALUES
    ('idle', NOW ())
ON CONFLICT DO NOTHING;

ALTER TABLE user_sessions
    ADD COLUMN IF NOT EXISTS flags jsonb;