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

package dbutil_test

import (
	"context"
	"errors"
	"testing"

	"github.com/huandu/go-sqlbuilder"
	"github.com/specterops/bloodhound/packages/go/dbutil"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/stretchr/testify/assert"
)

func TestApplyFilters(t *testing.T) {
	t.Parallel()

	type args struct {
		filters params.Filters
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
			name: "success_-_combines_filters_in_column_order",
			args: args{
				filters: params.Filters{
					"k.id": {
						{Operator: params.GreaterThan, Value: "1"},
						{Operator: params.LessThan, Value: "10", SetOperator: params.FilterAnd},
						{Operator: params.Equals, Value: "42", SetOperator: params.FilterOr},
					},
					"k.name": {
						{Operator: params.Equals, Value: "User"},
					},
				},
			},
			wantSQL:  "SELECT * FROM kind AS k WHERE (((k.id > $1) AND k.id < $2) OR k.id = $3) AND (k.name = $4)",
			wantArgs: []any{"1", "10", "42", "User"},
		},
		{
			name: "error_-_propagates_unsupported_operator",
			args: args{
				filters: params.Filters{
					"k.name": {
						{Operator: "unsupported", Value: "User"},
					},
				},
			},
			wantSQL: "SELECT * FROM kind AS k",
			wantErr: errors.New("invalid operator \"unsupported\""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
			builder.Select("*").From("kind AS k")
			if err := dbutil.ApplyFilters(builder, test.args.filters); test.wantErr != nil {
				assert.EqualError(t, err, test.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			sqlQuery, queryArgs := builder.Build()
			assert.Equal(t, test.wantSQL, sqlQuery)
			assert.Equal(t, test.wantArgs, queryArgs)
		})
	}
}

func TestApplySorts(t *testing.T) {
	t.Parallel()

	type args struct {
		sorts params.SortItems
	}

	type testCase struct {
		name    string
		args    args
		wantSQL string
	}

	tests := []testCase{
		{
			name: "success_-_applies_sorts_in_order",
			args: args{
				sorts: params.SortItems{
					{Field: "k.name", Direction: params.Ascending},
					{Field: "k.id", Direction: params.Descending},
				},
			},
			wantSQL: "SELECT * FROM kind AS k ORDER BY k.name ASC, k.id DESC",
		},
		{
			name: "success_-_applies_single_sort",
			args: args{
				sorts: params.SortItems{
					{Field: "k.name", Direction: params.Ascending},
				},
			},
			wantSQL: "SELECT * FROM kind AS k ORDER BY k.name ASC",
		},
		{
			name:    "success_-_applies_no_sorts",
			args:    args{sorts: params.SortItems{}},
			wantSQL: "SELECT * FROM kind AS k",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
			builder.Select("*").From("kind AS k")
			dbutil.ApplySorts(builder, test.args.sorts)

			sqlQuery, queryArgs := builder.Build()
			assert.Equal(t, test.wantSQL, sqlQuery)
			assert.Empty(t, queryArgs)
		})
	}
}

func TestApplyLimitSkip(t *testing.T) {
	t.Parallel()

	type args struct {
		limit int
		skip  int
	}

	type testCase struct {
		name     string
		args     args
		wantSQL  string
		wantArgs []any
	}

	tests := []testCase{
		{
			name:     "success_-_applies_limit_and_skip",
			args:     args{limit: 10, skip: 5},
			wantSQL:  "SELECT * FROM kind AS k LIMIT $1 OFFSET $2",
			wantArgs: []any{10, 5},
		},
		{
			name:     "success_-_applies_limit_without_skip",
			args:     args{limit: 10},
			wantSQL:  "SELECT * FROM kind AS k LIMIT $1",
			wantArgs: []any{10},
		},
		{
			name:     "success_-_applies_skip_without_limit",
			args:     args{skip: 5},
			wantSQL:  "SELECT * FROM kind AS k OFFSET $1",
			wantArgs: []any{5},
		},
		{
			name:    "success_-_ignores_nonpositive_values",
			args:    args{limit: -1, skip: 0},
			wantSQL: "SELECT * FROM kind AS k",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
			builder.Select("*").From("kind AS k")
			dbutil.ApplyLimitSkip(builder, test.args.limit, test.args.skip)

			sqlQuery, queryArgs := builder.Build()
			assert.Equal(t, test.wantSQL, sqlQuery)
			assert.Equal(t, test.wantArgs, queryArgs)
		})
	}
}

