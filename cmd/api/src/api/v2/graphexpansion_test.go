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

package v2_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	querymocks "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_ExpandGraph(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockGraphQuery := querymocks.NewMockGraph(ctrl)
	mockDatabase := dbmocks.NewMockDatabase(ctrl)
	primaryDisplayKinds := graphschema.PrimaryDisplayKinds{graph.StringKind("KindA"): graphschema.DisplayKind{}}

	nodeID := int64(42)
	payload := v2.GraphExpansionPayload{
		NodeID:            &nodeID,
		Direction:         "outbound",
		Limit:             1,
		IncludeProperties: true,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
	}

	req := &http.Request{
		URL: &url.URL{
			Path: "/api/v2/graphs/expand",
		},
		Body: io.NopCloser(bytes.NewReader(jsonPayload)),
		Header: http.Header{
			headers.ContentType.String(): []string{"application/json"},
		},
		Method: http.MethodPost,
	}
	req = req.WithContext(setupUserCtx(model.User{AllEnvironments: true}))

	mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).Return(appcfg.FeatureFlag{Enabled: false}, nil)
	mockGraphQuery.EXPECT().PrepareCypherQuery(gomock.Any(), int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
		StrippedQuery: "query",
		HasMutation:   false,
	}, nil)
	mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(primaryDisplayKinds, nil)
	mockGraphQuery.EXPECT().RawCypherQuery(gomock.Any(), gomock.Eq(primaryDisplayKinds), gomock.Any(), true).Return(model.UnifiedGraph{
		Nodes: map[string]model.UnifiedNode{
			"1": {
				Label:      "one",
				Properties: map[string]any{"node_key": "value"},
			},
			"2": {
				Label:      "two",
				Properties: map[string]any{"node_key": "value"},
			},
			"3": {
				Label:      "three",
				Properties: map[string]any{"unused": "value"},
			},
		},
		Edges: []model.UnifiedEdge{
			{Source: "1", Target: "2", Properties: map[string]any{"edge_key": "value"}},
			{Source: "2", Target: "3", Properties: map[string]any{"unused": "value"}},
		},
		Literals: graph.Literals{},
	}, nil)

	resources := v2.Resources{
		GraphQuery: mockGraphQuery,
		DB:         mockDatabase,
		Authorizer: auth.NewAuthorizer(mockDatabase),
		DogTags:    dogtags.NewTestService(dogtags.TestOverrides{}),
	}

	response := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/v2/graphs/expand", resources.ExpandGraph).Methods(req.Method)
	router.ServeHTTP(response, req)

	status, header, body := test.ProcessResponse(t, response)

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, header)
	assert.JSONEq(t, `{"data":{"node_keys":["node_key"],"edge_keys":["edge_key"],"nodes":{"1":{"label":"one","kind":"","kinds":null,"objectId":"","isTierZero":false,"isOwnedObject":false,"lastSeen":"0001-01-01T00:00:00Z","properties":{"node_key":"value"}},"2":{"label":"two","kind":"","kinds":null,"objectId":"","isTierZero":false,"isOwnedObject":false,"lastSeen":"0001-01-01T00:00:00Z","properties":{"node_key":"value"}}},"edges":[{"id":"","source":"1","target":"2","label":"","kind":"","lastSeen":"0001-01-01T00:00:00Z","properties":{"edge_key":"value"}}],"literals":[],"limit":1,"truncated":true}}`, body)
}

func TestResources_ExpandGraphRequiresNodeID(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"direction": "outbound",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
	}

	req := &http.Request{
		URL: &url.URL{
			Path: "/api/v2/graphs/expand",
		},
		Body: io.NopCloser(bytes.NewReader(jsonPayload)),
		Header: http.Header{
			headers.ContentType.String(): []string{"application/json"},
		},
		Method: http.MethodPost,
	}
	req = req.WithContext(setupUserCtx(model.User{AllEnvironments: true}))

	resources := v2.Resources{
		DogTags: dogtags.NewTestService(dogtags.TestOverrides{}),
	}

	response := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/v2/graphs/expand", resources.ExpandGraph).Methods(req.Method)
	router.ServeHTTP(response, req)

	status, _, body := test.ProcessResponse(t, response)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "node_id is required")
}
