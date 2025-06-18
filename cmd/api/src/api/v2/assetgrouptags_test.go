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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	uuid2 "github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/dawgs/graph"
	graphmocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/database"
	mocks_db "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/queries"
	mocks_graph "github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_GetAssetGroupTags(t *testing.T) {
	const queryParamTagType = "type"
	var (
		mockCtrl      = gomock.NewController(t)
		mockDB        = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb   = mocks_graph.NewMockGraph(mockCtrl)
		resourcesInst = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
	)

	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resourcesInst.GetAssetGroupTags).
		Run([]apitest.Case{
			{
				Name: "InvalidTagType",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, queryParamTagType, "123456")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTags(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTags{}, database.ErrNotFound)
				},
				Test: func(output apitest.Output) {
					resp := v2.GetAssetGroupTagsResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.Equal(output, 0, len(resp.Tags))
				},
			},
			{
				Name: "InvalidIncludeCounts",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, api.QueryParameterIncludeCounts, "blah")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "DatabaseError",
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTags(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTags{}, errors.New("failure"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, api.ErrorResponseDetailsInternalServerError)
				},
			},
			{
				Name: "NoResults",
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTags(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTags{}, database.ErrNotFound)
				},
				Test: func(output apitest.Output) {
					resp := v2.GetAssetGroupTagsResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.Equal(output, 0, len(resp.Tags))
				},
			},
			{
				Name: "TagTypeLabel",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, queryParamTagType, "eq:2") // model.AssetGroupTagTypeLabel
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTags(gomock.Any(), model.SQLFilter{
							SQLString: "type = ?",
							Params:    []any{strconv.Itoa(int(model.AssetGroupTagTypeLabel))},
						}).
						Return(model.AssetGroupTags{
							model.AssetGroupTag{ID: 1, Type: model.AssetGroupTagTypeLabel},
							model.AssetGroupTag{ID: 2, Type: model.AssetGroupTagTypeLabel},
						}, nil)
				},
				Test: func(output apitest.Output) {
					resp := v2.GetAssetGroupTagsResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.Equal(output, 2, len(resp.Tags))
					for _, t := range resp.Tags {
						apitest.Equal(output, model.AssetGroupTagTypeLabel, t.Type)
					}
				},
			},
			{
				Name: "TagTypeTier",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, queryParamTagType, "eq:1") // model.AssetGroupTagTypeTier
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTags(gomock.Any(), model.SQLFilter{
							SQLString: "type = ?",
							Params:    []any{strconv.Itoa(int(model.AssetGroupTagTypeTier))},
						}).
						Return(model.AssetGroupTags{
							model.AssetGroupTag{ID: 1, Type: model.AssetGroupTagTypeTier},
							model.AssetGroupTag{ID: 2, Type: model.AssetGroupTagTypeTier},
						}, nil)
				},
				Test: func(output apitest.Output) {
					resp := v2.GetAssetGroupTagsResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.Equal(output, 2, len(resp.Tags))
					for _, t := range resp.Tags {
						apitest.Equal(output, model.AssetGroupTagTypeTier, t.Type)
					}
				},
			},
			{
				Name: "TagTypeDefault",
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTags(gomock.Any(), model.SQLFilter{}).
						Return(model.AssetGroupTags{
							model.AssetGroupTag{ID: 1, Type: model.AssetGroupTagTypeLabel},
							model.AssetGroupTag{ID: 2, Type: model.AssetGroupTagTypeLabel},
							model.AssetGroupTag{ID: 3, Type: model.AssetGroupTagTypeTier},
							model.AssetGroupTag{ID: 4, Type: model.AssetGroupTagTypeTier},
						}, nil)
				},
				Test: func(output apitest.Output) {
					resp := v2.GetAssetGroupTagsResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.Equal(output, 4, len(resp.Tags))
					tierCount := 0
					for _, t := range resp.Tags {
						if t.Type == model.AssetGroupTagTypeTier {
							apitest.Equal(output, model.AssetGroupTagTypeTier, t.Type)
							tierCount++
						} else {
							apitest.Equal(output, model.AssetGroupTagTypeLabel, t.Type)
						}
					}
					apitest.Equal(output, 2, tierCount)
				},
			},
			{
				Name: "IncludeCounts Selector counts",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, api.QueryParameterIncludeCounts, "true")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTags(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTags{
							model.AssetGroupTag{ID: 1, Type: model.AssetGroupTagTypeTier},
							model.AssetGroupTag{ID: 2, Type: model.AssetGroupTagTypeTier},
							model.AssetGroupTag{ID: 3, Type: model.AssetGroupTagTypeLabel},
							model.AssetGroupTag{ID: 4, Type: model.AssetGroupTagTypeLabel},
						}, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorCounts(gomock.Any(), []int{1, 2, 3, 4}).
						Return(map[int]int{
							1: 5,
							2: 10,
							3: 0,
							4: 8,
						}, nil)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(0), nil).Times(4)
				},
				Test: func(output apitest.Output) {
					expectedCounts := map[int]int{
						1: 5,
						2: 10,
						3: 0,
						4: 8,
					}
					resp := v2.GetAssetGroupTagsResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.Equal(output, 4, len(resp.Tags))
					for _, t := range resp.Tags {
						expCount, ok := expectedCounts[t.ID]
						apitest.Equal(output, true, ok)
						apitest.Equal(output, false, t.Counts == nil)
						apitest.Equal(output, expCount, t.Counts.Selectors)
					}
				},
			},
			{
				Name: "IncludeCounts member counts",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, api.QueryParameterIncludeCounts, "true")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTags(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTags{
							model.AssetGroupTag{ID: 1, Name: "testlabel", Type: model.AssetGroupTagTypeLabel},
							model.AssetGroupTag{ID: 2, Name: "testtier", Type: model.AssetGroupTagTypeTier},
						}, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorCounts(gomock.Any(), []int{1, 2}).
						Return(map[int]int{
							1: 1,
							2: 1,
						}, nil)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(6), nil).
						Times(1)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(4), nil).
						Times(1)
				},
				Test: func(output apitest.Output) {
					expectedMemberCounts := map[int]int64{
						1: 6,
						2: 4,
					}
					resp := v2.GetAssetGroupTagsResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.Equal(output, 2, len(resp.Tags))
					for _, t := range resp.Tags {
						expCount, ok := expectedMemberCounts[t.ID]
						apitest.Equal(output, true, ok)
						apitest.Equal(output, false, t.Counts == nil)
						apitest.Equal(output, expCount, t.Counts.Members)
					}
				},
			},
		})
}