func TestNormalizeFilterSort(t *testing.T) {
	t.Parallel()

	type args struct {
		setup   dbutil.FilterSortSetup
		filters params.Filters
		sorts   params.SortItems
	}

	type testCase struct {
		name        string
		args        args
		wantFilters params.Filters
		wantSorts   params.SortItems
		wantErr     string
	}

	tests := []testCase{
		{
			name: "success_-_aliases_filters_and_sorts_in_order",
			args: args{
				setup: dbutil.FilterSortSetup{
					ColumnAliases: map[string]string{
						"name": "k.name",
						"id":   "k.id",
					},
					DefaultSortField: "k.id",
				},
				filters: params.Filters{
					"name": {
						{Operator: params.Equals, Value: "User", SetOperator: params.FilterAnd},
					},
				},
				sorts: params.SortItems{
					{Field: "name", Direction: params.Descending},
					{Field: "id", Direction: params.Ascending},
				},
			},
			wantFilters: params.Filters{
				"k.name": {
					{Operator: params.Equals, Value: "User", SetOperator: params.FilterAnd},
				},
			},
			wantSorts: params.SortItems{
				{Field: "k.name", Direction: params.Descending},
				{Field: "k.id", Direction: params.Ascending},
			},
		},
		{
			name: "success_-_uses_default_sort_for_empty_input",
			args: args{
				setup: dbutil.FilterSortSetup{
					ColumnAliases:    map[string]string{},
					DefaultSortField: "k.id",
				},
				filters: params.Filters{},
			},
			wantFilters: params.Filters{},
			wantSorts:   params.SortItems{{Field: "k.id", Direction: params.Ascending}},
		},
		{
			name: "success_-_does_not_add_default_sort_when_sorts_are_provided",
			args: args{
				setup: dbutil.FilterSortSetup{
					ColumnAliases:    map[string]string{"name": "k.name"},
					DefaultSortField: "k.id",
				},
				sorts: params.SortItems{{Field: "name", Direction: params.Descending}},
			},
			wantFilters: params.Filters{},
			wantSorts:   params.SortItems{{Field: "k.name", Direction: params.Descending}},
		},
		{
			name: "error_-_rejects_unknown_filter_field",
			args: args{
				setup: dbutil.FilterSortSetup{
					ColumnAliases: map[string]string{},
				},
				filters: params.Filters{"unknown": {{Operator: params.Equals, Value: "User"}}},
			},
			wantErr: `filter references unknown field "unknown"`,
		},
		{
			name: "error_-_rejects_unknown_sort_field",
			args: args{
				setup: dbutil.FilterSortSetup{
					ColumnAliases: map[string]string{},
				},
				sorts: params.SortItems{{Field: "unknown", Direction: params.Ascending}},
			},
			wantErr: `sort references unknown field "unknown"`,
		},
		{
			name: "error_-_rejects_invalid_sort_direction",
			args: args{
				setup: dbutil.FilterSortSetup{
					ColumnAliases: map[string]string{"name": "k.name"},
				},
				sorts: params.SortItems{{Field: "name", Direction: "sideways"}},
			},
			wantErr: `invalid sort direction "sideways"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filters, sorts, err := dbutil.NormalizeFilterSort(test.args.setup, test.args.filters, test.args.sorts)
			if test.wantErr != "" {
				assert.EqualError(t, err, test.wantErr)
				assert.Nil(t, filters)
				assert.Nil(t, sorts)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wantFilters, filters)
			assert.Equal(t, test.wantSorts, sorts)
		})
	}
}

func TestEntityListerApplyFilterSortPagination(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		setup          dbutil.FilterSortSetup
		input          dbutil.FilterSortPaginationInput
		wantSelectSQL  string
		wantSelectArgs []any
		wantCountSQL   string
		wantCountArgs  []any
		wantErr        string
	}

	tests := []testCase{
		{
			name: "success_-_applies_filters_to_both_and_sort_pagination_to_select",
			setup: dbutil.FilterSortSetup{
				ColumnAliases: map[string]string{
					"name": "k.name",
					"id":   "k.id",
				},
				DefaultSortField: "k.id",
			},
			input: dbutil.FilterSortPaginationInput{
				Filters: params.Filters{
					"name": {{Operator: params.Equals, Value: "User"}},
				},
				Sort:  params.SortItems{{Field: "id", Direction: params.Descending}},
				Limit: 10,
				Skip:  5,
			},
			wantSelectSQL:  "SELECT * FROM kind AS k WHERE (k.name = $1) ORDER BY k.id DESC LIMIT $2 OFFSET $3",
			wantSelectArgs: []any{"User", 10, 5},
			wantCountSQL:   "SELECT COUNT(*) FROM kind AS k WHERE (k.name = $1)",
			wantCountArgs:  []any{"User"},
		},
		{
			name: "success_-_uses_default_sort_without_filters_or_pagination",
			setup: dbutil.FilterSortSetup{
				ColumnAliases:    map[string]string{},
				DefaultSortField: "k.id",
			},
			input:         dbutil.FilterSortPaginationInput{},
			wantSelectSQL: "SELECT * FROM kind AS k ORDER BY k.id ASC",
			wantCountSQL:  "SELECT COUNT(*) FROM kind AS k",
		},
		{
			name: "error_-_rejects_unknown_filter_before_mutating_builders",
			setup: dbutil.FilterSortSetup{
				ColumnAliases: map[string]string{},
			},
			input: dbutil.FilterSortPaginationInput{
				Filters: params.Filters{"unknown": {{Operator: params.Equals, Value: "User"}}},
				Limit:   10,
			},
			wantSelectSQL: "SELECT * FROM kind AS k",
			wantCountSQL:  "SELECT COUNT(*) FROM kind AS k",
			wantErr:       `filter references unknown field "unknown"`,
		},
		{
			name: "error_-_rejects_unknown_sort_before_mutating_builders",
			setup: dbutil.FilterSortSetup{
				ColumnAliases: map[string]string{},
			},
			input: dbutil.FilterSortPaginationInput{
				Sort: params.SortItems{{Field: "unknown", Direction: params.Ascending}},
				Skip: 5,
			},
			wantSelectSQL: "SELECT * FROM kind AS k",
			wantCountSQL:  "SELECT COUNT(*) FROM kind AS k",
			wantErr:       `sort references unknown field "unknown"`,
		},
		{
			name: "error_-_propagates_invalid_filter_operator_without_sort_or_pagination",
			setup: dbutil.FilterSortSetup{
				ColumnAliases:    map[string]string{"name": "k.name"},
				DefaultSortField: "k.id",
			},
			input: dbutil.FilterSortPaginationInput{
				Filters: params.Filters{
					"name": {{Operator: "unsupported", Value: "User"}},
				},
				Limit: 10,
				Skip:  5,
			},
			wantSelectSQL: "SELECT * FROM kind AS k",
			wantCountSQL:  "SELECT COUNT(*) FROM kind AS k",
			wantErr:       "invalid operator \"unsupported\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
			selectBuilder.Select("*").From("kind AS k")

			countBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
			countBuilder.Select("COUNT(*)").From("kind AS k")

			err := dbutil.EntityListerApplyFilterSortPagination(
				context.Background(),
				selectBuilder,
				countBuilder,
				test.setup,
				test.input,
			)
			if test.wantErr != "" {
				assert.EqualError(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}

			selectSQL, selectArgs := selectBuilder.Build()
			assert.Equal(t, test.wantSelectSQL, selectSQL)
			assert.Equal(t, test.wantSelectArgs, selectArgs)

			countSQL, countArgs := countBuilder.Build()
			assert.Equal(t, test.wantCountSQL, countSQL)
			assert.Equal(t, test.wantCountArgs, countArgs)
		})
	}
}
