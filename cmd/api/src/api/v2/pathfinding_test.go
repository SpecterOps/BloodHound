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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	mocks_graph "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
		mockCtrl       = gomock.NewController(t)
		mockGraph      = mocks_graph.NewMockGraph(mockCtrl)
		mockDB         = mocks.NewMockDatabase(mockCtrl)
		dogTagsService = dogtags.NewTestService(dogtags.TestOverrides{})
		resources      = v2.Resources{GraphQuery: mockGraph, DB: mockDB, DogTags: dogTagsService}

		user    = setupUser()
		userCtx = setupUserCtx(user)

		opengraphKinds                      = graph.Kinds{graph.StringKind("OpenGraphKindA"), graph.StringKind("OpenGraphKindB")}
		validBuiltInKinds                   = graph.Kinds(ad.Relationships()).Concatenate(azure.Relationships())
		validBuiltInTraversableKinds        = graph.Kinds(ad.PathfindingRelationshipsMatchFrontend()).Concatenate(azure.PathfindingRelationships())
		allKindsWithOpenGraph               = validBuiltInKinds.Concatenate(opengraphKinds)
		traversableKindsWithOpenGraph       = validBuiltInTraversableKinds.Concatenate(opengraphKinds)
		traversableKindsFilter              = query.KindIn(query.Relationship(), validBuiltInTraversableKinds...)
		allKindsFilter                      = query.KindIn(query.Relationship(), validBuiltInKinds...)
		traversableKindsFilterWithOpenGraph = query.KindIn(query.Relationship(), traversableKindsWithOpenGraph...)
		allKindsFilterWithOpenGraph         = query.KindIn(query.Relationship(), allKindsWithOpenGraph...)

		openGraphEdges = model.GraphSchemaRelationshipKinds{{SchemaExtensionId: 1, Name: "OpenGraphKindA", Description: "OpenGraph Kind A", IsTraversable: true}, {SchemaExtensionId: 2, Name: "OpenGraphKindB", Description: "OpenGraph Kind B", IsTraversable: true}}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetShortestPath).
		Run([]apitest.Case{
			// OpenGraph Feature Flag Off

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
				Name: "GetFlagByKeyError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "wrx")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{}, errors.New("database error"))

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "an internal error has occurred that is preventing the service from servicing this request")
				},
			},
			{
				Name: "InvalidRelationshipKindsQuery",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "wrx")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)

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
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)

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
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)

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
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
				},
			},
			{
				Name: "InvalidCombinationOfQueryParams",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "only_traversable", "true")
					apitest.AddQueryParam(input, "relationship_kinds", fmt.Sprintf("nin:%s", validBuiltInTraversableKinds))
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)

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
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "nin:Owns,GenericAll,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns,GenericAll,GenericWrite,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "nin:Owns,GenericAll,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns,GenericAll,GenericWrite,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
				},
			},
			{
				Name: "InvalidOnlyTraversableParamDefaultsToFalse",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "only_traversable", "notABool")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), "someID", "someOtherID", allKindsFilter).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Path not found")
				},
			},
			{
				Name: "OnlyTraversableParamReturnsOnlyTraversableRelationships",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "only_traversable", "true")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), "someID", "someOtherID", traversableKindsFilter).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Path not found")
				},
			},
			{
				Name: "OnlyTraversableParamWithRelationshipKindsReturnsOnlyTraversableRelationships",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Contains,AZScopedTo")
					apitest.AddQueryParam(input, "only_traversable", "true")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockGraph.EXPECT().
						GetAllShortestPaths(gomock.Any(), "someID", "someOtherID", query.KindIn(query.Relationship(), ad.Contains)).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Path not found")
				},
			},
			// OpenGraph Feature Flag On
			{
				Name: "GetGraphSchemaRelationshipKindsError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(model.GraphSchemaRelationshipKinds{}, 0, errors.New("database error"))

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "an internal error has occurred that is preventing the service from servicing this request")
				},
			},
			{
				Name: "InvalidRelationshipKindsQuery With OpenGraph",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "wrx")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "invalid query parameter 'relationship_kinds': acceptable values should match the format: in|nin:Kind1,Kind2")
				},
			},
			{
				Name: "InvalidRelationshipKindsOperator With OpenGraph",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "abcd:Owns,GenericAll,GenericWrite")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "invalid query parameter 'relationship_kinds': acceptable values should match the format: in|nin:Kind1,Kind2")
				},
			},
			{
				Name: "EmptyRelationshipKindsValues With OpenGraph",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "abcd:")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
				},
			},
			{
				Name: "InvalidRelationshipKindsParam With OpenGraph",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns,avbcs,GenericAll")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
				},
			},
			{
				Name: "GraphDBGetShortestPathsWithOpenGraphError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", allKindsFilterWithOpenGraph).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "graph error")
				},
			},
			{
				Name: "Empty Result Set With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", allKindsFilterWithOpenGraph).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Path not found")
				},
			},
			{
				Name: "NotFoundNin With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "nin:Owns,GenericAll,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", query.KindIn(query.Relationship(), allKindsWithOpenGraph.Exclude(graph.Kinds{}.Add(ad.Owns, ad.GenericAll, azure.ServicePrincipalEndpointReadWriteAll))...)).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
				},
			},
			{
				Name: "NotFoundIn With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns,GenericAll,GenericWrite,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", query.KindIn(query.Relationship(), graph.Kinds{}.Add(ad.Owns, ad.GenericAll, ad.GenericWrite, azure.ServicePrincipalEndpointReadWriteAll)...)).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
				},
			},
			{
				Name: "SuccessNin With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "nin:Owns,GenericAll,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", query.KindIn(query.Relationship(), allKindsWithOpenGraph.Exclude(graph.Kinds{}.Add(ad.Owns, ad.GenericAll, azure.ServicePrincipalEndpointReadWriteAll))...)).
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
				Name: "SuccessIn With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns,GenericAll,GenericWrite,AZMGServicePrincipalEndpoint_ReadWrite_All")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", query.KindIn(query.Relationship(), graph.Kinds{}.Add(ad.Owns, ad.GenericAll, ad.GenericWrite, azure.ServicePrincipalEndpointReadWriteAll)...)).
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
				Name: "NotFoundSingleKind With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Owns")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", query.KindIn(query.Relationship(), ad.Owns)).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
				},
			},
			{
				Name: "OnlyTraversableParamReturnsOnlyTraversableRelationships With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "only_traversable", "true")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{"is_traversable": []model.Filter{{Operator: model.Equals, Value: "true"}}}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", traversableKindsFilterWithOpenGraph).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Path not found")
				},
			},
			{
				Name: "OnlyTraversableParamWithINRelationshipKindsReturnsOnlyTraversableRelationships With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "in:Contains,AZScopedTo")
					apitest.AddQueryParam(input, "only_traversable", "true")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{"is_traversable": []model.Filter{{Operator: model.Equals, Value: "true"}}}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", query.KindIn(query.Relationship(), ad.Contains)).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Path not found")
				},
			},

			{
				Name: "OnlyTraversableParamWithNINRelationshipKindsReturnsOnlyTraversableRelationships With OpenGraph",
				Input: func(input *apitest.Input) {
					userCtx = setupUserCtx(user)
					apitest.SetContext(input, userCtx)

					apitest.AddQueryParam(input, "start_node", "someID")
					apitest.AddQueryParam(input, "end_node", "someOtherID")
					apitest.AddQueryParam(input, "relationship_kinds", "nin:Contains,AZScopedTo")
					apitest.AddQueryParam(input, "only_traversable", "true")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetGraphSchemaRelationshipKinds(gomock.Any(), model.Filters{"is_traversable": []model.Filter{{Operator: model.Equals, Value: "true"}}}, model.Sort{}, 0, 0).Return(openGraphEdges, 2, nil)
					mockGraph.EXPECT().
						GetAllShortestPathsWithOpenGraph(gomock.Any(), "someID", "someOtherID", query.KindIn(query.Relationship(), traversableKindsWithOpenGraph.Exclude(graph.Kinds{}.Add(ad.Contains))...)).
						Return(graph.NewPathSet(), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.UnmarshalBody(output, &api.ErrorWrapper{})
					apitest.BodyContains(output, "Path not found")
				},
			},
		})
}

