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

// filtersortpagination handles helpers for applying filters,
// sorts, and pagination to DB queries.

package dbutil

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/huandu/go-sqlbuilder"
	"github.com/specterops/bloodhound/packages/go/params"
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
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
		return "", &ValidationError{Message: fmt.Sprintf("invalid operator \"%s\"", filter.Operator)}
	}
}

// ApplyFilters combines each column's filters using their set operators,
// then ANDs the column groups. Sorting columns keeps SQL and argument order deterministic.
func ApplyFilters(builder *sqlbuilder.SelectBuilder, filters params.Filters) error {
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

// ApplySorts applies the given sorts to the builder in the order listed
func ApplySorts(builder *sqlbuilder.SelectBuilder, sorts params.SortItems) {
	for _, sortItem := range sorts {
		if sortItem.Direction == params.Descending {
			builder.OrderByDesc(sortItem.Field)
		} else {
			builder.OrderByAsc(sortItem.Field)
		}
	}
}

func ApplyLimitSkip(builder *sqlbuilder.SelectBuilder, limit, skip int) {
	if limit > 0 {
		builder.Limit(limit)
	}
	if skip > 0 {
		builder.Offset(skip)
	}
}

type FilterSortSetup struct {
	ColumnAliases    map[string]string
	DefaultSortField string
}

type FilterSortPaginationInput struct {
	Filters params.Filters
	Sort    params.SortItems
	Skip    int
	Limit   int
}

// NormalizeFilterSort takes the given filters and sorts and applies the columnAliases and defaults provided,
// returning a normalized version of the filter and sort parameter lists.
func NormalizeFilterSort(setup FilterSortSetup, filters params.Filters, sort params.SortItems) (params.Filters, params.SortItems, error) {
	aliasedFilters := make(params.Filters, len(filters))
	aliasedSorts := make(params.SortItems, 0, len(sort))

	if len(filters) > 0 {
		for filterColumn, filter := range filters {
			aliasedColumn, ok := setup.ColumnAliases[filterColumn]
			if !ok {
				return nil, nil, &ValidationError{Message: fmt.Sprintf("filter references unknown field %q", filterColumn)}
			}
			aliasedFilters[aliasedColumn] = filter
		}
	}
	if len(sort) > 0 {
		for _, sortItem := range sort {
			aliasedColumn, ok := setup.ColumnAliases[sortItem.Field]
			if !ok {
				return nil, nil, &ValidationError{Message: fmt.Sprintf("sort references unknown field %q", sortItem.Field)}
			}
			sortItem.Field = aliasedColumn
			if sortItem.Direction != params.Ascending && sortItem.Direction != params.Descending {
				return nil, nil, &ValidationError{Message: fmt.Sprintf("invalid sort direction %q", sortItem.Direction)}
			}
			aliasedSorts = append(aliasedSorts, sortItem)
		}
	} else {
		aliasedSorts = append(aliasedSorts, params.SortItem{Field: setup.DefaultSortField, Direction: params.Ascending})
	}

	return aliasedFilters, aliasedSorts, nil
}

// EntityListerApplyFilterSortPagination takes a selectBuilder and countBuilder for
// and entity lister and applies the provided filter, sort, and pagination rules and values
// as appropriate to the given selectBuilder and countBuilder.
func EntityListerApplyFilterSortPagination(ctx context.Context, selectBuilder, countBuilder *sqlbuilder.SelectBuilder, setup FilterSortSetup, input FilterSortPaginationInput) error {
	aliasedFilters, aliasedSort, err := NormalizeFilterSort(setup, input.Filters, input.Sort)
	if err != nil {
		return err
	}

	for _, builder := range []*sqlbuilder.SelectBuilder{selectBuilder, countBuilder} {
		if err := ApplyFilters(builder, aliasedFilters); err != nil {
			return err
		}
	}

	ApplySorts(selectBuilder, aliasedSort)
	ApplyLimitSkip(selectBuilder, input.Limit, input.Skip)

	return nil
}