func TestResources_CreateAssetGroupTagSelector(t *testing.T) {
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
		NewHarness(t, resourcesInst.CreateAssetGroupTagSelector).
		Run([]apitest.Case{
			{
				Name: "BadRequest",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyString(input, `{"name":["BadRequest"]}`)
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)
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
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "Name: property is required")
				},
			},
			{
				Name: "MissingSeedsList",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)
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
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "InvalidUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "non-numeric")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "NonExistentTagUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1234")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, database.ErrNotFound).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "DatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockGraphDb.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, nil).Times(1)
					mockDB.EXPECT().
						CreateAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{}, errors.New("failure")).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, api.ErrorResponseDetailsInternalServerError)
				},
			},
			{
				Name: "InvalidCypher",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "cypher that's too complex"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)
					mockGraphDb.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, queries.ErrCypherQueryTooComplex).Times(1)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "InvalidSeedType",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: 0, Value: ""},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					value, _ := types.NewJSONBObject(map[string]any{"enabled": true})
					mockDB.EXPECT().
						GetConfigurationParameter(gomock.Any(), gomock.Any()).
						Return(appcfg.Parameter{Key: appcfg.ScheduledAnalysis, Value: value}, nil).Times(1)
					mockDB.EXPECT().
						CreateAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{Name: "TestSelector"}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)

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

