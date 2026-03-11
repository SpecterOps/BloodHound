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

package v2_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	openGraphSchemaMocks "github.com/specterops/bloodhound/cmd/api/src/api/v2/mocks"
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	graphMocks "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_SearchHandler(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = graphMocks.NewMockGraph(mockCtrl)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph, DB: mockDB, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}
		userCtx   = setupUserCtx(model.User{PrincipalName: "user"})
	)

	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.SearchHandler).
		Run([]apitest.Case{
			{
				Name: "EmptySearchQueryFailure",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Invalid search parameter")
				},
			},
			{
				Name: "GetPageParamsFailure",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.AddQueryParam(input, "skip", "notAnInt")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Invalid query parameter")
				},
			},
			{
				Name: "GetFeatureFlagError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.AddQueryParam(input, "type", "invalidKind")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{}, errors.New("database error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "an internal error has occurred that is preventing the service from servicing this request")
				},
			},
			{
				Name: "ParseKindsError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.AddQueryParam(input, "type", "invalidKind")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Invalid type parameter")
				},
			},
			{
				Name: "Success -- GetCustomNodesKinds error does not cause request to fail ",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, errors.New("database error"))
					mockDB.EXPECT().GetDisplayNodeGraphKinds(gomock.Any())
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), make(model.CustomNodeKindMap), nil, graph.Kinds{ad.Entity, azure.Entity}, "search value", 0, 10).
						Return(nil, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},

			{
				Name: "Success -- Custom Node Icon is set correctly",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{
						"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}, nil)
					mockDB.EXPECT().GetDisplayNodeGraphKinds(gomock.Any())
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), model.CustomNodeKindMap{"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}, nil, graph.Kinds{}, "search value", 0, 10).
						Return([]model.SearchResult{{ObjectID: "0001", Type: "Person", Name: "TestPerson", DistinguishedName: "TestName", SystemTags: "tags"}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "Person")
				},
			},

			{
				Name: "GraphDBSearchNodesError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
					mockDB.EXPECT().GetDisplayNodeGraphKinds(gomock.Any())
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), gomock.Any(), nil, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "Graph error:")
				},
			},
			{
				Name: "GetDisplayNodeGraphKindsError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
					mockDB.EXPECT().GetDisplayNodeGraphKinds(gomock.Any()).Return(nil, errors.New("database error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "an internal error has occurred that is preventing the service from servicing this request")
				},
			},
			{
				Name: "Success -- OpenGraphSearch Feature Flag On",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
					mockDB.EXPECT().GetDisplayNodeGraphKinds(gomock.Any())
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), make(model.CustomNodeKindMap), nil, graph.Kinds{}, "search value", 0, 10).
						Return(nil, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},

			{
				Name: "Success -- OpenGraphSearch Feature Flag Off",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
					mockDB.EXPECT().GetDisplayNodeGraphKinds(gomock.Any())
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), make(model.CustomNodeKindMap), nil, graph.Kinds{ad.Entity, azure.Entity}, "search value", 0, 10).
						Return(nil, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_SearchHandler_ETAC(t *testing.T) {
	tests := []struct {
		name               string
		queryParams        map[string]string
		expectedMocks      func(mockDB *dbMocks.MockDatabase, mockGraph *graphMocks.MockGraph)
		expectedStatusCode int
		assertBody         func(t *testing.T, body string)
		dogTagsOverrides   dogtags.TestOverrides
		user               model.User
	}{
		{
			name: "Success -- ETAC Feature Flag On",
			user: model.User{
				PrincipalName: "etac user",
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "12345"},
					{EnvironmentID: "54321"},
				},
			},
			queryParams: map[string]string{
				"q": "search value",
			},
			expectedMocks: func(mockDB *dbMocks.MockDatabase, mockGraph *graphMocks.MockGraph) {
				mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
				mockDB.EXPECT().GetDisplayNodeGraphKinds(gomock.Any())
				mockGraph.EXPECT().
					SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), make(model.CustomNodeKindMap), []string{"12345", "54321"}, graph.Kinds{ad.Entity, azure.Entity}, "search value", 0, 10).
					Return(nil, nil)
			},
			expectedStatusCode: 200,
			assertBody: func(t *testing.T, body string) {

			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
		},
		{
			name: "Success -- ETAC Feature Flag On User all_environments = true",
			user: model.User{
				PrincipalName:   "etac user",
				AllEnvironments: true,
			},
			queryParams: map[string]string{
				"q": "search value",
			},
			expectedMocks: func(mockDB *dbMocks.MockDatabase, mockGraph *graphMocks.MockGraph) {
				mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mockDB.EXPECT().GetCustomNodeKindsMap(gomock.Any()).Return(model.CustomNodeKindMap{}, nil)
				mockDB.EXPECT().GetDisplayNodeGraphKinds(gomock.Any())
				mockGraph.EXPECT().
					SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), make(model.CustomNodeKindMap), nil, graph.Kinds{ad.Entity, azure.Entity}, "search value", 0, 10).
					Return(nil, nil)
			},
			expectedStatusCode: 200,
			assertBody: func(t *testing.T, body string) {

			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
		},
		{
			name: "Fail -- ETAC Feature Flag On, No User",
			user: model.User{},
			queryParams: map[string]string{
				"q": "search value",
			},
			expectedMocks: func(mockDB *dbMocks.MockDatabase, mockGraph *graphMocks.MockGraph) {
			},
			expectedStatusCode: http.StatusBadRequest,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "no associated user found with request")
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
		},
	}

	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			var (
				mockCtrl       = gomock.NewController(tt)
				mockDB         = dbMocks.NewMockDatabase(mockCtrl)
				mockGraph      = graphMocks.NewMockGraph(mockCtrl)
				dogTagsService = dogtags.NewTestService(tc.dogTagsOverrides)
				resources      = v2.Resources{GraphQuery: mockGraph, DB: mockDB, DogTags: dogTagsService}
				endpoint       = "/api/v2/search"
			)
			defer mockCtrl.Finish()

			ctx := context.Background()
			if tc.user.PrincipalName != "" {
				ctx = setupUserCtx(tc.user)
			}

			tc.expectedMocks(mockDB, mockGraph)

			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			require.NoError(t, err)

			queryParams := req.URL.Query()
			for key, value := range tc.queryParams {
				queryParams.Set(key, value)
			}
			req.URL.RawQuery = queryParams.Encode()

			router := mux.NewRouter()
			router.HandleFunc(endpoint, resources.SearchHandler).Methods(http.MethodGet)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(tt, tc.expectedStatusCode, rr.Code)
			tc.assertBody(tt, rr.Body.String())
		})
	}
}

