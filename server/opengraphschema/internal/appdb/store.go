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
	"fmt"
	"maps"
	"slices"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
)

// pgxQuerier lists only the pgx methods this package actually calls against PostgreSQL.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Store reads and writes OpenGraph extension schema data to/from PostgreSQL. Callers receive
// services-layer sentinels rather than raw driver errors.
type Store struct {
	db pgxQuerier
}

// NewStore returns a Store backed by the provided pgx connection pool.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

// buildFilterComparison builds a comparison for a validated, qualified SQL column.
func buildFilterComparison(builder *sqlbuilder.SelectBuilder, column string, filter params.Filter) (string, error) {
	switch filter.Operator {
	case params.Equals:
		if filter.Value == "null" {
			return builder.IsNull(column), nil
		}
		return builder.Equal(column, filter.Value), nil
	case params.NotEquals:
		if filter.Value == "null" {
			return builder.IsNotNull(column), nil
		}
		return builder.NotEqual(column, filter.Value), nil
	case params.IsNull:
		return builder.IsNull(column), nil
	case params.IsNotNull:
		return builder.IsNotNull(column), nil
	case params.GreaterThan:
		return builder.GreaterThan(column, filter.Value), nil
	case params.GreaterThanOrEquals:
		return builder.GreaterEqualThan(column, filter.Value), nil
	case params.LessThan:
		return builder.LessThan(column, filter.Value), nil
	case params.LessThanOrEquals:
		return builder.LessEqualThan(column, filter.Value), nil
	case params.ApproximatelyEquals:
		return builder.ILike(column, "%"+filter.Value+"%"), nil
	default:
		return "", fmt.Errorf("invalid operator specified")
	}
}

// applyFilters combines each column's filters using their set operators,
// then ANDs the column groups. Sorting columns keeps SQL and argument order deterministic.
func applyFilters(builder *sqlbuilder.SelectBuilder, filters params.Filters) error {
	for _, column := range slices.Sorted(maps.Keys(filters)) {
		var expression string
		for _, filter := range filters[column] {
			if comparison, err := buildFilterComparison(builder, column, filter); err != nil {
				return err
			} else if filter.SetOperator == params.FilterOr {
				expression = builder.Or(expression, comparison)
			} else {
				expression = builder.And(expression, comparison)
			}
		}

		builder.Where(expression)
	}

	return nil
}

// applySorts applies the given sorts to the builder in the order listed
func applySorts(builder *sqlbuilder.SelectBuilder, sorts params.SortItems) {
	for _, sortItem := range sorts {
		if sortItem.Direction == params.Descending {
			builder.OrderByDesc(sortItem.Field)
		} else {
			builder.OrderByAsc(sortItem.Field)
		}
	}
}

func applyLimitSkip(builder *sqlbuilder.SelectBuilder, limit, skip int) {
	if limit > 0 {
		builder.Limit(limit)
	}
	if skip > 0 {
		builder.Offset(skip)
	}
}

// buildEntityLister returns a function which applies filters, sorts, and pagination params
// to retrieve the given entity. RowType must be a DB-row-readable struct type corresponding with SelectBuilder.
func (s *Store) buildEntityLister[RowType any](
	selectBuilder, countBuilder *sqlbuilder.SelectBuilder,
	columnAliases map[string]string,
	defaultSortField string,
) func(ctx context.Context, extensionFilters params.Filters, sort params.SortItems, skip, limit int) ([]RowType, int, error) {
	return func(ctx context.Context, extensionFilters params.Filters, sort params.SortItems, skip, limit int) ([]RowType, int, error) {
		var (
			totalRowCount int
		)

		aliasedFilters, aliasedSort, err := normalizeFilterSort(columnAliases, defaultSortField, extensionFilters, sort)
		if err != nil {
			return nil, 0, err
		}

		for _, builder := range []*sqlbuilder.SelectBuilder{selectBuilder, countBuilder} {
			if err := applyFilters(builder, aliasedFilters); err != nil {
				return nil, 0, err
			}
		}

		applySorts(selectBuilder, aliasedSort)
		applyLimitSkip(selectBuilder, limit, skip)

		typedRows := []RowType{}
		sqlQuery, args := selectBuilder.Build()
		if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
			return nil, 0, fmt.Errorf("error fetching data: %w", err)
		} else if typedRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[RowType]); err != nil {
			return nil, 0, fmt.Errorf("error reading rows: %w", err)
		}

		totalRowCount = len(typedRows)
		if limit > 0 || skip > 0 {
			countQuery, countArgs := countBuilder.Build()
			if rows, err := s.db.Query(ctx, countQuery, countArgs...); err != nil {
				return nil, 0, fmt.Errorf("error counting data: %w", err)
			} else if totalRowCount, err = pgx.CollectOneRow(rows, pgx.RowTo[int]); err != nil {
				return nil, 0, fmt.Errorf("error reading row count: %w", err)
			}
		}

		return typedRows, totalRowCount, nil
	}
}

// normalizeFilterSort takes the given filters and sorts and applies the columnAliases and defaults provided,
// returning a normalized version of the filter and sort parameter lists.
func normalizeFilterSort(columnAliases map[string]string, defaultSortField string, filters params.Filters, sort params.SortItems) (params.Filters, params.SortItems, error) {
	aliasedFilters := make(params.Filters, len(filters))
	aliasedSorts := make(params.SortItems, 0, len(sort))

	if len(filters) > 0 {
		for filterColumn, filter := range filters {
			aliasedColumn, ok := columnAliases[filterColumn]
			if !ok {
				return nil, nil, fmt.Errorf("node kind filter references unknown field %q", filterColumn)
			}
			aliasedFilters[aliasedColumn] = filter
		}
	}
	if len(sort) > 0 {
		for _, sortItem := range sort {
			aliasedColumn, ok := columnAliases[sortItem.Field]
			if !ok {
				return nil, nil, fmt.Errorf("node kind sort references unknown field %q", sortItem.Field)
			}
			sortItem.Field = aliasedColumn
			if sortItem.Direction != params.Ascending && sortItem.Direction != params.Descending {
				return nil, nil, fmt.Errorf("%w: %q", services.ErrInvalidSortDirection, sortItem.Direction)
			}
			aliasedSorts = append(aliasedSorts, sortItem)
		}
	} else {
		aliasedSorts = append(aliasedSorts, params.SortItem{Field: defaultSortField, Direction: params.Ascending})
	}

	return aliasedFilters, aliasedSorts, nil
}