func TestResources_GetShortestPath_ETAC(t *testing.T) {
	tests := []struct {
		name               string
		queryParams        map[string]string
		expectedMocks      func(mockDB *mocks.MockDatabase, mockGraph *mocks_graph.MockGraph)
		expectedStatusCode int
		assertBody         func(t *testing.T, body string)
		dogTagsOverrides   dogtags.TestOverrides
		user               model.User
	}{
		{
			name: "FilterETACGraph",
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						UserID:        "",
						EnvironmentID: "12345",
						BigSerial:     model.BigSerial{},
					},
				},
			},
			queryParams: map[string]string{
				"start_node":         "0",
				"end_node":           "1",
				"relationship_kinds": "in:GenericWrite",
			},
			expectedMocks: func(mockDB *mocks.MockDatabase, mockGraph *mocks_graph.MockGraph) {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
					Return(appcfg.FeatureFlag{Enabled: false}, nil)

				mockGraph.EXPECT().
					GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(graph.NewPathSet(graph.Path{
						Nodes: []*graph.Node{
							{
								ID:    0,
								Kinds: graph.Kinds{ad.Entity, ad.Computer},
								Properties: graph.AsProperties(graph.PropertyMap{
									ad.DomainSID: "inaccessible",
									common.Name:  "invisible",
								}),
							},
							{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.User},
								Properties: graph.AsProperties(graph.PropertyMap{
									ad.DomainSID: "12345",
									common.Name:  "visible",
								}),
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
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Hidden")
				assert.Contains(t, body, "visible")
				assert.NotContains(t, body, "invisible")
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Filter ETAC With No Nodes To Filter",
			user: model.User{
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						UserID:        "",
						EnvironmentID: "12345",
						BigSerial:     model.BigSerial{},
					},
				},
			},
			queryParams: map[string]string{
				"start_node":         "0",
				"end_node":           "1",
				"relationship_kinds": "in:GenericWrite",
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expectedMocks: func(mockDB *mocks.MockDatabase, mockGraph *mocks_graph.MockGraph) {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphExtensionManagement).
					Return(appcfg.FeatureFlag{Enabled: false}, nil)
				mockGraph.EXPECT().
					GetAllShortestPaths(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(graph.NewPathSet(graph.Path{
						Nodes: []*graph.Node{
							{
								ID:    0,
								Kinds: graph.Kinds{ad.Entity, ad.Computer},
								Properties: graph.AsProperties(graph.PropertyMap{
									ad.DomainSID: "12345",
									common.Name:  "visible-2",
								}),
							},
							{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.User},
								Properties: graph.AsProperties(graph.PropertyMap{
									ad.DomainSID: "12345",
									common.Name:  "visible",
								}),
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
			expectedStatusCode: http.StatusOK,
			assertBody: func(t *testing.T, body string) {
				assert.NotContains(t, body, "Hidden")
				assert.Contains(t, body, "visible")
				assert.Contains(t, body, "visible-2")
			},
		},
	}

	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {

			var (
				mockCtrl       = gomock.NewController(tt)
				mockDB         = mocks.NewMockDatabase(mockCtrl)
				mockGraph      = mocks_graph.NewMockGraph(mockCtrl)
				dogTagsService = dogtags.NewTestService(tc.dogTagsOverrides)
				resources      = v2.Resources{GraphQuery: mockGraph, DB: mockDB, DogTags: dogTagsService}
				endpoint       = "/api/v2/resource/get-shortest-path"
				ctx            = setupUserCtx(tc.user)
			)
			defer mockCtrl.Finish()

			tc.expectedMocks(mockDB, mockGraph)

			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			require.NoError(tt, err)

			queryParams := req.URL.Query()
			for key, value := range tc.queryParams {
				queryParams.Set(key, value)
			}
			req.URL.RawQuery = queryParams.Encode()

			router := mux.NewRouter()
			router.HandleFunc(endpoint, resources.GetShortestPath).Methods(http.MethodGet)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(tt, tc.expectedStatusCode, rr.Code)
			tc.assertBody(tt, rr.Body.String())
		})
	}
}

func TestResources_GetSearchResult(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks_graph.NewMockGraph(mockCtrl)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph, DB: mockDB, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}
		user      = setupUser()
		userCtx   = setupUserCtx(user)
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetSearchResult).
		Run([]apitest.Case{
			{
				Name: "MissingSearchParam",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
				},
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
					apitest.SetContext(input, userCtx)
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
					apitest.SetContext(input, userCtx)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Expected only one search type.")
				},
			},
			{
				Name: "FeatureFlagDatabaseError -- Open Graph",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).
						Return(appcfg.FeatureFlag{}, errors.New("database error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "an internal error has occurred that is preventing the service from servicing this request")
				},
			},
			{
				Name: "GraphDBSearchByNameOrObjectIDError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockGraph.EXPECT().
						SearchByNameOrObjectID(gomock.Any(), true, "some query", queries.SearchTypeFuzzy).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Error getting search results:")
				},
			},
			{
				Name: "DBGetCustomNodeKindsError -- should still return results",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockGraph.EXPECT().
						SearchByNameOrObjectID(gomock.Any(), true, "some query", queries.SearchTypeFuzzy).
						Return(graph.NewNodeSet(), nil)
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, errors.New("error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success -- include OpenGraph results",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.AddQueryParam(input, "type", "fuzzy")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)

					nodeSet := graph.NewNodeSet()
					personNode := &graph.Node{
						ID:    1,
						Kinds: []graph.Kind{graph.StringKind("Person")},
						Properties: &graph.Properties{
							Map: map[string]any{},
						},
					}

					nodeSet.Add(personNode)

					mockGraph.EXPECT().
						SearchByNameOrObjectID(gomock.Any(), true, "some query", queries.SearchTypeFuzzy).
						Return(nodeSet, nil)
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{
						{ID: 1, KindName: "Person", Config: model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "fas fa-person-half-dress")
				},
			},
			{
				Name: "Success -- exclude OpenGraph results",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.AddQueryParam(input, "type", "fuzzy")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).
						Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockGraph.EXPECT().
						SearchByNameOrObjectID(gomock.Any(), false, "some query", queries.SearchTypeFuzzy).
						Return(graph.NewNodeSet(), nil)
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},

			{
				Name: "Success -- include OpenGraph results, set icon properly when multiple kinds in Kinds array",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "query", "some query")
					apitest.AddQueryParam(input, "type", "fuzzy")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().
						GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).
						Return(appcfg.FeatureFlag{Enabled: true}, nil)

					nodeSet := graph.NewNodeSet()
					personNode := &graph.Node{
						ID:    1,
						Kinds: []graph.Kind{graph.StringKind("OtherKind"), graph.StringKind("Person")},
						Properties: &graph.Properties{
							Map: map[string]any{},
						},
					}

					nodeSet.Add(personNode)

					mockGraph.EXPECT().
						SearchByNameOrObjectID(gomock.Any(), true, "some query", queries.SearchTypeFuzzy).
						Return(nodeSet, nil)
					mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{
						{ID: 1, KindName: "Person", Config: model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "person-half-dress", Color: "#ff91af"}}}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, `"nodetype":"Person"`)
				},
			},
		})
}

