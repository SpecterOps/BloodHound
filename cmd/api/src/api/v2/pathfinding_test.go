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
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	mocks_graph "github.com/specterops/bloodhound/src/queries/mocks"
	"go.uber.org/mock/gomock"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func TestResources_GetPathfindingResult(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks_graph.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetPathfindingResult).
		Run([]apitest.Case{
			{
				Name: "MissingStartNodeIDParam",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Missing query parameter: start_node")
				},
			},
			{
				Name: "MissingEndNodeIDParam",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Missing query parameter: end_node")
				},
			},
			{
				Name: "GraphDBGetShortestPathsError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "Error:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetShortestPath(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks_graph.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetShortestPath).
		Run([]apitest.Case{
			{
				Name: "MissingStartNodeIDParam",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Missing query parameter: start_node")
				},
			},
			{
				Name: "MissingEndNodeIDParam",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Missing query parameter: end_node")
				},
			},
			{
				Name: "InvalidRelationshipKindsQuery",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "wrx")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "invalid query parameter 'relationship_kinds': acceptable values should match the format: in|nin:Kind1,Kind2")
				},
			},
			{
				Name: "InvalidRelationshipKindsOperator",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "abcd:Owns,GenericAll,GenericWrite")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "invalid query parameter 'relationship_kinds': acceptable values should match the format: in|nin:Kind1,Kind2")
				},
			},
			{
				Name: "EmptyRelationshipKindsValues",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "abcd:")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
				},
			},
			{
				Name: "InvalidRelationshipKindsParam",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns,avbcs,GenericAll")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
				},
			},
			{
				Name: "GraphDBGetShortestPathsError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "graph error")
				},
			},
			{
				Name: "Empty Result Set",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Path not found")
				},
			},
			{
				Name: "NotFoundNin",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "nin:Owns,GenericAll,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
				},
			},
			{
				Name: "NotFoundIn",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns,GenericAll,GenericWrite,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
				},
			},
			{
				Name: "SuccessNin",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "nin:Owns,GenericAll,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewPathSet(graph.Path{
							Nodes: []*graph.Node{
								{
									ID:         0,
									Kinds:      graph.Kinds{ad.Entity, ad.Computer},
									Properties: graph.NewProperties(),
								},
								{
									ID:         1,
									Kinds:      graph.Kinds{ad.Entity, ad.User},
									Properties: graph.NewProperties(),
								},
							},
							Edges: []*graph.Relationship{
								{
									ID:         0,
									StartID:    0,
									EndID:      1,
									Kind:       ad.GenericWrite,
									Properties: graph.NewProperties(),
								},
							},
						}), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessIn",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns,GenericAll,GenericWrite,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewPathSet(graph.Path{
							Nodes: []*graph.Node{
								{
									ID:         0,
									Kinds:      graph.Kinds{ad.Entity, ad.Computer},
									Properties: graph.NewProperties(),
								},
								{
									ID:         1,
									Kinds:      graph.Kinds{ad.Entity, ad.User},
									Properties: graph.NewProperties(),
								},
							},
							Edges: []*graph.Relationship{
								{
									ID:         0,
									StartID:    0,
									EndID:      1,
									Kind:       ad.GenericWrite,
									Properties: graph.NewProperties(),
								},
							},
						}), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "NotFoundSingleKind",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns")
				},
				Setup: func() {
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
				},
			},
		})
}

func TestResources_GetSearchResult(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks_graph.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetSearchResult).
		Run([]apitest.Case{
			{
				Name: "MissingSearchParam",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Expected search parameter to be set.")
				},
			},
			{
				Name: "TooManySearchParams",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.AddQueryParam(input, "query", "some other invalid query")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Expected only one search value.")
				},
			},
			{
				Name: "TooManySearchTypeParams",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.AddQueryParam(input, "type", "search type")
					apitest.AddQueryParam(input, "type", "another search type")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Expected only one search type.")
				},
			},
			{
				Name: "GraphDBSearchByNameOrObjectIDError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
				},
				Setup: func() {
					mockGraph.EXPECT().
						SearchByNameOrObjectID(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Error getting search results:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.AddQueryParam(input, "type", "fuzzy")
				},
				Setup: func() {
					mockGraph.EXPECT().
						SearchByNameOrObjectID(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewNodeSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}