func TestDatabase_GetAssetGroupTag(t *testing.T) {
	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb = mocks_graph.NewMockGraph(mockCtrl)
		resources   = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
	)

	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/asset-group-tags/%s"
	assetGroupTagId := "5"

	t.Run("successfully got asset group tag", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), int(5)).Return(model.AssetGroupTag{
			ID:             5,
			Name:           "test tag 5",
			Description:    "some description",
			RequireCertify: null.BoolFrom(true),
		}, nil)
		mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
			EmailAddress: null.StringFrom("spam@exaple.com"),
		}, nil).Times(2)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", fmt.Sprintf(endpoint, assetGroupTagId), nil)
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupTagID: assetGroupTagId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.GetAssetGroupTag)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		bodyBytes, err := io.ReadAll(response.Body)
		require.Nil(t, err)

		var temp struct {
			Data struct {
				Tag model.AssetGroupTag `json:"tag"`
			} `json:"data"`
		}
		err = json.Unmarshal(bodyBytes, &temp)
		require.Nil(t, err)

		parsedTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		require.Nil(t, err)

		require.Equal(t, model.AssetGroupTag{
			ID:             5,
			Type:           0,
			KindId:         0,
			Name:           "test tag 5",
			Description:    "some description",
			CreatedAt:      parsedTime,
			CreatedBy:      "spam@exaple.com",
			UpdatedAt:      parsedTime,
			UpdatedBy:      "spam@exaple.com",
			DeletedAt:      null.Time{},
			DeletedBy:      null.String{},
			Position:       null.Int32{},
			RequireCertify: null.BoolFrom(true),
		}, temp.Data.Tag)
	})

	t.Run("asset group tag doesn't exist error", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).Return(model.AssetGroupTag{}, database.ErrNotFound)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", fmt.Sprintf(endpoint, assetGroupTagId), nil)
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupTagID: assetGroupTagId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.GetAssetGroupTag)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("id malformed error", func(t *testing.T) {
		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", fmt.Sprintf(endpoint, assetGroupTagId), nil)
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/asset-group-tags/{5}", resources.GetAssetGroupTag).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		require.Equal(t, http.StatusNotFound, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsIDMalformed)
	})
}

func TestDatabase_GetAssetGroupTagSelector(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks_db.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
		handler   = http.HandlerFunc(resources.GetAssetGroupTagSelector)
		endpoint  = fmt.Sprintf("/api/v2/asset-group-tags/{%s}/selectors/{%s}", api.URIPathVariableAssetGroupTagID, api.URIPathVariableAssetGroupTagSelectorID)

		assetGroupTagId = "5"
		selectorId      = "7"
		selector        = model.AssetGroupTagSelector{
			AssetGroupTagId: 5,
			ID:              7,
			Name:            "Selector 7",
			Description:     "777",
			CreatedBy:       "spam@exaple.com",
			UpdatedBy:       "spam@exaple.com",
		}
	)

	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.Nil(t, err)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	t.Run("successfully got asset group tag selector", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 5).Return(model.AssetGroupTag{ID: 5}, nil)
		mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), 7).Return(selector, nil)
		mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
			EmailAddress: null.StringFrom("spam@exaple.com"),
		}, nil).Times(2)

		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupTagID: assetGroupTagId, api.URIPathVariableAssetGroupTagSelectorID: selectorId})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		bodyBytes, err := io.ReadAll(response.Body)
		require.Nil(t, err)

		var result struct {
			Data v2.GetSelectorResponse `json:"data"`
		}
		err = json.Unmarshal(bodyBytes, &result)
		require.Nil(t, err)

		require.Equal(t, selector, result.Data.Selector)
	})

	t.Run("asset group tag doesn't exist error", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).Return(model.AssetGroupTag{}, database.ErrNotFound)

		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupTagID: assetGroupTagId, api.URIPathVariableAssetGroupTagSelectorID: selectorId})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("asset group tag selector doesn't exist error", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 5).Return(model.AssetGroupTag{ID: 5}, nil)
		mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), 7).Return(model.AssetGroupTagSelector{}, database.ErrNotFound)

		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupTagID: assetGroupTagId, api.URIPathVariableAssetGroupTagSelectorID: selectorId})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("asset group tag id malformed error", func(t *testing.T) {
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupTagID: "", api.URIPathVariableAssetGroupTagSelectorID: selectorId})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusNotFound, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsIDMalformed)
	})

	t.Run("selector id malformed error", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 5).Return(model.AssetGroupTag{}, nil)

		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupTagID: assetGroupTagId, api.URIPathVariableAssetGroupTagSelectorID: ""})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusNotFound, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsIDMalformed)
	})

	t.Run("asset group tag id does not equal selector id", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 5).Return(model.AssetGroupTag{ID: 5}, nil)
		mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), 7).Return(model.AssetGroupTagSelector{AssetGroupTagId: 7}, nil)

		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupTagID: assetGroupTagId, api.URIPathVariableAssetGroupTagSelectorID: selectorId})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusNotFound, response.Code)
		require.Contains(t, response.Body.String(), "selector is not part of asset group tag")
	})
}

