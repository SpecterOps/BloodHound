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
	"sort"
	"strconv"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/dataquality/internal/services"
)

const tableAggregations = "data_quality_aggregations"

const (
	aggregationColumnID                      = "id"
	aggregationColumnRunID                   = "run_id"
	aggregationColumnSchemaExtensionID       = "schema_extension_id"
	aggregationColumnSchemaEnvironmentKindID = "schema_environment_kind_id"
	aggregationColumnMetricType              = "metric_type"
	aggregationColumnMetricName              = "metric_name"
	aggregationColumnMetricValue             = "metric_value"
	aggregationColumnKindID                  = "kind_id"
	aggregationColumnCreatedAt               = "created_at"
	aggregationColumnUpdatedAt               = "updated_at"
	aggregationColumnDeletedAt               = "deleted_at"
)

var aggregationColumns = []string{
	aggregationColumnID,
	aggregationColumnRunID,
	aggregationColumnSchemaExtensionID,
	aggregationColumnSchemaEnvironmentKindID,
	aggregationColumnMetricType,
	aggregationColumnMetricName,
	aggregationColumnMetricValue,
	aggregationColumnKindID,
	aggregationColumnCreatedAt,
	aggregationColumnUpdatedAt,
}

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

func (s *Store) ListAggregations(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems, skip, limit int) ([]services.Aggregation, int, error) {
	var (
		builder      = sqlbuilder.PostgreSQL.NewSelectBuilder()
		order        = make([]string, 0, len(sortItems)+1)
		aggregations []services.Aggregation
	)

	builder.Select(aggregationColumns...).From(tableAggregations)
	builder.Where(builder.IsNull(aggregationColumnDeletedAt))
	if err := applyAggregationFilters(builder, queryFilters); err != nil {
		return nil, 0, err
	}

	sortItems = applyDeterministicDataQualitySort(sortItems)

	for _, item := range sortItems {
		if (item.Field != aggregationColumnCreatedAt && item.Field != aggregationColumnUpdatedAt && item.Field != aggregationColumnID) ||
			(item.Direction != params.Ascending && item.Direction != params.Descending) {
			return nil, 0, services.ErrInvalidQuery
		}
		order = append(order, item.Field+" "+string(item.Direction))
	}

	countBuilder := builder.Clone().Select("COUNT(*)")
	builder.OrderBy(order...)
	if skip > 0 {
		builder.Offset(skip)
	}
	// Legacy limit=0 requests every matching row.
	if limit > 0 {
		builder.Limit(limit)
	}
	sqlQuery, arguments := builder.Build()
	rows, err := s.db.Query(ctx, sqlQuery, arguments...)
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
	if skip == 0 && limit == 0 {
		return aggregations, len(aggregations), nil
	}
	sqlQuery, arguments = countBuilder.Build()
	rows, err = s.db.Query(ctx, sqlQuery, arguments...)
	if err != nil {
		return nil, 0, fmt.Errorf("counting data quality aggregations: %s", err)
	}
	count, err := pgx.CollectOneRow(rows, pgx.RowTo[int])
	if err != nil {
		return nil, 0, fmt.Errorf("reading data quality aggregation count: %s", err)
	}
	return aggregations, count, nil
}

func applyAggregationFilters(builder *sqlbuilder.SelectBuilder, queryFilters params.Filters) error {
	var fields = make([]string, 0, len(queryFilters))

	for field := range queryFilters {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	for _, field := range fields {
		for _, filter := range queryFilters[field] {
			var (
				value  any
				err    error
				clause string
			)

			if filter.Field != field || (filter.SetOperator != "" && filter.SetOperator != params.FilterAnd) {
				return services.ErrInvalidQuery
			}

			switch field {
			case aggregationColumnSchemaEnvironmentKindID, aggregationColumnSchemaExtensionID:
				var id int64
				id, err = strconv.ParseInt(filter.Value, 10, 32)
				value = int32(id)
				if filter.Operator != params.Equals {
					return services.ErrInvalidQuery
				}
			case aggregationColumnCreatedAt:
				value, err = time.Parse(time.RFC3339Nano, filter.Value)
				if filter.Operator != params.GreaterThanOrEquals && filter.Operator != params.LessThanOrEquals {
					return services.ErrInvalidQuery
				}
			default:
				return services.ErrInvalidQuery
			}
			if err != nil {
				return fmt.Errorf("%w: %s: %s", services.ErrInvalidQuery, field, err)
			}

			switch filter.Operator {
			case params.Equals:
				clause = builder.Equal(field, value)
			case params.GreaterThanOrEquals:
				clause = builder.GreaterEqualThan(field, value)
			case params.LessThanOrEquals:
				clause = builder.LessEqualThan(field, value)
			default:
				return services.ErrInvalidQuery
			}
			builder.Where(clause)
		}
	}
	return nil
}

// applyDeterministicDataQualitySort defaults to created_at DESC when no sort is provided, and appends id ASC as a tiebreaker
// so pagination stays deterministic when created_at or updated_at values are not unique
func applyDeterministicDataQualitySort(sortItems params.SortItems) params.SortItems {

	if len(sortItems) == 0 {
		sortItems = params.SortItems{{Field: aggregationColumnCreatedAt, Direction: params.Descending}}
	}
	// only apply the ID tiebreaker if sort doesn't already have it
	sortContainsIDColumn := false
	for _, item := range sortItems {
		if item.Field == aggregationColumnID {
			sortContainsIDColumn = true
			break
		}
	}
	if !sortContainsIDColumn {
		sortItems = append(sortItems, params.SortItem{Field: aggregationColumnID, Direction: params.Ascending})
	}
	return sortItems
}
