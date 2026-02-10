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

package queries_test

import (
	"context"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/lab/generic"

	"github.com/specterops/bloodhound/cmd/api/src/api/bloodhoundgraph"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func setupGraphDb(t *testing.T) IntegrationTestSuite {
	var (
		fixturesPath = path.Join("fixtures", "OpenGraphJSON", "raw")
		testSuite    = setupIntegrationTestSuite(t, fixturesPath)
	)

	// Populate the graph with data

	base, err := generic.LoadGraphFromFile(os.DirFS(testSuite.WorkDir), "base.json")
	require.NoError(t, err)

	err = generic.WriteGraphToDatabase(testSuite.GraphDB, &base)
	require.NoError(t, err)
	return testSuite
}

func TestSearchNodesByNameOrObjectId(t *testing.T) {
	type testData struct {
		name                      string
		queryString               string
		inputArguments            graph.Kinds
		includeOpenGraphNodes     bool
		expectedResults           int
		expectedResultExplanation string
		shouldMatchUser           bool
		matchUserField            string
		shouldMatchType           bool
		expectedType              string
		expectedTypeExplanation   string
	}
	var (
		testSuite  = setupGraphDb(t)
		graphQuery = queries.NewGraphQuery(testSuite.GraphDB, cache.Cache{}, config.Configuration{})
		testTable  = []testData{
			{
				name:                      "Exact Match",
				queryString:               "USER NUMBER ONE",
				inputArguments:            graph.Kinds{azure.Entity, ad.Entity},
				includeOpenGraphNodes:     false,
				expectedResults:           1,
				expectedResultExplanation: "There should be one exact match returned",
				shouldMatchUser:           true,
				matchUserField:            "Name",
			},
			{
				name:                      "Fuzzy Match",
				queryString:               "USER NUMBER",
				inputArguments:            graph.Kinds{azure.Entity, ad.Entity},
				includeOpenGraphNodes:     false,
				expectedResults:           5,
				expectedResultExplanation: "All users that contain `USER_NUMBER` should be returned",
				shouldMatchUser:           false,
			},
			{
				name:                      "No AD Local Group",
				queryString:               "Remote Desktop",
				inputArguments:            graph.Kinds{azure.Entity, ad.Entity},
				includeOpenGraphNodes:     false,
				expectedResults:           0,
				expectedResultExplanation: "No ADLocalGroup nodes should be returned",
				shouldMatchUser:           false,
			},
			{
				name:                      "Returns OpenGraph results",
				queryString:               "person",
				inputArguments:            nil,
				includeOpenGraphNodes:     true,
				expectedResults:           3,
				expectedResultExplanation: "All three OpenGraph nodes should be returned",
				shouldMatchUser:           false,
				shouldMatchType:           true,
				expectedType:              "Person",
				expectedTypeExplanation:   "All three OpenGraph nodes should have type Person",
			},
			{
				name:                      "Returns Nodes from all Graphs",
				queryString:               "two",
				inputArguments:            nil,
				includeOpenGraphNodes:     true,
				expectedResults:           2,
				expectedResultExplanation: "All nodes with `two` in the name should be returned",
				shouldMatchUser:           false,
			},
			{
				name:                      "Group Local Group Correct",
				queryString:               "Account Op",
				inputArguments:            nil,
				includeOpenGraphNodes:     false,
				expectedResults:           1,
				expectedResultExplanation: ":ADLocalGroup nodes should return if they are also :Group nodes",
				shouldMatchUser:           false,
			},
			{
				name:                      "Exact Match ObjectID",
				queryString:               "2",
				inputArguments:            graph.Kinds{azure.Entity, ad.Entity},
				expectedResults:           1,
				includeOpenGraphNodes:     false,
				expectedResultExplanation: "Only one user can match exactly one Object ID",
				shouldMatchUser:           true,
				matchUserField:            "ObjectID",
			},
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			results, err := graphQuery.SearchNodesByNameOrObjectId(testSuite.Context, testCase.inputArguments, testCase.queryString, testCase.includeOpenGraphNodes, 0, 10, nil)
			require.Nil(t, err)
			require.Equal(t, testCase.expectedResults, len(results), testCase.expectedResultExplanation)
			if testCase.shouldMatchUser {
				expectedUser := results[0]
				value := reflect.ValueOf(expectedUser)
				require.Equal(t, value.FieldByName(testCase.matchUserField).String(), testCase.queryString)
			}
			if testCase.shouldMatchType {
				for _, result := range results {
					require.Equal(t, testCase.expectedType, result.Type, testCase.expectedTypeExplanation)
				}
			}
		})
	}
}

