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

//go:build integration

package database_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateDataQualityStats(t *testing.T) {
	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		{
			name: "Success: creates a single data quality stat",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				stat := model.DataQualityStat{
					RunID:                   "run-1",
					SchemaExtensionID:       1,
					SchemaEnvironmentKindID: 1,
					EnvironmentID:           "env-1",
					MetricType:              model.DataQualityMetricTypeNode,
					MetricName:              "users",
					MetricValue:             42,
				}

				created, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, model.DataQualityStats{stat})
				require.NoError(t, err)
				require.Len(t, created, 1)
				assert.NotZero(t, created[0].ID)
				assert.Equal(t, "run-1", created[0].RunID)
				assert.Equal(t, float64(42), created[0].MetricValue)
			},
		},
		{
			name: "Success: creates multiple data quality stats",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				stats := model.DataQualityStats{
					{RunID: "run-2", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "groups", MetricValue: 10},
					{RunID: "run-2", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeRelationship, MetricName: "edges", MetricValue: 20},
				}

				created, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, stats)
				require.NoError(t, err)
				require.Len(t, created, 2)
				assert.NotZero(t, created[0].ID)
				assert.NotZero(t, created[1].ID)
				assert.NotEqual(t, created[0].ID, created[1].ID)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			testCase.assert(t, testSuite)
		})
	}
}

