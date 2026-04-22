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
-- Add OpenGraph permissions to permissions table
-- +goose Up
INSERT INTO permissions(created_at, updated_at, authority, name)
VALUES (
        current_timestamp,
        current_timestamp,
        'opengraph',
        'Read'
       ), 
       (
        current_timestamp,
        current_timestamp,
        'opengraph',
        'Write'
       )
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM permissions 
WHERE authority = 'opengraph' 
  AND name IN ('Read', 'Write');
