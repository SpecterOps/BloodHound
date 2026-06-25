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
	"errors"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
)

const tableSchemaRelationshipKinds = "schema_relationship_kinds"

// kindRow is the package-local DB row type for a resolved relationship kind. The id is
// the schema_relationship_kinds row id, which is null when the kind has no
// schema_relationship_kinds entry; the name is the kind name from the kind table.
// db: tags drive pgx.RowToStructByName scanning.
type kindRow struct {
	ID   *int32 `db:"id"`
	Name string `db:"name"`
}

// toKind translates a raw kind row into the domain model.
func toKind(row kindRow) services.Kind {
	return services.Kind{ID: row.ID, Name: row.Name}
}

// toRelationship translates a graph relationship into the domain model. The kind id is
// left zero-valued here; the service resolves it from the kind table.
func toRelationship(relationship *graph.Relationship) services.Relationship {
	return services.Relationship{
		ID:           int64(relationship.ID),
		SourceNodeID: int64(relationship.StartID),
		TargetNodeID: int64(relationship.EndID),
		Kind:         services.Kind{Name: relationship.Kind.String()},
		Properties:   relationship.Properties.MapOrEmpty(),
	}
}

// GetRelationship fetches a relationship by its graph-assigned id, returning
// ErrRelationshipNotFound when no relationship matches.
func (s *Store) GetRelationship(ctx context.Context, id int64) (services.Relationship, error) {
	var (
		relationship *graph.Relationship
		err          error
	)

	err = s.graph.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var fetchErr error
		relationship, fetchErr = ops.FetchRelationship(tx, graph.ID(id))
		return fetchErr
	})

	if err != nil {
		if graph.IsErrNotFound(err) {
			return services.Relationship{}, services.ErrRelationshipNotFound
		}

		return services.Relationship{}, fmt.Errorf("fetching relationship: %w", err)
	}

	if relationship == nil {
		return services.Relationship{}, services.ErrRelationshipNotFound
	}

	return toRelationship(relationship), nil
}

// GetKindByName resolves a relationship kind name to its schema_relationship_kinds entry,
// returning the schema_relationship_kinds row id as the kind id. The kind id is null when
// the kind exists in the kind table but has no schema_relationship_kinds row. ErrKindNotFound
// is returned when the kind name has no entry in the kind table.
func (s *Store) GetKindByName(ctx context.Context, name string) (services.Kind, error) {
	var (
		rows pgx.Rows
		row  kindRow
		err  error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select("rk.id", "k.name")
	selectBuilder.From(selectBuilder.As(tableKind, "k"))
	selectBuilder.JoinWithOption(sqlbuilder.LeftJoin, selectBuilder.As(tableSchemaRelationshipKinds, "rk"), "rk.kind_id = k.id")
	selectBuilder.Where(selectBuilder.Equal("k.name", name))
	selectBuilder.Limit(1)

	sqlQuery, args := selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return services.Kind{}, err
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[kindRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Kind{}, services.ErrKindNotFound
	}
	if err != nil {
		return services.Kind{}, fmt.Errorf("reading rows: %w", err)
	}

	return toKind(row), nil
}
