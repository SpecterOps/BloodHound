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

	"github.com/specterops/bloodhound/errors"
)

type FilterOperator string

const (
	GreaterThan         FilterOperator = "gt"
	GreaterThanOrEquals FilterOperator = "gte"
	LessThan            FilterOperator = "lt"
	LessThanOrEquals    FilterOperator = "lte"
	Equals              FilterOperator = "eq"
	NotEquals           FilterOperator = "neq"

	GreaterThanSymbol         string = ">"
	GreaterThanOrEqualsSymbol string = ">="
	LessThanSymbol            string = "<"
	LessThanOrEqualsSymbol    string = "<="
	EqualsSymbol              string = "="
	NotEqualsSymbol           string = "<>"

	TrueString     = "true"
	FalseString    = "false"
	IdString       = "id"
	ObjectIdString = "objectid"

	ErrNotFiltered = errors.Error("parameter value is not filtered")
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
type QueryParameterFilterMap map[string]QueryParameterFilters

func (s QueryParameterFilterMap) BuildSQLFilter() (SQLFilter, error) {
	var (
		result      strings.Builder
		firstFilter = true
		predicate   string
		params      []any
	)

	for _, filters := range s {
		for _, filter := range filters {
			if !firstFilter {
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
			case NotEquals:
				predicate = NotEqualsSymbol
			default:
				return SQLFilter{}, fmt.Errorf("invalid filter predicate specified")
			}

			result.WriteString(filter.Name)
			result.WriteString(" ")
			result.WriteString(predicate)
			result.WriteString(" ?")
			params = append(params, filter.Value)
			firstFilter = false
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
			switch filter.Operator {
			case GreaterThan:
				criteria = append(criteria, query.GreaterThan(query.NodeProperty(filter.Name), guessFilterValueType(filter.Value)))

			case GreaterThanOrEquals:
				criteria = append(criteria, query.GreaterThanOrEquals(query.NodeProperty(filter.Name), guessFilterValueType(filter.Value)))

			case LessThan:
				criteria = append(criteria, query.LessThan(query.NodeProperty(filter.Name), guessFilterValueType(filter.Value)))

			case LessThanOrEquals:
				criteria = append(criteria, query.LessThanOrEquals(query.NodeProperty(filter.Name), guessFilterValueType(filter.Value)))

			case Equals:
				criteria = append(criteria, query.Equals(query.NodeProperty(filter.Name), guessFilterValueType(filter.Value)))

			case NotEquals:
				criteria = append(criteria, query.Not(query.Equals(query.NodeProperty(filter.Name), guessFilterValueType(filter.Value))))
			}
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

	return QueryParameterFilter{}, ErrNotFiltered
}

func (s QueryParameterFilterParser) ParseQueryParameterFilters(request *http.Request) (QueryParameterFilterMap, error) {
	filters := make(QueryParameterFilterMap)

	for name, values := range request.URL.Query() {
		// ignore pagination query params
		if slices.Contains(AllPaginationQueryParameters(), name) {
			continue
		}

		for _, value := range values {
			if filter, err := s.ParseQueryParameterFilter(name, value); err != nil {
				if err != ErrNotFiltered {
					return nil, err
				}
			} else {
				filters.AddFilter(filter)
			}
		}
	}

	return filters, nil
}

func NewQueryParameterFilterParser() QueryParameterFilterParser {
	return QueryParameterFilterParser{
		re: regexp.MustCompile(`([\w]+):([\w\d\--_]+)`),
	}
}
