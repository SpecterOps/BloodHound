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

package queries

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	graph_mocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_ApplyTimeoutReduction(t *testing.T) {
	// Query Weight			Reduction Factor 		  Runtime
	// 	0-4						1						x
	// 	5-9						2						x/2
	//	10-14					3						x/3
	//	15-19					4						x/4
	//	20-24					5						x/5
	//	25-29					6						x/6
	//	30-34					7						x/7
	//	35-39					8						x/8
	//	40-44					9						x/9
	//	45-49					10						x/10
	// 	50						11						x/11
	//	>50						Too complex

	var (
		inputRuntime      = 15 * time.Minute
		expectedReduction int64
	)

	// Start with weight of 2, increase by 5 in each iteration until reduction factor = 11
	// This will run the function and assess the results for each range of permissible query
	// weights, against their respective expected reduction factor and runtime.
	weight := int64(2)
	for expectedReduction = 1; expectedReduction < 12; expectedReduction++ {
		expectedRuntime := int64(inputRuntime.Seconds()) / expectedReduction
		reducedRuntime, reduction := applyTimeoutReduction(weight, inputRuntime)

		require.Equal(t, expectedReduction, reduction)
		require.Equal(t, expectedRuntime, int64(reducedRuntime.Seconds()))

		weight += 5
	}
}

const cacheKey = "ad-entity-query_queryName_objectID_1"

func Test_runMaybeCachedEntityQuery(t *testing.T) {
	var (
		mockCtrl         = gomock.NewController(t)
		mockDB           = graph_mocks.NewMockDatabase(mockCtrl)
		node             = graph.NewNode(0, graph.NewProperties(), graph.StringKind("kind"))
		ctx              = context.Background()
		happyQueryName   = "happyQuery"
		happyObjectID    = "happyObjectID"
		failureQueryName = "failureQuery"
		failureObjectID  = "failureObjectID"

		// Ensure we can match on the exact error from the delegate to ensure that none of the intermediate layers are
		// producing errors unexpectedly
		failureDelegateErr = errors.New("failed")
		failureDelegate    = func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
			return nil, failureDelegateErr
		}

		happyPathDelegateReturn = graph.NewNodeSet()
		happyPathDelegate       = func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
			return happyPathDelegateReturn, nil
		}
	)
	defer mockCtrl.Finish()

	cacheInstance, err := cache.NewCache(cache.Config{MaxSize: 100})
	require.Nil(t, err)
	graphQueryInst := &GraphQuery{
		Graph:              mockDB,
		Cache:              cacheInstance,
		SlowQueryThreshold: 0,
	}

	mockDB.EXPECT().ReadTransaction(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
		return txDelegate(nil)
	}).Times(2)

	t.Run("runMaybeCachedEntityQuery with delegate failure", func(t *testing.T) {
		_, err = graphQueryInst.runMaybeCachedEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     failureQueryName,
			ObjectID:      failureObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  failureDelegate,
		}, true)

		require.Equal(t, failureDelegateErr, err)
	})

	t.Run("runMaybeCachedEntityQuery happy path with no cached results", func(t *testing.T) {
		result, err := graphQueryInst.runMaybeCachedEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     happyQueryName,
			ObjectID:      happyObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  happyPathDelegate,
		}, true)

		require.Nil(t, err)
		// Result set is empty so assert on that
		require.Len(t, result, 0)
	})

	t.Run("runMaybeCachedEntityQuery happy path with cached results", func(t *testing.T) {
		key := fmt.Sprintf("ad-entity-query_%s_%s_%d", happyQueryName, happyObjectID, model.DataTypeList)
		cacheInstance.Set(key, graph.NodeSet{})

		result, err := graphQueryInst.runMaybeCachedEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     happyQueryName,
			ObjectID:      happyObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  happyPathDelegate,
		}, true)

		require.Nil(t, err)
		// Result set is empty so assert on that
		require.Len(t, result, 0)
	})
}

func Test_cacheQueryResult(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		mockDB   = graph_mocks.NewMockDatabase(mockCtrl)
		result   = graph.NodeSet{}
	)

	cacheInstance, err := cache.NewCache(cache.Config{MaxSize: 100})
	require.Nil(t, err)

	graphQuery := &GraphQuery{
		Graph:              mockDB,
		SlowQueryThreshold: time.Minute.Milliseconds(),
		Cache:              cacheInstance,
	}

	// Happy path rejection for queries that run quick enough
	graphQuery.cacheQueryResult(time.Now().Add(-time.Second), cacheKey, result)

	// Happy path setting for queries that are slow enough
	graphQuery.cacheQueryResult(time.Now().Add(-time.Hour), cacheKey, result)

	// Force test for the error case
	graphQuery.cacheQueryResult(time.Now().Add(-time.Hour), cacheKey, result)

	// Force test for when the cache key is already set
	graphQuery.cacheQueryResult(time.Now().Add(-time.Hour), cacheKey, result)
}

func Test_formatSearchResults_sorting(t *testing.T) {
	var (
		matches = NodeSearchResults{
			ExactResults: []model.SearchResult{
				{Name: "b@c.com"},
			},
			FuzzyResults: []model.SearchResult{
				{Name: "bab@c.com"},
				{Name: "ab@c.com"},
			},
		}
		skip     = 0
		limit    = 10
		expected = []model.SearchResult{
			matches.ExactResults[0], matches.FuzzyResults[1], matches.FuzzyResults[0], // manually put fuzzyMatches' elements in alphabetical order for assertion
		}
	)

	actual := formatSearchResults(matches, limit, skip)

	require.Equal(t, 3, len(actual))
	require.Equal(t, actual, expected)
}

