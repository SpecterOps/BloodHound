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
LOCK TABLE roles IN SHARE MODE;
LOCK TABLE users_roles IN ACCESS EXCLUSIVE MODE;

-- +goose StatementBegin
DO $$
DECLARE
    role_priority CONSTANT text[] := ARRAY[
        'Read-Only', 'Upload-Only', 'User', 'Auditor', 'Power User', 'Administrator'
    ];
    unknown_roles text;
BEGIN
    SELECT string_agg(DISTINCT COALESCE(r.name, '<NULL>'), ', ')
    INTO unknown_roles
    FROM users_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE (r.name IS NULL OR NOT (r.name = ANY(role_priority)))
      AND ur.user_id IN (
          SELECT user_id FROM users_roles GROUP BY user_id HAVING count(*) > 1
      );

    IF unknown_roles IS NOT NULL THEN
        RAISE EXCEPTION 'Cannot enforce one role per user: duplicate assignments contain unrecognized roles: %. Define their cleanup priority before retrying.', unknown_roles;
    END IF;

    WITH ranked_assignments AS (
        SELECT ur.user_id, ur.role_id,
               row_number() OVER (
                   PARTITION BY ur.user_id
                   ORDER BY array_position(role_priority, r.name), ur.role_id
               ) AS assignment_rank
        FROM users_roles ur
        JOIN roles r ON r.id = ur.role_id
    )
    DELETE FROM users_roles ur
    USING ranked_assignments ranked
    WHERE ur.user_id = ranked.user_id
      AND ur.role_id = ranked.role_id
      AND ranked.assignment_rank > 1;
END $$;
-- +goose StatementEnd

ALTER TABLE users_roles ADD CONSTRAINT users_roles_user_id_key UNIQUE (user_id);

-- +goose Down
ALTER TABLE users_roles DROP CONSTRAINT users_roles_user_id_key;
