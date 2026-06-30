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

// Package filters provides a storage-agnostic model for parsing and validating query parameter filters.
// It intentionally carries no knowledge of how filters are ultimately applied (SQL, graph, etc.) so that
// it can be reused across the API without coupling to any particular persistence layer.
package filters

import (
	"errors"
	"fmt"
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
// classify failures with errors.Is and may extract the offending column/operator via a *ValidationError
// using errors.As. These are intentionally free of any transport- or presentation-specific wording.
var (
	ErrMalformedFilter      = errors.New("query parameter filter is malformed")
	ErrColumnNotFilterable  = errors.New("column cannot be filtered")
	ErrOperatorNotSupported = errors.New("filter operator is not supported for column")
)

// ValidationError describes a filter validation failure. It wraps one of the validation sentinels so it
// can be classified with errors.Is, while also carrying the offending column and operator so callers can
// build their own messaging without parsing strings.
type ValidationError struct {
	Err      error
	Column   string
	Operator FilterOperator
}

func (s *ValidationError) Error() string {
	switch {
	case s.Column != "" && s.Operator != "":
		return fmt.Sprintf("%s: %s %s", s.Err, s.Column, s.Operator)
	case s.Column != "":
		return fmt.Sprintf("%s: %s", s.Err, s.Column)
	default:
		return s.Err.Error()
	}
}

func (s *ValidationError) Unwrap() error {
	return s.Err
}

// ParseFilterOperator validates a raw operator string and returns the corresponding FilterOperator.
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
		return "", fmt.Errorf("%w: unknown predicate %q", ErrMalformedFilter, raw)
	}
}

// FilterSetOperator describes how multiple filters on the same column are combined.
type FilterSetOperator string

const (
	FilterAnd FilterSetOperator = "AND"
	FilterOr  FilterSetOperator = "OR"
)

// Filter is a single validated filter applied to a column.
type Filter struct {
	Operator     FilterOperator
	Value        string
	SetOperator  FilterSetOperator
	IsStringData bool
}

// Filters maps a column name to the set of filters applied to it.
type Filters map[string][]Filter

// FilterableColumn describes a single column that may be filtered on. It carries the set of operators the
// column supports, IsStringData (whether the column holds string data), and SetOperator (how multiple
// filters applied to the column are combined). An empty SetOperator defaults to FilterAnd.
type FilterableColumn struct {
	Operators    []FilterOperator
	IsStringData bool
	SetOperator  FilterSetOperator
}

// Filterable is implemented by types that expose the columns that may be filtered along with each
// column's supported operators and data typing.
type Filterable interface {
	ValidFilters() map[string]FilterableColumn
}