func Test_formatSearchResults_limit(t *testing.T) {
	var (
		matches = NodeSearchResults{
			ExactResults: []model.SearchResult{
				{Name: "b@c.com"},
				{Name: "b@c.com"},
				{Name: "b@c.com"},
			},
			FuzzyResults: []model.SearchResult{
				{Name: "ab@c.com"},
			},
		}
		skip     = 0
		limit    = 3
		expected = matches.ExactResults
	)

	actual := formatSearchResults(matches, limit, skip)

	require.Equal(t, 3, len(actual))
	require.Equal(t, actual, expected)
}

func Test_filterNodesToSearchResult(t *testing.T) {
	var (
		inputNodeProps = graph.NewProperties().
				Set("name", "this is a name").
				Set("objectid", "object id").
				Set("distinguishedname", "ze most distinguished")

		input = []*graph.Node{
			{Properties: inputNodeProps},
		}
		customNodeKindsMap = model.CustomNodeKindMap{"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}
	)

	actual, err := filterNodesToSearchResult(false, nil, customNodeKindsMap, input...)
	require.Nil(t, err)

	expectedName, _ := inputNodeProps.Get("name").String()
	expectedObjectId, _ := inputNodeProps.Get("objectid").String()
	expectedDistinguishedName, _ := inputNodeProps.Get("distinguishedname").String()

	require.Equal(t, 1, len(actual))
	require.Equal(t, expectedName, "this is a name")
	require.Equal(t, expectedObjectId, "object id")
	require.Equal(t, expectedDistinguishedName, "ze most distinguished")
}

func Test_filterNodesToSearchResult_default(t *testing.T) {
	var (
		input = []*graph.Node{
			{Properties: graph.NewProperties()},
		}
		expectedName              = graphschema.DefaultMissingName
		expectedObjectId          = graphschema.DefaultMissingObjectId
		expectedDistinguishedName = ""

		customNodeKindsMap = model.CustomNodeKindMap{"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}
	)

	actual, err := filterNodesToSearchResult(false, nil, customNodeKindsMap, input...)
	require.Nil(t, err)

	require.Equal(t, 1, len(actual))
	require.Equal(t, expectedName, actual[0].Name)
	require.Equal(t, expectedObjectId, actual[0].ObjectID)
	require.Equal(t, expectedDistinguishedName, actual[0].DistinguishedName)
}

func Test_filterNodesToSearchResult_includeOpenGraphNodes(t *testing.T) {
	var (
		customKind     = "CustomKind"
		inputNodeProps = graph.NewProperties().
				Set("name", "this is a name").
				Set("objectid", "object id")
		input = []*graph.Node{
			{Kinds: []graph.Kind{graph.StringKind("OtherKind"), graph.StringKind(customKind)},
				Properties: inputNodeProps},
		}

		customNodeKindsMap = model.CustomNodeKindMap{customKind: model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}
	)

	actual, err := filterNodesToSearchResult(true, nil, customNodeKindsMap, input...)
	require.Nil(t, err)

	require.Equal(t, 1, len(actual))
	require.Equal(t, customKind, actual[0].Type)
}

func Test_filterNodesToSearchResult_filterEnvironments(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2", ad.DomainSID.String(): "54321"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3", azure.TenantID.String(): "azure12345"},
			},
		}

		input = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}

		customNodeKindsMap = model.CustomNodeKindMap{"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}
	)

	actual, err := filterNodesToSearchResult(false, []string{"54321"}, customNodeKindsMap, input...)
	require.Nil(t, err)

	expectedName, _ := inputNodeProp2.Properties.Get(common.Name.String()).String()
	expectedObjectId, _ := inputNodeProp2.Properties.Get(common.ObjectID.String()).String()

	require.Equal(t, 1, len(actual))
	actualResult := actual[0]
	require.Equal(t, expectedName, actualResult.Name)
	require.Equal(t, expectedObjectId, actualResult.ObjectID)
}

func Test_filterNodesToSearchResult_filterEnvironmentsEmpty(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2", ad.DomainSID.String(): "54321"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3", azure.TenantID.String(): "azure12345"},
			},
		}

		input = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}

		customNodeKindsMap = model.CustomNodeKindMap{"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}
	)

	actual, err := filterNodesToSearchResult(false, []string{}, customNodeKindsMap, input...)
	require.Nil(t, err)

	require.Empty(t, actual)
}

func Test_filterNodesToSearchResult_filterEnvironments_domainSIDFail(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3", azure.TenantID.String(): "azure12345"},
			},
		}

		input = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}

		customNodeKindsMap = model.CustomNodeKindMap{"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}
	)

	result, err := filterNodesToSearchResult(false, []string{"54321"}, customNodeKindsMap, input...)
	require.NoError(t, err)
	require.Len(t, result, 0)
}

func Test_filterNodesToSearchResult_filterEnvironments_tenantIDFail(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2", ad.DomainSID.String(): "54321"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3"},
			},
		}

		input = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}

		customNodeKindsMap = model.CustomNodeKindMap{"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}
	)

	result, err := filterNodesToSearchResult(false, []string{"azure12345"}, customNodeKindsMap, input...)
	require.NoError(t, err)
	require.Len(t, result, 0)
}
