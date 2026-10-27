package appdb

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
			wantErr: errors.New("invalid operator specified"),
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
			wantErr: errors.New("invalid operator specified"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
			builder.Select("*").From("kind AS k")
			if err := applyFilters(builder, test.args.filters); test.wantErr != nil {
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
