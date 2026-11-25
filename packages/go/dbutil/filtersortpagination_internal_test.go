// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package dbutil

import (
	"errors"
	"testing"

	"github.com/huandu/go-sqlbuilder"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/stretchr/testify/assert"
)

func TestBuildFilterComparison(t *testing.T) {
	t.Parallel()

	type args struct {
		column string
		filter params.Filter
	}

	type testCase struct {
		name     string
		args     args
		wantSQL  string
		wantArgs []any
		wantErr  error
	}

	tests := []testCase{
		{
			name: "success_-_equals",
			args: args{
				column: "k.name",
				filter: params.Filter{Operator: params.Equals, Value: "User"},
			},
			wantSQL:  "SELECT * FROM kind AS k WHERE k.name = $1",
			wantArgs: []any{"User"},
		},
		{
			name: "success_-_equals_null",
			args: args{
				column: "k.name",
				filter: params.Filter{Operator: params.Equals, Value: "null"},
			},
			wantSQL: "SELECT * FROM kind AS k WHERE k.name IS NULL",
		},
		{
			name: "success_-_not_equals_null",
			args: args{
				column: "k.name",
				filter: params.Filter{Operator: params.NotEquals, Value: "null"},
			},
			wantSQL: "SELECT * FROM kind AS k WHERE k.name IS NOT NULL",
		},
		{
			name:    "success_-_is_null",
			args:    args{column: "k.name", filter: params.Filter{Operator: params.IsNull}},
			wantSQL: "SELECT * FROM kind AS k WHERE k.name IS NULL",
		},
		{
			name:    "success_-_is_not_null",
			args:    args{column: "k.name", filter: params.Filter{Operator: params.IsNotNull}},
			wantSQL: "SELECT * FROM kind AS k WHERE k.name IS NOT NULL",
		},
		{
			name: "success_-_approximately_equals",
			args: args{
				column: "k.name",
				filter: params.Filter{Operator: params.ApproximatelyEquals, Value: "User"},
			},
			wantSQL:  "SELECT * FROM kind AS k WHERE k.name ILIKE $1",
			wantArgs: []any{"%User%"},
		},
		{
			name: "error_-_unsupported_operator",
			args: args{
				column: "k.name",
				filter: params.Filter{Operator: "unsupported", Value: "User"},
			},
			wantSQL: "SELECT * FROM kind AS k",
			wantErr: errors.New("invalid operator \"unsupported\""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
			comparison, err := buildFilterComparison(builder, test.args.column, test.args.filter)
			if test.wantErr != nil {
				assert.EqualError(t, err, test.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			builder.Select("*").From("kind AS k").Where(comparison)

			sqlQuery, queryArgs := builder.Build()
			assert.Equal(t, test.wantSQL, sqlQuery)
			assert.Equal(t, test.wantArgs, queryArgs)
		})
	}
}
