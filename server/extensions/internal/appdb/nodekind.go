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
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
)

const tableSchemaNodeKinds = "schema_node_kinds"

// nodeKindRow is the package-local DB row type for a schema_node_kinds entry joined
// with the kind table for the kind name. db: tags drive pgx.RowToStructByName scanning.
type nodeKindRow struct {
	ID                int32     `db:"id"`
	SchemaExtensionID int32     `db:"schema_extension_id"`
	KindID            int32     `db:"kind_id"`
	Name              string    `db:"name"`
	DisplayName       string    `db:"display_name"`
	Description       string    `db:"description"`
	IsDisplayKind     bool      `db:"is_display_kind"`
	Icon              string    `db:"icon"`
	IconColor         string    `db:"icon_color"`
	CreatedAt         null.Time `db:"created_at"`
	UpdatedAt         null.Time `db:"updated_at"`
	DeletedAt         null.Time `db:"deleted_at"`
}

// toNodeKind translates a raw node kind row into the domain model. icon_color maps to
// the domain Color, timestamps resolve via ValueOrZero, and deleted_at resolves via Ptr
// (nil when null).
func toNodeKind(row nodeKindRow) services.NodeKind {
	return services.NodeKind{
		ID:                row.ID,
		SchemaExtensionID: row.SchemaExtensionID,
		KindID:            row.KindID,
		Name:              row.Name,
		DisplayName:       row.DisplayName,
		Description:       row.Description,
		IsDisplayKind:     row.IsDisplayKind,
		Icon:              row.Icon,
		Color:             row.IconColor,
		CreatedAt:         row.CreatedAt.ValueOrZero(),
		UpdatedAt:         row.UpdatedAt.ValueOrZero(),
		DeletedAt:         row.DeletedAt.Ptr(),
	}
}

// GetNodeKind fetches a node kind by its schema_node_kinds row id, returning
// ErrNodeKindNotFound when no row matches.
func (s *Store) GetNodeKind(ctx context.Context, id int32) (services.NodeKind, error) {
	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()

	selectBuilder.Select(
		"nk.id",
		"nk.schema_extension_id",
		"nk.kind_id",
		"k.name",
		"nk.display_name",
		"nk.description",
		"nk.is_display_kind",
		"nk.icon",
		"nk.icon_color",
		"nk.created_at",
		"nk.updated_at",
		"nk.deleted_at",
	)
	selectBuilder.From(selectBuilder.As(tableSchemaNodeKinds, "nk"))
	selectBuilder.Join(selectBuilder.As(tableKind, "k"), "nk.kind_id = k.id")
	selectBuilder.Where(selectBuilder.Equal("nk.id", id))

	sqlQuery, args := selectBuilder.Build()

	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return services.NodeKind{}, fmt.Errorf("fetching node kind: %w", err)
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[nodeKindRow]); errors.Is(err, pgx.ErrNoRows) {
		return services.NodeKind{}, services.ErrNodeKindNotFound
	} else if err != nil {
		return services.NodeKind{}, fmt.Errorf("reading rows: %w", err)
	} else {
		return toNodeKind(row), nil
	}
}
