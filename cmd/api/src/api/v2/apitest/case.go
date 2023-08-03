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

package apitest

import (
	"net/http"

	"github.com/specterops/bloodhound/src/api"
)

type Case struct {
	Name  string
	Input InputFunc
	Setup func()
	Test  OutputFunc
}

// NewSortingErrorCase fails by giving an invalid `sort_by` column
func NewSortingErrorCase() Case {
	return Case{
		Name: "SortingError",
		Input: func(input *Input) {
			AddQueryParam(input, "sort_by", "definitelyInvalidColumn")
		},
		Test: func(output Output) {
			StatusCode(output, http.StatusBadRequest)
			BodyContains(output, api.ErrorResponseDetailsNotSortable)
		},
	}
}

// NewColumnNotFilterableCase fails by giving an invalid column for filtering
func NewColumnNotFilterableCase() Case {
	return Case{
		Name: "ColumnNotFilterable",
		Input: func(input *Input) {
			AddQueryParam(input, "definitelyInvalidColumn", "gt:0")
		},
		Test: func(output Output) {
			StatusCode(output, http.StatusBadRequest)
			BodyContains(output, api.ErrorResponseDetailsColumnNotFilterable)
		},
	}
}

// NewInvalidFilterPredicateCase fails when a valid column is given with an invalid predicate
func NewInvalidFilterPredicateCase(validColumn string) Case {
	return Case{
		Name: "InvalidFilterPredicate",
		Input: func(input *Input) {
			AddQueryParam(input, validColumn, "definitelyInvalidPredicate:0")
		},
		Test: func(output Output) {
			StatusCode(output, http.StatusBadRequest)
			BodyContains(output, api.ErrorResponseDetailsBadQueryParameterFilters)
		},
	}
}

// NewFilterPredicateMismatch fails when a valid column is given with a valid predicate that does not match
func NewFilterPredicateMismatch(validColumn string, mismatchedPredicate string) Case {
	return Case{
		Name: "FilterPredicateMismatch",
		Input: func(input *Input) {
			AddQueryParam(input, validColumn, mismatchedPredicate)
		},
		Test: func(output Output) {
			StatusCode(output, http.StatusBadRequest)
			BodyContains(output, api.ErrorResponseDetailsFilterPredicateNotSupported)
		},
	}
}
