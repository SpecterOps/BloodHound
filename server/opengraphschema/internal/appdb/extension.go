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
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
)

const (
	tableSchemaExtensions = "schema_extensions"
)

const (
	extensionColumnID          = "id"
	extensionColumnName        = "name"
	extensionColumnDisplayName = "display_name"
	extensionColumnVersion     = "version"
	extensionColumnIsBuiltin   = "is_builtin"
	extensionColumnNamespace   = "namespace"
)

var extensionColumns = []string{
	extensionColumnID,
	extensionColumnName,
	extensionColumnDisplayName,
	extensionColumnVersion,
	extensionColumnIsBuiltin,
	extensionColumnNamespace,
}

// extensionRow is the package-local DB row type for a schema_extensions entry. db: tags
// drive pgx.RowToStructByName scanning.
type extensionRow struct {
	ID          int32  `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Version     string `db:"version"`
	IsBuiltin   bool   `db:"is_builtin"`
	Namespace   string `db:"namespace"`
}

// toExtension translates a raw extension row into the domain model.
func toExtension(row extensionRow) services.Extension {
	return services.Extension{
		ID:          row.ID,
		Name:        row.Name,
		DisplayName: row.DisplayName,
		Namespace:   row.Namespace,
		IsBuiltin:   row.IsBuiltin,
		Version:     row.Version,
	}
}

// GetExtension fetches an extension by its schema_extensions row id, returning
// ErrExtensionNotFound when no row matches.
func (s *Store) GetExtension(ctx context.Context, id int32) (services.Extension, error) {
	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()

	selectBuilder.Select(extensionColumns...)
	selectBuilder.From(tableSchemaExtensions)
	selectBuilder.Where(selectBuilder.Equal("id", id))

	sqlQuery, args := selectBuilder.Build()

	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return services.Extension{}, fmt.Errorf("fetching extension: %s", err)
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[extensionRow]); errors.Is(err, pgx.ErrNoRows) {
		return services.Extension{}, services.ErrExtensionNotFound
	} else if err != nil {
		return services.Extension{}, fmt.Errorf("reading rows: %s", err)
	} else {
		return toExtension(row), nil
	}
}

func (s *Store) ListExtensions(ctx context.Context) ([]services.Extension, error) {
	var (
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
		extensions    = make([]services.Extension, 0)
	)
	selectBuilder.Select(extensionColumns...).From(tableSchemaExtensions)
	selectBuilder.OrderByAsc(extensionColumnName)
	var sqlQuery, args = selectBuilder.Build()
	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return nil, fmt.Errorf("listing extensions: %s", err)
	} else if extensionRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[extensionRow]); err != nil {
		return nil, fmt.Errorf("reading extensions: %s", err)
	} else {
		for _, row := range extensionRows {
			extensions = append(extensions, toExtension(row))
		}
		return extensions, nil
	}
}
