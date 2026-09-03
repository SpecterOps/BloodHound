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
CREATE TABLE IF NOT EXISTS red_team_notes (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id    TEXT,
    title      TEXT NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    note_type  TEXT NOT NULL DEFAULT 'general',
    tags       TEXT[] NOT NULL DEFAULT '{}',
    url        TEXT NOT NULL DEFAULT '',
    object_id  TEXT,
    edge_kind  TEXT
);

CREATE INDEX IF NOT EXISTS idx_red_team_notes_object_id ON red_team_notes (object_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_red_team_notes_edge_kind ON red_team_notes (edge_kind) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_red_team_notes_note_type ON red_team_notes (note_type) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_red_team_notes_note_type;
DROP INDEX IF EXISTS idx_red_team_notes_edge_kind;
DROP INDEX IF EXISTS idx_red_team_notes_object_id;
DROP TABLE IF EXISTS red_team_notes;
