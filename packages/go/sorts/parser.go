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

package sorts

import (
	"net/url"
	"strings"
)

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
// referenced field; its absence selects an ascending ordering. On failure it returns a *ValidationError
// wrapping one of the validation sentinels.
func (s QueryParameterSortParser) ParseAndValidate(values url.Values, sortable Sortable) (SortItems, error) {
	var (
		requestedFields = values[QueryParameter]
		parsedSort      = make(SortItems, 0, len(requestedFields))
	)

	for _, requestedField := range requestedFields {
		if requestedField == "" || requestedField == "-" {
			return nil, &ValidationError{Err: ErrFieldEmpty}
		}

		sortItem := SortItem{Field: requestedField, Direction: Ascending}
		if strings.HasPrefix(requestedField, DescendingPrefix) {
			sortItem.Field = strings.TrimPrefix(requestedField, DescendingPrefix)
			sortItem.Direction = Descending
		}

		if !sortable.IsSortable(sortItem.Field) {
			return nil, &ValidationError{Err: ErrFieldNotSortable, Field: sortItem.Field}
		}

		parsedSort = append(parsedSort, sortItem)
	}

	return parsedSort, nil
}
