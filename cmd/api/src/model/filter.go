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
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
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

var ErrNotFiltered = errors.New("parameter value is not filtered")

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

// ToFiltersModel converts the query parameter filter map model into the newer, more generic filter
// model. This is useful when accessing newer functions that expect the newer model without having
// refactor multiple sites all-at-once.
func (s QueryParameterFilterMap) ToFiltersModel() Filters {
	convertedFilters := Filters{}

	for name, oldModelFilters := range s {
		newModelFilters := make([]Filter, len(oldModelFilters))

		for idx, oldModelFilter := range oldModelFilters {
			newModelFilters[idx] = Filter{
				Operator: oldModelFilter.Operator,
				Value:    oldModelFilter.Value,
			}
		}

		convertedFilters[name] = newModelFilters
	}

	return convertedFilters
}

// filterValueAsPGLiteral takes a string value and returns a PG SQL literal that represents the value in the form of a
// Go struct. This function will attempt to parse the string value into different types but otherwise defaults to the
// given string value as a wrapped literal.
func filterValueAsPGLiteral(valueStr string, isNullValue bool) (pgsql.Literal, error) {
	if isNullValue {
		return pgsql.NullLiteral(), nil
	}

	if valueInt64, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return pgsql.NewLiteral(valueInt64, pgsql.Int8), nil
	}

	if valueFloat64, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return pgsql.NewLiteral(valueFloat64, pgsql.Float8), nil
	}

	if valueBool, err := strconv.ParseBool(valueStr); err == nil {
		return pgsql.NewLiteral(valueBool, pgsql.Boolean), nil
	}

	// Default to the pgsql package type conversion
	return pgsql.AsLiteral(valueStr)
}

// BuildSQLFilter builds a PGSQL syntax-correct SQLFilter result from the given Filters struct. This function
// uses the PGSQL AST to ensure formatted SQL correctness.
func BuildSQLFilter(filters Filters) (SQLFilter, error) {
	var whereClauseFragment pgsql.Expression

	for name, filterOperations := range filters {
		for _, filter := range filterOperations {
			var (
				operator    pgsql.Operator
				filterValue = filter.Value
				isNullValue = filterValue == NullString
			)

			switch filter.Operator {
			case GreaterThan:
				operator = pgsql.OperatorGreaterThan

			case GreaterThanOrEquals:
				operator = pgsql.OperatorGreaterThanOrEqualTo

			case LessThan:
				operator = pgsql.OperatorLessThan

			case LessThanOrEquals:
				operator = pgsql.OperatorLessThanOrEqualTo

			case Equals:
				if isNullValue {
					operator = pgsql.OperatorIs
				} else {
					operator = pgsql.OperatorEquals
				}

			case NotEquals:
				if isNullValue {
					operator = pgsql.OperatorIsNot
				} else {
					operator = pgsql.OperatorNotEquals
				}

			case ApproximatelyEquals:
				operator = pgsql.OperatorLike
				filterValue = "%" + filterValue + "%"

			default:
				return SQLFilter{}, fmt.Errorf("invalid operator specified")
			}

			if literalValue, err := filterValueAsPGLiteral(filterValue, isNullValue); err != nil {
				return SQLFilter{}, fmt.Errorf("invalid filter value specified for %s: %w", name, err)
			} else {
				whereClauseFragment = pgsql.OptionalAnd(whereClauseFragment, pgsql.NewBinaryExpression(
					pgsql.Identifier(name),
					operator,
					literalValue,
				))
			}
		}
	}

	var filter SQLFilter

	if whereClauseFragment != nil {
		if sqlFragment, err := format.SyntaxNode(whereClauseFragment); err != nil {
			return filter, fmt.Errorf("failed formatting SQL filter: %w", err)
		} else {
			filter = SQLFilter{
				SQLString: sqlFragment,
			}
		}
	}

	return filter, nil
}

func (s QueryParameterFilterMap) BuildSQLFilter() (SQLFilter, error) {
	return BuildSQLFilter(s.ToFiltersModel())
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

	return QueryParameterFilter{}, ErrNotFiltered
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
				if !errors.Is(err, ErrNotFiltered) {
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
