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


-- Update the 'auth_tokens' table adding created_by column
-- As of current, we don't have a way to backfill the data, so we are leaving this field optional for now
ALTER TABLE auth_tokens
  ADD COLUMN IF NOT EXISTS created_by text;

-- Create fk_auth_tokens_created_by to auth_tokens referencing users.id if it doesn't exist
DO $$
  BEGIN
    IF NOT EXISTS (
      SELECT 1
      FROM pg_constraint
      WHERE conname = 'fk_auth_tokens_created_by'
    ) THEN
      ALTER TABLE auth_tokens
        ADD CONSTRAINT fk_auth_tokens_created_by
          FOREIGN KEY (created_by)
            REFERENCES users(id);
    END IF;
  END $$;
