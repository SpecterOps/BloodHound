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

package dataquality

import (
	"context"
	"testing"

	databasemocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOpenGraphStatsEnvironmentProperties(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                  string
		builtin               bool
		environmentKind       graph.Kind
		sourceKind            graph.Kind
		nodeKind              graph.Kind
		relationshipKind      graph.Kind
		environmentIDProperty string
		membershipProperty    string
	}{
		{name: "AD", builtin: true, environmentKind: ad.Domain, sourceKind: ad.Entity, nodeKind: ad.User, relationshipKind: ad.MemberOf, environmentIDProperty: "objectid", membershipProperty: "domainsid"},
		{name: "Azure", builtin: true, environmentKind: azure.Tenant, sourceKind: azure.Entity, nodeKind: azure.User, relationshipKind: azure.MemberOf, environmentIDProperty: "objectid", membershipProperty: "tenantid"},
		{name: "custom OpenGraph", environmentKind: graph.StringKind("test_Environment"), sourceKind: graph.StringKind("test_Base"), nodeKind: graph.StringKind("test_User"), relationshipKind: graph.StringKind("test_MemberOf"), environmentIDProperty: "environmentid", membershipProperty: "environmentid"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var (
				ctx               = context.Background()
				controller        = gomock.NewController(t)
				database          = databasemocks.NewMockDatabase(controller)
				graphDatabase     = graphmocks.NewMockDatabase(controller)
				transaction       = graphmocks.NewMockTransaction(controller)
				extension         = model.GraphSchemaExtension{Serial: model.Serial{ID: 44}, IsBuiltin: testCase.builtin}
				environment       = model.SchemaEnvironment{SchemaExtensionId: 44, EnvironmentKindId: 22, EnvironmentKindName: testCase.environmentKind.String(), SourceKindId: 11}
				environmentQuery  = graphmocks.NewMockNodeQuery(controller)
				environmentCursor = graphmocks.NewMockCursor[*graph.Node](controller)
			)
			database.EXPECT().GetGraphSchemaExtensions(ctx, model.Filters{}, model.Sort{}, 0, 0).Return(model.GraphSchemaExtensions{extension}, 1, nil)
			database.EXPECT().GetEnvironmentsByExtensionId(ctx, extension.ID).Return([]model.SchemaEnvironment{environment}, nil)
			database.EXPECT().GetGraphSchemaRelationshipKinds(ctx, model.Filters{"schema_extension_id": {{Operator: model.Equals, Value: "44", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0).Return(model.GraphSchemaRelationshipKinds{{Name: testCase.relationshipKind.String()}}, 1, nil)
			database.EXPECT().GetKindsByIDs(ctx, int32(11)).Return([]model.Kind{{ID: 11, Name: testCase.sourceKind.String()}}, nil)
			database.EXPECT().GetKindsByNames(ctx, testCase.nodeKind.String(), testCase.relationshipKind.String()).Return([]model.Kind{{ID: 101, Name: testCase.nodeKind.String()}, {ID: 102, Name: testCase.relationshipKind.String()}}, nil)
			graphDatabase.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...graph.TransactionOption) error {
				return delegate(transaction)
			})
			transaction.EXPECT().Nodes().Return(environmentQuery)
			environmentQuery.EXPECT().Filterf(gomock.Any()).DoAndReturn(func(provider graph.CriteriaProvider) graph.NodeQuery {
				require.Equal(t, query.Kind(query.Node(), testCase.environmentKind), provider())
				return environmentQuery
			})
			environmentQuery.EXPECT().Fetch(gomock.Any()).DoAndReturn(func(delegate func(graph.Cursor[*graph.Node]) error, _ ...graph.Criteria) error {
				return delegate(environmentCursor)
			})
			environmentCursor.EXPECT().Chan().Return(pocCursorChannel([]*graph.Node{
				graph.NewNode(1, graph.AsProperties(map[string]any{testCase.environmentIDProperty: "env-a"}), testCase.environmentKind),
				graph.NewNode(2, graph.AsProperties(map[string]any{testCase.environmentIDProperty: "env-b"}), testCase.environmentKind),
				// Missing IDs must be skipped rather than counting an empty environment ID.
				graph.NewNode(3, graph.AsProperties(map[string]any{}), testCase.environmentKind),
			}))
			environmentCursor.EXPECT().Error().Return(nil)
			for _, environmentID := range []string{"env-a", "env-b"} {
				var (
					nodeQuery          = graphmocks.NewMockNodeQuery(controller)
					nodeCursor         = graphmocks.NewMockCursor[graph.KindsResult](controller)
					relationshipQuery  = graphmocks.NewMockRelationshipQuery(controller)
					relationshipCursor = graphmocks.NewMockCursor[graph.RelationshipKindsResult](controller)
				)
				transaction.EXPECT().Nodes().Return(nodeQuery)
				nodeQuery.EXPECT().Filterf(gomock.Any()).DoAndReturn(func(provider graph.CriteriaProvider) graph.NodeQuery {
					require.Equal(t, query.And(query.Kind(query.Node(), testCase.sourceKind), query.Equals(query.NodeProperty(testCase.membershipProperty), environmentID)), provider())
					return nodeQuery
				})
				nodeQuery.EXPECT().FetchKinds(gomock.Any()).DoAndReturn(func(delegate func(graph.Cursor[graph.KindsResult]) error) error { return delegate(nodeCursor) })
				nodeCursor.EXPECT().Chan().Return(pocCursorChannel([]graph.KindsResult{
					{ID: 4, Kinds: graph.Kinds{testCase.sourceKind, testCase.nodeKind, graphschema.Meta}},
					{ID: 5, Kinds: graph.Kinds{testCase.sourceKind, testCase.nodeKind}},
				}))
				nodeCursor.EXPECT().Error().Return(nil)
				transaction.EXPECT().Relationships().Return(relationshipQuery)
				relationshipQuery.EXPECT().Filterf(gomock.Any()).DoAndReturn(func(provider graph.CriteriaProvider) graph.RelationshipQuery {
					require.Equal(t, query.And(query.Kind(query.Start(), testCase.sourceKind), query.Equals(query.StartProperty(testCase.membershipProperty), environmentID), query.KindIn(query.Relationship(), testCase.relationshipKind)), provider())
					return relationshipQuery
				})
				relationshipQuery.EXPECT().FetchKinds(gomock.Any()).DoAndReturn(func(delegate func(graph.Cursor[graph.RelationshipKindsResult]) error) error {
					return delegate(relationshipCursor)
				})
				relationshipCursor.EXPECT().Chan().Return(pocCursorChannel([]graph.RelationshipKindsResult{{Kind: testCase.relationshipKind}}))
				relationshipCursor.EXPECT().Error().Return(nil)
			}

			stats, aggregations, err := openGraphStats(ctx, database, graphDatabase)
			require.NoError(t, err)
			require.Len(t, stats, 4)
			require.NotEmpty(t, stats[0].RunID)
			expectedStats := model.DataQualityStats{}
			for _, environmentID := range []string{"env-a", "env-b"} {
				expectedStats = append(expectedStats,
					model.DataQualityStat{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, EnvironmentID: environmentID, MetricType: model.DataQualityMetricTypeNode, MetricName: testCase.nodeKind.String(), MetricValue: 2, KindID: null.Int32From(101)},
					model.DataQualityStat{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, EnvironmentID: environmentID, MetricType: model.DataQualityMetricTypeRelationship, MetricName: testCase.relationshipKind.String(), MetricValue: 1, KindID: null.Int32From(102)},
				)
			}
			require.ElementsMatch(t, expectedStats, stats)
			require.ElementsMatch(t, model.DataQualityAggregations{
				{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, MetricType: model.DataQualityMetricTypeNode, MetricName: testCase.nodeKind.String(), MetricValue: 4, KindID: null.Int32From(101)},
				{RunID: stats[0].RunID, SchemaExtensionID: 44, SchemaEnvironmentKindID: 22, MetricType: model.DataQualityMetricTypeRelationship, MetricName: testCase.relationshipKind.String(), MetricValue: 2, KindID: null.Int32From(102)},
			}, aggregations)
		})
	}
}

func pocCursorChannel[T any](items []T) chan T {
	var channel = make(chan T, len(items))
	for _, item := range items {
		channel <- item
	}
	close(channel)
	return channel
}
