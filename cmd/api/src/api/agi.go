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

package api

import (
	"fmt"
	"slices"
	"sort"
	"strconv"

	"github.com/specterops/bloodhound/src/model"
)

type AssetGroupMember struct {
	AssetGroupID    int      `json:"asset_group_id"`
	ObjectID        string   `json:"object_id"`
	PrimaryKind     string   `json:"primary_kind"`
	Kinds           []string `json:"kinds"`
	EnvironmentID   string   `json:"environment_id"`
	EnvironmentKind string   `json:"environment_kind"`
	Name            string   `json:"name"`
	CustomMember    bool     `json:"custom_member"`
}

type AssetGroupMembers []AssetGroupMember

// This model does not exist in the DBs and is just a display model. hence we cannot use
// the regular sorting logic used in most other spots as that hinges on the DB query.
func (s AssetGroupMembers) SortBy(columns []string) (AssetGroupMembers, error) {
	for _, column := range columns {
		if column == "" {
			return AssetGroupMembers{}, fmt.Errorf(ErrorResponseEmptySortParameter)
		}

		descending := false
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		switch column {
		case "object_id":
			sort.SliceStable(s, func(i, j int) bool {
				return s[i].ObjectID < s[j].ObjectID
			})
		case "asset_group_id":
			sort.SliceStable(s, func(i, j int) bool {
				return s[i].ObjectID < s[j].ObjectID
			})
		case "primary_kind":
			sort.SliceStable(s, func(i, j int) bool {
				return s[i].PrimaryKind < s[j].PrimaryKind
			})
		case "environment_id":
			sort.SliceStable(s, func(i, j int) bool {
				return s[i].EnvironmentID < s[j].EnvironmentID
			})
		case "environment_kind":
			sort.SliceStable(s, func(i, j int) bool {
				return s[i].EnvironmentKind < s[j].EnvironmentKind
			})
		case "name":
			sort.SliceStable(s, func(i, j int) bool {
				return s[i].Name < s[j].Name
			})
		default:
			return AssetGroupMembers{}, fmt.Errorf(ErrorResponseDetailsNotSortable)
		}

		if descending {
			slices.Reverse(s)
		}
	}
	return s, nil
}

// This model does not exist in the DBs and is just a display model. hence we cannot use
// the regular filtering logic used in most other spots as that hinges on the DB query..
func (s AssetGroupMembers) Filter(filterMap model.QueryParameterFilterMap) (AssetGroupMembers, error) {
	result := s
	for column, filters := range filterMap {
		if validPredicates, err := s.GetValidFilterPredicatesAsStrings(column); err != nil {
			return AssetGroupMembers{}, fmt.Errorf("%s: %s", model.ErrorResponseDetailsColumnNotFilterable, column)
		} else {
			for _, filter := range filters {
				if !slices.Contains(validPredicates, string(filter.Operator)) {
					return AssetGroupMembers{}, fmt.Errorf("%s: %s, %s", model.ErrorResponseDetailsFilterPredicateNotSupported, column, string(filter.Operator))
				} else if conditional, err := s.BuildFilteringConditional(column, filter.Operator, filter.Value); err != nil {
					return AssetGroupMembers{}, err
				} else {
					result = FilterStructSlice(result, conditional)
				}
			}
		}
	}
	return result, nil
}

func (s AssetGroupMembers) BuildFilteringConditional(column string, operator model.FilterOperator, value string) (func(t AssetGroupMember) bool, error) {
	switch column {
	case "object_id":
		if operator == model.Equals {
			return func(t AssetGroupMember) bool { return t.ObjectID == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMember) bool { return t.ObjectID != value }, nil
		} else {
			return nil, fmt.Errorf(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "primary_kind":
		if operator == model.Equals {
			return func(t AssetGroupMember) bool { return t.PrimaryKind == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMember) bool { return t.PrimaryKind != value }, nil
		} else {
			return nil, fmt.Errorf(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "environment_id":
		if operator == model.Equals {
			return func(t AssetGroupMember) bool { return t.EnvironmentID == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMember) bool { return t.EnvironmentID != value }, nil
		} else {
			return nil, fmt.Errorf(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "environment_kind":
		if operator == model.Equals {
			return func(t AssetGroupMember) bool { return t.EnvironmentKind == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMember) bool { return t.EnvironmentKind != value }, nil
		} else {
			return nil, fmt.Errorf(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "name":
		if operator == model.Equals {
			return func(t AssetGroupMember) bool { return t.Name == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMember) bool { return t.Name != value }, nil
		} else {
			return nil, fmt.Errorf(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "asset_group_id":
		if intValue, err := strconv.Atoi(value); err != nil {
			return nil, fmt.Errorf(ErrorResponseDetailsBadQueryParameterFilters)
		} else if operator == model.Equals {
			return func(t AssetGroupMember) bool { return t.AssetGroupID == intValue }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMember) bool { return t.AssetGroupID != intValue }, nil
		} else {
			return nil, fmt.Errorf(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "custom_member":
		if boolValue, err := strconv.ParseBool(value); err != nil {
			return nil, fmt.Errorf(ErrorResponseDetailsBadQueryParameterFilters)
		} else if operator == model.Equals {
			return func(t AssetGroupMember) bool { return t.CustomMember == boolValue }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMember) bool { return t.CustomMember != boolValue }, nil
		} else {
			return nil, fmt.Errorf(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	default:
		return nil, fmt.Errorf(ErrorResponseDetailsColumnNotFilterable)
	}
}

func (s AssetGroupMembers) ValidFilters() map[string][]model.FilterOperator {
	return map[string][]model.FilterOperator{
		"object_id":        {model.Equals, model.NotEquals},
		"primary_kind":     {model.Equals, model.NotEquals},
		"custom_member":    {model.Equals, model.NotEquals},
		"environment_id":   {model.Equals, model.NotEquals},
		"environment_kind": {model.Equals, model.NotEquals},
		"name":             {model.Equals, model.NotEquals},
	}
}

func (s AssetGroupMembers) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s AssetGroupMembers) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

type ListAssetGroupMembersResponse struct {
	Members AssetGroupMembers `json:"members"`
}

type ListAssetGroupMemberCountsResponse struct {
	TotalCount int            `json:"total_count"`
	Counts     map[string]int `json:"counts"`
}
