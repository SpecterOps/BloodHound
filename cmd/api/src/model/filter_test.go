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

package model_test

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
)

func TestModel_BuildSQLFilter_Failure(t *testing.T) {
	filter1 := model.QueryParameterFilter{
		Name:         "filtercolumn1",
		Operator:     model.FilterOperator("foo"), // invalid predicate
		Value:        "0",
		IsStringData: false,
	}

	queryParameterFilterMap := model.QueryParameterFilterMap{
		filter1.Name: model.QueryParameterFilters{filter1},
	}

	_, err := queryParameterFilterMap.BuildSQLFilter()
	require.Contains(t, err.Error(), "invalid filter predicate")
}

func TestModel_BuildSQLFilter_Success(t *testing.T) {
	numericMin := model.QueryParameterFilter{
		Name:         "filtercolumn1",
		Operator:     model.GreaterThan,
		Value:        "0",
		IsStringData: false,
	}

	numericMax := model.QueryParameterFilter{
		Name:         "filtercolumn2",
		Operator:     model.LessThan,
		Value:        "10",
		IsStringData: false,
	}

	stringValue := model.QueryParameterFilter{
		Name:         "filtercolumn3",
		Operator:     model.Equals,
		Value:        "stringValue",
		IsStringData: true,
	}

	boolEquals := model.QueryParameterFilter{
		Name:         "filtercolumn4",
		Operator:     model.Equals,
		Value:        "true",
		IsStringData: false,
	}

	boolNotEquals := model.QueryParameterFilter{
		Name:         "filtercolumn5",
		Operator:     model.NotEquals,
		Value:        "false",
		IsStringData: false,
	}

	expectedResults := map[string]model.SQLFilter{
		"numericMin":    {SQLString: fmt.Sprintf("%s > ?", numericMin.Name), Params: []any{numericMin.Value}},
		"numericMax":    {SQLString: fmt.Sprintf("%s < ?", numericMax.Name), Params: []any{numericMax.Value}},
		"stringValue":   {SQLString: fmt.Sprintf("%s = ?", stringValue.Name), Params: []any{stringValue.Value}},
		"boolEquals":    {SQLString: fmt.Sprintf("%s = ?", boolEquals.Name), Params: []any{boolEquals.Value}},
		"boolNotEquals": {SQLString: fmt.Sprintf("%s <> ?", boolNotEquals.Name), Params: []any{boolNotEquals.Value}},
	}

	queryParameterFilterMap := model.QueryParameterFilterMap{
		numericMax.Name:    model.QueryParameterFilters{numericMin, numericMax},
		stringValue.Name:   model.QueryParameterFilters{stringValue},
		boolEquals.Name:    model.QueryParameterFilters{boolEquals},
		boolNotEquals.Name: model.QueryParameterFilters{boolNotEquals},
	}

	result, err := queryParameterFilterMap.BuildSQLFilter()
	require.Nil(t, err)

	for _, val := range expectedResults {
		require.Contains(t, result.SQLString, val.SQLString)
		require.EqualValues(t, result.Params, result.Params)
	}
}