func TestResources_ListAvailableEnvironments(t *testing.T) {

	t.Parallel()

	type mock struct {
		mockDatabase               *dbMocks.MockDatabase
		mockGraphQuery             *graphMocks.MockGraph
		mockOpenGraphSchemaService *openGraphSchemaMocks.MockOpenGraphSchemaService
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}

	tests := []struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}{
		{
			name: "fail - invalid query parameter",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/available-domains",
						RawQuery: "sort_by=invalid",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"column format does not support sorting"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "fail - Get OpenGraph Findings feature flag error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/available-domains",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{}, fmt.Errorf("Some error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "fail - GetEnvironmentKindsAndEnvironmentExtensionDisplayNames error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/available-domains",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mock.mockOpenGraphSchemaService.EXPECT().GetEnvironmentKindsAndEnvironmentExtensionDisplayNames(gomock.Any(), true).Return(graph.Kinds{}, map[string]string{}, fmt.Errorf("Some error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "fail - GraphQueryError",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/available-domains",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mock.mockOpenGraphSchemaService.EXPECT().GetEnvironmentKindsAndEnvironmentExtensionDisplayNames(gomock.Any(), true).Return(graph.Kinds{
					ad.Domain, azure.Tenant,
				}, map[string]string{
					ad.Domain.String():    "Active Directory",
					azure.Tenant.String(): "Azure",
				}, nil)
				mock.mockGraphQuery.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return([]*graph.Node{}, fmt.Errorf("Some error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"Some error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "success - empty response",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/available-domains",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mock.mockOpenGraphSchemaService.EXPECT().GetEnvironmentKindsAndEnvironmentExtensionDisplayNames(gomock.Any(), true).Return(graph.Kinds{}, map[string]string{}, nil)
				mock.mockGraphQuery.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return([]*graph.Node{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":[]}`,
			},
		},
		{
			name: "success - list built-in environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/available-domains",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mock.mockOpenGraphSchemaService.EXPECT().GetEnvironmentKindsAndEnvironmentExtensionDisplayNames(gomock.Any(), true).Return(graph.Kinds{ad.Domain}, map[string]string{ad.Domain.String(): "Active Directory"}, nil)
				mock.mockGraphQuery.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return([]*graph.Node{
					{
						Properties: graph.AsProperties(map[string]any{
							common.Name.String():      "Domain1",
							common.ObjectID.String():  "1",
							common.Collected.String(): false,
						}),
						Kinds: graph.Kinds{ad.Domain},
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":[{"type":"active-directory","name":"Domain1","id":"1","collected":false}]}`,
			},
		},
		{
			name: "success - list OpenGraph environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/available-domains",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockOpenGraphSchemaService.EXPECT().GetEnvironmentKindsAndEnvironmentExtensionDisplayNames(gomock.Any(), false).Return(graph.Kinds{graph.StringKind("HeeHaw Kind")}, map[string]string{graph.StringKind("HeeHaw Kind").String(): "HeeHaw"}, nil)
				mock.mockGraphQuery.EXPECT().
					GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).
					Return([]*graph.Node{
						{
							Properties: graph.AsProperties(map[string]any{
								common.Name.String():      "HeeHaw Name",
								common.ObjectID.String():  "1",
								common.Collected.String(): true,
							}),
							Kinds: graph.Kinds{graph.StringKind("HeeHaw Kind")},
						},
					}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":[{"type":"HeeHaw","name":"HeeHaw Name","id":"1","collected":true}]}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase:               dbMocks.NewMockDatabase(ctrl),
				mockGraphQuery:             graphMocks.NewMockGraph(ctrl),
				mockOpenGraphSchemaService: openGraphSchemaMocks.NewMockOpenGraphSchemaService(ctrl),
			}

			request := tt.buildRequest()
			tt.setupMocks(t, mocks)

			resources := v2.Resources{
				DB:                     mocks.mockDatabase,
				GraphQuery:             mocks.mockGraphQuery,
				OpenGraphSchemaService: mocks.mockOpenGraphSchemaService,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/available-domains", resources.ListAvailableEnvironments).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, tt.expected.responseCode, status)
			assert.Equal(t, tt.expected.responseHeader, header)
			assert.JSONEq(t, tt.expected.responseBody, body)
		})
	}
}
