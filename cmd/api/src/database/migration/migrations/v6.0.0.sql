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

-- Add Citrix RDP
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('analysis.citrix_rdp_support', 'Citrix RDP Support',
        'This configuration parameter toggles Citrix support during post-processing. When enabled, computers identified with a ''Direct Access Users'' local group will assume that Citrix is installed and CanRDP edges will require membership of both ''Direct Access Users'' and ''Remote Desktop Users'' local groups on the computer.',
        '{
          "enabled": false
        }', current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

-- Add Prune TTLs
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('prune.ttl', 'Prune Retention TTL Configuration Parameters',
        'This configuration parameter sets the retention TTLs during analysis pruning.', '{
    "base_ttl": "P7D",
    "has_session_edge_ttl": "P3D"
  }', current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

-- Add Reconciliation to parameters and remove from feature_flags
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('analysis.reconciliation', 'Reconciliation',
        'This configuration parameter enables / disables reconciliation during analysis.', format('{"enabled": %s}',
                                                                                                  (SELECT COALESCE(
                                                                                                            (SELECT enabled FROM feature_flags WHERE key = 'reconciliation'),
                                                                                                            TRUE))::text)::json,
        current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;
-- must occur after insert to ensure reconciliation flag is set to whatever current value is
DELETE
FROM feature_flags
WHERE key = 'reconciliation';

-- Grant the Read-Only user SavedQueriesRead permissions
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Read-Only'),
        (SELECT id FROM permissions WHERE permissions.authority = 'saved_queries' and permissions.name = 'Read'))
ON CONFLICT DO NOTHING;

-- Add OIDC Support feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'oidc_support',
        'OIDC Support',
        'Enables OpenID Connect authentication support for SSO Authentication.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Update existing Edge tables with an additional constraint to support ON CONFLICT upserts
do
$$
  begin
    -- Update existing Edge tables with an additional constraint to support ON CONFLICT upserts
    alter table edge
      drop constraint if exists edge_graph_id_start_id_end_id_kind_id_key;
    alter table edge
      add constraint edge_graph_id_start_id_end_id_kind_id_key unique (graph_id, start_id, end_id, kind_id);
  exception
    -- This guards against the possibility that the edge table doesn't exist, in which case there's no constraint to
    -- migrate
    when undefined_table then null;
  end
$$;

-- Set Dark Mode to default to enabled and hide the flag from users in the UI
UPDATE feature_flags set enabled = true, user_updatable = false where key = 'dark_mode';
