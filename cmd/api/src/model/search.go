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
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

// SchemaEnvironmentReader defines the interface for reading schema environments.
type SchemaEnvironmentReader interface {
	GetSchemaEnvironments(ctx context.Context) ([]SchemaEnvironment, error)
}

const (
	ErrResponseDetailsBadQueryParameterFilters    = "there are errors in the query parameter filters specified"
	ErrResponseDetailsFilterPredicateNotSupported = "the specified filter predicate is not supported for this column"
	ErrResponseDetailsColumnNotFilterable         = "the specified column cannot be filtered"
	ErrResponseDetailsColumnNotSortable           = "the specified column cannot be sorted"
)

type EnvironmentSelector struct {
	Type               string    `json:"type"`
	Name               string    `json:"name"`
	ObjectID           string    `json:"id"`
	Collected          bool      `json:"collected"`
	ImpactValue        *int      `json:"impactValue,omitempty"`
	HygieneAttackPaths *int64    `json:"hygiene_attack_paths,omitempty"` // caution: if value is bigger than maxsafeint, the UI will truncate the value
	Exposures          Exposures `json:"exposures,omitempty"`
}

type EnvironmentSelectors []EnvironmentSelector

type Exposure struct {
	ExposurePercent int           `json:"exposure_percent"`
	AssetGroupTag   AssetGroupTag `json:"asset_group_tag"`
}

type Exposures []Exposure

func (s EnvironmentSelectors) IsSortable(column string) bool {
	switch column {
	case "objectid",
		"name":
		return true
	default:
		return false
	}
}

func (s EnvironmentSelectors) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"objectid":  {Equals, NotEquals},
		"name":      {Equals, NotEquals},
		"collected": {Equals, NotEquals},
	}
}

func (s EnvironmentSelectors) IsString(column string) bool {
	switch column {
	case "name",
		"objectid",
		"collected":
		return true
	default:
		return false
	}
}

func (s EnvironmentSelectors) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s EnvironmentSelectors) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
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

func (s EnvironmentSelectors) GetFilterCriteria(request *http.Request, environmentKinds []graph.Kind) (graph.Criteria, error) {
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

		criteria = query.And(queryFilters.BuildGDBNodeFilter(), query.KindIn(query.Node(), environmentKinds...))
		return criteria, nil
	}
}
