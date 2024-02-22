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

ALTER TABLE IF EXISTS audit_logs
  RENAME COLUMN source TO source_ip_address;

ALTER TABLE IF EXISTS audit_logs
  DROP CONSTRAINT IF EXISTS audit_logs_status_check,
  ADD CONSTRAINT status_check
  CHECK (status IN ('intent', 'success', 'failure')),
  ALTER COLUMN status SET DEFAULT 'intent',
  ALTER COLUMN source_ip_address TYPE TEXT,
  ADD COLUMN IF NOT EXISTS commit_id TEXT;

-- Add indices for scalability
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_email ON audit_logs USING btree (actor_email);
CREATE INDEX IF NOT EXISTS idx_audit_logs_source_ip_address ON audit_logs USING btree (source_ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs USING btree (status);
UPDATE feature_flags SET enabled = false, user_updatable = false WHERE key = 'adcs';