func TestResources_UpdateAssetGroupTagSelector(t *testing.T) {
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
		NewHarness(t, resourcesInst.UpdateAssetGroupTagSelector).
		Run([]apitest.Case{
			{
				Name: "BadRequest",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.BodyString(input, `{"name":["BadRequest"]}`)
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponsePayloadUnmarshalError)
				},
			},
			{
				Name: "MissingTagUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						AutoCertify: null.BoolFrom(false),
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "InvalidTagUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "non-numeric")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						AutoCertify: null.BoolFrom(false),
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "MissingSelectorUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "InvalidSelectorUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "non-numeric")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "NonExistentTagUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1234")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, database.ErrNotFound).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "NonExistentSelectorUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1234")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1}, database.ErrNotFound).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "CannotUpdateNameOnDefaultSelector",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name: "TestSelector",
					})
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1, IsDefault: true}, nil).Times(1)

					mockGraphDb.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusForbidden)
					apitest.BodyContains(output, "default selectors only support modifying auto_certify and disabled_at")
				},
			},
			{
				Name: "DatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "this should be a string of cypher"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().UpdateAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{}, errors.New("failure")).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1}, nil).Times(1)
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
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					value, _ := types.NewJSONBObject(map[string]any{"enabled": true})
					mockDB.EXPECT().
						GetConfigurationParameter(gomock.Any(), gomock.Any()).
						Return(appcfg.Parameter{Key: appcfg.ScheduledAnalysis, Value: value}, nil).Times(1)
					mockDB.EXPECT().
						UpdateAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{Name: "TestSelector"}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1}, nil).Times(1)

					mockGraphDb.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SetEmptyDescription",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					value, _ := types.NewJSONBObject(map[string]any{"enabled": true})
					mockDB.EXPECT().
						GetConfigurationParameter(gomock.Any(), gomock.Any()).
						Return(appcfg.Parameter{Key: appcfg.ScheduledAnalysis, Value: value}, nil).Times(1)
					mockDB.EXPECT().
						UpdateAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Cond(func(s model.AssetGroupTagSelector) bool {
							return s.Description == ""
						})).
						Return(model.AssetGroupTagSelector{Name: "TestSelector"}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1}, nil).Times(1)

					mockGraphDb.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetAssetGroupTagSelectors(t *testing.T) {
	var (
		mockCtrl      = gomock.NewController(t)
		mockDB        = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb   = mocks_graph.NewMockGraph(mockCtrl)
		resourcesInst = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resourcesInst.GetAssetGroupTagSelectors).
		Run([]apitest.Case{
			{
				Name: "Bad Request - Invalid Asset Group Tag ID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "DB error - GetAssetGroupTag",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, database.ErrNotFound).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "DB error - GetAssetGroupTagSelectorsByTagId",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTagSelectorsByTagId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelectors{}, errors.New("some error")).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.BodyStruct(input, model.AssetGroupTagSelector{
						Name:        "TestSelector",
						Description: "Test selector description",
						Seeds: []model.SelectorSeed{
							{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
						},
						IsDefault:   false,
						AutoCertify: null.BoolFrom(false),
					})
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTagSelectorsByTagId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelectors{{Name: "Test1", AssetGroupTagId: 1}}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "Test1")
				},
			},
			{
				Name: "Success with counts",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, api.QueryParameterIncludeCounts, "true")
				},
				Setup: func() {
					assetGroupTag := model.AssetGroupTag{
						KindId: 50,
					}
					agtSelectors := model.AssetGroupTagSelectors{
						{ID: 1, Name: "TestSelector1", AssetGroupTagId: 1},
					}
					agtSelectorNodes := []model.AssetGroupSelectorNode{
						{SelectorId: 1, NodeId: 100, Source: model.AssetGroupSelectorNodeSourceSeed},
						{SelectorId: 1, NodeId: 101, Source: model.AssetGroupSelectorNodeSourceSeed},
					}

					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil).Times(1)

					mockDB.EXPECT().
						GetAssetGroupTagSelectorsByTagId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(agtSelectors, nil).Times(1)

					mockDB.EXPECT().
						GetSelectorNodesBySelectorIds(gomock.Any(), agtSelectors[0].ID).
						Return(agtSelectorNodes, nil).Times(1)

					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), query.And(
							query.KindIn(query.Node(), assetGroupTag.ToKind()),
							query.InIDs(query.NodeID(), agtSelectorNodes[0].NodeId, agtSelectorNodes[1].NodeId),
						)).
						Return(int64(2), nil)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, "\"counts\":{\"members\":2}")
				},
			},
		})
}

