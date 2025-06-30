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

//go:build integration
// +build integration

package queries_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"

	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/cmd/api/src/api/bloodhoundgraph"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func TestSearchNodesByName_ExactMatch(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			harness.SearchHarness.Setup(testContext)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			var (
				userWanted = "USER NUMBER ONE"
				skip       = 0
				limit      = 10
				graphQuery = queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})
			)

			results, err := graphQuery.SearchNodesByName(context.Background(), graph.Kinds{azure.Entity, ad.Entity}, userWanted, skip, limit)
			require.Equal(t, 1, len(results), "There should be one exact match returned")
			require.Nil(t, err)
			expectedUser := results[0]
			require.Equal(t, expectedUser.Name, userWanted)
		})
}

func TestSearchNodesByName_FuzzyMatch(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			harness.SearchHarness.Setup(testContext)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			var (
				userWanted = "USER NUMBER"
				skip       = 0
				limit      = 10
				graphQuery = queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})
			)

			results, err := graphQuery.SearchNodesByName(context.Background(), graph.Kinds{azure.Entity, ad.Entity}, userWanted, skip, limit)

			require.Nil(t, err)
			require.Equal(t, 5, len(results), "All users that contain `USER NUMBER` should be returned ")
		})
}

func TestSearchNodesByName_NoADLocalGroup(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			harness.SearchHarness.Setup(testContext)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			var (
				userWanted = "Remote Desktop"
				skip       = 0
				limit      = 10
				graphQuery = queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})
			)

			results, err := graphQuery.SearchNodesByName(context.Background(), graph.Kinds{azure.Entity, ad.Entity}, userWanted, skip, limit)

			require.Nil(t, err)
			require.Equal(t, 0, len(results), "No ADLocalGroup nodes should be returned ")
		})
}

func TestSearchNodesByName_GroupLocalGroupCorrect(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			harness.SearchHarness.Setup(testContext)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			var (
				groupWanted = "Account Op"
				skip        = 0
				limit       = 10
				graphQuery  = queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})
			)

			results, err := graphQuery.SearchNodesByName(context.Background(), graph.Kinds{azure.Entity, ad.Entity}, groupWanted, skip, limit)

			require.Nil(t, err)
			require.Equal(t, 1, len(results), ":ADLocalGroup nodes should return if they are also :Group nodes")
		})
}

func TestSearchNodesByName_ExactMatch_ObjectID(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			harness.SearchHarness.Setup(testContext)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			var (
				userObjectId = harness.SearchHarness.User1.Properties.Get(common.ObjectID.String())
				skip         = 0
				limit        = 10
				graphQuery   = queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})
			)

			searchQuery, _ := userObjectId.String()

			results, err := graphQuery.SearchNodesByName(context.Background(), graph.Kinds{azure.Entity, ad.Entity}, searchQuery, skip, limit)

			actual := results[0]

			require.Nil(t, err)
			require.Equal(t, 1, len(results), "Only one user can match exactly one Object ID")
			require.Equal(t, searchQuery, actual.ObjectID)
		})
}

func TestGetEntityResults(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	queryCache, err := cache.NewCache(cache.Config{MaxSize: 1})
	require.Nil(t, err)

	testContext.SetupActiveDirectory()
	testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
		objectID, err := harness.InboundControl.ControlledUser.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)

		graphQuery := queries.GraphQuery{
			Graph: db,
			Cache: queryCache,
		}
		params := queries.EntityQueryParameters{
			QueryName:     "InboundADEntityController",
			ObjectID:      objectID,
			Skip:          1,
			Limit:         2,
			RequestedType: 1,
			PathDelegate:  adAnalysis.FetchInboundADEntityControllerPaths,
			ListDelegate:  adAnalysis.FetchInboundADEntityControllers,
		}

		results, count, err := graphQuery.GetADEntityQueryResult(context.Background(), params, false)
		require.Nil(t, err)

		require.Equal(t, 4, count)
		require.Len(t, results, 2)
		require.Equal(t, 0, queryCache.Len())
	})
}

func TestGetEntityResults_QueryShorterThanSlowQueryThreshold(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	queryCache, err := cache.NewCache(cache.Config{MaxSize: 1})
	require.Nil(t, err)

	testContext.SetupActiveDirectory()
	testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
		objectID, err := harness.InboundControl.ControlledUser.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)

		graphQuery := queries.GraphQuery{
			Graph:              db,
			Cache:              queryCache,
			SlowQueryThreshold: 1500000,
		}
		params := queries.EntityQueryParameters{
			QueryName:     "InboundADEntityController",
			ObjectID:      objectID,
			Skip:          1,
			Limit:         2,
			RequestedType: 1,
			PathDelegate:  adAnalysis.FetchInboundADEntityControllerPaths,
			ListDelegate:  adAnalysis.FetchInboundADEntityControllers,
		}

		results, count, err := graphQuery.GetADEntityQueryResult(context.Background(), params, true)
		require.Nil(t, err)

		require.Equal(t, 4, count)
		require.Len(t, results, 2)
		require.Equal(t, 0, queryCache.Len())
	})
}

