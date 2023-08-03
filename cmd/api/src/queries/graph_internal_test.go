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
	"fmt"
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/cache"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/errors"
)

const cacheKey = "ad-entity-query_queryName_objectID_1"

func Test_getOrRunListQuery(t *testing.T) {
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

	t.Run("getOrRunListQuery with delegate failure", func(t *testing.T) {
		_, err = graphQueryInst.runListEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     failureQueryName,
			ObjectID:      failureObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  failureDelegate,
		}, true)

		require.Equal(t, failureDelegateErr, err)
	})

	t.Run("getOrRunListQuery happy path with no cached results", func(t *testing.T) {
		result, err := graphQueryInst.runListEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     happyQueryName,
			ObjectID:      happyObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  happyPathDelegate,
		}, true)

		require.Nil(t, err)

		// Assert on result type first
		typedResult, typeOK := result.(api.ResponseWrapper)
		require.True(t, typeOK)

		// Result set is empty so assert on that
		require.Equal(t, 0, typedResult.Count)
		require.Equal(t, 0, typedResult.Skip)
		require.Equal(t, 0, typedResult.Limit)
	})

	t.Run("getOrRunListQuery happy path with cached results", func(t *testing.T) {
		key := fmt.Sprintf("ad-entity-query_%s_%s_%d", happyQueryName, happyObjectID, model.DataTypeList)
		cacheInstance.Set(key, graph.NodeSet{})

		result, err := graphQueryInst.runListEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     happyQueryName,
			ObjectID:      happyObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  happyPathDelegate,
		}, true)

		require.Nil(t, err)

		// Assert on result type first
		typedResult, typeOK := result.(api.ResponseWrapper)
		require.True(t, typeOK)

		// Result set is empty so assert on that
		require.Equal(t, 0, typedResult.Count)
		require.Equal(t, 0, typedResult.Skip)
		require.Equal(t, 0, typedResult.Limit)
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
		exactMatches = []model.SearchResult{
			{Name: "b@c.com"},
		}
		fuzzyMatches = []model.SearchResult{
			{Name: "bab@c.com"},
			{Name: "ab@c.com"},
		}
		skip     = 0
		limit    = 10
		expected = []model.SearchResult{
			exactMatches[0], fuzzyMatches[1], fuzzyMatches[0], // manually put fuzzyMatches' elements in alphabetical order for assertion
		}
	)

	actual := formatSearchResults(exactMatches, fuzzyMatches, limit, skip)

	require.Equal(t, 3, len(actual))
	require.Equal(t, actual, expected)
}

func Test_formatSearchResults_limit(t *testing.T) {
	var (
		exactMatches = []model.SearchResult{
			{Name: "b@c.com"},
			{Name: "b@c.com"},
			{Name: "b@c.com"},
		}
		fuzzyMatches = []model.SearchResult{
			{Name: "ab@c.com"},
		}
		skip     = 0
		limit    = 3
		expected = exactMatches
	)

	actual := formatSearchResults(exactMatches, fuzzyMatches, limit, skip)

	require.Equal(t, 3, len(actual))
	require.Equal(t, actual, expected)
}

func Test_nodesToSearchResult(t *testing.T) {
	var (
		inputNodeProps = graph.NewProperties().
				Set("name", "this is a name").
				Set("objectid", "object id").
				Set("distinguishedname", "ze most distinguished")

		input = []*graph.Node{
			{Properties: inputNodeProps},
		}
	)

	actual := nodesToSearchResult(input...)

	expectedName, _ := inputNodeProps.Get("name").String()
	expectedObjectId, _ := inputNodeProps.Get("objectid").String()
	expectedDistinguishedName, _ := inputNodeProps.Get("distinguishedname").String()

	require.Equal(t, 1, len(actual))
	require.Equal(t, expectedName, "this is a name")
	require.Equal(t, expectedObjectId, "object id")
	require.Equal(t, expectedDistinguishedName, "ze most distinguished")
}

func Test_nodesToSearchResult_default(t *testing.T) {
	var (
		input = []*graph.Node{
			{Properties: graph.NewProperties()},
		}
		expectedName              = "NO NAME"
		expectedObjectId          = "NO OBJECT ID"
		expectedDistinguishedName = ""
	)

	actual := nodesToSearchResult(input...)

	require.Equal(t, 1, len(actual))
	require.Equal(t, expectedName, actual[0].Name)
	require.Equal(t, expectedObjectId, actual[0].ObjectID)
	require.Equal(t, expectedDistinguishedName, actual[0].DistinguishedName)
}
