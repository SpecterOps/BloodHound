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
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/common"
)

type FilterOperator string

const (
	GreaterThan         FilterOperator = "gt"
	GreaterThanOrEquals FilterOperator = "gte"
	LessThan            FilterOperator = "lt"
	LessThanOrEquals    FilterOperator = "lte"
	Equals              FilterOperator = "eq"
	NotEquals           FilterOperator = "neq"
	ApproximatelyEquals FilterOperator = "~eq"

	GreaterThanSymbol         string = ">"
	GreaterThanOrEqualsSymbol string = ">="
	LessThanSymbol            string = "<"
	LessThanOrEqualsSymbol    string = "<="
	EqualsSymbol              string = "="
	NotEqualsSymbol           string = "<>"
	ApproximatelyEqualSymbol  string = "ILIKE"

	NullString     = "null"
	TrueString     = "true"
	FalseString    = "false"
	IdString       = "id"
	ObjectIdString = "objectid"
)

type Filtered interface {
	ValidFilters() map[string][]FilterOperator
}

func ParseFilterOperator(raw string) (FilterOperator, error) {
	switch FilterOperator(raw) {
	case GreaterThan:
		return GreaterThan, nil

	case GreaterThanOrEquals:
		return GreaterThanOrEquals, nil

	case LessThan:
		return LessThan, nil

	case LessThanOrEquals:
		return LessThanOrEquals, nil

	case Equals:
		return Equals, nil

	case NotEquals:
		return NotEquals, nil

	case ApproximatelyEquals:
		return ApproximatelyEquals, nil

	default:
		return "", fmt.Errorf("unknown query parameter filter predicate: %s", raw)
	}
}

type SQLFilter struct {
	SQLString string
	Params    []any
}

type QueryParameterFilter struct {
	Name         string
	Operator     FilterOperator
	Value        string
	IsStringData bool
}

type QueryParameterFilters []QueryParameterFilter

func (s QueryParameterFilter) BuildGDBNodeFilter() graph.Criteria {
	var (
		propertyRef = query.NodeProperty(s.Name)
		value       = guessFilterValueType(s.Value)
	)

	// TODO: Investigate whether we can set the collected property for domains that originate from trusts in ParseDomainTrusts
	switch {
	case s.Name == common.Collected.String() && s.Operator == Equals:
		switch s.Value {
		case FalseString:
			return query.Or(
				query.Equals(propertyRef, false),
				query.Not(query.Exists(propertyRef)),
			)
		case TrueString:
			return query.Equals(propertyRef, true)
		}
	}

	switch s.Operator {
	case GreaterThan:
		return query.GreaterThan(propertyRef, value)
	case GreaterThanOrEquals:
		return query.GreaterThanOrEquals(propertyRef, value)
	case LessThan:
		return query.LessThan(propertyRef, value)
	case LessThanOrEquals:
		return query.LessThanOrEquals(propertyRef, value)
	case Equals:
		return query.Equals(propertyRef, value)
	case NotEquals:
		return query.Not(query.Equals(propertyRef, value))
	default:
		return nil
	}
}

type QueryParameterFilterMap map[string]QueryParameterFilters

func (s QueryParameterFilterMap) BuildSQLFilter() (SQLFilter, error) {
	var (
		result    strings.Builder
		predicate string
		params    []any
	)

	for _, filters := range s {
		for _, filter := range filters {
			if result.Len() > 0 {
				result.WriteString(" AND ")
			}

			switch filter.Operator {
			case GreaterThan:
				predicate = GreaterThanSymbol
			case GreaterThanOrEquals:
				predicate = GreaterThanOrEqualsSymbol
			case LessThan:
				predicate = LessThanSymbol
			case LessThanOrEquals:
				predicate = LessThanOrEqualsSymbol
			case Equals:
				predicate = EqualsSymbol
				// This would need to be updated for a nullable string column
				if filter.Value == NullString && !filter.IsStringData {
					result.WriteString(filter.Name + " IS NULL")
					continue
				}
			case NotEquals:
				predicate = NotEqualsSymbol
				// This would need to be updated for a nullable string column
				if filter.Value == NullString && !filter.IsStringData {
					result.WriteString(filter.Name + " IS NOT NULL")
					continue
				}
			case ApproximatelyEquals:
				predicate = ApproximatelyEqualSymbol
				filter.Value = fmt.Sprintf("%%%s%%", filter.Value)
			default:
				return SQLFilter{}, fmt.Errorf("invalid filter predicate specified")
			}

			_, _ = result.WriteString(fmt.Sprintf("%s %s ?", filter.Name, predicate))
			params = append(params, filter.Value)
		}
	}

	return SQLFilter{SQLString: result.String(), Params: params}, nil
}

