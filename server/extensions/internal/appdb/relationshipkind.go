// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
)

const tableSchemaRelationshipKinds = "schema_relationship_kinds"

// todo: read and join on schema_extension_id to pull in extension fields?
type relationshipKindRow struct {
	ID            int32     `db:"id"`
	KindID        int32     `db:"kind_id"`
	Name          string    `db:"name"`
	Description   string    `db:"description"`
	IsTraversable bool      `db:"is_traversable"`
	CreatedAt     null.Time `db:"created_at"`
	UpdatedAt     null.Time `db:"updated_at"`

	SchemaExtensionID int32 `db:"schema_extension_id"`
}

func toRelationshipKind(row relationshipKindRow) services.RelationshipKind {
	return services.RelationshipKind{
		ID:            row.ID,
		KindID:        row.KindID,
		Name:          row.Name,
		Description:   row.Description,
		IsTraversable: row.IsTraversable,
		CreatedAt:     row.CreatedAt.ValueOrZero(),
		UpdatedAt:     row.UpdatedAt.ValueOrZero(),
		Extension: services.Extension{
			ID: row.SchemaExtensionID,
		},
	}
}

// GetRelationshipKind reads one row from schema_relationship_kinds
func (s *Store) GetRelationshipKind(ctx context.Context, id int32) (services.RelationshipKind, error) {
	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()

	query, args := selectBuilder.Select(
		"rk.id",
		"rk.schema_extension_id",
		"rk.kind_id",
		"k.name",
		"rk.description",
		"rk.is_traversable",
		"rk.created_at",
		"rk.updated_at",
	).
		From(selectBuilder.As(tableSchemaRelationshipKinds, "rk")).
		Join(selectBuilder.As(tableKind, "k"), "rk.kind_id = k.id").
		Where(selectBuilder.Equal("rk.id", id)).
		Build()

	if rows, err := s.db.Query(ctx, query, args...); err != nil {
		return services.RelationshipKind{}, fmt.Errorf("fetching relationship kind: %w", err)
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[relationshipKindRow]); errors.Is(err, pgx.ErrNoRows) {
		return services.RelationshipKind{}, services.ErrRelationshipKindNotFound
	} else if err != nil {
		return services.RelationshipKind{}, fmt.Errorf("reading rows: %w", err)
	} else {
		return toRelationshipKind(row), nil
	}
}
