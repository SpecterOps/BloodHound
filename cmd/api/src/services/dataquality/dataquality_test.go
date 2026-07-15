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

package dataquality_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/dataquality"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSaveDataQuality(t *testing.T) {
	var (
		ctx                    = context.Background()
		environmentKind        = graph.StringKind("dogpark_LargeDogArea")
		sourceKind             = graph.StringKind("dogpark_Entity")
		dogKind                = graph.StringKind("dogpark_Dog")
		patronKind             = graph.StringKind("dogpark_Patron")
		relationshipKind       = graph.StringKind("dogpark_WalksWith")
		schemaExtension        = model.GraphSchemaExtension{Serial: model.Serial{ID: 44}}
		schemaEnvironment      = model.SchemaEnvironment{Serial: model.Serial{ID: 88}, SchemaExtensionId: 44, EnvironmentKindId: 22, EnvironmentKindName: environmentKind.String(), SourceKindId: 11}
		expectedKindIDs        = map[string]null.Int32{"dogpark_Dog": null.Int32From(101), "dogpark_Patron": null.Int32From(102), "dogpark_WalksWith": null.Int32From(103)}
		builtInExtensionFilter = model.Filters{"is_builtin": []model.Filter{{Operator: model.Equals, Value: "false"}}}
		expectedError          = errors.New("expected error")
	)

	tests := []struct {
		name          string
		setupMocks    func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase)
		expectedError string
	}{
		{
			name: "saves opengraph node and relationship counts",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				expectOpenGraphStatsRead(t, mockGraph, environmentKind, sourceKind, dogKind, patronKind, relationshipKind)

				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{schemaExtension}, 1, nil)
				mockDB.EXPECT().
					GetEnvironmentsByExtensionId(gomock.Any(), schemaExtension.ID).
					Return([]model.SchemaEnvironment{schemaEnvironment}, nil)
				mockDB.EXPECT().
					GetKindsByIDs(gomock.Any(), schemaEnvironment.SourceKindId).
					Return([]model.Kind{{ID: schemaEnvironment.SourceKindId, Name: sourceKind.String()}}, nil)
				expectOpenGraphRelationshipKinds(mockDB, schemaExtension, relationshipKind)
				mockDB.EXPECT().
					GetKindsByNames(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, names ...string) ([]model.Kind, error) {
						if !gomock.InAnyOrder([]string{dogKind.String(), patronKind.String(), relationshipKind.String()}).Matches(names) {
							t.Fatalf("expected kind names to match, got %#v", names)
						}

						return []model.Kind{
							{ID: 101, Name: dogKind.String()},
							{ID: 102, Name: patronKind.String()},
							{ID: 103, Name: relationshipKind.String()},
						}, nil
					})
				mockDB.EXPECT().
					CreateDataQualityStats(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, stats model.DataQualityStats) (model.DataQualityStats, error) {
						require.Len(t, stats, 5)
						require.NotEmpty(t, stats[0].RunID)

						expectedStats := model.DataQualityStats{
							{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, EnvironmentID: "env-a", MetricType: model.DataQualityMetricTypeNode, MetricName: "dogpark_Dog", MetricValue: 1, KindID: expectedKindIDs["dogpark_Dog"]},
							{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, EnvironmentID: "env-a", MetricType: model.DataQualityMetricTypeNode, MetricName: "dogpark_Patron", MetricValue: 2, KindID: expectedKindIDs["dogpark_Patron"]},
							{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, EnvironmentID: "env-b", MetricType: model.DataQualityMetricTypeNode, MetricName: "dogpark_Patron", MetricValue: 3, KindID: expectedKindIDs["dogpark_Patron"]},
							{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, EnvironmentID: "env-a", MetricType: model.DataQualityMetricTypeRelationship, MetricName: "dogpark_WalksWith", MetricValue: 2, KindID: expectedKindIDs["dogpark_WalksWith"]},
							{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, EnvironmentID: "env-b", MetricType: model.DataQualityMetricTypeRelationship, MetricName: "dogpark_WalksWith", MetricValue: 1, KindID: expectedKindIDs["dogpark_WalksWith"]},
						}

						require.ElementsMatch(t, expectedStats, stats)
						return stats, nil
					})
				mockDB.EXPECT().
					CreateDataQualityAggregations(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, aggregations model.DataQualityAggregations) (model.DataQualityAggregations, error) {
						require.Len(t, aggregations, 3)
						require.NotEmpty(t, aggregations[0].RunID)

						expectedAggregations := model.DataQualityAggregations{
							{RunID: aggregations[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, MetricType: model.DataQualityMetricTypeNode, MetricName: "dogpark_Dog", MetricValue: 1, KindID: expectedKindIDs["dogpark_Dog"]},
							{RunID: aggregations[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, MetricType: model.DataQualityMetricTypeNode, MetricName: "dogpark_Patron", MetricValue: 5, KindID: expectedKindIDs["dogpark_Patron"]},
							{RunID: aggregations[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, MetricType: model.DataQualityMetricTypeRelationship, MetricName: "dogpark_WalksWith", MetricValue: 3, KindID: expectedKindIDs["dogpark_WalksWith"]},
						}

						require.ElementsMatch(t, expectedAggregations, aggregations)
						return aggregations, nil
					})
			},
		},
		{
			name: "skips opengraph inserts when there are no custom extensions",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{}, 0, nil)
			},
		},
		{
			name: "skips opengraph collection when extension management feature flag is disabled",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, false, nil)
			},
		},
		{
			name: "returns error when active directory stats collection fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectReadTransactionError(mockGraph, expectedError)
			},
			expectedError: "could not get active directory data quality stats",
		},
		{
			name: "returns error when azure stats collection fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADStatsRead(t, mockGraph)
				expectReadTransactionError(mockGraph, expectedError)
			},
			expectedError: "could not get azure data quality stats",
		},
		{
			name: "returns error when opengraph extension lookup fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{}, 0, expectedError)
			},
			expectedError: "could not get graph schema extensions",
		},
		{
			name: "returns error when opengraph extension management feature flag lookup fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, false, expectedError)
			},
			expectedError: "could not get open graph extension management feature flag",
		},
		{
			name: "returns error when opengraph environment lookup fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{schemaExtension}, 1, nil)
				mockDB.EXPECT().
					GetEnvironmentsByExtensionId(gomock.Any(), schemaExtension.ID).
					Return(nil, expectedError)
			},
			expectedError: "could not get environments for extension",
		},
		{
			name: "returns error when opengraph source kind lookup fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{schemaExtension}, 1, nil)
				mockDB.EXPECT().
					GetEnvironmentsByExtensionId(gomock.Any(), schemaExtension.ID).
					Return([]model.SchemaEnvironment{schemaEnvironment}, nil)
				expectOpenGraphRelationshipKinds(mockDB, schemaExtension, relationshipKind)
				mockDB.EXPECT().
					GetKindsByIDs(gomock.Any(), schemaEnvironment.SourceKindId).
					Return(nil, expectedError)
			},
			expectedError: "could not get source kind for schema environment",
		},
		{
			name: "returns error when opengraph graph read fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{schemaExtension}, 1, nil)
				mockDB.EXPECT().
					GetEnvironmentsByExtensionId(gomock.Any(), schemaExtension.ID).
					Return([]model.SchemaEnvironment{schemaEnvironment}, nil)
				mockDB.EXPECT().
					GetKindsByIDs(gomock.Any(), schemaEnvironment.SourceKindId).
					Return([]model.Kind{{ID: schemaEnvironment.SourceKindId, Name: sourceKind.String()}}, nil)
				expectOpenGraphRelationshipKinds(mockDB, schemaExtension, relationshipKind)
				expectReadTransactionError(mockGraph, expectedError)
			},
			expectedError: "could not count open graph nodes",
		},
		{
			name: "returns error when opengraph relationship kind lookup fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)

				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{schemaExtension}, 1, nil)
				mockDB.EXPECT().
					GetEnvironmentsByExtensionId(gomock.Any(), schemaExtension.ID).
					Return([]model.SchemaEnvironment{schemaEnvironment}, nil)
				mockDB.EXPECT().
					GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{"schema_extension_id": []model.Filter{{
						Operator:    model.Equals,
						Value:       fmt.Sprintf("%d", schemaExtension.ID),
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 0).
					Return(nil, 0, expectedError)
			},
			expectedError: "could not get relationship kinds for extension",
		},
		{
			name: "returns error when counted kind lookup fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				expectOpenGraphStatsRead(t, mockGraph, environmentKind, sourceKind, dogKind, patronKind, relationshipKind)

				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{schemaExtension}, 1, nil)
				mockDB.EXPECT().
					GetEnvironmentsByExtensionId(gomock.Any(), schemaExtension.ID).
					Return([]model.SchemaEnvironment{schemaEnvironment}, nil)
				mockDB.EXPECT().
					GetKindsByIDs(gomock.Any(), schemaEnvironment.SourceKindId).
					Return([]model.Kind{{ID: schemaEnvironment.SourceKindId, Name: sourceKind.String()}}, nil)
				expectOpenGraphRelationshipKinds(mockDB, schemaExtension, relationshipKind)
				mockDB.EXPECT().
					GetKindsByNames(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, expectedError)
			},
			expectedError: "could not get data quality metric kinds",
		},
		{
			name: "returns error when opengraph stats insert fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				expectOpenGraphStatsRead(t, mockGraph, environmentKind, sourceKind, dogKind, patronKind, relationshipKind)

				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{schemaExtension}, 1, nil)
				mockDB.EXPECT().
					GetEnvironmentsByExtensionId(gomock.Any(), schemaExtension.ID).
					Return([]model.SchemaEnvironment{schemaEnvironment}, nil)
				mockDB.EXPECT().
					GetKindsByIDs(gomock.Any(), schemaEnvironment.SourceKindId).
					Return([]model.Kind{{ID: schemaEnvironment.SourceKindId, Name: sourceKind.String()}}, nil)
				expectOpenGraphRelationshipKinds(mockDB, schemaExtension, relationshipKind)
				mockDB.EXPECT().
					GetKindsByNames(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]model.Kind{{ID: 101, Name: dogKind.String()}, {ID: 102, Name: patronKind.String()}, {ID: 103, Name: relationshipKind.String()}}, nil)
				mockDB.EXPECT().
					CreateDataQualityStats(gomock.Any(), gomock.Any()).
					Return(model.DataQualityStats{}, expectedError)
			},
			expectedError: "could not save open graph stats",
		},
		{
			name: "returns error when opengraph aggregation insert fails",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase, mockGraph *graphmocks.MockDatabase) {
				expectEmptyADAndAzureStatsReads(t, mockGraph)
				expectOpenGraphExtensionManagementFlag(mockDB, true, nil)
				expectOpenGraphStatsRead(t, mockGraph, environmentKind, sourceKind, dogKind, patronKind, relationshipKind)

				mockDB.EXPECT().
					GetGraphSchemaExtensions(gomock.Any(), builtInExtensionFilter, model.Sort{}, 0, 0).
					Return(model.GraphSchemaExtensions{schemaExtension}, 1, nil)
				mockDB.EXPECT().
					GetEnvironmentsByExtensionId(gomock.Any(), schemaExtension.ID).
					Return([]model.SchemaEnvironment{schemaEnvironment}, nil)
				mockDB.EXPECT().
					GetKindsByIDs(gomock.Any(), schemaEnvironment.SourceKindId).
					Return([]model.Kind{{ID: schemaEnvironment.SourceKindId, Name: sourceKind.String()}}, nil)
				expectOpenGraphRelationshipKinds(mockDB, schemaExtension, relationshipKind)
				mockDB.EXPECT().
					GetKindsByNames(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]model.Kind{{ID: 101, Name: dogKind.String()}, {ID: 102, Name: patronKind.String()}, {ID: 103, Name: relationshipKind.String()}}, nil)
				mockDB.EXPECT().
					CreateDataQualityStats(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, stats model.DataQualityStats) (model.DataQualityStats, error) {
						return stats, nil
					})
				mockDB.EXPECT().
					CreateDataQualityAggregations(gomock.Any(), gomock.Any()).
					Return(model.DataQualityAggregations{}, expectedError)
			},
			expectedError: "could not save open graph aggregations",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			mockDB := mocks.NewMockDatabase(mockController)
			mockGraph := graphmocks.NewMockDatabase(mockController)

			test.setupMocks(t, mockDB, mockGraph)

			err := dataquality.SaveDataQuality(ctx, mockDB, mockGraph)
			if test.expectedError != "" {
				require.ErrorContains(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func expectOpenGraphExtensionManagementFlag(mockDB *mocks.MockDatabase, enabled bool, err error) {
	mockDB.EXPECT().
		GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
		Return(appcfg.FeatureFlag{Enabled: enabled}, err)
}

func expectOpenGraphRelationshipKinds(mockDB *mocks.MockDatabase, schemaExtension model.GraphSchemaExtension, relationshipKinds ...graph.Kind) {
	schemaRelationshipKinds := make(model.GraphSchemaRelationshipKinds, 0, len(relationshipKinds))
	for _, relationshipKind := range relationshipKinds {
		schemaRelationshipKinds = append(schemaRelationshipKinds, model.GraphSchemaRelationshipKind{Name: relationshipKind.String()})
	}

	mockDB.EXPECT().
		GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{"schema_extension_id": []model.Filter{{
			Operator:    model.Equals,
			Value:       fmt.Sprintf("%d", schemaExtension.ID),
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 0).
		Return(schemaRelationshipKinds, len(schemaRelationshipKinds), nil)
}

func expectEmptyADAndAzureStatsReads(t *testing.T, mockGraph *graphmocks.MockDatabase) {
	t.Helper()

	expectEmptyADStatsRead(t, mockGraph)

	expectReadTransaction(t, mockGraph, func(mockController *gomock.Controller, mockTransaction *graphmocks.MockTransaction) {
		expectFetchNodes(mockController, mockTransaction, nil)
	})
}

func expectEmptyADStatsRead(t *testing.T, mockGraph *graphmocks.MockDatabase) {
	t.Helper()

	expectReadTransaction(t, mockGraph, func(mockController *gomock.Controller, mockTransaction *graphmocks.MockTransaction) {
		expectFetchNodes(mockController, mockTransaction, nil)
		expectFetchNodes(mockController, mockTransaction, nil)
		expectFetchNodes(mockController, mockTransaction, nil)
	})
}

func expectOpenGraphStatsRead(t *testing.T, mockGraph *graphmocks.MockDatabase, environmentKind graph.Kind, sourceKind graph.Kind, dogKind graph.Kind, patronKind graph.Kind, relationshipKind graph.Kind) {
	t.Helper()

	expectReadTransaction(t, mockGraph, func(mockController *gomock.Controller, mockTransaction *graphmocks.MockTransaction) {
		expectFetchNodes(mockController, mockTransaction, []*graph.Node{
			graph.NewNode(1, graph.AsProperties(map[string]any{graphschema.EnvironmentIDKey: "env-a"}), environmentKind),
			graph.NewNode(2, graph.AsProperties(map[string]any{graphschema.EnvironmentIDKey: "env-b"}), environmentKind),
		})
		expectFetchKinds(mockController, mockTransaction, []graph.KindsResult{
			{ID: 3, Kinds: graph.Kinds{sourceKind, dogKind, patronKind, graphschema.Meta}},
			{ID: 4, Kinds: graph.Kinds{sourceKind, patronKind}},
		})
		expectFetchRelationshipKinds(mockController, mockTransaction, []graph.RelationshipKindsResult{
			{RelationshipTripleResult: graph.RelationshipTripleResult{ID: 8, StartID: 3, EndID: 4}, Kind: relationshipKind},
			{RelationshipTripleResult: graph.RelationshipTripleResult{ID: 9, StartID: 4, EndID: 3}, Kind: relationshipKind},
		})
		expectFetchKinds(mockController, mockTransaction, []graph.KindsResult{
			{ID: 5, Kinds: graph.Kinds{sourceKind, patronKind}},
			{ID: 6, Kinds: graph.Kinds{sourceKind, patronKind}},
			{ID: 7, Kinds: graph.Kinds{sourceKind, patronKind}},
		})
		expectFetchRelationshipKinds(mockController, mockTransaction, []graph.RelationshipKindsResult{
			{RelationshipTripleResult: graph.RelationshipTripleResult{ID: 10, StartID: 5, EndID: 6}, Kind: relationshipKind},
		})
	})
}

func expectReadTransactionError(mockGraph *graphmocks.MockDatabase, err error) {
	mockGraph.EXPECT().
		ReadTransaction(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectReadTransaction(t *testing.T, mockGraph *graphmocks.MockDatabase, setupTransaction func(mockController *gomock.Controller, mockTransaction *graphmocks.MockTransaction)) {
	t.Helper()

	mockController := gomock.NewController(t)
	mockTransaction := graphmocks.NewMockTransaction(mockController)
	setupTransaction(mockController, mockTransaction)

	mockGraph.EXPECT().
		ReadTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...graph.TransactionOption) error {
			return delegate(mockTransaction)
		})
}

func expectFetchNodes(mockController *gomock.Controller, mockTransaction *graphmocks.MockTransaction, nodes []*graph.Node) {
	mockNodeQuery := graphmocks.NewMockNodeQuery(mockController)
	mockTransaction.EXPECT().Nodes().Return(mockNodeQuery)
	mockNodeQuery.EXPECT().Filterf(gomock.Any()).Return(mockNodeQuery)
	mockNodeQuery.EXPECT().
		Fetch(gomock.Any()).
		DoAndReturn(func(delegate func(graph.Cursor[*graph.Node]) error, _ ...graph.Criteria) error {
			mockCursor := graphmocks.NewMockCursor[*graph.Node](mockController)
			mockCursor.EXPECT().Chan().Return(cursorChan(nodes))
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})
}

func expectFetchKinds(mockController *gomock.Controller, mockTransaction *graphmocks.MockTransaction, kindsResults []graph.KindsResult) {
	mockNodeQuery := graphmocks.NewMockNodeQuery(mockController)
	mockTransaction.EXPECT().Nodes().Return(mockNodeQuery)
	mockNodeQuery.EXPECT().Filterf(gomock.Any()).Return(mockNodeQuery)
	mockNodeQuery.EXPECT().
		FetchKinds(gomock.Any()).
		DoAndReturn(func(delegate func(graph.Cursor[graph.KindsResult]) error) error {
			mockCursor := graphmocks.NewMockCursor[graph.KindsResult](mockController)
			mockCursor.EXPECT().Chan().Return(cursorChan(kindsResults))
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})
}

func expectFetchRelationshipKinds(mockController *gomock.Controller, mockTransaction *graphmocks.MockTransaction, kindsResults []graph.RelationshipKindsResult) {
	mockRelationshipQuery := graphmocks.NewMockRelationshipQuery(mockController)
	mockTransaction.EXPECT().Relationships().Return(mockRelationshipQuery)
	mockRelationshipQuery.EXPECT().Filterf(gomock.Any()).Return(mockRelationshipQuery)
	mockRelationshipQuery.EXPECT().
		FetchKinds(gomock.Any()).
		DoAndReturn(func(delegate func(graph.Cursor[graph.RelationshipKindsResult]) error) error {
			mockCursor := graphmocks.NewMockCursor[graph.RelationshipKindsResult](mockController)
			mockCursor.EXPECT().Chan().Return(cursorChan(kindsResults))
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})
}

func cursorChan[T any](items []T) chan T {
	ch := make(chan T, len(items))
	for _, item := range items {
		ch <- item
	}
	close(ch)
	return ch
}
