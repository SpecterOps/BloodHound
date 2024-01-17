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
	"fmt"
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"

	"go.uber.org/mock/gomock"

	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	graphMocks "github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/specterops/bloodhound/errors"
)

func TestResources_SearchHandler(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = graphMocks.NewMockGraph(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.SearchHandler).
		Run([]apitest.Case{
			{
				Name: "EmptySearchQueryFailure",
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
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Invalid query parameter")
				},
			},
			{
				Name: "ParseKindsError",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
					apitest.AddQueryParam(input, "type", "invalidKind")
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
				},
				Setup: func() {
					mockGraph.EXPECT().
						SearchNodesByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, errors.New("graph error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "Graph error:")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "q", "search value")
				},
				Setup: func() {
					mockGraph.EXPECT().
						SearchNodesByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, nil)
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
					mockGraphQueries.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return(graph.NodeSet{}, fmt.Errorf("Some error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Setup: func() {
					mockGraphQueries.EXPECT().GetFilteredAndSortedNodes(gomock.Any(), gomock.Any()).Return(graph.NodeSet{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "[]")
				},
			},
		})
}
