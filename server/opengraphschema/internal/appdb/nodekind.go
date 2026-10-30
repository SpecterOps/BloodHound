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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
)

const (
	tableSchemaNodeKinds = "schema_node_kinds"
	// tableKind is the kind table joined when resolving node kinds to their names.
	tableKind                                    = "kind"
	graphSchemaNodeKindConstraintUniqueKindID    = "schema_node_kinds_kind_id_key"
	graphSchemaNodeKindConstraintExtensionIDFKey = "schema_node_kinds_schema_extension_id_fkey"
	// PostgresUniqueViolationCode is the SQLSTATE code for unique constraint violations.
	PostgresUniqueViolationCode = "23505"
	// PostgresForeignKeyViolationCode is the SQLSTATE code for foreign key constraint violations.
	PostgresForeignKeyViolationCode = "23503"
)

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

// CreateGraphSchemaNodeKind creates a schema node kind and its DAWGS kind if needed.
// Callers must refresh the DAWGS kind map after creating a kind.
func (s *Store) CreateGraphSchemaNodeKind(ctx context.Context, extensionID int32, nodeKind services.NodeInput) (services.NodeKind, error) {
	var (
		kindSelectBuilder   = sqlbuilder.PostgreSQL.NewSelectBuilder()
		insertBuilder       = sqlbuilder.PostgreSQL.NewInsertBuilder()
		insertSelectBuilder = insertBuilder.Select("dk.id")
		selectBuilder       = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	// The CTEs register the kind and insert its schema row atomically, then join
	// their results to include the kind name.
	kindSelectBuilder.Select("id", "name")
	kindSelectBuilder.From("upsert_kind(" + kindSelectBuilder.Var(nodeKind.Name) + ")")

	insertBuilder.InsertInto(tableSchemaNodeKinds)
	insertBuilder.Cols(
		"kind_id", "schema_extension_id", "display_name", "description",
		"is_display_kind", "icon", "icon_color",
	)
	insertSelectBuilder.SelectMore(
		insertSelectBuilder.Var(extensionID),
		insertSelectBuilder.Var(nodeKind.DisplayName),
		insertSelectBuilder.Var(nodeKind.Description),
		insertSelectBuilder.Var(nodeKind.IsDisplayKind),
		insertSelectBuilder.Var(nodeKind.Icon),
		insertSelectBuilder.Var(nodeKind.IconColor),
	)
	insertSelectBuilder.From(insertSelectBuilder.As("dawgs_kind", "dk"))
	insertBuilder.Returning(
		"id", "kind_id", "schema_extension_id", "display_name", "description",
		"is_display_kind", "icon", "icon_color", "created_at", "updated_at", "deleted_at",
	)

	selectBuilder.With(sqlbuilder.PostgreSQL.NewCTEBuilder().With(
		sqlbuilder.PostgreSQL.NewCTEQueryBuilder().Table("dawgs_kind", "id", "name").As(kindSelectBuilder),
		sqlbuilder.PostgreSQL.NewCTEQueryBuilder().Table("inserted_schema_node").As(insertBuilder),
	))
	selectBuilder.Select(
		"isn.id", "isn.kind_id", "isn.schema_extension_id", "dk.name",
		"isn.display_name", "isn.description", "isn.is_display_kind",
		"isn.icon", "isn.icon_color", "isn.created_at", "isn.updated_at", "isn.deleted_at",
	)
	selectBuilder.From(selectBuilder.As("inserted_schema_node", "isn"))
	selectBuilder.Join(selectBuilder.As("dawgs_kind", "dk"), "isn.kind_id = dk.id")

	sqlQuery, args := selectBuilder.Build()

	// pgx can defer statement errors until rows are consumed, so both query
	// and collection errors need translation.
	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return services.NodeKind{}, checkNodeKindError(fmt.Errorf("creating node kind: %w", err), nodeKind.Name)
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[nodeKindRow]); err != nil {
		return services.NodeKind{}, checkNodeKindError(fmt.Errorf("creating node kind: %w", err), nodeKind.Name)
	} else {
		return toNodeKind(row), nil
	}
}

// TODO: Port custom node kind CRUD from cmd/api/src/database/customnode.go:
// CreateCustomNodeKinds, GetCustomNodeKinds, UpdateCustomNodeKind, and DeleteCustomNodeKind.
// Preserve auditing and schema-node icon synchronization.

