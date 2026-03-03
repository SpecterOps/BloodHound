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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	mocks_db "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_GetComputerEntityInfo(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}

		bheCtx = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner:               model.User{},
				Session:             model.UserSession{},
			},
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetComputerEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, bheCtx.ConstructGoContext())
				},
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
		resources = v2.Resources{GraphQuery: mockGraph, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}
		bheCtx    = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner:               model.User{},
				Session:             model.UserSession{},
			},
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetDomainEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, bheCtx.ConstructGoContext())
				},
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
		resources = v2.Resources{GraphQuery: mockGraph, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}
		bheCtx    = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner:               model.User{},
				Session:             model.UserSession{},
			},
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetGPOEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, bheCtx.ConstructGoContext())
				},
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
		resources = v2.Resources{GraphQuery: mockGraph, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}
		bheCtx    = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner:               model.User{},
				Session:             model.UserSession{},
			},
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetOUEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, bheCtx.ConstructGoContext())
				},
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
		resources = v2.Resources{GraphQuery: mockGraph, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}
		bheCtx    = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner:               model.User{},
				Session:             model.UserSession{},
			},
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetUserEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, bheCtx.ConstructGoContext())
				},
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
		resources = v2.Resources{GraphQuery: mockGraph, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}
		bheCtx    = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner:               model.User{},
				Session:             model.UserSession{},
			},
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetGroupEntityInfo).
		Run([]apitest.Case{
			{
				Name: "NoObjectID",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, bheCtx.ConstructGoContext())
				},
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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
					apitest.SetContext(input, bheCtx.ConstructGoContext())
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

func TestResources_GetBaseEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *mocks_db.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		dogTagsOverrides dogtags.TestOverrides
		user             model.User
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",

			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/base/id",
						RawQuery: "counts=bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/base/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("Base")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/base/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("Base")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/base/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("Base")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/base/id",
						RawQuery: "counts=false",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null, "kinds":[]}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("Base")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/base/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.Entity).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity},
					Properties: props,
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, ad.Entity), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/base/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.Entity).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity},
					Properties: props,
				}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, graph.StringsToKinds([]string{ad.Entity.String()})...), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/base/id",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.Entity).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			bheCtx := ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/base/{%s}", api.URIPathVariableObjectID), resources.GetBaseEntityInfo).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetContainerEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *mocks_db.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		dogTagsOverrides dogtags.TestOverrides
		user             model.User
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/containers/id",
						RawQuery: "counts=`",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/containers/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("Container")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/containers/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("Container")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/containers/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("Container")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/containers/id",
						RawQuery: "counts=false",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null, "kinds": []}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("Container")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/containers/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.Container).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Container},
					Properties: props,
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, ad.Container), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/containers/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.Container).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity, ad.Container},
					Properties: props,
				}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, graph.StringsToKinds([]string{ad.Entity.String(), ad.Container.String()})...), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/containers/id",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.Container).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			bheCtx := ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/containers/{%s}", api.URIPathVariableObjectID), resources.GetContainerEntityInfo).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetAIACAEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *mocks_db.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/aiacas/id",
						RawQuery: "counts=`",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/aiacas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("AIACA")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/aiacas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("AIACA")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/aiacas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("AIACA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/aiacas/id",
						RawQuery: "counts=false",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null, "kinds": []}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("AIACA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/aiacas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.AIACA).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.AIACA},
					Properties: props,
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, ad.AIACA), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/aiacas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.AIACA).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity, ad.AIACA},
					Properties: props,
				}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, graph.StringsToKinds([]string{ad.Entity.String(), ad.AIACA.String()})...), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/aiacas/id",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.AIACA).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
			}
			request := testCase.buildRequest()

			bheCtx := ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/aiacas/{%s}", api.URIPathVariableObjectID), resources.GetAIACAEntityInfo).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetRootCAEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *mocks_db.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "counts=`",
						Path:     "/api/v2/rootcas/id",
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
						Path:     "/api/v2/rootcas/id",
					},
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("RootCA")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
						Path:     "/api/v2/rootcas/id",
					},
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("RootCA")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "counts=true",
						Path:     "/api/v2/rootcas/id",
					},
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("RootCA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "counts=false",
						Path:     "/api/v2/rootcas/id",
					},
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null, "kinds": []}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("RootCA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/rootcas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.RootCA).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.RootCA},
					Properties: props,
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, ad.RootCA), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/rootcas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.RootCA).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity, ad.RootCA},
					Properties: props,
				}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, graph.StringsToKinds([]string{ad.Entity.String(), ad.RootCA.String()})...), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/rootcas/id",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.RootCA).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()

			bheCtx := ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/rootcas/{%s}", api.URIPathVariableObjectID), resources.GetRootCAEntityInfo).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetEnterpriseCAEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *mocks_db.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/enterprisecas/id",
						RawQuery: "counts=`",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/enterprisecas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("EnterpriseCA")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/enterprisecas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("EnterpriseCA")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/enterprisecas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("EnterpriseCA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/enterprisecas/id",
						RawQuery: "counts=false",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null, "kinds": []}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("EnterpriseCA")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/enterprisecas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.EnterpriseCA).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.EnterpriseCA},
					Properties: props,
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, ad.EnterpriseCA), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/enterprisecas/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.EnterpriseCA).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity, ad.EnterpriseCA},
					Properties: props,
				}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, graph.StringsToKinds([]string{ad.Entity.String(), ad.EnterpriseCA.String()})...), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/enterprisecas/id",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.EnterpriseCA).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			bheCtx := ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/enterprisecas/{%s}", api.URIPathVariableObjectID), resources.GetEnterpriseCAEntityInfo).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetNTAuthStoreEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *mocks_db.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/ntauthstores/id",
						RawQuery: "counts=`",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/ntauthstores/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("NTAuthStore")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/ntauthstores/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("NTAuthStore")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/ntauthstores/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("NTAuthStore")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/ntauthstores/id",
						RawQuery: "counts=false",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null, "kinds": []}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("NTAuthStore")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/ntauthstores/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.NTAuthStore).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.NTAuthStore},
					Properties: props,
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, ad.NTAuthStore), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/ntauthstores/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.NTAuthStore).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity, ad.NTAuthStore},
					Properties: props,
				}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, graph.StringsToKinds([]string{ad.Entity.String(), ad.NTAuthStore.String()})...), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/ntauthstores/id",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.NTAuthStore).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			bheCtx := ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/ntauthstores/{%s}", api.URIPathVariableObjectID), resources.GetNTAuthStoreEntityInfo).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetCertTemplateEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *mocks_db.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/certtemplates/id",
						RawQuery: "counts=`",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/certtemplates/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("CertTemplate")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/certtemplates/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("CertTemplate")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/certtemplates/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("CertTemplate")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/certtemplates/id",
						RawQuery: "counts=false",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null, "kinds": []}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("CertTemplate")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/certtemplates/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.CertTemplate).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.CertTemplate},
					Properties: props,
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, ad.CertTemplate), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/certtemplates/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.CertTemplate).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity, ad.CertTemplate},
					Properties: props,
				}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, graph.StringsToKinds([]string{ad.Entity.String(), ad.CertTemplate.String()})...), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/certtemplates/id",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.CertTemplate).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			bheCtx := ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/certtemplates/{%s}", api.URIPathVariableObjectID), resources.GetCertTemplateEntityInfo).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetIssuancePolicyEntityInfo(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *mocks_db.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: ParseOptionalBool - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/issuancepolicies/id",
						RawQuery: "counts=`",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: GetEntityByObjectId - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/issuancepolicies/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("IssuancePolicy")).Return(nil, graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"node not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetEntityByObjectId - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/issuancepolicies/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("IssuancePolicy")).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error getting node: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/issuancepolicies/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("IssuancePolicy")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
				mocks.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(1), graph.NewProperties()), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: !hydrateCounts - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/issuancepolicies/id",
						RawQuery: "counts=false",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"props":null, "kinds": []}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", graph.StringKind("IssuancePolicy")).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/issuancepolicies/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.IssuancePolicy).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.IssuancePolicy},
					Properties: props,
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, ad.IssuancePolicy), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/issuancepolicies/id",
						RawQuery: "counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				props := graph.AsProperties(map[string]any{
					"domainsid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.IssuancePolicy).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{ad.Entity, ad.IssuancePolicy},
					Properties: props,
				}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
				mock.mockGraphQuery.EXPECT().GetEntityCountResults(gomock.Any(), graph.NewNode(graph.ID(16), props, graph.StringsToKinds([]string{ad.Entity.String(), ad.IssuancePolicy.String()})...), gomock.Any()).Return(map[string]any{"results": "output"})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"results":"output"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/issuancepolicies/id",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.IssuancePolicy).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			bheCtx := ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/issuancepolicies/{%s}", api.URIPathVariableObjectID), resources.GetIssuancePolicyEntityInfo).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
