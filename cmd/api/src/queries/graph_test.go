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
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/dawgs/graph"
	graphMocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	bhCtx "github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGraphQuery_PrepareCypherQuery(t *testing.T) {
	var (
		mockCtrl     = gomock.NewController(t)
		mockGraphDB  = graphMocks.NewMockDatabase(mockCtrl)
		gq           = queries.NewGraphQuery(mockGraphDB, cache.Cache{}, config.Configuration{EnableCypherMutations: true})
		gqMutDisable = queries.NewGraphQuery(mockGraphDB, cache.Cache{}, config.Configuration{EnableCypherMutations: false})

		rawCypherRead     = "MATCH (n:Label) return n"
		rawCypherMutation = "DETACH DELETE (n:Label)"
		rawCypherInvalid  = "derp"
	)

	t.Run("invalid cypher", func(t *testing.T) {
		_, err := gq.PrepareCypherQuery(rawCypherInvalid)
		assert.ErrorContains(t, err, "mismatched input 'derp'")
	})

	t.Run("valid cypher with mutation while mutations disabled", func(t *testing.T) {
		_, err := gqMutDisable.PrepareCypherQuery(rawCypherMutation)
		assert.ErrorContains(t, err, "not supported")
	})

	t.Run("valid cypher without mutation", func(t *testing.T) {
		preparedQuery, err := gq.PrepareCypherQuery(rawCypherRead)
		require.Nil(t, err)
		assert.Equal(t, preparedQuery.HasMutation, false)
	})

	t.Run("valid cypher with mutation", func(t *testing.T) {
		preparedQuery, err := gq.PrepareCypherQuery(rawCypherMutation)
		require.Nil(t, err)
		assert.Equal(t, preparedQuery.HasMutation, true)
	})

	t.Run("valid cypher without mutation while mutations disabled", func(t *testing.T) {
		preparedQuery, err := gq.PrepareCypherQuery(rawCypherRead)
		require.Nil(t, err)
		assert.Equal(t, preparedQuery.HasMutation, false)
	})
}

func TestGraphQuery_RawCypherQuery(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		mockGraphDB    = graphMocks.NewMockDatabase(mockCtrl)
		gq             = queries.NewGraphQuery(mockGraphDB, cache.Cache{}, config.Configuration{})
		outerBHCtxInst = &bhCtx.Context{
			StartTime: time.Now(),
			Timeout:   time.Second * 5,
			RequestID: must.NewUUIDv4().String(),
			AuthCtx:   auth.Context{},
			Host: &url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}
	)

	t.Run("RawCypherQuery user set timeouts successfully", func(t *testing.T) {

		// First validate that user set timeouts work
		mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
			innerBHCtxInst := bhCtx.Get(ctx)

			require.Equal(t, outerBHCtxInst.Timeout, innerBHCtxInst.Timeout)

			// Validate that the options are being set correctly
			if len(options) != 1 {
				t.Fatalf("Expected only one transaction option for RawCypherQuery but saw: %d", len(options))
			}

			// Create a new transaction config to capture the query timeout logic
			txConfig := &graph.TransactionConfig{}
			options[0](txConfig)

			require.Equal(t, outerBHCtxInst.Timeout, txConfig.Timeout)

			return nil
		})

		preparedQuery, err := gq.PrepareCypherQuery("match (n:Label) return n;")
		require.Nil(t, err)

		_, err = gq.RawCypherQuery(outerBHCtxInst.ConstructGoContext(), preparedQuery, false)
		require.Nil(t, err)

	})

	t.Run("RawCypherQuery query complexity controls", func(t *testing.T) {
		// Validate that query complexity controls are working
		// Scenario 1:
		mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
			// Validate that the options are being set correctly
			if len(options) != 1 {
				t.Fatalf("Expected only one transaction option for RawCypherQuery but saw: %d", len(options))
			}

			// Create a new transaction config to capture the query timeout logic
			txConfig := &graph.TransactionConfig{}
			options[0](txConfig)

			require.Equal(t, time.Second*225, txConfig.Timeout)

			return nil
		}).Times(1)

		// Scenario 2:
		mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
			// Validate that the options are being set correctly
			if len(options) != 1 {
				t.Fatalf("Expected only one transaction option for RawCypherSearch but saw: %d", len(options))
			}

			// Create a new transaction config to capture the query timeout logic
			txConfig := &graph.TransactionConfig{}
			options[0](txConfig)

			require.Equal(t, time.Second*5, txConfig.Timeout)

			return nil
		}).Times(1)

		// Scenario 1:
		// Unset the user-set timeout in the BH context to validate QC runtime reduction of a complex query
		// This will be set to a default of 15 min or 900 sec
		// availableRuntime = 900 sec, query cost = 15
		// reductionFactor = 1 + (15/5) = 4
		// Therefore actual timeout = availableRuntime/reductionFactor : 900/4 = 225sec

		outerBHCtxInst.Timeout = 0
		preparedQuery, err := gq.PrepareCypherQuery("match ()-[:HasSession*..]->()-[:MemberOf*..]->() return n;")
		require.Nil(t, err)
		_, err = gq.RawCypherQuery(outerBHCtxInst.ConstructGoContext(), preparedQuery, false)
		require.Nil(t, err)

		// Scenario 2:
		// Prove that overriding QC with a user-preference works
		// This will be directly used as the config timeout, without any reduction factor
		outerBHCtxInst.Timeout = time.Second * 5

		preparedQuery, err = gq.PrepareCypherQuery("match ()-[:HasSession*..]->()-[:MemberOf*..]->() return n;")
		require.Nil(t, err)

		_, err = gq.RawCypherQuery(outerBHCtxInst.ConstructGoContext(), preparedQuery, false)
		require.Nil(t, err)
	})

	t.Run("RawCypherQuery read query leverages read tx", func(t *testing.T) {
		mockGraphDB.EXPECT().WriteTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

		preparedQuery, err := gq.PrepareCypherQuery("match (b) where b.name = 'harley' return b;")
		require.Nil(t, err)

		_, err = gq.RawCypherQuery(outerBHCtxInst.ConstructGoContext(), preparedQuery, false)
		require.Nil(t, err)
	})

	t.Run("RawCypherQuery mutation query leverages write tx if enable_cypher_mutations is true", func(t *testing.T) {
		mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockGraphDB.EXPECT().WriteTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

		qgWMut := queries.NewGraphQuery(mockGraphDB, cache.Cache{}, config.Configuration{EnableCypherMutations: true})
		preparedQuery, err := qgWMut.PrepareCypherQuery("match (b) where b.name = 'bruce' remove b.prop return b;")
		require.Nil(t, err)

		_, err = qgWMut.RawCypherQuery(outerBHCtxInst.ConstructGoContext(), preparedQuery, false)
		require.Nil(t, err)
	})
}

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
	graphQuery, listQuery, countQuery := url.Values{}, url.Values{}, url.Values{}
	graphQuery.Add("type", "graph")
	listQuery.Add("type", "list")
	countQuery.Add("type", "somethingElse")

	require.Equal(t, model.DataTypeGraph, queries.GetRequestedType(graphQuery))
	require.Equal(t, model.DataTypeList, queries.GetRequestedType(listQuery))
	require.Equal(t, model.DataTypeCount, queries.GetRequestedType(countQuery))
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

	results, count, err := graphQuery.GetEntityResults(context.Background(), node, params, true)
	require.Nil(t, err)
	require.Len(t, results, 10)
	require.Equal(t, count, 20)
}