// GetGraphSchemaNodeKinds returns matching node kinds and the total count before pagination.
func (s *Store) GetGraphSchemaNodeKinds(ctx context.Context, filters params.Filters, sort params.SortItems, skip, limit int) ([]services.NodeKind, int, error) {
	var (
		selectBuilder   = sqlbuilder.PostgreSQL.NewSelectBuilder()
		countBuilder    = sqlbuilder.PostgreSQL.NewSelectBuilder()
		schemaNodeKinds = []services.NodeKind{}
		nodeKindRows    []nodeKindRow
		totalRowCount   int
		aliasedFilters  = make(params.Filters, len(filters))
		aliasedSorts    = make(params.SortItems, 0, len(sort))

		nodeKindColumnAliases = map[string]string{
			"schema_extension_id": "nk.schema_extension_id",
			"name":                "k.name",
			"id":                  "nk.id",
			"display_name":        "nk.display_name",
			"description":         "nk.description",
			"is_display_kind":     "nk.is_display_kind",
			"icon":                "nk.icon",
			"icon_color":          "nk.icon_color",
			"created_at":          "nk.created_at",
			"updated_at":          "nk.updated_at",
			"deleted_at":          "nk.deleted_at",
		}
	)

	selectBuilder.Select(
		"nk.id", "nk.kind_id", "k.name", "nk.schema_extension_id",
		"nk.display_name", "nk.description", "nk.is_display_kind", "nk.icon",
		"nk.icon_color", "nk.created_at", "nk.updated_at", "nk.deleted_at",
	)
	selectBuilder.From(selectBuilder.As(tableSchemaNodeKinds, "nk"))
	selectBuilder.Join(selectBuilder.As(tableKind, "k"), "nk.kind_id = k.id")

	countBuilder.Select("COUNT(*)")
	countBuilder.From(countBuilder.As(tableSchemaNodeKinds, "nk"))
	countBuilder.Join(countBuilder.As(tableKind, "k"), "nk.kind_id = k.id")

	if len(filters) > 0 {
		for filterColumn, filter := range filters {
			aliasedColumn, ok := nodeKindColumnAliases[filterColumn]
			if !ok {
				return schemaNodeKinds, 0, fmt.Errorf("node kind filter references unknown field %q", filterColumn)
			}
			aliasedFilters[aliasedColumn] = filter
		}
	}
	if len(sort) > 0 {
		for _, sortItem := range sort {
			aliasedColumn, ok := nodeKindColumnAliases[sortItem.Field]
			if !ok {
				return schemaNodeKinds, 0, fmt.Errorf("node kind sort references unknown field %q", sortItem.Field)
			}
			sortItem.Field = aliasedColumn
			if sortItem.Direction != params.Ascending && sortItem.Direction != params.Descending {
				return schemaNodeKinds, 0, fmt.Errorf("%w: %q", services.ErrInvalidSortDirection, sortItem.Direction)
			}
			aliasedSorts = append(aliasedSorts, sortItem)
		}
	} else {
		aliasedSorts = append(aliasedSorts, params.SortItem{Field: "nk.id", Direction: params.Ascending})
	}

	for _, builder := range []*sqlbuilder.SelectBuilder{selectBuilder, countBuilder} {
		if err := applyFilters(builder, aliasedFilters); err != nil {
			return schemaNodeKinds, 0, err
		}
	}

	for _, sortItem := range aliasedSorts {
		if sortItem.Direction == params.Descending {
			selectBuilder.OrderByDesc(sortItem.Field)
		} else {
			selectBuilder.OrderByAsc(sortItem.Field)
		}
	}

	if limit > 0 {
		selectBuilder.Limit(limit)
	}
	if skip > 0 {
		selectBuilder.Offset(skip)
	}

	sqlQuery, args := selectBuilder.Build()
	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("error fetching node kinds: %w", err)
	} else if nodeKindRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[nodeKindRow]); err != nil {
		return nil, 0, fmt.Errorf("error reading rows: %w", err)
	}

	// LIMIT/OFFSET applies only to the returned nodes. The total must count every
	// node matching the filters, so paginated requests need a separate COUNT query.
	totalRowCount = len(nodeKindRows)
	if limit > 0 || skip > 0 {
		countQuery, countArgs := countBuilder.Build()
		if rows, err := s.db.Query(ctx, countQuery, countArgs...); err != nil {
			return []services.NodeKind{}, 0, fmt.Errorf("error counting node kinds: %w", err)
		} else if totalRowCount, err = pgx.CollectOneRow(rows, pgx.RowTo[int]); err != nil {
			return []services.NodeKind{}, 0, fmt.Errorf("error reading node kind count: %w", err)
		}
	}

	for _, row := range nodeKindRows {
		schemaNodeKinds = append(schemaNodeKinds, toNodeKind(row))
	}

	return schemaNodeKinds, totalRowCount, nil
}

