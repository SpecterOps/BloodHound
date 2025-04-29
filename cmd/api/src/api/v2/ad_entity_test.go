// Copyright 2024 Specter Ops, Inc.
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
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_GetComputerEntityInfo(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetComputerEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "error reading objectid:")
				},
			},
			{
				Name: "InvalidCountsParam",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsBadQueryParameterFilters)
				},
			},
			{
				Name: "GraphDBNotFoundError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, graph.ErrNoResultsFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, "node not found")
				},
			},
			{
				Name: "GraphDBGetEntityByObjectIdError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "error getting node:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, nil)
					mockGraph.EXPECT().
						GetEntityCountResults(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessWithoutCounts",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "false")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetDomainEntityInfo(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetDomainEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "error reading objectid:")
				},
			},
			{
				Name: "InvalidCountsParam",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsBadQueryParameterFilters)
				},
			},
			{
				Name: "GraphDBNotFoundError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, graph.ErrNoResultsFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, "node not found")
				},
			},
			{
				Name: "GraphDBGetEntityByObjectIdError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "error getting node:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, nil)
					mockGraph.EXPECT().
						GetEntityCountResults(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessWithoutCounts",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "false")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_PatchDomain(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
		node      = graph.NewNode(graph.ID(1), graph.NewProperties())
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.PatchDomain).
		WithCommonRequest(func(input *apitest.Input) {
			apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
		}).
		Run([]apitest.Case{
			{
				Name: "RequestMarshalError",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponsePayloadUnmarshalError)
				},
			},
			{
				Name: "NoCollectionValueError",
				Input: func(input *apitest.Input) {
					apitest.BodyString(input, `{}`)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "no domain fields sent for patching")
				},
			},
			{
				Name: "MissingObjectIDError",
				Input: func(input *apitest.Input) {
					apitest.BodyString(input, `{"collected":true}`)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "error reading objectid:")
				},
			},
			{
				Name: "GraphGetEntityByObjectIdNotFoundError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableObjectID, "1")
					apitest.BodyString(input, `{"collected":true}`)
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, graph.ErrNoResultsFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, "node not found")
				},
			},
			{
				Name: "GraphGetEntityByObjectIdUnknownError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableObjectID, "1")
					apitest.BodyString(input, `{"collected":true}`)
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("generic graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "error getting node:")
				},
			},
			{
				Name: "GraphBatchNodeUpdateError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableObjectID, "1")
					apitest.BodyString(input, `{"collected":true}`)
				},
				Setup: func() {
					mockGraph.EXPECT().GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).Return(node, nil)
					mockGraph.EXPECT().BatchNodeUpdate(gomock.Any(), gomock.Any()).Return(errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "error updating node:")
				},
			},

			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableObjectID, "1")
					apitest.BodyString(input, `{"collected":true}`)
				},
				Setup: func() {
					mockGraph.EXPECT().GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).Return(node, nil)
					mockGraph.EXPECT().BatchNodeUpdate(gomock.Any(), gomock.Any()).Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetGPOEntityInfo(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetGPOEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "error reading objectid:")
				},
			},
			{
				Name: "InvalidCountsParam",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsBadQueryParameterFilters)
				},
			},
			{
				Name: "GraphDBNotFoundError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, graph.ErrNoResultsFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, "node not found")
				},
			},
			{
				Name: "GraphDBGetEntityByObjectIdError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "error getting node:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, nil)
					mockGraph.EXPECT().
						GetEntityCountResults(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessWithoutCounts",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "false")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetOUEntityInfo(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetOUEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "error reading objectid:")
				},
			},
			{
				Name: "InvalidCountsParam",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsBadQueryParameterFilters)
				},
			},
			{
				Name: "GraphDBNotFoundError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, graph.ErrNoResultsFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, "node not found")
				},
			},
			{
				Name: "GraphDBGetEntityByObjectIdError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "error getting node:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, nil)
					mockGraph.EXPECT().
						GetEntityCountResults(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessWithoutCounts",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "false")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetUserEntityInfo(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetUserEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "error reading objectid:")
				},
			},
			{
				Name: "InvalidCountsParam",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsBadQueryParameterFilters)
				},
			},
			{
				Name: "GraphDBNotFoundError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, graph.ErrNoResultsFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, "node not found")
				},
			},
			{
				Name: "GraphDBGetEntityByObjectIdError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "error getting node:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, nil)
					mockGraph.EXPECT().
						GetEntityCountResults(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessWithoutCounts",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "false")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetGroupEntityInfo(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetGroupEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "error reading objectid:")
				},
			},
			{
				Name: "InvalidCountsParam",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsBadQueryParameterFilters)
				},
			},
			{
				Name: "GraphDBNotFoundError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, graph.ErrNoResultsFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, "node not found")
				},
			},
			{
				Name: "GraphDBGetEntityByObjectIdError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "error getting node:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, nil)
					mockGraph.EXPECT().
						GetEntityCountResults(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessWithoutCounts",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, "object_id", "1")
					apitest.AddQueryParam(input, "counts", "false")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetEntityByObjectId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestManagementResource_GetBaseEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=`"}},
			},
		},
		{
			name: "Error: GetEntityObjectIDFromRequestPath - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"not_object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error reading objectid: no object ID found in request"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("Base")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("Base")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("Base")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(req.Context(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=false"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("Base")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			resources.GetBaseEntityInfo(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetContainerEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=`"}},
			},
		},
		{
			name: "Error: GetEntityObjectIDFromRequestPath - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"not_object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error reading objectid: no object ID found in request"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("Container")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("Container")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("Container")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(req.Context(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=false"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("Container")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			resources.GetContainerEntityInfo(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetAIACAEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=`"}},
			},
		},
		{
			name: "Error: GetEntityObjectIDFromRequestPath - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"not_object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error reading objectid: no object ID found in request"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("AIACA")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("AIACA")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("AIACA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(req.Context(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=false"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("AIACA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}
			request := testCase.buildRequest()

			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			resources.GetAIACAEntityInfo(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetRootCAEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=`"}},
			},
		},
		{
			name: "Error: GetEntityObjectIDFromRequestPath - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"not_object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error reading objectid: no object ID found in request"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("RootCA")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("RootCA")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("RootCA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(req.Context(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=false"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("RootCA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()

			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			resources.GetRootCAEntityInfo(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetEnterpriseCAEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=`"}},
			},
		},
		{
			name: "Error: GetEntityObjectIDFromRequestPath - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"not_object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error reading objectid: no object ID found in request"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("EnterpriseCA")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("EnterpriseCA")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("EnterpriseCA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(req.Context(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=false"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("EnterpriseCA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			resources.GetEnterpriseCAEntityInfo(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetNTAuthStoreEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=`"}},
			},
		},
		{
			name: "Error: GetEntityObjectIDFromRequestPath - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"not_object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error reading objectid: no object ID found in request"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("NTAuthStore")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("NTAuthStore")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("NTAuthStore")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(req.Context(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=false"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("NTAuthStore")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			resources.GetNTAuthStoreEntityInfo(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetCertTemplateEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=`"}},
			},
		},
		{
			name: "Error: GetEntityObjectIDFromRequestPath - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"not_object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error reading objectid: no object ID found in request"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("CertTemplate")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("CertTemplate")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("CertTemplate")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(req.Context(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=false"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("CertTemplate")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			resources.GetCertTemplateEntityInfo(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetIssuancePolicyEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=`"}},
			},
		},
		{
			name: "Error: GetEntityObjectIDFromRequestPath - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"not_object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error reading objectid: no object ID found in request"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("IssuancePolicy")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("IssuancePolicy")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("IssuancePolicy")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(req.Context(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=true"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
					},
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?counts=false"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(req.Context(), "id", graph.StringKind("IssuancePolicy")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			resources.GetIssuancePolicyEntityInfo(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
