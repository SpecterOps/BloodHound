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

// ParseAndValidate parses the supplied query parameters and validates them against the Filterable schema
// in a single pass, returning a fully-enriched Filters value. Every returned Filter has its SetOperator
// and IsStringData populated from the column definition. On failure it returns a *ValidationError wrapping
// one of the validation sentinels.
func (s QueryParameterFilterParser) ParseAndValidate(values url.Values, filterable Filterable) (Filters, error) {
	var (
		validFields   = filterable.ValidFilters()
		parsedFilters = Filters{}
	)

	for name, fieldValues := range values {
		if slices.Contains(s.ignoredParameters, name) {
			continue
		}

		for _, value := range fieldValues {
			if subgroups := s.valuePattern.FindStringSubmatch(value); len(subgroups) == 0 {
				continue
			} else if operator, err := ParseFilterOperator(subgroups[1]); err != nil {
				return nil, &ValidationError{Err: err}
			} else if field, isFilterable := validFields[name]; !isFilterable {
				return nil, &ValidationError{Err: ErrFieldNotFilterable, Field: name}
			} else if !slices.Contains(field.Operators, operator) {
				return nil, &ValidationError{Err: ErrOperatorNotSupported, Field: name, Operator: operator}
			} else {
				setOperator := field.SetOperator
				if setOperator == "" {
					setOperator = FilterAnd
				}

				parsedFilters[name] = append(parsedFilters[name], Filter{
					Field:        name,
					Operator:     operator,
					Value:        subgroups[2],
					SetOperator:  setOperator,
					IsStringData: field.IsStringData,
				})
			}
		}
	}

	return parsedFilters, nil
}
