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
	"github.com/specterops/bloodhound/server/extensions/internal/services"
)

const tableSchemaRelationshipKinds = "schema_relationship_kinds"

type relationshipKindRow struct {
	ID int32 `db:"id"`
}

func toRelationshipKind(row relationshipKindRow) services.RelationshipKind {
	return services.RelationshipKind{
		ID: row.ID,
	}
}

// GetRelationshipKind reads one row from schema_relationship_kinds
func (s *Store) GetRelationshipKind(ctx context.Context, id int32) (services.RelationshipKind, error) {
	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()

	query, args := selectBuilder.Select(
		"rk.id",
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