func TestDatabase_GetDataQualityStats(t *testing.T) {
	type args struct {
		filters     model.Filters
		sort        model.Sort
		skip, limit int
	}

	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, testSuite IntegrationTestSuite, args args)
	}{
		{
			name: "Success: returns stats with no filters or sorting",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				stats := model.DataQualityStats{
					{RunID: "run-3", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "users", MetricValue: 5},
					{RunID: "run-3", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "groups", MetricValue: 15},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, stats)
				require.NoError(t, err)

				results, total, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err)
				assert.Equal(t, 2, total)
				assert.Len(t, results, 2)
			},
		},
		{
			name: "Success: returns stats with equals filter",
			args: args{
				filters: model.Filters{
					"metric_name": []model.Filter{
						{Operator: model.Equals, Value: "users", IsStringData: true},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				stats := model.DataQualityStats{
					{RunID: "run-4", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "users", MetricValue: 5},
					{RunID: "run-4", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "groups", MetricValue: 15},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, stats)
				require.NoError(t, err)

				results, total, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err)
				assert.Equal(t, 1, total)
				require.Len(t, results, 1)
				assert.Equal(t, "users", results[0].MetricName)
			},
		},
		{
			name: "Success: filters by created_at date range",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				var (
					yesterday = time.Now().UTC().Add(-24 * time.Hour)
					now       = time.Now().UTC()
				)

				stats := model.DataQualityStats{
					{RunID: "run-date", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "a", MetricValue: 1, Serial: model.Serial{Basic: model.Basic{CreatedAt: yesterday}}},
					{RunID: "run-date", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "b", MetricValue: 2, Serial: model.Serial{Basic: model.Basic{CreatedAt: now}}},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, stats)
				require.NoError(t, err)

				rangeStart := now.Add(-1 * time.Hour).Format(time.RFC3339)
				rangeEnd := now.Add(1 * time.Hour).Format(time.RFC3339)

				dateFilters := model.Filters{
					"created_at": []model.Filter{
						{Operator: model.GreaterThanOrEquals, Value: rangeStart},
						{Operator: model.LessThanOrEquals, Value: rangeEnd},
					},
				}

				results, total, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, dateFilters, args.sort, args.skip, args.limit)
				require.NoError(t, err)
				assert.Equal(t, 1, total)
				require.Len(t, results, 1)
				assert.Equal(t, "b", results[0].MetricName)
			},
		},
		{
			name: "Success: filters by environment_id",
			args: args{
				filters: model.Filters{
					"environment_id": []model.Filter{
						{Operator: model.Equals, Value: "env-target", IsStringData: true},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				stats := model.DataQualityStats{
					{RunID: "run-env", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-target", MetricType: model.DataQualityMetricTypeNode, MetricName: "users", MetricValue: 10},
					{RunID: "run-env", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-other", MetricType: model.DataQualityMetricTypeNode, MetricName: "groups", MetricValue: 20},
					{RunID: "run-env", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-target", MetricType: model.DataQualityMetricTypeRelationship, MetricName: "edges", MetricValue: 30},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, stats)
				require.NoError(t, err)

				results, total, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err)
				assert.Equal(t, 2, total)
				require.Len(t, results, 2)
				for _, result := range results {
					assert.Equal(t, "env-target", result.EnvironmentID)
				}
			},
		},
		{
			name: "Success: returns stats with ascending sort",
			args: args{
				sort: model.Sort{
					{Column: "metric_value", Direction: model.AscendingSortDirection},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				stats := model.DataQualityStats{
					{RunID: "run-5", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "a", MetricValue: 100},
					{RunID: "run-5", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "b", MetricValue: 1},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, stats)
				require.NoError(t, err)

				results, _, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err)
				require.Len(t, results, 2)
				assert.Equal(t, float64(1), results[0].MetricValue)
				assert.Equal(t, float64(100), results[1].MetricValue)
			},
		},
		{
			name: "Success: returns stats with pagination",
			args: args{
				sort:  model.Sort{{Column: "metric_value", Direction: model.AscendingSortDirection}},
				skip:  1,
				limit: 1,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				stats := model.DataQualityStats{
					{RunID: "run-6", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "a", MetricValue: 10},
					{RunID: "run-6", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "b", MetricValue: 20},
					{RunID: "run-6", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "c", MetricValue: 30},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, stats)
				require.NoError(t, err)

				results, total, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err)
				assert.Equal(t, 3, total, "total count should reflect all matching rows")
				require.Len(t, results, 1, "page should contain 1 row")
				assert.Equal(t, float64(20), results[0].MetricValue, "skip=1 should return the second row")
			},
		},
		{
			name: "Success: soft-deleted rows are excluded",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				stats := model.DataQualityStats{
					{RunID: "run-7", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "visible", MetricValue: 1},
					{RunID: "run-7", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "hidden", MetricValue: 2},
				}
				created, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, stats)
				require.NoError(t, err)
				require.Len(t, created, 2)

				// Soft-delete one row via raw SQL
				testSuite.DB.Exec("UPDATE data_quality_stats SET deleted_at = NOW() WHERE id = ?", created[1].ID)

				results, total, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err)
				assert.Equal(t, 1, total)
				require.Len(t, results, 1)
				assert.Equal(t, "visible", results[0].MetricName)
			},
		},
		{
			name: "Success: returns empty results when no stats exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				results, total, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err)
				assert.Equal(t, 0, total)
				assert.Empty(t, results)
			},
		},
		{
			name: "Error: invalid filter operator",
			args: args{
				filters: model.Filters{
					"`": []model.Filter{{}},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, _, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.Error(t, err)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			testCase.assert(t, testSuite, testCase.args)
		})
	}
}

func TestDatabase_CreateDataQualityAggregations(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		{
			name: "Success: create empty aggregations with empty input",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				created, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, model.DataQualityAggregations{})
				require.NoError(t, err)
				assert.Empty(t, created)
			},
		},
		{
			name: "Success: create aggregations and populate IDs",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				// schema_environment_kind_id is a foreign key, register a real kind and use its ID
				envKind := registerAndGetKind(t, testSuite, "RandomKind")

				aggregations := model.DataQualityAggregations{
					{RunID: "run-abc", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "computers", MetricValue: 5},
					{RunID: "run-abc", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "sessions", MetricValue: 15},
				}

				created, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, aggregations)
				require.NoError(t, err)
				require.Len(t, created, 2)

				for _, aggregation := range created {
					assert.NotZero(t, aggregation.ID, "create should populate the generated ID")
				}

				_, total, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, nil, nil, 0, 0)
				require.NoError(t, err)
				assert.Equal(t, 2, total)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			testCase.test(t, testSuite)
		})
	}
}

