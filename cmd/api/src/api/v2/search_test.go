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
	"testing"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	graphMocks "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"go.uber.org/mock/gomock"
)

func TestResources_SearchHandler(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = graphMocks.NewMockGraph(mockCtrl)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph, DB: mockDB}
		user      = setupUser()
		userCtx   = setupUserCtx(user)
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
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureETAC).Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureETAC).Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{}, errors.New("database error"))
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureETAC).Return(appcfg.FeatureFlag{Enabled: false}, nil)
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
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureETAC).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Invalid type parameter")
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
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), nil).
						Return(nil, errors.New("graph error"))
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureETAC).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "Graph error:")
				},
			},
			{
				Name: "Success -- Feature Flag On",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), graph.Kinds{}, "search value", true, 0, 10, nil).
						Return(nil, nil)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureETAC).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success -- Feature Flag Off",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureOpenGraphSearch).Return(appcfg.FeatureFlag{Enabled: false}, nil)
					mockGraph.EXPECT().
						SearchNodesByNameOrObjectId(gomock.Any(), graph.Kinds{ad.Entity, azure.Entity}, "search value", false, 0, 10, nil).
						Return(nil, nil)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureETAC).Return(appcfg.FeatureFlag{Enabled: false}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetAvailableDomains(t *testing.T) {
	var (
		mockCtrl         = gomock.NewController(t)
		mockGraphQueries = graphMocks.NewMockGraph(mockCtrl)
		resources        = v2.Resources{GraphQuery: mockGraphQueries}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetAvailableDomains).
		Run([]apitest.Case{
			{
				Name: "GraphQueryError",
				Setup: func() {
					mockGraphQueries.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return([]*graph.Node{}, fmt.Errorf("Some error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Setup: func() {
					mockGraphQueries.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return([]*graph.Node{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "[]")
				},
			},
		})
}
