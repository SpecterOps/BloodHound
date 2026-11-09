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

package v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	databaseMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	graphMocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandleAdRelatedEntityQuery_CacheIsolation(t *testing.T) {
	var testCases = []struct {
		name            string
		cacheEnabled    bool
		expectedFetches int
	}{
		{name: "cache disabled", expectedFetches: 3},
		{name: "cache miss then shared hits", cacheEnabled: true, expectedFetches: 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				mockController  = gomock.NewController(t)
				mockDatabase    = databaseMocks.NewMockDatabase(mockController)
				mockGraph       = graphMocks.NewMockDatabase(mockController)
				mockTransaction = graphMocks.NewMockTransaction(mockController)
				mockNodeQuery   = graphMocks.NewMockNodeQuery(mockController)
				anchor          = graph.NewNode(100, graph.AsProperties(map[string]any{
					common.ObjectID.String(): "anchor",
					ad.DomainSID.String():    "anchor-domain",
				}), ad.Entity, ad.User)
				rawNodes = graph.NewNodeSet(
					graph.NewNode(3, graph.AsProperties(map[string]any{
						common.ObjectID.String(): "object-a",
						common.Name.String():     "Object A",
						ad.DomainSID.String():    "domain-a",
					}), ad.Entity, ad.Computer),
					graph.NewNode(2, graph.AsProperties(map[string]any{
						common.ObjectID.String(): "object-b",
						common.Name.String():     "Object B",
						ad.DomainSID.String():    "domain-b",
					}), ad.Entity, ad.Computer),
				)
				fetchCount int
			)

			cacheInstance, err := cache.NewCache(cache.Config{MaxSize: 10})
			require.NoError(t, err)

			var (
				resources = Resources{
					DB: mockDatabase,
					GraphQuery: &queries.GraphQuery{
						Graph:              mockGraph,
						Cache:              cacheInstance,
						SlowQueryThreshold: -1,
					},
					DogTags: dogtags.NewTestService(dogtags.TestOverrides{
						Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: true},
					}),
				}
				listDelegate = func(_ graph.Transaction, _ *graph.Node, skip, limit int) (graph.NodeSet, error) {
					fetchCount++
					require.Zero(t, skip)
					require.Zero(t, limit)
					return rawNodes, nil
				}
			)

			mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, transactionDelegate graph.TransactionDelegate, _ ...graph.TransactionOption) error {
					return transactionDelegate(mockTransaction)
				}).Times(6 + testCase.expectedFetches)
			mockTransaction.EXPECT().Nodes().Return(mockNodeQuery).Times(6)
			mockNodeQuery.EXPECT().Filterf(gomock.Any()).Return(mockNodeQuery).Times(6)
			mockNodeQuery.EXPECT().First().Return(anchor, nil).Times(6)
			mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureEntityPanelCaching).
				Return(appcfg.FeatureFlag{Enabled: testCase.cacheEnabled}, nil).Times(3)
			mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Times(3)

			for _, requestCase := range []struct {
				domainID string
				objectID string
			}{
				{domainID: "domain-a", objectID: "object-a"},
				{domainID: "domain-b", objectID: "object-b"},
				{domainID: "domain-a", objectID: "object-a"},
			} {
				var (
					requestContext = bhctx.Context{AuthCtx: auth.Context{Owner: model.User{
						EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
							{EnvironmentID: "anchor-domain"},
							{EnvironmentID: requestCase.domainID},
						},
					}}}
					request  = mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/related?limit=1", nil).WithContext(requestContext.ConstructGoContext()), map[string]string{"object_id": "anchor"})
					response = httptest.NewRecorder()
					body     struct {
						Count int                        `json:"count"`
						Data  []model.PagedNodeListEntry `json:"data"`
					}
				)

				resources.handleAdRelatedEntityQuery(response, request, "cache-isolation", nil, listDelegate)
				require.Equal(t, http.StatusOK, response.Code, response.Body.String())
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
				require.Equal(t, 1, body.Count)
				require.Len(t, body.Data, 1)
				require.Equal(t, requestCase.objectID, body.Data[0].ObjectID)
			}

			require.Equal(t, testCase.expectedFetches, fetchCount)
			require.Equal(t, 2, rawNodes.Len(), "response filtering must preserve the shared raw result")
		})
	}
}
