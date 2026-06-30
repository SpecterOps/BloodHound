// Copyright 2026 Specter Ops, Inc.
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

package filters

import (
	"net/url"
	"regexp"
	"slices"
)

// queryParameterFilter is a single filter parsed from a request's query parameters. The column it applies
// to is carried as the key of the queryParameterFilterMap rather than duplicated on the struct. It is an
// internal staging type: ParseAndValidate is the only producer of a validated, enriched Filters value.
type queryParameterFilter struct {
	Operator FilterOperator
	Value    string
}

// queryParameterFilters is a collection of filters parsed for a single column.
type queryParameterFilters []queryParameterFilter

// queryParameterFilterMap maps a column name to the filters parsed for it.
type queryParameterFilterMap map[string]queryParameterFilters

// addFilter appends the given filter to the map under the supplied column name.
func (s queryParameterFilterMap) addFilter(name string, filter queryParameterFilter) {
	s[name] = append(s[name], filter)
}

// QueryParameterFilterParser extracts filters from request query parameters. Parameters whose names are
// listed in ignoredParameters are skipped, allowing callers to exclude application-specific concerns such
// as pagination parameters without coupling this package to them.
type QueryParameterFilterParser struct {
	valuePattern      *regexp.Regexp
	ignoredParameters []string
}

// NewQueryParameterFilterParser returns a parser that ignores the supplied query parameter names.
func NewQueryParameterFilterParser(ignoredParameters ...string) QueryParameterFilterParser {
	return QueryParameterFilterParser{
		valuePattern:      regexp.MustCompile(`([~\w]+):([\w\--_ ]+)`),
		ignoredParameters: ignoredParameters,
	}
}

// parseQueryParameterFilters parses all eligible query parameters into a filter map.
func (s QueryParameterFilterParser) parseQueryParameterFilters(values url.Values) (queryParameterFilterMap, error) {
	filters := make(queryParameterFilterMap)

	for name, columnValues := range values {
		if slices.Contains(s.ignoredParameters, name) {
			continue
		}

		for _, value := range columnValues {
			if subgroups := s.valuePattern.FindStringSubmatch(value); len(subgroups) > 0 {
				if filterPredicate, err := ParseFilterOperator(subgroups[1]); err != nil {
					return nil, err
				} else {
					filters.addFilter(name, queryParameterFilter{
						Operator: filterPredicate,
						Value:    subgroups[2],
					})
				}
			}
		}
	}

	return filters, nil
}

// ParseAndValidate parses the supplied query parameters and validates them against the Filterable schema,
// returning a fully-enriched Filters value. Every returned Filter has its IsStringData populated from the
// column definition. On failure it returns a *ValidationError wrapping one of the validation sentinels.
func (s QueryParameterFilterParser) ParseAndValidate(values url.Values, filterable Filterable) (Filters, error) {
	var (
		validColumns  = filterable.ValidFilters()
		parsedFilters = Filters{}
	)

	queryFilters, err := s.parseQueryParameterFilters(values)
	if err != nil {
		return nil, &ValidationError{Err: err}
	}

	for name, columnFilters := range queryFilters {
		column, isFilterable := validColumns[name]
		if !isFilterable {
			return nil, &ValidationError{Err: ErrColumnNotFilterable, Column: name}
		}

		setOperator := column.SetOperator
		if setOperator == "" {
			setOperator = FilterAnd
		}

		for _, columnFilter := range columnFilters {
			if !slices.Contains(column.Operators, columnFilter.Operator) {
				return nil, &ValidationError{Err: ErrOperatorNotSupported, Column: name, Operator: columnFilter.Operator}
			}

			parsedFilters[name] = append(parsedFilters[name], Filter{
				Operator:     columnFilter.Operator,
				Value:        columnFilter.Value,
				SetOperator:  setOperator,
				IsStringData: column.IsStringData,
			})
		}
	}

	return parsedFilters, nil
}
