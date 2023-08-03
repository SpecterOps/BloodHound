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

package queries_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	"github.com/specterops/bloodhound/cache"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"github.com/specterops/bloodhound/dawgs/graph"
	graphMocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func TestQueries_GetEntityObjectIDFromRequestPath(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v2/users/S-1-5-21-570004220-2248230615-4072641716-4001/admin-rights", nil)
	require.Nil(t, err)

	_, err = queries.GetEntityObjectIDFromRequestPath(req)
	require.Equal(t, "no object ID found in request", err.Error())

	expectedObjectID := "1"
	req = mux.SetURLVars(req, map[string]string{"object_id": expectedObjectID})

	objectID, err := queries.GetEntityObjectIDFromRequestPath(req)
	require.Nil(t, err)
	require.Equal(t, expectedObjectID, objectID)
}

func TestQueries_GetRequestedType(t *testing.T) {
	graph, list, count := url.Values{}, url.Values{}, url.Values{}
	graph.Add("type", "graph")
	list.Add("type", "list")
	count.Add("type", "somethingElse")

	require.Equal(t, model.DataTypeGraph, queries.GetRequestedType(graph))
	require.Equal(t, model.DataTypeList, queries.GetRequestedType(list))
	require.Equal(t, model.DataTypeCount, queries.GetRequestedType(count))
}

func TestQueries_BuildEntityQueryParams_MissingObjectID(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v2/users/S-1-5-21-570004220-2248230615-4072641716-4001/admin-rights", nil)
	require.Nil(t, err)

	_, err = queries.BuildEntityQueryParams(req, "", nil, nil)
	require.Contains(t, err.Error(), "error getting objectid")
}

func TestQueries_BuildEntityQueryParams_InvalidSkip(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v2/users/S-1-5-21-570004220-2248230615-4072641716-4001/admin-rights", nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{"object_id": "1"})

	q := url.Values{}
	q.Add("skip", "-1")
	req.URL.RawQuery = q.Encode()

	_, err = queries.BuildEntityQueryParams(req, "", nil, nil)
	require.Contains(t, err.Error(), "invalid skip")
}

func TestQueries_BuildEntityQueryParams_InvalidLimit(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v2/users/S-1-5-21-570004220-2248230615-4072641716-4001/admin-rights", nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{"object_id": "1"})

	q := url.Values{}
	q.Add("limit", "-1")
	req.URL.RawQuery = q.Encode()

	_, err = queries.BuildEntityQueryParams(req, "", nil, nil)
	require.Contains(t, err.Error(), "invalid limit")
}

func TestQueries_BuildEntityQueryParams(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v2/users/S-1-5-21-570004220-2248230615-4072641716-4001/admin-rights", nil)
	require.Nil(t, err)

	objectID := "S-1-5-21-570004220-2248230615-4072641716-4001"
	req = mux.SetURLVars(req, map[string]string{"object_id": objectID})

	q := url.Values{}
	q.Add("skip", "5")
	q.Add("limit", "120")
	req.URL.RawQuery = q.Encode()

	params, err := queries.BuildEntityQueryParams(req, "", nil, nil)
	require.Nil(t, err)
	require.Equal(t, 5, params.Skip)
	require.Equal(t, 120, params.Limit)
	require.Equal(t, objectID, params.ObjectID)
}

func TestQueries_BuildEntityQueryParams_DataTypeCount(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v2/users/S-1-5-21-570004220-2248230615-4072641716-4001/admin-rights", nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{"object_id": "1"})

	q := url.Values{}
	q.Add("type", "count")
	req.URL.RawQuery = q.Encode()

	params, err := queries.BuildEntityQueryParams(req, "", nil, nil)
	require.Nil(t, err)
	require.Equal(t, 0, params.Skip)
	require.Equal(t, 0, params.Limit)
}

func TestQueries_GetEntityResults(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = graphMocks.NewMockDatabase(mockCtrl)
		node      = graph.NewNode(100, graph.AsProperties(map[string]any{
			common.Name.String(): "foo",
		}), ad.Entity)

		params = queries.EntityQueryParameters{
			ObjectID:      "100",
			RequestedType: 1,
			Skip:          0,
			Limit:         10,
			PathDelegate:  nil,
			ListDelegate: func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
				set := make([]*graph.Node, 0)
				for i := 0; i < 20; i++ {
					set = append(set, graph.NewNode(graph.ID(i), graph.AsProperties(map[string]any{
						common.Name.String(): fmt.Sprintf("Node %d", i),
					}), ad.Entity))
				}

				return graph.NewNodeSet(set...), nil
			},
		}
	)

	defer mockCtrl.Finish()

	cacheInstance, err := cache.NewCache(cache.Config{MaxSize: 100})
	require.Nil(t, err)

	graphQuery := queries.GraphQuery{
		Graph:              mockGraph,
		Cache:              cacheInstance,
		SlowQueryThreshold: 200000, // Setting high to prevent any caching logic
	}

	mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, delegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
		return delegate(nil)
	})

	results, err := graphQuery.GetEntityResults(context.Background(), node, params, true)
	require.Nil(t, err)
	castResult, ok := results.(api.ResponseWrapper)
	require.True(t, ok)
	require.Equal(t, 0, castResult.Skip)
	require.Equal(t, 10, castResult.Limit)
	require.Len(t, castResult.Data, 10)
}
