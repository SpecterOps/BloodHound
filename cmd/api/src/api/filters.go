// Copyright 2024 Specter Ops, Inc.
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

package api

import (
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/src/model"
)

var (
	ErrColumnUnfilterable          = errors.New("the specified column cannot be filtered")
	ErrFilterPredicateNotSupported = errors.New("the specified filter predicate is not supported for this column")
	ErrNoFindingType               = errors.New("no finding type specified")
	ErrColumnFormatNotSupported    = errors.New("column format does not support sorting")
	ErrInvalidFindingType          = errors.New("invalid finding type specified")
	ErrInvalidAcceptedFilter       = errors.New("invalid finding type specified")
)

type Filterable interface {
	ValidFilters() map[string][]model.FilterOperator
}

func GetFilterableColumns(f Filterable) []string {
	var columns = make([]string, 0)
	for column := range f.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func GetValidFilterPredicatesAsStrings(f Filterable, column string) ([]string, error) {
	if predicates, validColumn := f.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}