func TestSearchByNameOrObjectId(t *testing.T) {
	type testData struct {
		name                      string
		includeOpenGraph          bool
		searchValue               string
		searchType                string
		expectedResults           int
		expectedResultExplanation string
		shouldMatchUser           bool
		matchUserField            string
	}
	var (
		testSuite  = setupGraphDb(t)
		graphQuery = queries.NewGraphQuery(testSuite.GraphDB, cache.Cache{}, config.Configuration{})
		testTable  = []testData{
			{
				name:                      "Exact Match",
				includeOpenGraph:          false,
				searchValue:               "USER NUMBER ONE",
				searchType:                queries.SearchTypeExact,
				expectedResults:           1,
				expectedResultExplanation: "There should be one exact match returned",
				shouldMatchUser:           true,
				matchUserField:            "name",
			},
			{
				name:                      "Fuzzy Match",
				includeOpenGraph:          false,
				searchValue:               "USER NUMBER",
				searchType:                queries.SearchTypeFuzzy,
				expectedResults:           5,
				expectedResultExplanation: "All users that start with `USER_NUMBER` should be returned",
				shouldMatchUser:           false,
			},
			{
				name:                      "Returns OpenGraph results",
				includeOpenGraph:          true,
				searchValue:               "person",
				searchType:                queries.SearchTypeFuzzy,
				expectedResults:           3,
				expectedResultExplanation: "All three OpenGraph nodes should be returned",
				shouldMatchUser:           false,
			},
			{
				name:                      "Returns Nodes from all Graphs",
				includeOpenGraph:          true,
				searchValue:               "a",
				searchType:                queries.SearchTypeFuzzy,
				expectedResults:           2,
				expectedResultExplanation: "All nodes starting with `a` should be returned",
				shouldMatchUser:           false,
			},
			{
				name:                      "Exact Match ObjectID",
				includeOpenGraph:          false,
				searchValue:               "2",
				searchType:                queries.SearchTypeExact,
				expectedResults:           1,
				expectedResultExplanation: "Only one user can match exactly one Object ID",
				shouldMatchUser:           true,
				matchUserField:            "objectid",
			},
			{
				name:                      "Exact Match OpenGraph Node",
				includeOpenGraph:          true,
				searchValue:               "PERSON ONE",
				searchType:                queries.SearchTypeExact,
				expectedResults:           1,
				expectedResultExplanation: "There should be one exact match returned",
				shouldMatchUser:           true,
				matchUserField:            "name",
			},
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			results, err := graphQuery.SearchByNameOrObjectID(testSuite.Context, testCase.includeOpenGraph, testCase.searchValue, testCase.searchType)
			require.Nil(t, err)
			require.Equal(t, testCase.expectedResults, len(results), testCase.expectedResultExplanation)
			if testCase.shouldMatchUser {
				var actualResult *graph.Node
				for _, node := range results {
					actualResult = node
					break
				}
				actualValue := actualResult.Properties.Map[testCase.matchUserField]
				require.Equal(t, actualValue, testCase.searchValue)
			}
		})
	}
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

		groupACategory := comboNode[groupAObjectID].(bloodhoundgraph.BloodHoundGraphNode).Data["category"]
		groupBCategory := comboNode[groupBObjectID].(bloodhoundgraph.BloodHoundGraphNode).Data["category"]

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