func TestResources_GetSearchResult_ETAC(t *testing.T) {
	tests := []struct {
		name               string
		queryParams        map[string]string
		expectedMocks      func(mockDB *mocks.MockDatabase, mockGraph *mocks_graph.MockGraph)
		expectedStatusCode int
		assertBody         func(t *testing.T, body string)
		dogTagsOverrides   dogtags.TestOverrides
		user               model.User
	}{
		{
			name: "Success -- ETAC enabled,user has all environments",
			user: model.User{
				AllEnvironments: true,
			},
			queryParams: map[string]string{
				"query": "some query",
				"type":  "fuzzy",
			},
			expectedMocks: func(mockDB *mocks.MockDatabase, mockGraph *mocks_graph.MockGraph) {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).
					Return(appcfg.FeatureFlag{Enabled: true}, nil)

				mockGraph.EXPECT().
					SearchByNameOrObjectID(gomock.Any(), true, "some query", queries.SearchTypeFuzzy).
					Return(graph.NewNodeSet(), nil)

				mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)

			},
			expectedStatusCode: http.StatusOK,
			assertBody: func(t *testing.T, body string) {

			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
		},
		{
			name: "Success -- ETAC enabled,user has limited access",
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "testenv"},
				},
			},
			queryParams: map[string]string{
				"query": "some query",
				"type":  "fuzzy",
			},
			expectedMocks: func(mockDB *mocks.MockDatabase, mockGraph *mocks_graph.MockGraph) {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).
					Return(appcfg.FeatureFlag{Enabled: true}, nil)

				nodeSet := graph.NewNodeSet()

				accessibleNode := &graph.Node{
					ID: 1,
					Properties: &graph.Properties{
						Map: map[string]any{
							"domainsid": "testenv"},
					},
				}
				hiddenNode := &graph.Node{
					ID:    2,
					Kinds: graph.Kinds{ad.Computer},
					Properties: &graph.Properties{
						Map: map[string]any{
							"domainsid": "restricted",
							"name":      "restricted",
						},
					},
				}
				nodeSet.Add(accessibleNode)
				nodeSet.Add(hiddenNode)
				mockGraph.EXPECT().
					SearchByNameOrObjectID(gomock.Any(), true, "some query", queries.SearchTypeFuzzy).
					Return(nodeSet, nil)

				mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return([]model.CustomNodeKind{}, nil)
			},
			expectedStatusCode: http.StatusOK,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "testenv")
				assert.Contains(t, body, "** Hidden Computer Object **")
				assert.NotContains(t, body, "restricted")
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
				mockGraph      = mocks_graph.NewMockGraph(mockCtrl)
				dogTagsService = dogtags.NewTestService(tc.dogTagsOverrides)
				resources      = v2.Resources{GraphQuery: mockGraph, DB: mockDB, DogTags: dogTagsService}
				endpoint       = "/api/v2/graph-search"
				ctx            = setupUserCtx(tc.user)
			)
			defer mockCtrl.Finish()

			tc.expectedMocks(mockDB, mockGraph)

			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			require.NoError(tt, err)

			queryParams := req.URL.Query()
			for key, value := range tc.queryParams {
				queryParams.Set(key, value)
			}
			req.URL.RawQuery = queryParams.Encode()

			router := mux.NewRouter()
			router.HandleFunc(endpoint, resources.GetSearchResult).Methods(http.MethodGet)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(tt, tc.expectedStatusCode, rr.Code)
			tc.assertBody(tt, rr.Body.String())
		})
	}
}
