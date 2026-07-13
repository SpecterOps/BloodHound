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

// This file provides a storage-agnostic model for parsing and validating query parameter filters.
// It intentionally carries no knowledge of how filters are ultimately applied (SQL, graph, etc.) so that
// it can be reused across the API without coupling to any particular persistence layer.

package params

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// FilterOperator identifies the comparison predicate applied to a filtered column.
type FilterOperator string

const (
	GreaterThan         FilterOperator = "gt"
	GreaterThanOrEquals FilterOperator = "gte"
	LessThan            FilterOperator = "lt"
	LessThanOrEquals    FilterOperator = "lte"
	Equals              FilterOperator = "eq"
	NotEquals           FilterOperator = "neq"
	ApproximatelyEquals FilterOperator = "~eq"
)

// Validation sentinels classify why a set of query parameter filters failed validation. Callers should
// classify failures with errors.Is and may extract the offending field/operator via a *FilterValidationError
// using errors.As. These are intentionally free of any transport- or presentation-specific wording.
var (
	ErrMalformedFilter      = errors.New("query parameter filter is malformed")
	ErrFieldNotFilterable   = errors.New("column cannot be filtered")
	ErrOperatorNotSupported = errors.New("filter operator is not supported for column")
)

// FilterValidationError describes a filter validation failure. It wraps one of the validation sentinels so it
// can be classified with errors.Is, while also carrying the offending field and operator so callers can
// build their own messaging without parsing strings.
type FilterValidationError struct {
	Err      error
	Field    string
	Operator FilterOperator
}

func (s *FilterValidationError) Error() string {
	switch {
	case s.Field != "" && s.Operator != "":
		return fmt.Sprintf("%s: %s %s", s.Err, s.Field, s.Operator)
	case s.Field != "":
		return fmt.Sprintf("%s: %s", s.Err, s.Field)
	default:
		return s.Err.Error()
	}
}

func (s *FilterValidationError) Unwrap() error {
	return s.Err
}

// ParseFilterOperator validates a raw operator string and returns the corresponding FilterOperator.
func ParseFilterOperator(raw string) (FilterOperator, error) {
	switch operator := FilterOperator(raw); operator {
	case GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, Equals, NotEquals, ApproximatelyEquals:
		return operator, nil
	default:
		return "", fmt.Errorf("%w: unknown predicate %q", ErrMalformedFilter, raw)
	}
}

// FilterSetOperator describes how multiple filters on the same column are combined.
type FilterSetOperator string

const (
	FilterAnd FilterSetOperator = "AND"
	FilterOr  FilterSetOperator = "OR"
)

// Filter is a single validated filter applied to a field.
type Filter struct {
	Field        string
	Operator     FilterOperator
	Value        string
	SetOperator  FilterSetOperator
	IsStringData bool
}

// Filters maps a field name to the set of filters applied to it.
type Filters map[string][]Filter

// FilterableField describes a single field that may be filtered on. It carries the set of operators the
// field supports, IsStringData (whether the field holds string data), and SetOperator (how multiple
// filters applied to the field are combined). An empty SetOperator defaults to FilterAnd.
type FilterableField struct {
	Operators    []FilterOperator
	IsStringData bool
	SetOperator  FilterSetOperator
}

// Filterable is implemented by types that expose the fields that may be filtered along with each
// field's supported operators and data typing.
type Filterable interface {
	ValidFilters() map[string]FilterableField
}

// QueryParameterFilterParser extracts filters from request query parameters. Parameters whose names are
// listed in ignoredParameters are skipped, allowing callers to exclude application-specific concerns such
// as pagination parameters without coupling this package to them.
type QueryParameterFilterParser struct {
	ignoredParameters []string
}

// NewQueryParameterFilterParser returns a parser that ignores the supplied query parameter names.
func NewQueryParameterFilterParser(ignoredParameters ...string) QueryParameterFilterParser {
	return QueryParameterFilterParser{
		ignoredParameters: ignoredParameters,
	}
}

// ParseAndValidate parses the supplied query parameters and validates them against the Filterable schema
// in a single pass, returning a fully-enriched Filters value. Every returned Filter has its SetOperator
// and IsStringData populated from the column definition. On failure it returns a *FilterValidationError
// wrapping one of the validation sentinels.
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
			if subgroups := strings.SplitN(value, ":", 2); len(subgroups) != 2 {
				continue
			} else if operator, err := ParseFilterOperator(subgroups[0]); err != nil {
				return nil, &FilterValidationError{Err: err}
			} else if field, isFilterable := validFields[name]; !isFilterable {
				return nil, &FilterValidationError{Err: ErrFieldNotFilterable, Field: name}
			} else if !slices.Contains(field.Operators, operator) {
				return nil, &FilterValidationError{Err: ErrOperatorNotSupported, Field: name, Operator: operator}
			} else {
				setOperator := field.SetOperator
				if setOperator == "" {
					setOperator = FilterAnd
				}

				parsedFilters[name] = append(parsedFilters[name], Filter{
					Field:        name,
					Operator:     operator,
					Value:        subgroups[1],
					SetOperator:  setOperator,
					IsStringData: field.IsStringData,
				})
			}
		}
	}

	return parsedFilters, nil
}