func TestGetAllShortestPathsWithOpenGraph(t *testing.T) {
	type testData struct {
		name        string
		startNodeID string
		endNodeID   string
		filter      graph.Criteria
		expected    graph.PathSet
	}
	var (
		testSuite     = setupGraphDb(t)
		graphQuery    = queries.NewGraphQuery(testSuite.GraphDB, cache.Cache{}, config.Configuration{})
		allValidKinds = graph.Kinds(graph.StringsToKinds([]string{"Knows", "Contains", "IsParent"}))
		testTable     = []testData{
			{
				name:        "Find shortest path between 2 OpenGraph Nodes",
				startNodeID: "7",
				endNodeID:   "12",
				filter:      query.KindIn(query.Relationship(), allValidKinds...),
				expected: graph.NewPathSet(graph.Path{
					Nodes: []*graph.Node{
						{
							ID:    7,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "PERSON ONE",
								"objectid": "7",
							}),
						},
						{
							ID:    12,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"name":     "ALICE",
								"objectid": "12",
							}),
						},
					},
					Edges: []*graph.Relationship{
						{
							ID:         4,
							StartID:    7,
							EndID:      12,
							Kind:       graph.StringKind("IsParent"),
							Properties: graph.NewPropertiesRed(),
						},
					},
				}),
			},
			{
				name:        "Find shortest path between 2 AD Nodes",
				startNodeID: "5",
				endNodeID:   "10",
				filter:      query.KindIn(query.Relationship(), allValidKinds...),
				expected: graph.NewPathSet(graph.Path{
					Nodes: []*graph.Node{
						{
							ID:    5,
							Kinds: graph.Kinds{graph.StringKind("Base"), graph.StringKind("User")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "USER NUMBER FOUR",
								"objectid": "5",
							}),
						},
						{
							ID:    10,
							Kinds: graph.Kinds{graph.StringKind("ADLocalGroup"), graph.StringKind("Base")},
							Properties: graph.AsProperties(map[string]any{
								"name":     "REMOTE DESKTOP USERS",
								"objectid": "10",
							}),
						},
					},
					Edges: []*graph.Relationship{
						{
							ID:         5,
							StartID:    5,
							EndID:      10,
							Kind:       graph.StringKind("Contains"),
							Properties: graph.NewPropertiesRed(),
						},
					},
				}),
			},
			{
				name:        "Find shortest path between OpenGraph and AD Node",
				startNodeID: "7",
				endNodeID:   "10",
				filter:      query.KindIn(query.Relationship(), allValidKinds...),
				expected: graph.NewPathSet(graph.Path{
					Nodes: []*graph.Node{
						{
							ID:    7,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "PERSON ONE",
								"objectid": "7",
							}),
						},
						{
							ID:    10,
							Kinds: graph.Kinds{graph.StringKind("ADLocalGroup"), graph.StringKind("Base")},
							Properties: graph.AsProperties(map[string]any{
								"name":     "REMOTE DESKTOP USERS",
								"objectid": "10",
							}),
						},
					},
					Edges: []*graph.Relationship{
						{
							ID:         7,
							StartID:    7,
							EndID:      10,
							Kind:       graph.StringKind("Contains"),
							Properties: graph.NewPropertiesRed(),
						},
					},
				}),
			},
			{
				name:        "Find shortest path between nodes, filter out specific kinds",
				startNodeID: "7",
				endNodeID:   "12",
				filter:      query.KindIn(query.Relationship(), allValidKinds.Exclude(graph.StringsToKinds([]string{"IsParent"}))...),
				expected: graph.NewPathSet(graph.Path{
					Nodes: []*graph.Node{
						{
							ID:    7,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "PERSON ONE",
								"objectid": "7",
							}),
						},
						{
							ID:    8,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "PERSON TWO",
								"objectid": "8",
							}),
						},
						{
							ID:    9,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "PERSON THREE",
								"objectid": "9",
							}),
						},
						{
							ID:    12,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"name":     "ALICE",
								"objectid": "12",
							}),
						},
					},
					Edges: []*graph.Relationship{

						{
							ID:         1,
							StartID:    7,
							EndID:      8,
							Kind:       graph.StringKind("Knows"),
							Properties: graph.NewPropertiesRed(),
						},
						{
							ID:         2,
							StartID:    8,
							EndID:      9,
							Kind:       graph.StringKind("Knows"),
							Properties: graph.NewPropertiesRed(),
						},
						{
							ID:         3,
							StartID:    9,
							EndID:      12,
							Kind:       graph.StringKind("Knows"),
							Properties: graph.NewPropertiesRed(),
						},
					},
				}),
			},
			{
				name:        "Find shortest path between nodes, only include specific kinds",
				startNodeID: "7",
				endNodeID:   "12",
				filter:      query.KindIn(query.Relationship(), graph.StringKind("Knows")),
				expected: graph.NewPathSet(graph.Path{
					Nodes: []*graph.Node{
						{
							ID:    7,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "PERSON ONE",
								"objectid": "7",
							}),
						},
						{
							ID:    8,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "PERSON TWO",
								"objectid": "8",
							}),
						},
						{
							ID:    9,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"hello":    "world",
								"name":     "PERSON THREE",
								"objectid": "9",
							}),
						},
						{
							ID:    12,
							Kinds: graph.Kinds{graph.StringKind("Person")},
							Properties: graph.AsProperties(map[string]any{
								"name":     "ALICE",
								"objectid": "12",
							}),
						},
					},
					Edges: []*graph.Relationship{

						{
							ID:         1,
							StartID:    7,
							EndID:      8,
							Kind:       graph.StringKind("Knows"),
							Properties: graph.NewPropertiesRed(),
						},
						{
							ID:         2,
							StartID:    8,
							EndID:      9,
							Kind:       graph.StringKind("Knows"),
							Properties: graph.NewPropertiesRed(),
						},
						{
							ID:         3,
							StartID:    9,
							EndID:      12,
							Kind:       graph.StringKind("Knows"),
							Properties: graph.NewPropertiesRed(),
						},
					},
				}),
			},
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := graphQuery.GetAllShortestPathsWithOpenGraph(testSuite.Context, testCase.startNodeID, testCase.endNodeID, testCase.filter)
			require.Nil(t, err)
			require.Equal(t, testCase.expected, actual)
		})
	}
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

