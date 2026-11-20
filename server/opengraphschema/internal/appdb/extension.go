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
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/packages/go/dbutil"
	"github.com/specterops/bloodhound/packages/go/params"
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
	extensionColumnCreatedAt   = "created_at"
	extensionColumnUpdatedAt   = "updated_at"
)

var extensionColumns = []string{
	extensionColumnID,
	extensionColumnName,
	extensionColumnDisplayName,
	extensionColumnVersion,
	extensionColumnIsBuiltin,
	extensionColumnNamespace,
	extensionColumnCreatedAt,
	extensionColumnUpdatedAt,
}

// extensionRow is the package-local DB row type for a schema_extensions entry. db: tags
// drive pgx.RowToStructByName scanning.
type extensionRow struct {
	ID          int32     `db:"id"`
	Name        string    `db:"name"`
	DisplayName string    `db:"display_name"`
	Version     string    `db:"version"`
	IsBuiltin   bool      `db:"is_builtin"`
	Namespace   string    `db:"namespace"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
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
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
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

// ListExtensions gets all the rows from the extensions table that match the given filter,
// sorted by the given sorts and paged by the given pagination params.
// It returns a slice of Extension structs populated with the data,
// as well as an integer giving the total number of rows returned by the query (excluding any given pagination)
func (s *Store) ListExtensions(ctx context.Context, extensionFilters params.Filters, sort params.SortItems, skip, limit int) ([]services.Extension, int, error) {
	var (
		extensions = []services.Extension{}
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(extensionColumns...)
	selectBuilder.From(tableSchemaExtensions)

	countBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countBuilder.Select("COUNT(*)")
	countBuilder.From(tableSchemaExtensions)

	// defines the columns we support filter/pagination on
	columnAliases := map[string]string{
		"display_name": "display_name",
		"name":         "name",
		"is_builtin":   "is_builtin",
	}

	err := dbutil.EntityListerApplyFilterSortPagination(ctx, selectBuilder, countBuilder,
		dbutil.FilterSortSetup{
			ColumnAliases:    columnAliases,
			DefaultSortField: "id",
		},
		dbutil.FilterSortPaginationInput{
			Filters: extensionFilters,
			Sort:    sort,
			Skip:    skip,
			Limit:   limit,
		})
	if err != nil {
		return nil, 0, err
	}

	var extensionRows []extensionRow
	sqlQuery, args := selectBuilder.Build()
	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("error fetching data: %w", err)
	} else if extensionRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[extensionRow]); err != nil {
		return nil, 0, fmt.Errorf("error reading rows: %w", err)
	}

	totalRowCount := len(extensionRows)
	if limit > 0 || skip > 0 {
		countQuery, countArgs := countBuilder.Build()
		if rows, err := s.db.Query(ctx, countQuery, countArgs...); err != nil {
			return nil, 0, fmt.Errorf("error counting data: %w", err)
		} else if totalRowCount, err = pgx.CollectOneRow(rows, pgx.RowTo[int]); err != nil {
			return nil, 0, fmt.Errorf("error reading row count: %w", err)
		}
	}

	for _, row := range extensionRows {
		extensions = append(extensions, toExtension(row))
	}

	return extensions, totalRowCount, nil
}

// CreateExtension creates a new row in the extensions table.
// Returns an Extension populated with the value as it stands in the database.
func (s *Store) CreateExtension(ctx context.Context, extension services.Extension) (services.Extension, error) {
	// TODO BED-10577: also insert audit log

	var insertBuilder = sqlbuilder.PostgreSQL.NewInsertBuilder()

	insertBuilder.InsertInto(tableSchemaExtensions)
	insertBuilder.Cols(
		"name", "display_name", "version", "is_builtin", "namespace", "created_at", "updated_at",
	)
	insertBuilder.Values(
		extension.Name,
		extension.DisplayName,
		extension.Version,
		false, // extensions created here are not "builtin" by definition
		extension.Namespace,
		time.Now(),
		time.Now(),
	)
	insertBuilder.Returning(extensionColumns...)

	sqlQuery, args := insertBuilder.Build()

	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return services.Extension{}, fmt.Errorf("creating extension: %w", checkExtensionError(err))
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[extensionRow]); err != nil {
		return services.Extension{}, fmt.Errorf("reading created extension: %w", checkExtensionError(err))
	} else {
		return toExtension(row), nil
	}
}

var (
	graphSchemaExtensionConstraintUniqueNamespace = "schema_extensions_namespace_unique"
	graphSchemaExtensionConstraintUniqueName      = "schema_extensions_name_key"
)

func checkExtensionError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case graphSchemaExtensionConstraintUniqueNamespace:
			return fmt.Errorf("%w: %w", services.ErrDuplicateExtensionNamespace, err)
		case graphSchemaExtensionConstraintUniqueName:
			return fmt.Errorf("%w: %w", services.ErrDuplicateExtensionName, err)
		}
	}

	return err
}

// UpdateGraphSchemaExtension updates an existing Graph Schema Extension.
// Only the `name`, `display_name`, and `version` fields are updatable.
// It returns the updated extension, or an error if the update violates schema constraints or did not succeed.
func (s *Store) UpdateExtension(ctx context.Context, extension services.Extension) (services.Extension, error) {
	updateBuilder := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	updateBuilder.Update(tableSchemaExtensions)
	updateBuilder.Set(
		updateBuilder.Assign("name", extension.Name),
		updateBuilder.Assign("display_name", extension.DisplayName),
		updateBuilder.Assign("version", extension.Version),
		updateBuilder.Assign("namespace", extension.Namespace),
		updateBuilder.Assign("updated_at", time.Now()),
	)

	updateBuilder.Where(updateBuilder.Equal("id", extension.ID))
	updateBuilder.Returning(extensionColumns...)

	sqlQuery, args := updateBuilder.Build()

	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return services.Extension{}, fmt.Errorf("updating extension: %w", checkExtensionError(err))
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[extensionRow]); errors.Is(err, pgx.ErrNoRows) {
		return services.Extension{}, services.ErrExtensionNotFound
	} else if err != nil {
		return services.Extension{}, fmt.Errorf("reading updated extension: %w", checkExtensionError(err))
	} else {
		return toExtension(row), nil
	}
}

func (s *Store) DeleteExtension(ctx context.Context, extensionID int32) error {
	// TODO BED-10577: also insert audit log

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %s", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Retrieve the extension to check if it exists and if it's built-in
	selectBuiltinBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuiltinBuilder.Select("is_builtin")
	selectBuiltinBuilder.From(tableSchemaExtensions)
	selectBuiltinBuilder.Where(selectBuiltinBuilder.Equal("id", extensionID))
	sql, args := selectBuiltinBuilder.Build()

	var isBuiltin bool
	if err = tx.QueryRow(ctx, sql, args...).Scan(&isBuiltin); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return services.ErrExtensionNotFound
		}
		return fmt.Errorf("failure preparing to delete extension: %w", err)
	}

	// Prevent deletion of built-in extensions
	if isBuiltin {
		return services.ErrGraphExtensionBuiltIn
	}

	// Before deleting the extension (which cascades to schema_node_kinds), create stubs in custom_node_kinds for all non-display
	// node kinds. Non-display kinds are not tracked in custom_node_kinds while the schema is active. It is possible that a schemaless
	// node kind of the same name existed before this extension upserted it to a non-display node kind, which removes it from the custom
	// node table. Because of this edge case, inserting stubs here for the non-display node kinds ensures that those node kinds will
	// remain tracked even after the schema is gone, in the event any nodes of those kinds remain in the graph.
	nodeKinds, err := getGraphSchemaNodeKindsByExtensionId(ctx, tx, extensionID)
	if err != nil {
		return fmt.Errorf("failed to fetch node kinds for extension %d: %w", extensionID, err)
	}

	for _, nodeKind := range nodeKinds {
		if !nodeKind.IsDisplayKind {
			// Ignore ErrDuplicateCustomNodeKindName. This indicates the kind is already tracked, nothing to do.
			_, err := insertCustomNodeKind(ctx, tx, services.CustomNodeKind{KindID: nodeKind.KindID, KindName: nodeKind.Name, Config: CustomNodeKindStubConfig})
			if err != nil && !errors.Is(err, services.ErrDuplicateCustomNodeKindName) {
				return fmt.Errorf("failed to create stub for non-display node kind %q: %w", nodeKind.Name, err)
			}
		}
	}

	// Now delete the extension
	deleteBuilder := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	deleteBuilder.DeleteFrom(tableSchemaExtensions)
	deleteBuilder.Where(deleteBuilder.Equal("id", extensionID))
	deleteBuilder.Returning("id")
	sqlQuery, args := deleteBuilder.Build()
	if rows, err := tx.Query(ctx, sqlQuery, args...); err != nil {
		return fmt.Errorf("deleting extension %d: %w", extensionID, err)
	} else if _, err := pgx.CollectOneRow(rows, pgx.RowTo[int32]); errors.Is(err, pgx.ErrNoRows) {
		return services.ErrExtensionNotFound
	} else if err != nil {
		return fmt.Errorf("reading deleted extension %d: %w", extensionID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing extension deletion: %w", err)
	}

	return nil
}
