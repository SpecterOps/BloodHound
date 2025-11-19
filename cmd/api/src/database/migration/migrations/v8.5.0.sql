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


-- OpenGraph graph schema - extensions (collectors)
CREATE TABLE IF NOT EXISTS schema_extensions (
    id SERIAL NOT NULL,
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    version TEXT NOT NULL,
    is_builtin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    PRIMARY KEY (id)
);


create table if not exists schema_node_kinds (
    id smallserial primary key,
    schema_extension_id int not null references schema_extensions (id) on delete cascade, -- indicates which extension this node kind belongs to
    name text unique not null, -- unique is required by the DAWGS kind table
    display_name text not null, -- can be different from name but usually isn't other than Base/Entity
    description text not null, -- human-readable description of the kind
    is_display_kind bool not null default false,
    icon text not null, -- font-awesome icon
    icon_color text not null default '#00000000', -- default to a transparent hex color
    created_at timestamp with time zone not null default current_timestamp,
    updated_at timestamp with time zone not null default current_timestamp,
    deleted_at timestamp with time zone default null
);

create index idx_graph_schema_node_kinds_extensions_id on schema_node_kinds (schema_extension_id);
