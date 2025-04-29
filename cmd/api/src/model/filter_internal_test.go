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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryParameterFilterParser_ParseQueryParameterFilter(t *testing.T) {
	t.Parallel()

	parser := NewQueryParameterFilterParser()

	type args struct {
		name  string
		value string
	}

	type expected struct {
		name     string
		value    string
		operator FilterOperator
	}

	tests := []struct {
		name     string
		args     args
		expected expected
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "success - parser should parse a parameter filter",
			args: args{
				name:  "parameter",
				value: "eq:auth.value",
			},
			expected: expected{
				name:     "parameter",
				value:    "auth.value",
				operator: Equals,
			},
			wantErr: assert.NoError,
		},
		{
			name: "success - parser should parse a parameter with ~ filter",
			args: args{
				name:  "parameter",
				value: "~eq:auth.value",
			},
			expected: expected{
				name:     "parameter",
				value:    "auth.value",
				operator: ApproximatelyEquals,
			},
			wantErr: assert.NoError,
		},
		{
			name: "success - parser should parse a parameter filter with spacing",
			args: args{
				name:  "parameter",
				value: "eq:hello world",
			},
			expected: expected{
				name:     "parameter",
				value:    "hello world",
				operator: Equals,
			},
			wantErr: assert.NoError,
		},
		{
			name: "fail - error when parsing an invalid parameter",
			args: args{
				name:  "parameter",
				value: "eq : hello world",
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if res, err := parser.ParseQueryParameterFilter(tt.args.name, tt.args.value); !tt.wantErr(t, err) {
				t.Errorf("ParseQueryParameterFilter() returned an unexpected error = %v", err)
			} else {
				assert.Equal(t, tt.expected.name, res.Name)
				assert.Equal(t, tt.expected.value, res.Value)
				assert.Equal(t, tt.expected.operator, res.Operator)
			}
		})
	}
}
