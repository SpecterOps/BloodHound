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

// This file provides a storage-agnostic model for parsing and validating sort query parameters.
// It intentionally carries no knowledge of how the resulting order is ultimately applied (SQL, graph,
// etc.) so that it can be reused across the API without coupling to any particular persistence layer.

package params

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// SortDirection identifies the ordering applied to a sorted field.
type SortDirection string

const (
	Ascending  SortDirection = "asc"
	Descending SortDirection = "desc"
)

// DescendingPrefix is the leading character on a sort_by value that selects a descending ordering for the
// referenced field. Its absence selects an ascending ordering.
const DescendingPrefix = "-"

// Validation sentinels classify why a set of sort query parameters failed validation. Callers should
// classify failures with errors.Is and may extract the offending field via a *SortValidationError using
// errors.As. These are intentionally free of any transport- or presentation-specific wording.
var (
	ErrFieldEmpty       = errors.New("sort column cannot be empty")
	ErrFieldNotSortable = errors.New("column cannot be sorted")
)

// SortValidationError describes a sort validation failure. It wraps one of the validation sentinels so it can
// be classified with errors.Is, while also carrying the offending field so callers can build their own
// messaging without parsing strings.
type SortValidationError struct {
	Err   error
	Field string
}

func (s *SortValidationError) Error() string {
	if s.Field != "" {
		return fmt.Sprintf("%s: %s", s.Err, s.Field)
	}

	return s.Err.Error()
}

func (s *SortValidationError) Unwrap() error {
	return s.Err
}

// SortItem is a single validated sort applied to a field.
type SortItem struct {
	Field     string
	Direction SortDirection
}

// SortItems is an ordered set of SortItem values, preserving the order in which the fields were requested.
type SortItems []SortItem

// Sortable is implemented by types that expose which fields may be sorted on.
type Sortable interface {
	IsSortable(field string) bool
}

// QueryParameter is the query parameter name from which sort fields are read.
const QueryParameter = "sort_by"

// QueryParameterSortParser extracts sort items from request query parameters.
type QueryParameterSortParser struct{}

// NewQueryParameterSortParser returns a parser ready to parse and validate sort query parameters.
func NewQueryParameterSortParser() QueryParameterSortParser {
	return QueryParameterSortParser{}
}

// ParseAndValidate parses the sort_by query parameters and validates them against the Sortable schema in a
// single pass, returning an ordered SortItems value. A leading "-" selects a descending ordering for the
// referenced field; its absence selects an ascending ordering. On failure it returns a *SortValidationError
// wrapping one of the validation sentinels.
func (s QueryParameterSortParser) ParseAndValidate(values url.Values, sortable Sortable) (SortItems, error) {
	var (
		requestedFields = values[QueryParameter]
		parsedSort      = make(SortItems, 0, len(requestedFields))
	)

	for _, requestedField := range requestedFields {
		if requestedField == "" || requestedField == "-" {
			return nil, &SortValidationError{Err: ErrFieldEmpty}
		}

		sortItem := SortItem{Field: requestedField, Direction: Ascending}
		if strings.HasPrefix(requestedField, DescendingPrefix) {
			sortItem.Field = strings.TrimPrefix(requestedField, DescendingPrefix)
			sortItem.Direction = Descending
		}

		if !sortable.IsSortable(sortItem.Field) {
			return nil, &SortValidationError{Err: ErrFieldNotSortable, Field: sortItem.Field}
		}

		parsedSort = append(parsedSort, sortItem)
	}

	return parsedSort, nil
}
