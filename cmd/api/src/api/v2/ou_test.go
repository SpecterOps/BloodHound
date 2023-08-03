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

	"go.uber.org/mock/gomock"

	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/errors"
)

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
