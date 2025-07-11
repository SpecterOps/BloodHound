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
)

func TestBuildSQLFilter(t *testing.T) {
	testCases := []struct {
		name    string
		input   model.Filters
		output  model.SQLFilter
		wantErr bool
	}{
		{
			name: "greater than",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator: "gt",
					Value:    "12",
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
					Operator: "gte",
					Value:    "12",
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
					Operator: "lt",
					Value:    "12",
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
					Operator: "lte",
					Value:    "12",
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
					Operator: "eq",
					Value:    "12",
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
					Operator: "eq",
					Value:    "12.215",
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
					Operator: "eq",
					Value:    "1notanumber2",
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = '1notanumber2'",
			},
		},
		{
			name: "equals boolean",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator: "eq",
					Value:    "false",
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo = false",
			},
		},
		{
			name: "equals null",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator: "eq",
					Value:    "null",
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
					Operator: "neq",
					Value:    "12",
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo != 12",
			},
		},
		{
			name: "not equals null",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator: "neq",
					Value:    "null",
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo is not null",
			},
		},
		{
			name: "aprox equals",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator: "~eq",
					Value:    "12",
				}},
			},
			output: model.SQLFilter{
				SQLString: "foo like '%12%'",
			},
		},
		{
			name: "broken operator",
			input: model.Filters{
				"foo": []model.Filter{{
					Operator: "NOT OPERATOR",
					Value:    "12",
				}},
			},
			wantErr: true,
		},
	}

	// Run each test case and compare the output with the expected result
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOutput, err := model.BuildSQLFilter(tc.input)

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
