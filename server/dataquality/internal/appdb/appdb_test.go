// Copyright 2026 Specter Ops, Inc.
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

package appdb_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/dataquality/internal/appdb"
	"github.com/specterops/bloodhound/server/dataquality/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	listSQL  = "SELECT id, run_id, schema_extension_id, schema_environment_kind_id, metric_type, metric_name, metric_value, kind_id, created_at, updated_at FROM data_quality_aggregations WHERE deleted_at IS NULL AND schema_environment_kind_id = $1 AND created_at >= $2 AND created_at <= $3"
	countSQL = "SELECT COUNT(*) FROM data_quality_aggregations WHERE deleted_at IS NULL AND schema_environment_kind_id = $1 AND created_at >= $2 AND created_at <= $3"
)

func TestListAggregations(t *testing.T) {
	t.Parallel()
	var (
		ctx         = context.Background()
		start       = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		end         = start.Add(time.Hour)
		driverError = &pgconn.PgError{Code: "08006", Message: "connection failed"}
		columns     = []string{"id", "run_id", "schema_extension_id", "schema_environment_kind_id", "metric_type", "metric_name", "metric_value", "kind_id", "created_at", "updated_at"}
		record      = services.Aggregation{ID: 1, RunID: "run", SchemaExtensionID: 42, SchemaEnvironmentKindID: 10, MetricType: "count", MetricName: "nodes", MetricValue: 1.5, CreatedAt: start, UpdatedAt: end}
		query       = services.AggregationQuery{EnvironmentKindID: 10, Start: start, End: end, Limit: 7, Skip: 3}
	)
	testCases := []struct {
		name      string
		query     services.AggregationQuery
		setupMock func(pgxmock.PgxPoolIface)
		expected  []services.Aggregation
		count     int
		errText   string
	}{
		{name: "default sort and total count", query: query, setupMock: func(pool pgxmock.PgxPoolIface) {
			pool.ExpectQuery(listSQL+" ORDER BY created_at desc, id ASC LIMIT $4 OFFSET $5").WithArgs(int32(10), start, end, 7, 3).WillReturnRows(pool.NewRows(columns).AddRow(int32(1), "run", int32(42), int32(10), "count", "nodes", 1.5, nil, start, end))
			pool.ExpectQuery(countSQL).WithArgs(int32(10), start, end).WillReturnRows(pool.NewRows([]string{"count"}).AddRow(42))
		}, expected: []services.Aggregation{record}, count: 42},
		{name: "empty results stay nil", query: query, setupMock: func(pool pgxmock.PgxPoolIface) {
			pool.ExpectQuery(listSQL+" ORDER BY created_at desc, id ASC LIMIT $4 OFFSET $5").WithArgs(int32(10), start, end, 7, 3).WillReturnRows(pool.NewRows(columns))
			pool.ExpectQuery(countSQL).WithArgs(int32(10), start, end).WillReturnRows(pool.NewRows([]string{"count"}).AddRow(0))
		}},
		{name: "query error is opaque", query: query, setupMock: func(pool pgxmock.PgxPoolIface) {
			pool.ExpectQuery(listSQL+" ORDER BY created_at desc, id ASC LIMIT $4 OFFSET $5").WithArgs(int32(10), start, end, 7, 3).WillReturnError(driverError)
		}, errText: "listing data quality aggregations"},
		{name: "scan error is opaque", query: query, setupMock: func(pool pgxmock.PgxPoolIface) {
			pool.ExpectQuery(listSQL+" ORDER BY created_at desc, id ASC LIMIT $4 OFFSET $5").WithArgs(int32(10), start, end, 7, 3).WillReturnRows(pool.NewRows(columns).AddRow(int32(1), "run", int32(42), int32(10), "count", "nodes", 1.5, nil, start, end).RowError(0, driverError))
		}, errText: "reading data quality aggregations"},
		{name: "count error is opaque", query: query, setupMock: func(pool pgxmock.PgxPoolIface) {
			pool.ExpectQuery(listSQL+" ORDER BY created_at desc, id ASC LIMIT $4 OFFSET $5").WithArgs(int32(10), start, end, 7, 3).WillReturnRows(pool.NewRows(columns))
			pool.ExpectQuery(countSQL).WithArgs(int32(10), start, end).WillReturnError(driverError)
		}, errText: "counting data quality aggregations"},
		{name: "count scan error is opaque", query: query, setupMock: func(pool pgxmock.PgxPoolIface) {
			pool.ExpectQuery(listSQL+" ORDER BY created_at desc, id ASC LIMIT $4 OFFSET $5").WithArgs(int32(10), start, end, 7, 3).WillReturnRows(pool.NewRows(columns))
			pool.ExpectQuery(countSQL).WithArgs(int32(10), start, end).WillReturnRows(pool.NewRows([]string{"count"}).AddRow(0).RowError(0, driverError))
		}, errText: "reading data quality aggregation count"},
		{name: "rejects unsafe sort field", query: services.AggregationQuery{Sort: params.SortItems{{Field: "created_at; DROP TABLE data_quality_aggregations", Direction: params.Ascending}}}, errText: services.ErrInvalidQuery.Error()},
		{name: "rejects unsafe sort direction", query: services.AggregationQuery{Sort: params.SortItems{{Field: "created_at", Direction: "desc; DROP TABLE data_quality_aggregations"}}}, errText: services.ErrInvalidQuery.Error()},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer pool.Close()
			if testCase.setupMock != nil {
				testCase.setupMock(pool)
			}
			records, count, err := appdb.NewStore(pool).ListAggregations(ctx, testCase.query)
			if testCase.errText != "" {
				require.ErrorContains(t, err, testCase.errText)
				assert.NotErrorIs(t, err, driverError)
				var postgresError *pgconn.PgError
				assert.False(t, errors.As(err, &postgresError))
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.expected, records)
			assert.Equal(t, testCase.count, count)
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
