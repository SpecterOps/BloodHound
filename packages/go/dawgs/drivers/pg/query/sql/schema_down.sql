-- Copyright 2023 Specter Ops, Inc.
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

-- Drop triggers
drop trigger if exists delete_node_edges on node;
drop function if exists delete_node_edges;

-- Drop functions
drop function if exists query_perf;
drop function if exists lock_details;
drop function if exists table_sizes;
drop function if exists get_node;
drop function if exists node_prop;
drop function if exists kinds;
drop function if exists has_kind;
drop function if exists mt_get_root;
drop function if exists index_utilization;
drop function if exists _format_asp_where_clause;
drop function if exists _format_asp_query;
drop function if exists _all_shortest_paths;
drop function if exists all_shortest_paths;
drop function if exists traversal_step;
drop function if exists _format_traversal_continuation_termination;
drop function if exists _format_traversal_query;
drop function if exists _format_traversal_initial_query;
drop function if exists expand_traversal_step;
drop function if exists traverse;
drop function if exists edges_to_path;
drop function if exists traverse_paths;

-- Drop all tables in order of dependency.
drop table if exists node;
drop table if exists edge;
drop table if exists kind;
drop table if exists graph;

-- Remove custom types
do
$$
  begin
    drop type pathComposite;
  exception
    when undefined_object then null;
  end
$$;

do
$$
  begin
    drop type nodeComposite;
  exception
    when undefined_object then null;
  end
$$;

do
$$
  begin
    drop type edgeComposite;
  exception
    when undefined_object then null;
  end
$$;

do
$$
  begin
    drop type _traversal_step;
  exception
    when undefined_object then null;
  end
$$;

-- Pull the tri-gram and intarray extensions.
drop
  extension if exists pg_trgm;
drop
  extension if exists intarray;
drop
  extension if exists pg_stat_statements;
