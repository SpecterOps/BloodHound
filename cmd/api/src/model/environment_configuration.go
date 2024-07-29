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

package model

import (
	"encoding/json"
	"fmt"
)

type EnvironmentConfiguration struct {
	Name string          `json:"name" gorm:"primaryKey"`
	Data json.RawMessage `json:"data" gorm:"type:jsonb"`
}

type EnvironmentConfigurations []EnvironmentConfiguration

func (e EnvironmentConfigurations) IsSortable(column string) bool {
	switch column {
	case "name":
		return true
	default:
		return false
	}
}

func (e EnvironmentConfigurations) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"name": {Equals, NotEquals},
	}
}

func (e EnvironmentConfigurations) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range e.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (e EnvironmentConfigurations) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := e.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf(ErrorResponseDetailsColumnNotFilterable)
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func (e EnvironmentConfigurations) IsString(column string) bool {
	switch column {
	case "name":
		return true
	default:
		return false
	}
}