// GetGraphSchemaNodeKindsByExtensionId - retrieves all node kinds belonging to the given extension.
func (s *Store) GetGraphSchemaNodeKindsByExtensionId(ctx context.Context, extensionID int32) ([]services.NodeKind, error) {
	if nodeKinds, _, err := s.GetGraphSchemaNodeKinds(ctx,
		params.Filters{"schema_extension_id": {{
			Operator:    params.Equals,
			Value:       fmt.Sprintf("%d", extensionID),
			SetOperator: params.FilterAnd,
		}}}, params.SortItems{}, 0, 0); err != nil {
		return nil, err
	} else {
		return nodeKinds, nil
	}
}

// UpdateGraphSchemaNodeKind - updates a row in the schema_node_kinds table based on the provided id. It will return an
// error if the target schema node kind does not exist or if any of the updates violate the schema constraints.
//
// This function does NOT update the DAWGS name column since the schema_node_kinds table FKs to the DAWGS kind table, and that
// table is append only. A new node kind should be created instead.
func (s *Store) UpdateGraphSchemaNodeKind(ctx context.Context, nodeKind services.NodeKind) (services.NodeKind, error) {
	updateBuilder := sqlbuilder.PostgreSQL.NewUpdateBuilder()

	updateBuilder.Update(tableSchemaNodeKinds + " AS nk")
	updateBuilder.Set(
		updateBuilder.Assign("schema_extension_id", nodeKind.SchemaExtensionID),
		updateBuilder.Assign("display_name", nodeKind.DisplayName),
		updateBuilder.Assign("description", nodeKind.Description),
		updateBuilder.Assign("is_display_kind", nodeKind.IsDisplayKind),
		updateBuilder.Assign("icon", nodeKind.Icon),
		updateBuilder.Assign("icon_color", nodeKind.Color),
		updateBuilder.Assign("updated_at", sqlbuilder.Raw("NOW()")),
	)
	updateBuilder.From(tableKind + " AS k")
	updateBuilder.Where(
		updateBuilder.Equal("nk.id", nodeKind.ID),
		"nk.kind_id = k.id",
	)

	updateBuilder.Returning(
		"nk.id", "nk.kind_id", "nk.schema_extension_id", "k.name",
		"nk.display_name", "nk.description", "nk.is_display_kind", "nk.icon",
		"nk.icon_color", "nk.created_at", "nk.updated_at", "nk.deleted_at",
	)

	sqlQuery, args := updateBuilder.Build()

	// pgx can defer statement errors until rows are consumed, so both query
	// and collection errors need translation.
	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return services.NodeKind{}, checkNodeKindError(fmt.Errorf("updating node kind: %w", err), nodeKind.Name)
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[nodeKindRow]); errors.Is(err, pgx.ErrNoRows) {
		return services.NodeKind{}, services.ErrNodeKindNotFound
	} else if err != nil {
		return services.NodeKind{}, checkNodeKindError(fmt.Errorf("reading updated node kind: %w", err), nodeKind.Name)
	} else {
		return toNodeKind(row), nil
	}
}

// DeleteGraphSchemaNodeKind - deletes a schema_node_kinds row based on the provided id. Will return an error if that id does not exist.
func (s *Store) DeleteGraphSchemaNodeKind(ctx context.Context, nodeKindID int32) error {
	deleteBuilder := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	deleteBuilder.DeleteFrom(tableSchemaNodeKinds)
	deleteBuilder.Where(deleteBuilder.Equal("id", nodeKindID))

	sqlQuery, args := deleteBuilder.Build()

	if commandTag, err := s.db.Exec(ctx, sqlQuery, args...); err != nil {
		return fmt.Errorf("error deleting node kind: %w", err)
	} else if commandTag.RowsAffected() == 0 {
		return services.ErrNodeKindNotFound
	}

	return nil
}

func checkNodeKindError(err error, name string) error {
	var postgresError *pgconn.PgError

	if errors.As(err, &postgresError) {
		switch {
		case postgresError.Code == PostgresUniqueViolationCode &&
			postgresError.ConstraintName == graphSchemaNodeKindConstraintUniqueKindID:
			return fmt.Errorf("%w: %s", services.ErrDuplicateSchemaNodeKindName, name)
		case postgresError.Code == PostgresForeignKeyViolationCode &&
			postgresError.ConstraintName == graphSchemaNodeKindConstraintExtensionIDFKey:
			return fmt.Errorf("%w: %w", services.ErrExtensionNotFound, err)
		}
	}

	return err
}
