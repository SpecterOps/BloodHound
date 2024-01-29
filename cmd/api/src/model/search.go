// Copyright 2023 Specter Ops, Inc.
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

package model

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
)

type DomainSelector struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	ObjectID  string `json:"id"`
	Collected bool   `json:"collected"`
}

type DomainSelectors []DomainSelector

func (s DomainSelectors) IsSortable(column string) bool {
	switch column {
	case "objectid",
		"name":
		return true
	default:
		return false
	}
}

func (s DomainSelectors) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"objectid":  {Equals, NotEquals},
		"name":      {Equals, NotEquals},
		"collected": {Equals, NotEquals},
	}
}

func (s DomainSelectors) IsString(column string) bool {
	switch column {
	case "name",
		"objectid",
		"collected":
		return true
	default:
		return false
	}
}

func (s DomainSelectors) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s DomainSelectors) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf(ErrorResponseDetailsColumnNotFilterable)
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

type OrderCriterion struct {
	Property string
	Order    graph.Criteria
}

type OrderCriteria []OrderCriterion

const (
	ErrorResponseDetailsBadQueryParameterFilters    = "there are errors in the query parameter filters specified"
	ErrorResponseDetailsFilterPredicateNotSupported = "the specified filter predicate is not supported for this column"
	ErrorResponseDetailsColumnNotFilterable         = "the specified column cannot be filtered"
	ErrorResponseDetailsColumnNotSortable           = "the specified column cannot be sorted"
)

func (s DomainSelectors) GetOrderCriteria(params url.Values) (OrderCriteria, error) {
	var (
		sortByColumns = params["sort_by"]
		orderCriteria OrderCriteria
	)

	for _, column := range sortByColumns {
		criterion := OrderCriterion{}

		if string(column[0]) == "-" {
			column = column[1:]
			criterion.Order = query.Descending()
		} else {
			criterion.Order = query.Ascending()
		}
		criterion.Property = column

		if !s.IsSortable(column) {
			return OrderCriteria{}, fmt.Errorf(ErrorResponseDetailsColumnNotSortable)
		}

		orderCriteria = append(orderCriteria, criterion)
	}
	return orderCriteria, nil
}

func (s DomainSelectors) GetFilterCriteria(request *http.Request) (graph.Criteria, error) {
	var (
		queryParameterFilterParser = NewQueryParameterFilterParser()
		criteria                   graph.Criteria
	)

	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		return nil, fmt.Errorf(ErrorResponseDetailsBadQueryParameterFilters)
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(s.GetFilterableColumns(), name); !valid {
				return nil, fmt.Errorf(ErrorResponseDetailsColumnNotFilterable)
			}

			if validPredicates, err := s.GetValidFilterPredicatesAsStrings(name); err != nil {
				return nil, fmt.Errorf(ErrorResponseDetailsColumnNotFilterable)
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						return nil, fmt.Errorf(ErrorResponseDetailsFilterPredicateNotSupported)
					}

					queryFilters[name][i].IsStringData = s.IsString(filter.Name)
				}
			}
		}
		// ignoring the error here as this would've failed at ParseQueryParameterFilters before getting here
		criteria = query.And(queryFilters.BuildGDBNodeFilter(), query.KindIn(query.Node(), ad.Domain, azure.Tenant))

		return criteria, nil
	}
}