func TestResources_DeleteAssetGroupTagSelector(t *testing.T) {
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
		NewHarness(t, resourcesInst.DeleteAssetGroupTagSelector).
		Run([]apitest.Case{
			{
				Name: "MissingTagUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "InvalidTagUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "non-numeric")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "MissingSelectorUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "InvalidSelectorUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "non-numeric")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "NonExistentTagUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1234")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, database.ErrNotFound).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "NonExistentSelectorUrlId",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1234")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1}, database.ErrNotFound).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "CannotDeleteDefaultSelector",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1, IsDefault: true}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusForbidden)
					apitest.BodyContains(output, "cannot delete a default selector")
				},
			},
			{
				Name: "DatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().DeleteAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(errors.New("failure")).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1}, nil).Times(1)
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
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Setup: func() {
					value, _ := types.NewJSONBObject(map[string]any{"enabled": true})
					mockDB.EXPECT().
						GetConfigurationParameter(gomock.Any(), gomock.Any()).
						Return(appcfg.Parameter{Key: appcfg.ScheduledAnalysis, Value: value}, nil).Times(1)
					mockDB.EXPECT().
						DeleteAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{ID: 1}, nil).Times(1)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{AssetGroupTagId: 1}, nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
		})
}

func TestResources_GetAssetGroupTagSelectorsByTagId(t *testing.T) {
	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb = mocks_graph.NewMockGraph(mockCtrl)
		resources   = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
		assetGroupTag = model.AssetGroupTag{ID: 1, Name: "Tier Zero"}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetAssetGroupTagMemberCountsByKind).
		Run([]apitest.Case{
			{
				Name: "InvalidAssetGroupTagID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "invalid")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "DatabaseGetAssetGroupTagError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, errors.New("GetAssetGroupTag fail"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, api.ErrorResponseDetailsInternalServerError)
				},
			},
			{
				Name: "GraphDatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetPrimaryNodeKindCounts(gomock.Any(), gomock.Any()).
						Return(map[string]int{}, fmt.Errorf("GetAssetGroupTag Nodes fail"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetPrimaryNodeKindCounts(gomock.Any(), gomock.Any()).
						Return(map[string]int{ad.Domain.String(): 2}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					result := v2.GetAssetGroupTagMemberCountsResponse{}
					apitest.UnmarshalData(output, &result)
					require.Equal(t, 2, result.TotalCount)
					require.Equal(t, 2, result.Counts[ad.Domain.String()])
				},
			},
		})
}

