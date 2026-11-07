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