func TestGetPrimaryNodeKindCounts(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.SetupActiveDirectory()
	testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
		graphQuery := queries.GraphQuery{
			Graph: db,
		}

		results, err := graphQuery.GetPrimaryNodeKindCounts(context.Background(), ad.Entity)
		require.Nil(t, err)

		// While this is a very thin test, any more specificity would require constant updates each time the harness added new kind
		require.Greater(t, len(results), 1)
	})
}

func TestFetchNodeByGraphId(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZBaseHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		graphQuery := queries.GraphQuery{Graph: db}

		node, err := graphQuery.FetchNodeByGraphId(context.Background(), harness.AZBaseHarness.User.ID)
		require.Nil(t, err)
		require.NotNil(t, node)

		require.Equal(t, harness.AZBaseHarness.User.ID, node.ID)
	})
}

func TestGetEntityResults_Cache(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	queryCache, err := cache.NewCache(cache.Config{MaxSize: 2})
	require.Nil(t, err)

	testContext.SetupActiveDirectory()
	testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
		objectID, err := harness.InboundControl.ControlledUser.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)

		graphQuery := queries.GraphQuery{
			Graph:              db,
			Cache:              queryCache,
			SlowQueryThreshold: 0, // Setting 0 ensures it will cache any queries
		}

		params := queries.EntityQueryParameters{
			QueryName:     "InboundADEntityController",
			ObjectID:      objectID,
			Skip:          1,
			Limit:         2,
			RequestedType: 1,
			PathDelegate:  adAnalysis.FetchInboundADEntityControllerPaths,
			ListDelegate:  adAnalysis.FetchInboundADEntityControllers,
		}

		// Get results and check that an entry was added to the cache
		results, count, err := graphQuery.GetADEntityQueryResult(context.Background(), params, true)
		require.Nil(t, err)

		require.Equal(t, 4, count)
		require.Len(t, results, 2)
		require.Equal(t, 1, queryCache.Len())

		// Ensure that after being cached, we get the same results
		results, count, err = graphQuery.GetADEntityQueryResult(context.Background(), params, true)
		require.Nil(t, err)

		require.Equal(t, 4, count)
		require.Len(t, results, 2)
		require.Equal(t, 1, queryCache.Len())
	})
}

func TestGetAssetGroupComboNode(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.SetupActiveDirectory()
	testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
		graphQuery := queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})
		comboNode, err := graphQuery.GetAssetGroupComboNode(context.Background(), "", ad.AdminTierZero)
		require.Nil(t, err)

		groupBObjectID := harness.AssetGroupComboNodeHarness.GroupB.ID.String()
		groupAObjectID := harness.AssetGroupComboNodeHarness.GroupA.ID.String()

		groupACategory := comboNode[groupAObjectID].(bloodhoundgraph.BloodHoundGraphNode).BloodHoundGraphItem.Data["category"]
		groupBCategory := comboNode[groupBObjectID].(bloodhoundgraph.BloodHoundGraphNode).BloodHoundGraphItem.Data["category"]

		// ensure that nodes from within T0 as well as from other domains all have the category tagged
		require.Equal(t, "Asset Groups", groupACategory)
		require.Equal(t, "Asset Groups", groupBCategory)
	})
}

func TestGetAssetGroupNodes(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AssetGroupNodesHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		graphQuery := queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})

		tierZeroNodes, err := graphQuery.GetAssetGroupNodes(context.Background(), harness.AssetGroupNodesHarness.TierZeroTag, true)
		require.Nil(t, err)

		customGroup1Nodes, err := graphQuery.GetAssetGroupNodes(context.Background(), harness.AssetGroupNodesHarness.CustomTag1, false)
		require.Nil(t, err)

		customGroup2Nodes, err := graphQuery.GetAssetGroupNodes(context.Background(), harness.AssetGroupNodesHarness.CustomTag2, false)
		require.Nil(t, err)

		require.True(t, tierZeroNodes.Contains(harness.AssetGroupNodesHarness.GroupB))
		require.True(t, tierZeroNodes.Contains(harness.AssetGroupNodesHarness.GroupC))
		require.Equal(t, 2, len(tierZeroNodes))

		require.True(t, customGroup1Nodes.Contains(harness.AssetGroupNodesHarness.GroupD))
		require.Equal(t, 1, len(customGroup1Nodes))

		require.True(t, customGroup2Nodes.Contains(harness.AssetGroupNodesHarness.GroupE))
		require.Equal(t, 1, len(customGroup2Nodes))
	})
}