func TestResources_GetAssetGroupTagMemberInfo(t *testing.T) {
	var (
		mockCtrl      = gomock.NewController(t)
		mockDB        = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb   = mocks_graph.NewMockGraph(mockCtrl)
		resourcesInst = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
		testNode = &graph.Node{
			ID:           0,
			Kinds:        graph.StringsToKinds([]string{"kind"}),
			AddedKinds:   graph.StringsToKinds([]string{"added kind"}),
			DeletedKinds: graph.StringsToKinds([]string{"deleted kind"}),
			Properties:   &graph.Properties{Map: map[string]any{"prop": 1}},
		}
		testNode2 = &graph.Node{
			ID:           0,
			Kinds:        graph.StringsToKinds([]string{"kind"}),
			AddedKinds:   graph.StringsToKinds([]string{"added kind"}),
			DeletedKinds: graph.StringsToKinds([]string{"deleted kind"}),
			Properties:   &graph.Properties{},
		}
		testSelectors = model.AssetGroupTagSelectors{model.AssetGroupTagSelector{Name: "test"}}
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resourcesInst.GetAssetGroupTagMemberInfo).
		Run([]apitest.Case{
			{
				Name: "Bad Request - Invalid Asset Group Tag ID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "Bad Request - Invalid Member ID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagMemberID, "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "DB error - GetAssetGroupTag",
				Input: func(input *apitest.Input) {

					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagMemberID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, database.ErrNotFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "DB error - GetSelectorsByMemberId",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagMemberID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().GetSelectorsByMemberId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelectors{}, database.ErrNotFound)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)

				},
			},
			{
				Name: "Not found error - no selectors",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagMemberID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().GetSelectorsByMemberId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelectors{}, nil)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)

				},
			},
			{
				Name: "Success - properties in response",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagMemberID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().GetSelectorsByMemberId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testSelectors, nil)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil)
					mockGraphDb.EXPECT().FetchNodeByGraphId(gomock.Any(), gomock.Any()).
						Return(testNode, nil)
				},
				Test: func(output apitest.Output) {
					resp := v2.MemberInfoResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.BodyContains(output, "prop")
					apitest.BodyContains(output, "test")
					apitest.Equal(output, 1, len(resp.Member.Properties))
				},
			},
			{
				Name: "Success - no props in response",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagMemberID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().GetSelectorsByMemberId(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(testSelectors, nil)
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil)
					mockGraphDb.EXPECT().FetchNodeByGraphId(gomock.Any(), gomock.Any()).
						Return(testNode2, nil)
				},
				Test: func(output apitest.Output) {
					resp := v2.MemberInfoResponse{}
					apitest.StatusCode(output, http.StatusOK)
					apitest.UnmarshalData(output, &resp)
					apitest.BodyContains(output, "test")
					apitest.Equal(output, 0, len(resp.Member.Properties))
				},
			},
		})
}

