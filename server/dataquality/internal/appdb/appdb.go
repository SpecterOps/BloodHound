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

package appdb

import (
	"context"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/dataquality/internal/services"
)

const tableAggregations = "data_quality_aggregations"

var aggregationColumns = []string{"id", "run_id", "schema_extension_id", "schema_environment_kind_id", "metric_type", "metric_name", "metric_value", "kind_id", "created_at", "updated_at"}

type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Store struct{ database pgxQuerier }

func NewStore(database pgxQuerier) *Store { return &Store{database: database} }

type aggregationRow struct {
	ID                      int32     `db:"id"`
	RunID                   string    `db:"run_id"`
	SchemaExtensionID       int32     `db:"schema_extension_id"`
	SchemaEnvironmentKindID int32     `db:"schema_environment_kind_id"`
	MetricType              string    `db:"metric_type"`
	MetricName              string    `db:"metric_name"`
	MetricValue             float64   `db:"metric_value"`
	KindID                  *int32    `db:"kind_id"`
	CreatedAt               time.Time `db:"created_at"`
	UpdatedAt               time.Time `db:"updated_at"`
}

func toAggregation(row aggregationRow) services.Aggregation {
	return services.Aggregation{
		ID: row.ID, RunID: row.RunID, SchemaExtensionID: row.SchemaExtensionID,
		SchemaEnvironmentKindID: row.SchemaEnvironmentKindID, MetricType: row.MetricType,
		MetricName: row.MetricName, MetricValue: row.MetricValue, KindID: row.KindID,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func (s *Store) ListAggregations(ctx context.Context, query services.AggregationQuery) ([]services.Aggregation, int, error) {
	var (
		builder      = sqlbuilder.PostgreSQL.NewSelectBuilder()
		sortItems    = query.Sort
		order        = make([]string, 0, len(sortItems)+1)
		aggregations []services.Aggregation
	)
	if query.Skip < 0 || query.Limit < 0 {
		return nil, 0, services.ErrInvalidQuery
	}
	if len(sortItems) == 0 {
		sortItems = params.SortItems{{Field: "created_at", Direction: params.Descending}}
	}
	for _, item := range sortItems {
		if (item.Field != "created_at" && item.Field != "updated_at") || (item.Direction != params.Ascending && item.Direction != params.Descending) {
			return nil, 0, services.ErrInvalidQuery
		}
		order = append(order, item.Field+" "+string(item.Direction))
	}
	order = append(order, "id ASC")
	builder.Select(aggregationColumns...).From(tableAggregations)
	builder.Where(builder.IsNull("deleted_at"), builder.Equal("schema_environment_kind_id", query.EnvironmentKindID), builder.GreaterEqualThan("created_at", query.Start), builder.LessEqualThan("created_at", query.End))
	if query.ExtensionID != nil {
		builder.Where(builder.Equal("schema_extension_id", *query.ExtensionID))
	}
	countBuilder := builder.Clone().Select("COUNT(*)")
	builder.OrderBy(order...)
	if query.Skip > 0 {
		builder.Offset(query.Skip)
	}
	// Legacy limit=0 requests every matching row.
	if query.Limit > 0 {
		builder.Limit(query.Limit)
	}
	sqlQuery, arguments := builder.Build()
	rows, err := s.database.Query(ctx, sqlQuery, arguments...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing data quality aggregations: %s", err)
	}
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[aggregationRow])
	if err != nil {
		return nil, 0, fmt.Errorf("reading data quality aggregations: %s", err)
	}
	// Keep nil for an empty result to preserve the legacy JSON null.
	for _, record := range records {
		aggregations = append(aggregations, toAggregation(record))
	}
	if query.Skip == 0 && query.Limit == 0 {
		return aggregations, len(aggregations), nil
	}
	sqlQuery, arguments = countBuilder.Build()
	rows, err = s.database.Query(ctx, sqlQuery, arguments...)
	if err != nil {
		return nil, 0, fmt.Errorf("counting data quality aggregations: %s", err)
	}
	count, err := pgx.CollectOneRow(rows, pgx.RowTo[int])
	if err != nil {
		return nil, 0, fmt.Errorf("reading data quality aggregation count: %s", err)
	}
	return aggregations, count, nil
}
