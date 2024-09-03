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

INSERT INTO parameters (key, name, description, value, created_at, updated_at) 
VALUES ('analysis.citrix_rdp_support', 'Citrix RDP Support', 'This configuration parameter toggles Citrix support during post-processing. When enabled, computers identified with a ''Direct Access Users'' local group will assume that Citrix is installed and CanRDP edges will require membership of both ''Direct Access Users'' and ''Remote Desktop Users'' local groups on the computer.', '{"enabled": false}',current_timestamp,current_timestamp) ON CONFLICT DO NOTHING;
-- -- Fix Parameter table missing autoincr
-- CREATE SEQUENCE IF NOT EXISTS parameter_id_seq
--     AS integer
--     START WITH 1
--     INCREMENT BY 1
--     NO MINVALUE
--     NO MAXVALUE
--     CACHE 1
--     OWNED BY parameters.id;

INSERT INTO parameters (id, key, name, description, value, created_at, updated_at) 
VALUES (3, 'analysis.citrix_rdp_support', 'Citrix RDP Support', 'This configuration parameter toggles Citrix support during post-processing. When on, CanRDP edges will come from the `Direct Access Users` group instead of the builtin `Remote Desktop Users` group.', '{"enabled": false}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Add Prune TTLs
INSERT INTO parameters (id, key, name, description, value, created_at, updated_at) VALUES (3, 'prune.ttl', 'Prune Retention TTL Configuration Parameters', 'This configuration parameter sets the retention TTLs during analysis pruning.', '{"base_ttl": "P7D", "has_session_edge_ttl": "P3D"}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;