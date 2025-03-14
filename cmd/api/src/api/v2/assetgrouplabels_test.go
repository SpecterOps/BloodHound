// Copyright 2025 Specter Ops, Inc.
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
	"testing"

	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	mocks_db "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	mocks_graph "github.com/specterops/bloodhound/src/queries/mocks"
	"go.uber.org/mock/gomock"
)

func TestResources_CreateAssetGroupLabelSelector(t *testing.T) {
	var (
		mockCtrl      = gomock.NewController(t)
		mockDB        = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb   = mocks_graph.NewMockGraph(mockCtrl)
		resourcesInst = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
		user    = setupUser()
		userCtx = setupUserCtx(user)
	)

	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resourcesInst.CreateAssetGroupLabelSelector).
		Run([]apitest.Case{
			{
				Name: "BadRequest",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupLabelID, "1")
					apitest.BodyString(input, "{\"name\":[\"BadRequest\"]}")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponsePayloadUnmarshalError)
				},
			},
			{
				Name: "MissingName",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupLabelID, "1")
					apitest.BodyStruct(input, model.AssetGroupLabelSelector{
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: 0, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: false,
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Name: property is required")
				},
			},
			{
				Name: "MissingSeeds",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupLabelID, "1")
					apitest.BodyStruct(input, model.AssetGroupLabelSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						IsDefault:   false,
						AutoCertify: false,
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Seeds: property is required")
				},
			},
			{
				Name: "MissingUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.BodyStruct(input, model.AssetGroupLabelSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: 0, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: false,
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "invalid asset group label id specified in url")
				},
			},
			{
				Name: "InvalidUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupLabelID, "non-numeric")
					apitest.BodyStruct(input, model.AssetGroupLabelSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: 0, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: false,
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "invalid asset group label id specified in url")
				},
			},
			{
				Name: "NonExistantLabelUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupLabelID, "1234")
					apitest.BodyStruct(input, model.AssetGroupLabelSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: 0, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: false,
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupLabel(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupLabel{}, errors.New("entity not found")).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "invalid asset group label id specified in url")
				},
			},
			{
				Name: "DatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupLabelID, "1")
					apitest.BodyStruct(input, model.AssetGroupLabelSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: 0, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: false,
					})
				},
				Setup: func() {
					mockDB.EXPECT().
						CreateAssetGroupLabelSelector(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupLabelSelector{}, errors.New("failure")).Times(1)
					mockDB.EXPECT().GetAssetGroupLabel(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupLabel{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, api.ErrorResponseDetailsInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupLabelID, "1")
					apitest.BodyStruct(input, model.AssetGroupLabelSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
						},
						IsDefault:   false,
						AutoCertify: false,
					})
				},
				Setup: func() {
					mockDB.EXPECT().
						CreateAssetGroupLabelSelector(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupLabelSelector{Name: "TestSelector"}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupLabel(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupLabel{}, nil).Times(1)

					mockGraphDb.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, nil).Times(1)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusCreated)
				},
			},
		})
}
