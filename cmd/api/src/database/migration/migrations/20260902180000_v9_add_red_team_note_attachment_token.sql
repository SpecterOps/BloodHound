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
ALTER TABLE red_team_note_attachments ADD COLUMN IF NOT EXISTS token TEXT NOT NULL DEFAULT '';
UPDATE red_team_note_attachments SET token = gen_random_uuid()::text WHERE token = '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_red_team_note_attachments_token ON red_team_note_attachments (token);

-- +goose Down
DROP INDEX IF EXISTS idx_red_team_note_attachments_token;
ALTER TABLE red_team_note_attachments DROP COLUMN IF EXISTS token;
