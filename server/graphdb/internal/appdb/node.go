// Copyright 2026 Specter Ops, Inc.
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

package appdb

import (
	"context"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
)

const tableSchemaNodeKinds = "schema_node_kinds"

// toNode translates a graph node into the domain model. The kind ids are
// left zero-valued here; the service resolves them from the kind table.
func toNode(node *graph.Node) services.Node {
	var kinds []services.Kind

	for _, kind := range node.Kinds {
		kinds = append(kinds, services.Kind{Name: kind.String()})
	}

	return services.Node{
		ID:         int64(node.ID),
		Kinds:      kinds,
		Properties: node.Properties.MapOrEmpty(),
	}
}

// GetNode fetches a node by its graph-assigned id, returning
// ErrNodeNotFound when no node matches.
func (s *Store) GetNode(ctx context.Context, id int64) (services.Node, error) {
	var (
		node *graph.Node
		err  error
	)

	if err = s.graph.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var fetchErr error
		node, fetchErr = ops.FetchNode(tx, graph.ID(id))
		return fetchErr
	}); graph.IsErrNotFound(err) {
		return services.Node{}, services.ErrNodeNotFound
	} else if err != nil {
		return services.Node{}, fmt.Errorf("fetching node: %w", err)
	} else if node == nil {
		return services.Node{}, services.ErrNodeNotFound
	} else {
		return toNode(node), nil
	}
}

// GetNodeKindsByNames resolves multiple kind names to their schema_node_kinds entries,
// returning the schema_node_kinds row id as the kind id. For kinds that do not have
// a schema_node_kinds entry, a Kind with ID=nil and the name is returned. This allows
// unregistered kinds to be included in the response (best-effort resolution).
func (s *Store) GetNodeKindsByNames(ctx context.Context, names []string) ([]services.Kind, error) {
	var (
		rows  pgx.Rows
		kinds []services.Kind
		err   error
	)

	if len(names) == 0 {
		return kinds, nil
	}

	var (
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
		sqlQuery      string
		args          []any
		kindRows      []kindRow
	)

	selectBuilder.Select("nk.id", "k.name")
	selectBuilder.From(selectBuilder.As(tableSchemaNodeKinds, "nk"))
	selectBuilder.Join(selectBuilder.As(tableKind, "k"), "nk.kind_id = k.id")
	selectBuilder.Where(selectBuilder.In("k.name", sqlbuilder.List(names)))

	sqlQuery, args = selectBuilder.Build()

	if rows, err = s.db.Query(ctx, sqlQuery, args...); err != nil {
		return nil, err
	} else if kindRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[kindRow]); err != nil {
		return nil, fmt.Errorf("reading rows: %w", err)
	}

	// Build a map of registered kinds by name for lookup
	kindsByName := make(map[string]services.Kind, len(kindRows))
	for _, row := range kindRows {
		kindsByName[row.Name] = toKind(row)
	}

	// Return kinds in input order, using nil ID for unregistered kinds
	for _, name := range names {
		if kind, found := kindsByName[name]; found {
			kinds = append(kinds, kind)
		} else {
			kinds = append(kinds, services.Kind{ID: nil, Name: name})
		}
	}

	return kinds, nil
}