func TestDatabase_GetDataQualityAggregations(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		{
			name: "Success: get all aggregations, no filters or sorting",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				// schema_environment_kind_id is a foreign key, register a real kind and use its ID
				envKind := registerAndGetKind(t, testSuite, "RandomKind")

				aggregations := model.DataQualityAggregations{
					{RunID: "run-abc", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "computers", MetricValue: 5},
					{RunID: "run-abc", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "sessions", MetricValue: 15},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, aggregations)
				require.NoError(t, err)

				results, total, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, nil, nil, 0, 0)
				require.NoError(t, err)
				assert.Equal(t, 2, total)
				assert.Len(t, results, 2)
			},
		},
		{
			name: "Success: get aggregations with all fields set",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				envKind := registerAndGetKind(t, testSuite, "RandomEnvKind")
				metricKind := registerAndGetKind(t, testSuite, "RandomMetricKind")

				aggregation := model.DataQualityAggregation{
					RunID:                   "run-111",
					SchemaExtensionID:       42,
					SchemaEnvironmentKindID: envKind.ID,
					MetricType:              model.DataQualityMetricTypeNode,
					MetricName:              "owned_principals",
					MetricValue:             1024.5,
					KindID:                  null.Int32From(metricKind.ID),
				}

				created, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, model.DataQualityAggregations{aggregation})
				require.NoError(t, err)
				require.Len(t, created, 1)
				require.NotZero(t, created[0].ID, "create should populate the generated ID")

				results, total, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, nil, nil, 0, 0)
				require.NoError(t, err)
				assert.Equal(t, 1, total)
				require.Len(t, results, 1)

				// assert each field is stored and returned unchanged
				got := results[0]
				assert.Equal(t, "run-111", got.RunID)
				assert.Equal(t, int32(42), got.SchemaExtensionID)
				assert.Equal(t, envKind.ID, got.SchemaEnvironmentKindID)
				assert.Equal(t, model.DataQualityMetricTypeNode, got.MetricType)
				assert.Equal(t, "owned_principals", got.MetricName)
				assert.Equal(t, 1024.5, got.MetricValue)
				assert.True(t, got.KindID.Valid, "kind_id should not be null")
				assert.Equal(t, metricKind.ID, got.KindID.Int32)
			},
		},
		{
			name: "Success: get aggregations with multiple filters",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				envKind := registerAndGetKind(t, testSuite, "RandomEnvKind")

				aggregations := model.DataQualityAggregations{
					{RunID: "run-123", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "containers", MetricValue: 5},
					{RunID: "run-123", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeRelationship, MetricName: "memberof", MetricValue: 20},
					{RunID: "run-999", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeRelationship, MetricName: "memberof", MetricValue: 30},
					// approximately equal case
					{RunID: "run-123", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeRelationship, MetricName: "members", MetricValue: 25},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, aggregations)
				require.NoError(t, err)

				filters := model.Filters{
					"run_id": []model.Filter{
						{Operator: model.Equals, Value: "run-123", IsStringData: true},
					},
					// two filters on the same column
					"metric_name": []model.Filter{
						{Operator: model.Equals, Value: "memberof", IsStringData: true},
						{Operator: model.ApproximatelyEquals, Value: "member", IsStringData: true},
					},
				}

				results, total, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, filters, nil, 0, 0)
				require.NoError(t, err)
				assert.Equal(t, 1, total)
				require.Len(t, results, 1)
				assert.Equal(t, "run-123", results[0].RunID)
				assert.Equal(t, "memberof", results[0].MetricName)
				assert.Equal(t, model.DataQualityMetricTypeRelationship, results[0].MetricType)
			},
		},
		{
			name: "Success: get aggregations sorted by metric_value ascending and with completeness metrics",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				envKind := registerAndGetKind(t, testSuite, "RandomEnvKind")

				aggregations := model.DataQualityAggregations{
					{RunID: "run-jkl", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeRelationship, MetricName: "session_completeness", MetricValue: 87.5},
					{RunID: "run-jkl", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeRelationship, MetricName: "local_group_completeness", MetricValue: 9.0},
				}
				_, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, aggregations)
				require.NoError(t, err)

				sort := model.Sort{
					{Column: "metric_value", Direction: model.AscendingSortDirection},
				}

				results, _, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, nil, sort, 0, 0)
				require.NoError(t, err)
				require.Len(t, results, 2)
				assert.Equal(t, 9.0, results[0].MetricValue)
				assert.Equal(t, 87.5, results[1].MetricValue)
				assert.Equal(t, "local_group_completeness", results[0].MetricName)
			},
		},
		{
			name: "Success: get aggregations with soft-deleted rows excluded from results and count",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				envKind := registerAndGetKind(t, testSuite, "RandomEnvKind")

				aggregations := model.DataQualityAggregations{
					{RunID: "run-xyz", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "computers", MetricValue: 1},
					{RunID: "run-xyz", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "users", MetricValue: 2},
					{RunID: "run-xyz", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "groups", MetricValue: 3},
					{RunID: "run-xyz", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "sessions", MetricValue: 4},
				}
				created, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, aggregations)
				require.NoError(t, err)
				require.Len(t, created, 4)

				// soft-delete one row
				deletedID := created[0].ID
				deleteResult := testSuite.DB.WithContext(testSuite.Context).Exec("UPDATE data_quality_aggregations SET deleted_at = NOW() WHERE id = ?", deletedID)
				require.NoError(t, deleteResult.Error)
				require.EqualValues(t, 1, deleteResult.RowsAffected)

				// run a GET with a limit
				results, total, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, nil, nil, 0, 2)
				require.NoError(t, err)

				// total count is 4 - 1 (soft-deleted row), count query respects the deleted_at WHERE clause
				assert.Equal(t, 3, total)
				assert.Len(t, results, 2)

				for _, aggregation := range results {
					assert.NotEqual(t, deletedID, aggregation.ID, "soft-deleted row must not appear in the results")
				}
			},
		},
		{
			name: "Success: get aggregations sorted by created_at descending when no sort is provided",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				envKind := registerAndGetKind(t, testSuite, "RandomEnvKind")

				aggregations := model.DataQualityAggregations{
					{RunID: "run-ts", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "computers", MetricValue: 1},
					{RunID: "run-ts", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "users", MetricValue: 2},
					{RunID: "run-ts", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "groups", MetricValue: 3},
				}
				created, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, aggregations)
				require.NoError(t, err)
				require.Len(t, created, 3)

				// modify created_at for each row so we have distinct created_at dates
				rowHourAgo := created[0]
				rowNow := created[1]
				rowTwoHoursAgo := created[2]
				// scramble the created_at order to ensure GetDataQualityAggregations sorts it correctly
				require.NoError(t, testSuite.DB.WithContext(testSuite.Context).Exec("UPDATE data_quality_aggregations SET created_at = NOW() - INTERVAL '1 hour' WHERE id = ?", rowHourAgo.ID).Error)
				require.NoError(t, testSuite.DB.WithContext(testSuite.Context).Exec("UPDATE data_quality_aggregations SET created_at = NOW() WHERE id = ?", rowNow.ID).Error)
				require.NoError(t, testSuite.DB.WithContext(testSuite.Context).Exec("UPDATE data_quality_aggregations SET created_at = NOW() - INTERVAL '2 hours' WHERE id = ?", rowTwoHoursAgo.ID).Error)

				results, _, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, nil, nil, 0, 0)
				require.NoError(t, err)
				require.Len(t, results, 3)

				gotOrder := []int32{results[0].ID, results[1].ID, results[2].ID}
				wantOrder := []int32{rowNow.ID, rowHourAgo.ID, rowTwoHoursAgo.ID}
				assert.Equal(t, wantOrder, gotOrder)
			},
		},
		{
			name: "Success: get aggregations sorted by id ascending when created_at values are equal",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				envKind := registerAndGetKind(t, testSuite, "RandomEnvKind")

				aggregations := model.DataQualityAggregations{
					{RunID: "run-tie", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "computers", MetricValue: 1},
					{RunID: "run-tie", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "users", MetricValue: 2},
					{RunID: "run-tie", SchemaExtensionID: 42, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "groups", MetricValue: 3},
				}
				created, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, aggregations)
				require.NoError(t, err)
				require.Len(t, created, 3)

				// name each row by its insert position; serial ids ascend with insert order
				rowFirst := created[0]
				rowSecond := created[1]
				rowThird := created[2]
				require.True(t, rowFirst.ID < rowSecond.ID && rowSecond.ID < rowThird.ID, "serial ids should ascend in insert order")
				// tie every created_at so only the id ascending tiebreaker can order the rows
				require.NoError(t, testSuite.DB.WithContext(testSuite.Context).Exec("UPDATE data_quality_aggregations SET created_at = NOW()").Error)

				results, _, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, nil, nil, 0, 0)
				require.NoError(t, err)
				require.Len(t, results, 3)

				gotOrder := []int32{results[0].ID, results[1].ID, results[2].ID}
				wantOrder := []int32{rowFirst.ID, rowSecond.ID, rowThird.ID}
				assert.Equal(t, wantOrder, gotOrder)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			testCase.test(t, testSuite)
		})
	}
}

