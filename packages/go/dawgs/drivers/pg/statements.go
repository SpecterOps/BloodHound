// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package pg

const (
	createNodeStatement               = `insert into node (graph_id, kind_ids, properties) values (@graph_id, @kind_ids, @properties) returning (id, kind_ids, properties)::nodeComposite;`
	createNodeWithoutIDBatchStatement = `insert into node (graph_id, kind_ids, properties) select $1, unnest($2::text[])::int2[], unnest($3::jsonb[])`
	createNodeWithIDBatchStatement    = `insert into node (graph_id, id, kind_ids, properties) select $1, unnest($2::int4[]), unnest($3::text[])::int2[], unnest($4::jsonb[])`
	deleteNodeWithIDStatement         = `delete from node where node.id = any($1)`

	createEdgeStatement       = `insert into edge (graph_id, start_id, end_id, kind_id, properties) values (@graph_id, @start_id, @end_id, @kind_id, @properties) returning (id, start_id, end_id, kind_id, properties)::edgeComposite;`
	createEdgeBatchStatement  = `insert into edge (graph_id, start_id, end_id, kind_id, properties) select $1::int4, unnest($2::int4[]), unnest($3::int4[]), unnest($4::int2[]), unnest($5::jsonb[]);`
	deleteEdgeWithIDStatement = `delete from edge as e where e.id = any($1)`

	edgePropertySetOnlyStatement      = `update edge set properties = properties || $1::jsonb where edge.id = $2`
	edgePropertyDeleteOnlyStatement   = `update edge set properties = properties - $1::text[] where edge.id = $2`
	edgePropertySetAndDeleteStatement = `update edge set properties = properties || $1::jsonb - $2::text[] where edge.id = $3`
)
