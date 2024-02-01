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

-- Add new columns for audit_logs
ALTER TABLE audit_logs
ADD COLUMN IF NOT EXISTS actor_email VARCHAR(330) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS source VARCHAR(40) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS status VARCHAR(15) CHECK (status IN ('success', 'failure')) DEFAULT 'success';

-- Populate actor_email for existing records by looking up the email address from the users table
UPDATE audit_logs
SET actor_email = COALESCE((SELECT email_address FROM users WHERE audit_logs.actor_id = users.id), 'unknown');