func TestDatabase_DeleteAllDataQuality(t *testing.T) {
	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		{
			name: "Success: deletes data quality stats",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				_, err := testSuite.BHDatabase.CreateDataQualityStats(testSuite.Context, model.DataQualityStats{
					{RunID: "run-del", SchemaExtensionID: 1, SchemaEnvironmentKindID: 1, EnvironmentID: "env-1", MetricType: model.DataQualityMetricTypeNode, MetricName: "users", MetricValue: 10},
				})
				require.NoError(t, err)

				err = testSuite.BHDatabase.DeleteAllDataQuality(testSuite.Context)
				require.NoError(t, err)

				results, total, err := testSuite.BHDatabase.GetDataQualityStats(testSuite.Context, nil, nil, 0, 0)
				require.NoError(t, err)
				assert.Equal(t, 0, total)
				assert.Empty(t, results)
			},
		},
		{
			name: "Success: deletes data quality aggregations",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				envKind := registerAndGetKind(t, testSuite, "RandomKind")

				_, err := testSuite.BHDatabase.CreateDataQualityAggregations(testSuite.Context, model.DataQualityAggregations{
					{RunID: "run-abc", SchemaExtensionID: 15, SchemaEnvironmentKindID: envKind.ID, MetricType: model.DataQualityMetricTypeNode, MetricName: "computers", MetricValue: 5},
				})
				require.NoError(t, err)

				err = testSuite.BHDatabase.DeleteAllDataQuality(testSuite.Context)
				require.NoError(t, err)

				results, total, err := testSuite.BHDatabase.GetDataQualityAggregations(testSuite.Context, nil, nil, 0, 0)
				require.NoError(t, err)
				assert.Equal(t, 0, total)
				assert.Empty(t, results)
			},
		},
		{
			name: "Success: no error when tables are already empty",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				err := testSuite.BHDatabase.DeleteAllDataQuality(testSuite.Context)
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			testCase.assert(t, testSuite)
		})
	}
}