func TestRawCypherQuery(t *testing.T) {
	var (
		testSuite  = setupGraphDb(t)
		graphQuery = queries.NewGraphQuery(testSuite.GraphDB, cache.Cache{}, config.Configuration{})
	)
	defer teardownIntegrationTestSuite(t, &testSuite)

	t.Run("Test return nodes", func(t *testing.T) {
		preparedQuery, err := graphQuery.PrepareCypherQuery("match (n:User) return n", queries.DefaultQueryFitnessLowerBoundExplore)
		require.Nil(t, err)

		results, err := graphQuery.RawCypherQuery(context.Background(), preparedQuery, false)
		require.Nil(t, err)
		require.Equal(t, 5, len(results.Nodes))
	})

	t.Run("Test return edges", func(t *testing.T) {
		preparedQuery, err := graphQuery.PrepareCypherQuery("match p = (m:Person)-[:Knows]->() return p", queries.DefaultQueryFitnessLowerBoundExplore)
		require.Nil(t, err)

		results, err := graphQuery.RawCypherQuery(context.Background(), preparedQuery, false)
		require.Nil(t, err)
		require.Equal(t, 3, len(results.Edges))
	})

	t.Run("Test return properties", func(t *testing.T) {
		preparedQuery, err := graphQuery.PrepareCypherQuery("match (m) where m.name = 'ALICE' return m", queries.DefaultQueryFitnessLowerBoundExplore)
		require.Nil(t, err)

		results, err := graphQuery.RawCypherQuery(context.Background(), preparedQuery, true)
		require.Nil(t, err)
		require.Equal(t, 1, len(results.Nodes))
		require.Equal(t, "ALICE", results.Nodes["12"].Properties["name"])
	})

	t.Run("Test return literals", func(t *testing.T) {
		preparedQuery, err := graphQuery.PrepareCypherQuery("match (m) where m.name = 'ALICE' return 7-6 = 1, count(m), max(m.name)", queries.DefaultQueryFitnessLowerBoundExplore)
		require.Nil(t, err)

		results, err := graphQuery.RawCypherQuery(context.Background(), preparedQuery, false)
		require.Nil(t, err)
		require.Equal(t, 3, len(results.Literals))
		require.Equal(t, true, results.Literals[0].Value)
		require.Equal(t, int64(1), results.Literals[1].Value)
		require.Equal(t, "ALICE", results.Literals[2].Value)

	})

	t.Run("Test return combination", func(t *testing.T) {
		preparedQuery, err := graphQuery.PrepareCypherQuery("match (m:User) return m, count(m)", queries.DefaultQueryFitnessLowerBoundExplore)
		require.Nil(t, err)

		results, err := graphQuery.RawCypherQuery(context.Background(), preparedQuery, false)
		require.Nil(t, err)
		require.Equal(t, 5, len(results.Nodes))
		require.Equal(t, 5, len(results.Literals))
		require.Equal(t, int64(1), results.Literals[0].Value)
	})
}
