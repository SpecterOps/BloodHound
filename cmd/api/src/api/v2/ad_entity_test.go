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
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/queries/mocks"
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
