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
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/cypher/models"
)

func TestBuildSQLFilter(t *testing.T) {
	testCases := []struct {
		name    string
		input   model.Filters
		alias   models.Optional[string]
		output  model.SQLFilter
		wantErr bool
	}{
		{
			name: "greater than",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "gt",
					Value:        "12",
					IsStringData: false,
				}},
			},
			alias: models.OptionalValue("f"),
			output: model.SQLFilter{
				SQLString: "f.foo > 12",
			},
			wantErr: false,
		},
		{
			name: "greater than or equals",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "gte",
					Value:        "12",
					IsStringData: false,
				}},
			},
			alias: models.OptionalValue("f"),
			output: model.SQLFilter{
				SQLString: "f.foo >= 12",
			},
		},
		{
			name: "greater than",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "gt",
					Value:        "12",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo > 12",
			},
		},
		{
			name: "greater than or equals",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "gte",
					Value:        "12",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo >= 12",
			},
		},
		{
			name: "less than",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "lt",
					Value:        "12",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo < 12",
			},
		},
		{
			name: "less than or equals",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "lte",
					Value:        "12",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo <= 12",
			},
		},
		{
			name: "equals int",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "12",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = 12",
			},
		},
		{
			name: "equals float",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "12.215",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = 12.215",
			},
		},
		{
			name: "equals string",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "1notanumber2",
					IsStringData: true,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = '1notanumber2'",
			},
		},
		{
			name: "equals numeric string",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "1",
					IsStringData: true,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = '1'",
			},
		},
		{
			name: "equals float-like string",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "1.1",
					IsStringData: true,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = '1.1'",
			},
		},
		{
			name: "equals boolean-like string",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "t",
					IsStringData: true,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = 't'",
			},
		},
		{
			name: "not equals boolean-like string",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "neq",
					Value:        "f",
					IsStringData: true,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo != 'f'",
			},
		},
		{
			name: "equals boolean",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "false",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = false",
			},
		},
		{
			name: "equals one-letter boolean",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "t",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = true",
			},
		},
		{
			name: "equals null",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "eq",
					Value:        "null",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo is null",
			},
		},
		{
			name: "not equals int",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "neq",
					Value:        "12",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo != 12",
			},
		},
		{
			name: "not equals one-letter boolean",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "neq",
					Value:        "f",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo != false",
			},
		},
		{
			name: "not equals null",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "neq",
					Value:        "null",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo is not null",
			},
		},
		{
			name: "approx equals",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "~eq",
					Value:        "12",
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo ilike '%12%'",
			},
		},
		{
			name: "or equals",
			input: model.Filters{
				"z": []model.Filter{{
					Operator:     "eq",
					Value:        "6",
					SetOperator:  model.FilterOr,
					IsStringData: false,
				}, {
					Operator:     "eq",
					Value:        "7",
					SetOperator:  model.FilterOr,
					IsStringData: false,
				}},
			},
			output: model.SQLFilter{
				SQLString: "(z = 6 or z = 7)",
			},
		},
		{
			name: "combined AND equals numeric strings",
			input: model.Filters{
				"some_column": []model.Filter{{
					Operator:     "eq",
					Value:        "6",
					SetOperator:  model.FilterAnd,
					IsStringData: true,
				}, {
					Operator:     "eq",
					Value:        "7",
					SetOperator:  model.FilterAnd,
					IsStringData: true,
				}},
			},
			output: model.SQLFilter{
				SQLString: "some_column = '6' and some_column = '7'",
			},
		},
		{
			name: "combined OR equals boolean-like strings",
			input: model.Filters{
				"some_column": []model.Filter{{
					Operator:     "eq",
					Value:        "t",
					SetOperator:  model.FilterOr,
					IsStringData: true,
				}, {
					Operator:     "eq",
					Value:        "false",
					SetOperator:  model.FilterOr,
					IsStringData: true,
				}},
			},
			output: model.SQLFilter{
				SQLString: "(some_column = 't' or some_column = 'false')",
			},
		},
		{
			name: "broken operator",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator:     "NOT OPERATOR",
					Value:        "12",
					IsStringData: false,
				}},
			},
			wantErr: true,
		},
	}

	// Run each test case and compare the output with the expected result
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOutput, err := model.BuildSQLFilter(tc.input, tc.alias)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if actualOutput.SQLString != tc.output.SQLString {
				t.Errorf("incorrect SQL string: got %q, want %q", actualOutput.SQLString, tc.output.SQLString)
			}
		})
	}
}
