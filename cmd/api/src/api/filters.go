// Copyright 2024 Specter Ops, Inc.
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

package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/src/model"
)

var (
	ErrColumnUnfilterable          = errors.New("the specified column cannot be filtered")
	ErrFilterPredicateNotSupported = errors.New("the specified filter predicate is not supported for this column")
	ErrNoFindingType               = errors.New("no finding type specified")
	ErrColumnFormatNotSupported    = errors.New("column format does not support sorting")
	ErrInvalidFindingType          = errors.New("invalid finding type specified")
	ErrInvalidAcceptedFilter       = errors.New("invalid finding type specified")
)

type Filterable interface {
	ValidFilters() map[string][]model.FilterOperator
}

func GetFilterableColumns(f Filterable) []string {
	var columns = make([]string, 0)
	for column := range f.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func GetValidFilterPredicatesAsStrings(f Filterable, column string) ([]string, error) {
	if predicates, validColumn := f.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func BuildSQLFilter(filters model.Filters) (model.SQLFilter, error) {
	var (
		result      strings.Builder
		firstFilter = true
		predicate   string
		params      []any
	)
	for name, filterOperations := range filters {
		for _, filter := range filterOperations {
			if !firstFilter {
				result.WriteString(" AND ")
			}
			switch filter.Operator {
			case model.GreaterThan:
				predicate = model.GreaterThanSymbol
			case model.GreaterThanOrEquals:
				predicate = model.GreaterThanOrEqualsSymbol
			case model.LessThan:
				predicate = model.LessThanSymbol
			case model.LessThanOrEquals:
				predicate = model.LessThanOrEqualsSymbol
			case model.Equals:
				predicate = model.EqualsSymbol
			case model.NotEquals:
				predicate = model.NotEqualsSymbol
			case model.ApproximatelyEquals:
				predicate = model.ApproximatelyEqualSymbol
				filter.Value = fmt.Sprintf("%%%s%%", filter.Value)
			default:
				return model.SQLFilter{}, fmt.Errorf("invalid filter predicate specified")
			}
			result.WriteString(name)
			result.WriteString(" ")
			result.WriteString(predicate)
			result.WriteString(" ?")
			params = append(params, filter.Value)
			firstFilter = false
		}
	}
	return model.SQLFilter{SQLString: result.String(), Params: params}, nil
}

func BuildSQLSort(sort model.Sort) []string {
	var sqlSort = make([]string, 0, len(sort))
	for _, sortItem := range sort {
		var column string
		if sortItem.Direction == model.DescendingSortDirection {
			column = sortItem.Column + " desc"
		} else {
			column = sortItem.Column
		}
		sqlSort = append(sqlSort, column)
	}
	return sqlSort
}