func Test_GetAssetGroupMembersByTag(t *testing.T) {
	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb = mocks_graph.NewMockGraph(mockCtrl)
		resources   = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
		assetGroupTag = model.AssetGroupTag{ID: 1, Name: "Tier Zero"}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetAssetGroupMembersByTag).
		Run([]apitest.Case{
			{
				Name: "InvalidAssetGroupTagID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "invalid")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "Fail with non-sortable column",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, "sort_by", "invalidColumn")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					params := url.Values{}
					params.Add("sort_by", "invalidColumn")
					_, err := api.ParseGraphSortParameters(v2.AssetGroupMember{}, params)
					require.ErrorIs(t, err, api.ErrResponseDetailsCriteriaNotSortable)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Success with sort by id",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, "sort_by", "id")
				},
				Setup: func() {
					params := url.Values{}
					params.Add("sort_by", "id")

					orderCriteria, err := api.ParseGraphSortParameters(v2.AssetGroupMember{}, params)
					require.Nil(t, err)

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(orderCriteria, gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with sort by objectid",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, "sort_by", "objectid")
				},
				Setup: func() {
					params := url.Values{}
					params.Add("sort_by", "objectid")

					orderCriteria, err := api.ParseGraphSortParameters(v2.AssetGroupMember{}, params)
					require.Nil(t, err)

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(orderCriteria, gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with sort by name",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, "sort_by", "name")
				},
				Setup: func() {
					params := url.Values{}
					params.Add("sort_by", "name")

					orderCriteria, err := api.ParseGraphSortParameters(v2.AssetGroupMember{}, params)
					require.Nil(t, err)

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(orderCriteria, gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with limit",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, "limit", "5")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(5)).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with skip",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, "skip", "100")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Eq(100), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{
							{
								ID:    1,
								Kinds: []graph.Kind{ad.User},
								Properties: graph.AsProperties(map[string]any{
									"objectid": "OID-1",
									"name":     "node1",
								})},
							{
								ID:    2,
								Kinds: []graph.Kind{ad.Group},
								Properties: graph.AsProperties(map[string]any{
									"objectid": "OID-2",
									"name":     "node2",
								})},
						}, nil)
					mockGraphDb.EXPECT().
						CountNodesByKind(gomock.Any(), gomock.Any()).
						Return(int64(2), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					expected := v2.GetAssetGroupMembersResponse{
						Members: []v2.AssetGroupMember{
							{
								NodeId:      1,
								ObjectID:    "OID-1",
								PrimaryKind: "User",
								Name:        "node1",
							},
							{
								NodeId:      2,
								ObjectID:    "OID-2",
								PrimaryKind: "Group",
								Name:        "node2",
							},
						},
					}
					result := v2.GetAssetGroupMembersResponse{}
					apitest.UnmarshalData(output, &result)
					require.Equal(t, expected, result)
				},
			},
		})
}

