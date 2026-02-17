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
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	graphMocks "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
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
		mockDB    = mocks.NewMockDatabase(mockCtrl)
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, errors.New("database error"))
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), graph.Kinds{ad.Entity, azure.Entity}, "search value", false, 0, 10, nil, make(model.CustomNodeKindMap)).
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{
						{ID: 1, KindName: "Person", Config: model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}}, nil)
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), graph.Kinds{}, "search value", true, 0, 10, nil, model.CustomNodeKindMap{"Person": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}).
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), nil, gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "Graph error:")
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), graph.Kinds{}, "search value", true, 0, 10, nil, make(model.CustomNodeKindMap)).
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
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), graph.Kinds{ad.Entity, azure.Entity}, "search value", false, 0, 10, nil, make(model.CustomNodeKindMap)).
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
		expectedMocks      func(mockDB *mocks.MockDatabase, mockGraph *graphMocks.MockGraph)
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
			expectedMocks: func(mockDB *mocks.MockDatabase, mockGraph *graphMocks.MockGraph) {
				mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
				mockGraph.EXPECT().
					SearchNodesByNameOrObjectId(gomock.Any(), graph.Kinds{ad.Entity, azure.Entity}, "search value", false, 0, 10, []string{"12345", "54321"}, make(model.CustomNodeKindMap)).
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
			expectedMocks: func(mockDB *mocks.MockDatabase, mockGraph *graphMocks.MockGraph) {
				mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
				mockGraph.EXPECT().
					SearchNodesByNameOrObjectId(gomock.Any(), graph.Kinds{ad.Entity, azure.Entity}, "search value", false, 0, 10, nil, make(model.CustomNodeKindMap)).
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
			expectedMocks: func(mockDB *mocks.MockDatabase, mockGraph *graphMocks.MockGraph) {
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
				mockDB         = mocks.NewMockDatabase(mockCtrl)
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
	var (
		mockCtrl         = gomock.NewController(t)
		mockGraphQueries = graphMocks.NewMockGraph(mockCtrl)
		mockDB           = mocks.NewMockDatabase(mockCtrl)
		resources        = v2.Resources{GraphQuery: mockGraphQueries, DB: mockDB}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListAvailableEnvironments).
		Run([]apitest.Case{
			{
				Name: "GraphQueryError",
				Setup: func() {
					mockDB.EXPECT().GetEnvironments(gomock.Any()).Return([]model.SchemaEnvironment{}, nil)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockGraphQueries.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return([]*graph.Node{}, fmt.Errorf("Some error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success: Empty response",
				Setup: func() {
					mockDB.EXPECT().GetEnvironments(gomock.Any()).Return([]model.SchemaEnvironment{}, nil)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: false}, nil)

					mockGraphQueries.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return([]*graph.Node{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "[]")
				},
			},
			{
				Name: "Success: Built-in AD environment",
				Setup: func() {
					mockDB.EXPECT().
						GetEnvironments(gomock.Any()).
						Return([]model.SchemaEnvironment{
							{
								SchemaExtensionDisplayName: "Active Directory",
								EnvironmentKindName:        "Domain",
							},
						}, nil)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockGraphQueries.EXPECT().
						GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).
						Return([]*graph.Node{
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
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "{\"data\":[{\"type\":\"active-directory\",\"name\":\"Domain1\",\"id\":\"1\",\"collected\":false}]}")
				},
			},
			{
				Name: "Success: OpenGraph rando environment",
				Setup: func() {
					mockDB.EXPECT().
						GetEnvironments(gomock.Any()).
						Return([]model.SchemaEnvironment{
							{
								SchemaExtensionDisplayName: "Rando",
								EnvironmentKindName:        "HeeHaw Kind",
							},
						}, nil)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphFindings).Return(appcfg.FeatureFlag{Enabled: false}, nil)

					mockGraphQueries.EXPECT().
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
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "{\"data\":[{\"type\":\"Rando\",\"name\":\"HeeHaw Name\",\"id\":\"1\",\"collected\":true}]}")
				},
			},
		})
}
