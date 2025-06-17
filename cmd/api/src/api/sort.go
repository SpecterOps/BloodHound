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
	"fmt"
	"net/url"

	"github.com/specterops/dawgs/query"
	"github.com/specterops/bloodhound/src/model"
)

var (
	ErrResponseDetailsColumnEmpty         = errors.New("the specified column cannot be sorted because it is empty")
	ErrResponseDetailsColumnNotSortable   = errors.New("the specified column cannot be sorted")
	ErrResponseDetailsCriteriaEmpty       = errors.New("the specified criteria cannot be sorted because it is empty")
	ErrResponseDetailsCriteriaNotSortable = errors.New("the specified criteria cannot be sorted")
)

type Sortable interface {
	IsSortable(column string) bool
}

func ParseSortParameters(s Sortable, params url.Values) (model.Sort, error) {
	var (
		sortByColumns = params["sort_by"]
		sort          = make(model.Sort, 0, len(sortByColumns))
	)

	for _, column := range sortByColumns {
		var sortItem model.SortItem

		if column == "" {
			return sort, fmt.Errorf("%w: %s", ErrResponseDetailsColumnEmpty, sortItem.Column)
		} else if string(column[0]) == "-" {
			sortItem.Direction = model.DescendingSortDirection
			sortItem.Column = column[1:]
		} else {
			sortItem.Direction = model.AscendingSortDirection
			sortItem.Column = column
		}

		if !s.IsSortable(sortItem.Column) {
			return sort, fmt.Errorf("%w: %s", ErrResponseDetailsColumnNotSortable, sortItem.Column)
		}

		sort = append(sort, sortItem)
	}

	return sort, nil
}

func ParseGraphSortParameters(s Sortable, params url.Values) (query.SortItems, error) {
	var (
		sortByCriteria = params["sort_by"]
		sortItems      query.SortItems
	)

	for _, criteria := range sortByCriteria {
		sortItem := query.SortItem{}

		if criteria == "" {
			return query.SortItems{}, fmt.Errorf("%w: %s", ErrResponseDetailsCriteriaEmpty, criteria)
		} else if string(criteria[0]) == "-" {
			criteria = criteria[1:]
			sortItem.Direction = query.SortDirectionDescending
		} else {
			sortItem.Direction = query.SortDirectionAscending
		}

		if criteria == "id" {
			sortItem.SortCriteria = query.NodeID()
		} else {
			sortItem.SortCriteria = query.NodeProperty(criteria)
		}

		if !s.IsSortable(criteria) {
			return query.SortItems{}, fmt.Errorf("%w: %s", ErrResponseDetailsCriteriaNotSortable, criteria)
		}

		sortItems = append(sortItems, sortItem)
	}

	return sortItems, nil
}