func TestGraphQuery_GetAllShortestPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			var (
				userA = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     "A",
					common.ObjectID: "A",
				}), ad.Entity, ad.User)

				groupA = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     "GA",
					common.ObjectID: "B",
				}), ad.Entity, ad.Group)

				computer = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     "C",
					common.ObjectID: "C",
				}), ad.Entity, ad.Computer)
			)

			testContext.NewRelationship(userA, groupA, ad.MemberOf)
			testContext.NewRelationship(groupA, computer, ad.GenericAll)
			testContext.NewRelationship(userA, computer, ad.GenericWrite)

			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			graphQuery := queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})
			paths, err := graphQuery.GetAllShortestPaths(context.Background(), "A", "C", query.KindIn(query.Relationship(), ad.Relationships()...))

			require.Nil(t, err)
			require.Equal(t, 1, len(paths))
			require.Equal(t, 1, len(paths[0].Edges))

			paths, err = graphQuery.GetAllShortestPaths(context.Background(), "A", "C", query.KindIn(query.Relationship(), graph.Kinds(ad.Relationships()).Exclude(graph.Kinds{ad.GenericWrite})...))

			require.Nil(t, err)
			require.Equal(t, 1, len(paths))
			require.Equal(t, 2, len(paths[0].Edges))

			paths, err = graphQuery.GetAllShortestPaths(context.Background(), "A", "C", query.KindIn(query.Relationship(), ad.HasSession))

			require.Nil(t, err)
			require.Equal(t, 0, len(paths))
		})
}

func TestGetFilteredAndSortedNodesPaginated(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.SetupActiveDirectory()

	t.Run("Test sort ascending", func(t *testing.T) {
		testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
			graphQuery := queries.GraphQuery{
				Graph: db,
			}

			results, err := graphQuery.GetFilteredAndSortedNodesPaginated(
				query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionAscending}}, // sort by node ID ascending
				query.KindIn(query.Node(), ad.User), // give me all the nodes of kind ad.User
				0,
				0)
			require.Nil(t, err)
			// verify some nodes were returned
			require.Greater(t, len(results), 10)

			// verify the nodes are sorted ascending by ID
			prevNodeId := int64(0)
			for _, node := range results {
				require.GreaterOrEqual(t, node.ID.Int64(), prevNodeId)
				prevNodeId = node.ID.Int64()
			}
		})
	})

	t.Run("Test sort descending", func(t *testing.T) {
		testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
			graphQuery := queries.GraphQuery{
				Graph: db,
			}

			results, err := graphQuery.GetFilteredAndSortedNodesPaginated(
				query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionDescending}}, // sort by node ID descending
				query.KindIn(query.Node(), ad.User), // give me all the nodes of kind ad.User
				0,
				0)
			require.Nil(t, err)
			// verify some nodes were returned
			require.Greater(t, len(results), 10)

			prevNodeId := results[0].ID.Int64()
			for _, node := range results {
				require.LessOrEqual(t, node.ID.Int64(), prevNodeId)
				prevNodeId = node.ID.Int64()
			}
		})
	})

	t.Run("Test limit", func(t *testing.T) {
		testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
			graphQuery := queries.GraphQuery{
				Graph: db,
			}

			results, err := graphQuery.GetFilteredAndSortedNodesPaginated(
				query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionAscending}}, // sort by node ID Ascending
				query.KindIn(query.Node(), ad.User), // give me all the nodes of kind ad.User
				0,
				5)
			require.Nil(t, err)
			require.Len(t, results, 5)
		})
	})

	t.Run("Test offset", func(t *testing.T) {
		testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
			graphQuery := queries.GraphQuery{
				Graph: db,
			}

			results, err := graphQuery.GetFilteredAndSortedNodesPaginated(
				query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionAscending}}, // sort by node ID Ascending
				query.KindIn(query.Node(), ad.User), // give me all the nodes of kind ad.User
				0,
				0)
			require.Nil(t, err)

			// call again with offset 10, making sure results are properly shifted
			savedNode := results[10]
			results, err = graphQuery.GetFilteredAndSortedNodesPaginated(
				query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionAscending}}, // sort by node ID Ascending
				query.KindIn(query.Node(), ad.User), // give me all the nodes of kind ad.User
				10,
				0)
			require.Nil(t, err)
			require.Equal(t, savedNode, results[0])
		})
	})
}
