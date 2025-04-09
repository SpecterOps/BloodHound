// Copyright 2025 Specter Ops, Inc.
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
	"net/url"

	"github.com/specterops/bloodhound/dawgs/query"
)

const (
	ErrResponseDetailsBadQueryParameterFilters    = "there are errors in the query parameter filters specified"
	ErrResponseDetailsFilterPredicateNotSupported = "the specified filter predicate is not supported for this column"
	ErrResponseDetailsColumnNotFilterable         = "the specified column cannot be filtered"
	ErrResponseDetailsColumnNotSortable           = "the specified column cannot be sorted"
)

type Sortable interface {
	IsSortable(column string) bool
}

func GetGraphSortItems(s Sortable, params url.Values) (query.SortItems, error) {
	var (
		sortByColumns = params["sort_by"]
		sortItems     query.SortItems
	)

	for _, column := range sortByColumns {
		sortItem := query.SortItem{}

		if string(column[0]) == "-" {
			column = column[1:]
			sortItem.Direction = query.SortDirectionDescending
		} else {
			sortItem.Direction = query.SortDirectionAscending
		}
		sortItem.SortCriteria = query.NodeProperty(column)

		if !s.IsSortable(column) {
			return query.SortItems{}, errors.New(ErrResponseDetailsColumnNotSortable)
		}

		sortItems = append(sortItems, sortItem)
	}
	return sortItems, nil
}
