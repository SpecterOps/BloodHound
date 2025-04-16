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

ALTER TABLE asset_group_history
	ADD COLUMN IF NOT EXISTS email VARCHAR(330) DEFAULT NULL;

-- Populate email for existing records by looking up the email address from the users table
UPDATE asset_group_history
	SET email = (SELECT email_address FROM users WHERE asset_group_history.actor = users.id)
	WHERE email IS NULL;