func Test_GetAssetGroupMembersBySelector(t *testing.T) {
	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb = mocks_graph.NewMockGraph(mockCtrl)
		resources   = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
		assetGroupSelector = model.AssetGroupTagSelector{ID: 1, AssetGroupTagId: 1, Name: "Enterprise Domain Controllers"}
		assetGroupTag      = model.AssetGroupTag{ID: 1, Name: "Tier Zero"}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.GetAssetGroupMembersBySelector).
		Run([]apitest.Case{
			{
				Name: "MissingAssetGroupTagID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "InvalidAssetGroupTagID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "non-numeric")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "MissingAssetGroupSelectorID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "InvalidAssetGroupSelectorID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "non-numeric")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "NonExistentAssetGroupTagID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1234")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, database.ErrNotFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "NonExistentAssetGroupSelectorID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1234")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTag{}, nil)
					mockDB.EXPECT().GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelector{}, database.ErrNotFound)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNotFound)
					apitest.BodyContains(output, api.ErrorResponseDetailsResourceNotFound)
				},
			},
			{
				Name: "DatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Setup: func() {
					params := url.Values{}
					_, err := api.ParseGraphSortParameters(v2.AssetGroupMember{}, params)
					require.Nil(t, err)

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIds(gomock.Any(), gomock.Any()).
						Return([]model.AssetGroupSelectorNode{}, errors.New("some error")).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success with sort by id",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.AddQueryParam(input, "sort_by", "id")
				},
				Setup: func() {
					params := url.Values{}
					params.Add("sort_by", "id")

					orderCriteria, err := api.ParseGraphSortParameters(v2.AssetGroupMember{}, params)
					require.Nil(t, err)

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIds(gomock.Any(), gomock.Any()).
						Return([]model.AssetGroupSelectorNode{}, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(orderCriteria, gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with sort by objectid",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.AddQueryParam(input, "sort_by", "objectid")
				},
				Setup: func() {
					params := url.Values{}
					params.Add("sort_by", "objectid")

					orderCriteria, err := api.ParseGraphSortParameters(v2.AssetGroupMember{}, params)
					require.Nil(t, err)

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIds(gomock.Any(), gomock.Any()).
						Return([]model.AssetGroupSelectorNode{}, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(orderCriteria, gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with sort by name",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.AddQueryParam(input, "sort_by", "name")
				},
				Setup: func() {
					params := url.Values{}
					params.Add("sort_by", "name")

					orderCriteria, err := api.ParseGraphSortParameters(v2.AssetGroupMember{}, params)
					require.Nil(t, err)

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIds(gomock.Any(), gomock.Any()).
						Return([]model.AssetGroupSelectorNode{}, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(orderCriteria, gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with limit",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.AddQueryParam(input, "limit", "5")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIds(gomock.Any(), gomock.Any()).
						Return([]model.AssetGroupSelectorNode{}, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(5)).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with skip",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.AddQueryParam(input, "skip", "100")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIds(gomock.Any(), gomock.Any()).
						Return([]model.AssetGroupSelectorNode{}, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Eq(100), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIds(gomock.Any(), gomock.Any()).
						Return([]model.AssetGroupSelectorNode{}, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{
							{
								ID:    1,
								Kinds: []graph.Kind{ad.User},
								Properties: graph.AsProperties(map[string]any{
									"objectid": "OID-1",
									"name":     "node1",
								})},
							{
								ID:    2,
								Kinds: []graph.Kind{ad.Group},
								Properties: graph.AsProperties(map[string]any{
									"objectid": "OID-2",
									"name":     "node2",
								})},
						}, nil)
					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(2), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					expected := v2.GetAssetGroupMembersResponse{
						Members: []v2.AssetGroupMember{
							{
								NodeId:      1,
								ObjectID:    "OID-1",
								PrimaryKind: "User",
								Name:        "node1",
							},
							{
								NodeId:      2,
								ObjectID:    "OID-2",
								PrimaryKind: "Group",
								Name:        "node2",
							},
						},
					}
					result := v2.GetAssetGroupMembersResponse{}
					apitest.UnmarshalData(output, &result)
					require.Equal(t, expected, result)
				},
			},
		})
}

func TestResources_PreviewSelectors(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		mockDB         = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphQuery = mocks_graph.NewMockGraph(mockCtrl)
		mockGraphDb    = graphmocks.NewMockDatabase(mockCtrl)
		resourcesInst  = v2.Resources{
			DB:         mockDB,
			Graph:      mockGraphDb,
			GraphQuery: mockGraphQuery,
		}
		user    = setupUser()
		userCtx = setupUserCtx(user)
	)

	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resourcesInst.PreviewSelectors).
		Run([]apitest.Case{
			{
				Name: "Bad Limit Query Param",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, model.PaginationQueryParameterLimit, "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Bad Request - Error Decoding Body",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.BodyString(input, `{"seeds":["BadRequest"]}`)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Bad Request - Error Validating Seeds",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.BodyStruct(input, v2.PreviewSelectorBody{
						Seeds: model.SelectorSeeds{{Type: model.SelectorTypeCypher, Value: "invalid cypher"}},
					})
				},
				Setup: func() {
					mockGraphQuery.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, errors.New("failure")).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Internal Server Error - Bad User ",
				Input: func(input *apitest.Input) {
					apitest.BodyStruct(input, v2.PreviewSelectorBody{Seeds: model.SelectorSeeds{}})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "unknown user")
				},
			},
			{
				Name: "Bad Request - validateSelectorSeeds",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.BodyStruct(input, v2.PreviewSelectorBody{
						Seeds: model.SelectorSeeds{},
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "seeds are required")
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.AddQueryParam(input, model.PaginationQueryParameterLimit, "10")
					apitest.BodyStruct(input, v2.PreviewSelectorBody{
						Seeds: model.SelectorSeeds{
							{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
						},
					})
				},
				Setup: func() {
					mockGraphQuery.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, nil).Times(1)
					mockGraphDb.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}