func guessFilterValueType(raw string) any {
	if strings.ToLower(raw) == TrueString {
		return true
	}

	if strings.ToLower(raw) == FalseString {
		return false
	}

	if intValue, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return intValue
	}

	if floatValue, err := strconv.ParseFloat(raw, 64); err == nil {
		return floatValue
	}

	return raw
}

func (s QueryParameterFilterMap) BuildGDBNodeFilter() graph.Criteria {
	var criteria []graph.Criteria

	for _, filters := range s {
		for _, filter := range filters {
			criteria = append(criteria, filter.BuildGDBNodeFilter())
		}
	}

	return query.And(criteria...)
}

func (s QueryParameterFilterMap) BuildNeo4jFilter() (string, error) {
	var (
		result      = ""
		firstFilter = true
		predicate   string
	)

	for _, filters := range s {
		for _, filter := range filters {
			if !firstFilter {
				result = result + " AND "
			}

			switch filter.Operator {
			case GreaterThan:
				predicate = GreaterThanSymbol
			case GreaterThanOrEquals:
				predicate = GreaterThanOrEqualsSymbol
			case LessThan:
				predicate = LessThanSymbol
			case LessThanOrEquals:
				predicate = LessThanOrEqualsSymbol
			case Equals:
				predicate = EqualsSymbol
			case NotEquals:
				predicate = NotEqualsSymbol
			default:
				return "", fmt.Errorf("invalid filter predicate specified")
			}

			// our structs hold the data as id but the cypher column is actually objectid
			if filter.Name == IdString {
				filter.Name = ObjectIdString
			}

			filter.Name = fmt.Sprintf("n.%s", filter.Name)

			if filter.IsStringData {
				// for strings, add single quotes
				result = result + fmt.Sprintf("%s %s '%s'", filter.Name, predicate, filter.Value)
			} else {
				// for booleans, change the predicate to IS or IS NOT
				if (filter.Value == TrueString || filter.Value == FalseString) && !filter.IsStringData {
					if predicate == "=" {
						predicate = "IS"
					} else {
						predicate = "IS NOT"
					}
				}

				result = result + fmt.Sprintf("%s %s %s", filter.Name, predicate, filter.Value)
			}

			firstFilter = false
		}
	}

	return result, nil
}

func (s QueryParameterFilterMap) FirstFilter(name string) (QueryParameterFilter, bool) {
	if filters, hasFilters := s[name]; hasFilters {
		return filters[0], true
	}

	return QueryParameterFilter{}, false
}

func (s QueryParameterFilterMap) AddFilter(filter QueryParameterFilter) {
	if existingFilters, hasExisting := s[filter.Name]; hasExisting {
		s[filter.Name] = append(existingFilters, filter)
	} else {
		s[filter.Name] = QueryParameterFilters{filter}
	}
}

func (s QueryParameterFilterMap) IsFiltered(name string) bool {
	_, isFiltered := s[name]
	return isFiltered
}

type QueryParameterFilterParser struct {
	re *regexp.Regexp
}

func (s QueryParameterFilterParser) ParseQueryParameterFilter(name, value string) (QueryParameterFilter, error) {
	if subgroups := s.re.FindStringSubmatch(value); len(subgroups) > 0 {
		if filterPredicate, err := ParseFilterOperator(subgroups[1]); err != nil {
			return QueryParameterFilter{}, err
		} else {
			return QueryParameterFilter{
				Name:     name,
				Operator: filterPredicate,
				Value:    subgroups[2],
			}, nil
		}
	}
	return QueryParameterFilter{
		Name:     name,
		Operator: Equals,
		Value:    value,
	}, nil
}

func (s QueryParameterFilterParser) ParseQueryParameterFilters(request *http.Request) (QueryParameterFilterMap, error) {
	filters := make(QueryParameterFilterMap)

	for name, values := range request.URL.Query() {
		// ignore pagination query params
		if slices.Contains(AllPaginationQueryParameters(), name) {
			continue
		}

		if slices.Contains(IgnoreFilters(), name) {
			continue
		}

		for _, value := range values {
			if filter, err := s.ParseQueryParameterFilter(name, value); err != nil {
				return nil, err
			} else {
				filters.AddFilter(filter)
			}
		}
	}

	return filters, nil
}

func NewQueryParameterFilterParser() QueryParameterFilterParser {
	return QueryParameterFilterParser{
		re: regexp.MustCompile(`([~\w]+):([\w\--_ ]+)`),
	}
}

type Filter struct {
	Operator FilterOperator
	Value    string
}

type Filters map[string][]Filter
type ValidFilters map[string][]FilterOperator
type SortDirection int

const (
	InvalidSortDirection SortDirection = iota
	AscendingSortDirection
	DescendingSortDirection
)

type SortItem struct {
	Direction SortDirection
	Column    string
}

type Sort []SortItem
