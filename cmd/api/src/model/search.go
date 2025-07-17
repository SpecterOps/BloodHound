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
	"errors"
	"net/http"
	"slices"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

const (
	ErrResponseDetailsBadQueryParameterFilters    = "there are errors in the query parameter filters specified"
	ErrResponseDetailsFilterPredicateNotSupported = "the specified filter predicate is not supported for this column"
	ErrResponseDetailsColumnNotFilterable         = "the specified column cannot be filtered"
	ErrResponseDetailsColumnNotSortable           = "the specified column cannot be sorted"
)

type DomainSelector struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	ObjectID    string `json:"id"`
	Collected   bool   `json:"collected"`
	ImpactValue *int   `json:"impactValue,omitempty"`
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
		return []string{}, errors.New(ErrResponseDetailsColumnNotFilterable)
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func (s DomainSelectors) GetFilterCriteria(request *http.Request) (graph.Criteria, error) {
	var (
		queryParameterFilterParser = NewQueryParameterFilterParser()
		criteria                   graph.Criteria
	)

	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		return nil, errors.New(ErrResponseDetailsBadQueryParameterFilters)
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(s.GetFilterableColumns(), name); !valid {
				return nil, errors.New(ErrResponseDetailsColumnNotFilterable)
			}
			if validPredicates, err := s.GetValidFilterPredicatesAsStrings(name); err != nil {
				return nil, errors.New(ErrResponseDetailsColumnNotFilterable)
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						return nil, errors.New(ErrResponseDetailsFilterPredicateNotSupported)
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
