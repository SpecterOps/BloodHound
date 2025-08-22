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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	uuid2 "github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	mocks_db "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	mocks_graph "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
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
							SQLString: "type = " + strconv.Itoa(int(model.AssetGroupTagTypeLabel)),
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
							SQLString: "type = " + strconv.Itoa(int(model.AssetGroupTagTypeTier)),
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
					mockDB.EXPECT().UpdateAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
						UpdateAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
						UpdateAssetGroupTagSelector(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Cond(func(s model.AssetGroupTagSelector) bool {
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
						GetAssetGroupTagSelectorsByTagId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelectors{}, 0, errors.New("some error")).Times(1)
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
						GetAssetGroupTagSelectorsByTagId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(model.AssetGroupTagSelectors{{Name: "Test1", AssetGroupTagId: 1}}, 1, nil).Times(1)
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
						GetAssetGroupTagSelectorsByTagId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(agtSelectors, 0, nil).Times(1)

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

func TestResources_GetAssetGroupTagMemberCountsByKind(t *testing.T) {
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
				Name: "Success with environments",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, "environments", "testenv")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetPrimaryNodeKindCounts(gomock.Any(), gomock.Any(), []graph.Criteria{
							query.Or(
								query.In(query.NodeProperty(ad.DomainSID.String()), []string{"testenv"}),
								query.In(query.NodeProperty(azure.TenantID.String()), []string{"testenv"}),
							),
						}).
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
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsColumnNotFilterable)
				},
			},
			{
				Name: "Fail with graph list results failure",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, fmt.Errorf("graph err"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "Error getting members")
				},
			},
			{
				Name: "Fail with graph total count failure",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(0), fmt.Errorf("graph err"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "Error getting member count")
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
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(0), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success with environments",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.AddQueryParam(input, "environments", "testenv")
				},
				Setup: func() {
					nodeFilter := query.And(
						query.KindIn(query.Node(), assetGroupTag.ToKind()),
						query.Or(
							query.In(query.NodeProperty(ad.DomainSID.String()), []string{"testenv"}),
							query.In(query.NodeProperty(azure.TenantID.String()), []string{"testenv"}),
						),
					)
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockGraphDb.EXPECT().
						GetFilteredAndSortedNodesPaginated(gomock.Any(), nodeFilter, gomock.Any(), gomock.Any()).
						Return([]*graph.Node{}, nil)
					mockGraphDb.EXPECT().
						CountFilteredNodes(gomock.Any(), nodeFilter).
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
						CountFilteredNodes(gomock.Any(), gomock.Any()).
						Return(int64(2), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					expected := v2.GetAssetGroupMembersResponse{
						Members: []v2.AssetGroupMember{
							{
								NodeId:          1,
								ObjectID:        "OID-1",
								PrimaryKind:     "User",
								Name:            "node1",
								AssetGroupTagId: 1,
							},
							{
								NodeId:          2,
								ObjectID:        "OID-2",
								PrimaryKind:     "Group",
								Name:            "node2",
								AssetGroupTagId: 1,
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
		defaultSort        = model.Sort{model.SortItem{Column: "node_id", Direction: model.AscendingSortDirection}}
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
						GetSelectorNodesBySelectorIdsFilteredAndPaginated(gomock.Any(), gomock.Any(), defaultSort, 0, v2.AssetGroupTagDefaultLimit, 1).
						Return([]model.AssetGroupSelectorNode{}, 0, errors.New("db error"))
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

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIdsFilteredAndPaginated(gomock.Any(), gomock.Any(), defaultSort, 0, v2.AssetGroupTagDefaultLimit, 1).
						Return([]model.AssetGroupSelectorNode{}, 0, nil)
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

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIdsFilteredAndPaginated(gomock.Any(), gomock.Any(), model.Sort{model.SortItem{Column: "node_object_id", Direction: model.AscendingSortDirection}}, 0, v2.AssetGroupTagDefaultLimit, 1).
						Return([]model.AssetGroupSelectorNode{}, 0, nil)
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

					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIdsFilteredAndPaginated(gomock.Any(), gomock.Any(), model.Sort{model.SortItem{Column: "node_name", Direction: model.AscendingSortDirection}}, 0, v2.AssetGroupTagDefaultLimit, 1).
						Return([]model.AssetGroupSelectorNode{}, 0, nil)
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
						GetSelectorNodesBySelectorIdsFilteredAndPaginated(gomock.Any(), gomock.Any(), defaultSort, 0, 5, 1).
						Return([]model.AssetGroupSelectorNode{}, 0, nil)
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
						GetSelectorNodesBySelectorIdsFilteredAndPaginated(gomock.Any(), gomock.Any(), defaultSort, 100, v2.AssetGroupTagDefaultLimit, 1).
						Return([]model.AssetGroupSelectorNode{}, 0, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Success - environment filter",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagID, "1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupTagSelectorID, "1")
					apitest.AddQueryParam(input, "environments", "testenv")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupTag(gomock.Any(), gomock.Any()).
						Return(assetGroupTag, nil)
					mockDB.EXPECT().
						GetAssetGroupTagSelectorBySelectorId(gomock.Any(), gomock.Any()).
						Return(assetGroupSelector, nil)
					mockDB.EXPECT().
						GetSelectorNodesBySelectorIdsFilteredAndPaginated(gomock.Any(), model.SQLFilter{SQLString: "AND node_environment_id IN ?", Params: []any{[]string{"testenv"}}}, defaultSort, 0, v2.AssetGroupTagDefaultLimit, 1).
						Return([]model.AssetGroupSelectorNode{}, 0, nil)
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
						GetSelectorNodesBySelectorIdsFilteredAndPaginated(gomock.Any(), gomock.Any(), defaultSort, 0, v2.AssetGroupTagDefaultLimit, 1).
						Return([]model.AssetGroupSelectorNode{
							{
								NodeId:          1,
								NodePrimaryKind: ad.User.String(),
								NodeObjectId:    "OID-1",
								NodeName:        "node1",
							},
							{
								NodeId:          2,
								NodePrimaryKind: ad.Group.String(),
								NodeObjectId:    "OID-2",
								NodeName:        "node2",
							}}, 2, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					expected := v2.GetAssetGroupMembersResponse{
						Members: []v2.AssetGroupMember{
							{
								NodeId:          1,
								ObjectID:        "OID-1",
								PrimaryKind:     "User",
								Name:            "node1",
								AssetGroupTagId: 1,
							},
							{
								NodeId:          2,
								ObjectID:        "OID-2",
								PrimaryKind:     "Group",
								Name:            "node2",
								AssetGroupTagId: 1,
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
		user              = setupUser()
		userCtx           = setupUserCtx(user)
		badExpansionValue = model.AssetGroupExpansionMethod(5)
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
				Name: "Bad Request - Invalid Expansion Method",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
					apitest.BodyStruct(input, v2.PreviewSelectorBody{
						Seeds:     model.SelectorSeeds{{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"}},
						Expansion: &badExpansionValue,
					})
				},
				Setup: func() {
					mockGraphQuery.EXPECT().
						PrepareCypherQuery(gomock.Any(), gomock.Any()).
						Return(queries.PreparedQuery{}, nil).Times(1)
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

func TestResources_SearchAssetGroupTags(t *testing.T) {
	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = mocks_db.NewMockDatabase(mockCtrl)
		mockGraphDb = mocks_graph.NewMockGraph(mockCtrl)
		resources   = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraphDb,
		}
		handler  = http.HandlerFunc(resources.SearchAssetGroupTags)
		endpoint = "/api/v2/asset-group-tags/search"
	)

	type WrappedResponse struct {
		Data v2.SearchAssetGroupTagsResponse `json:"data"`
	}

	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, nil)
	require.Nil(t, err)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	t.Run("cannot decode request body error", func(t *testing.T) {

		reqBody := `{"query":`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusBadRequest, response.Code)
	})
	t.Run("invalid tag type error", func(t *testing.T) {

		reqBody := `{"query": "test", "tag_type": 5}`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "valid tag_type is required")
	})
	t.Run("empty query error", func(t *testing.T) {

		reqBody := `{"query": "", "tag_type": 2}`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "search query must be at least 3 characters long")
	})
	t.Run("get tags db error", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTags(gomock.Any(), gomock.Any()).Return(model.AssetGroupTags{}, errors.New("db error"))

		reqBody := `{"query": "test", "tag_type": 1}`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusInternalServerError, response.Code)
	})
	t.Run("get selectors db error", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTags(gomock.Any(), gomock.Any()).Return(model.AssetGroupTags{{Name: "test tier", Type: 1}}, nil)
		mockDB.EXPECT().GetAssetGroupTagSelectors(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AssetGroupTagSelectors{}, errors.New("db error"))

		reqBody := `{"query": "test", "tag_type": 1}`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("success - query by name type tier", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTags(gomock.Any(), gomock.Any()).Return(model.AssetGroupTags{{Name: "test tier", Type: 1}}, nil)
		mockDB.EXPECT().GetAssetGroupTagSelectors(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AssetGroupTagSelectors{{Name: "test selector"}}, nil)
		mockGraphDb.EXPECT().GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*graph.Node{
				{
					ID:    1,
					Kinds: []graph.Kind{graph.StringKind("Tag_test_tier")},
					Properties: graph.AsProperties(map[string]any{
						"objectid": "ID-1",
						"name":     "test1",
					})},
				{
					ID:    2,
					Kinds: []graph.Kind{graph.StringKind("Tag_test_tier")},
					Properties: graph.AsProperties(map[string]any{
						"objectid": "ID-2",
						"name":     "test2",
					})},
			}, nil)

		reqBody := `{"query": "test", "tag_type": 1}`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		expected := WrappedResponse{v2.SearchAssetGroupTagsResponse{
			Tags:      model.AssetGroupTags{{Name: "test tier", Type: 1}},
			Selectors: model.AssetGroupTagSelectors{{Name: "test selector"}},
			Members: []v2.AssetGroupMember{
				{
					NodeId:      1,
					ObjectID:    "ID-1",
					PrimaryKind: "Unknown",
					Name:        "test1",
				},
				{
					NodeId:      2,
					ObjectID:    "ID-2",
					PrimaryKind: "Unknown",
					Name:        "test2",
				},
			},
		},
		}

		wrappedResp := WrappedResponse{}
		err := json.Unmarshal(response.Body.Bytes(), &wrappedResp)
		require.NoError(t, err)
		require.Equal(t, expected, wrappedResp)
		require.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("success - query by name type label", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTags(gomock.Any(), gomock.Any()).Return(model.AssetGroupTags{{Name: "test label", Type: 2}}, nil)
		mockDB.EXPECT().GetAssetGroupTagSelectors(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AssetGroupTagSelectors{{Name: "test selector"}}, nil)
		mockGraphDb.EXPECT().GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*graph.Node{
				{
					ID:    1,
					Kinds: []graph.Kind{graph.StringKind("Tag_test_label")},
					Properties: graph.AsProperties(map[string]any{
						"objectid": "ID-1",
						"name":     "test1",
					})},
				{
					ID:    2,
					Kinds: []graph.Kind{graph.StringKind("Tag_test_label")},
					Properties: graph.AsProperties(map[string]any{
						"objectid": "ID-2",
						"name":     "test2",
					})},
			}, nil)

		reqBody := `{"query": "test", "tag_type": 2}`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		expected := WrappedResponse{v2.SearchAssetGroupTagsResponse{
			Tags:      model.AssetGroupTags{{Name: "test label", Type: 2}},
			Selectors: model.AssetGroupTagSelectors{{Name: "test selector"}},
			Members: []v2.AssetGroupMember{
				{
					NodeId:      1,
					ObjectID:    "ID-1",
					PrimaryKind: "Unknown",
					Name:        "test1",
				},
				{
					NodeId:      2,
					ObjectID:    "ID-2",
					PrimaryKind: "Unknown",
					Name:        "test2",
				},
			},
		},
		}

		wrappedResp := WrappedResponse{}
		err := json.Unmarshal(response.Body.Bytes(), &wrappedResp)
		require.NoError(t, err)
		require.Equal(t, expected, wrappedResp)
		require.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("success - query by name and type label and include owned type", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTags(gomock.Any(), gomock.Any()).Return(model.AssetGroupTags{{Name: "test owned label", Type: 2}, {Name: "owned", Type: 3}}, nil)
		mockDB.EXPECT().GetAssetGroupTagSelectors(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AssetGroupTagSelectors{}, nil)
		mockGraphDb.EXPECT().GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*graph.Node{
				{
					ID:    1,
					Kinds: []graph.Kind{graph.StringKind("Tag_test_owned_label")},
					Properties: graph.AsProperties(map[string]any{
						"objectid": "ID-1",
						"name":     "test1",
					})},
				{
					ID:    2,
					Kinds: []graph.Kind{graph.StringKind("Tag_test_owned_label")},
					Properties: graph.AsProperties(map[string]any{
						"objectid": "ID-2",
						"name":     "test2",
					})},
			}, nil)

		reqBody := `{"query": "owned", "tag_type": 2}`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		expected := WrappedResponse{v2.SearchAssetGroupTagsResponse{
			Tags:      model.AssetGroupTags{{Name: "test owned label", Type: 2}, {Name: "owned", Type: 3}},
			Selectors: model.AssetGroupTagSelectors{},
			Members: []v2.AssetGroupMember{
				{
					NodeId:      1,
					ObjectID:    "ID-1",
					PrimaryKind: "Unknown",
					Name:        "test1",
				},
				{
					NodeId:      2,
					ObjectID:    "ID-2",
					PrimaryKind: "Unknown",
					Name:        "test2",
				},
			},
		},
		}

		wrappedResp := WrappedResponse{}
		err := json.Unmarshal(response.Body.Bytes(), &wrappedResp)
		require.NoError(t, err)
		require.Equal(t, expected, wrappedResp)
		require.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("success - query by object id", func(t *testing.T) {
		mockDB.EXPECT().GetAssetGroupTags(gomock.Any(), gomock.Any()).Return(model.AssetGroupTags{{Name: "test tier", Type: 1}}, nil)
		mockDB.EXPECT().GetAssetGroupTagSelectors(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AssetGroupTagSelectors{}, nil)
		mockGraphDb.EXPECT().GetFilteredAndSortedNodesPaginated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*graph.Node{
				{
					ID:    1,
					Kinds: []graph.Kind{graph.StringKind("Tag_test_tier")},
					Properties: graph.AsProperties(map[string]any{
						"objectid": "ID-1234",
						"name":     "test1",
					})},
				{
					ID:    2,
					Kinds: []graph.Kind{graph.StringKind("Tag_test_tier")},
					Properties: graph.AsProperties(map[string]any{
						"objectid": "ID-123456",
						"name":     "test2",
					})},
			}, nil)

		reqBody := `{"query": "123", "tag_type": 1}`

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		expected := WrappedResponse{v2.SearchAssetGroupTagsResponse{
			Tags:      model.AssetGroupTags{},
			Selectors: model.AssetGroupTagSelectors{},
			Members: []v2.AssetGroupMember{
				{
					NodeId:      1,
					ObjectID:    "ID-1234",
					PrimaryKind: "Unknown",
					Name:        "test1",
				},
				{
					NodeId:      2,
					ObjectID:    "ID-123456",
					PrimaryKind: "Unknown",
					Name:        "test2",
				},
			},
		},
		}

		wrappedResp := WrappedResponse{}
		err := json.Unmarshal(response.Body.Bytes(), &wrappedResp)
		require.NoError(t, err)
		require.Equal(t, expected, wrappedResp)
		require.Equal(t, http.StatusOK, response.Code)
	})
}

func TestResources_GetAssetGroupTagHistory(t *testing.T) {
	var (
		mockCtrl      = gomock.NewController(t)
		mockDB        = mocks_db.NewMockDatabase(mockCtrl)
		resourcesInst = v2.Resources{
			DB: mockDB,
		}

		expectedHistoryRecs = []model.AssetGroupHistory{
			{ID: 1, CreatedAt: time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC), Actor: "UUID1", Email: null.StringFrom("user1@domain.com"), Action: model.AssetGroupHistoryActionCreateTag},
			{ID: 2, CreatedAt: time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC), Actor: "UUID2", Email: null.StringFrom("user2@domain.com"), Action: model.AssetGroupHistoryActionUpdateTag},
			{ID: 3, CreatedAt: time.Date(2025, 6, 12, 0, 0, 0, 0, time.UTC), Actor: "UUID1", Email: null.StringFrom("user1@domain.com"), Action: model.AssetGroupHistoryActionCreateSelector},
			{ID: 4, CreatedAt: time.Date(2025, 6, 12, 2, 0, 0, 0, time.UTC), Actor: "UUID2", Email: null.StringFrom("user2@domain.com"), Action: model.AssetGroupHistoryActionDeleteSelector},
		}
	)

	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resourcesInst.GetAssetGroupTagHistory).
		Run([]apitest.Case{
			{
				Name: "Invalid Filter Column",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "invalid_column", "eq:2")
				},
				Setup: func() {

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Filter on created_at",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "created_at", "gt:2025-06-17T00:00:00Z") // model.AssetGroupHistory
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupHistoryRecords(gomock.Any(),
							model.SQLFilter{SQLString: "created_at > '2025-06-17T00:00:00Z'"},
							model.Sort{{Column: "created_at", Direction: model.DescendingSortDirection}},
							0,
							v2.AssetGroupTagDefaultLimit).
						Return([]model.AssetGroupHistory{}, 0, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Bad limit query param",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, model.PaginationQueryParameterLimit, "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Success with limit",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, model.PaginationQueryParameterLimit, "5")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupHistoryRecords(gomock.Any(), gomock.Any(),
							model.Sort{{Column: "created_at", Direction: model.DescendingSortDirection}},
							0,
							5).
						Return([]model.AssetGroupHistory{}, 0, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Bad skip query param",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, model.PaginationQueryParameterSkip, "foo")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Success with skip",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, model.PaginationQueryParameterSkip, "10")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupHistoryRecords(gomock.Any(), gomock.Any(),
							model.Sort{{Column: "created_at", Direction: model.DescendingSortDirection}},
							10,
							v2.AssetGroupTagDefaultLimit).
						Return([]model.AssetGroupHistory{}, 0, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "Parse sort parameters error",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "sort_by", "invalid_column")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "Success with sort",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "sort_by", "created_at")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroupHistoryRecords(gomock.Any(), gomock.Any(),
							model.Sort{{Column: "created_at", Direction: model.AscendingSortDirection}},
							0,
							v2.AssetGroupTagDefaultLimit).
						Return([]model.AssetGroupHistory{}, 0, nil)
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
						GetAssetGroupHistoryRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(expectedHistoryRecs, len(expectedHistoryRecs), nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)

					wrapper := api.ResponseWrapper{}
					apitest.UnmarshalBody(output, &wrapper)

					// unmarshall data as v2.AssetGroupHistoryResp
					dataBytes, err := json.Marshal(wrapper.Data)
					require.NoError(t, err)
					historyRecordsResp := v2.AssetGroupHistoryResp{}
					err = json.Unmarshal(dataBytes, &historyRecordsResp)
					require.NoError(t, err)

					// verify skip, limit and count
					require.Equal(t, 0, wrapper.Skip)
					require.Equal(t, 50, wrapper.Limit)
					require.Equal(t, len(expectedHistoryRecs), wrapper.Count)

					// verify the records are as expected
					require.Equal(t, expectedHistoryRecs, historyRecordsResp.Records)
				},
			},
		})
}

func TestResources_SearchAssetGroupTagHistory(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks_db.MockDatabase
	}

	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}

	type testData struct {
		name         string
		buildRequest func(testName string) *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "cannot decode request body error",
			buildRequest: func(name string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/asset-group-tags-history",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				request.Body = io.NopCloser(strings.NewReader(`Not real json`))

				request.Header.Set("Content-Type", "application/json")

				return request
			},

			setupMocks: func(t *testing.T, mock *mock) {

			},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: `{"errors":[{"context":"","message":"error unmarshalling JSON payload"}],
									"http_status":400,"request_id":"",
									"timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "query less than 3 characters error",
			buildRequest: func(name string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/asset-group-tags-history",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				request.Body = io.NopCloser(strings.NewReader(`{"query": ""}`))

				request.Header.Set("Content-Type", "application/json")

				return request
			},

			setupMocks: func(t *testing.T, mock *mock) {

			},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: `{"errors":[{"context":"","message":"search query must be at least 3 characters long"}],
									"http_status":400,"request_id":"",
									"timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "get asset group history records db error",
			buildRequest: func(name string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/asset-group-tags-history",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				request.Body = io.NopCloser(strings.NewReader(`{"query": "test"}`))

				request.Header.Set("Content-Type", "application/json")

				return request
			},

			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupHistoryRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.AssetGroupHistory{}, 0, errors.New("entity not found"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],
									"http_status":500,"request_id":"",
									"timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "success - query by action",
			buildRequest: func(name string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/asset-group-tags-history",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				request.Body = io.NopCloser(strings.NewReader(`{"query": "UpdateTag"}`))

				request.Header.Set("Content-Type", "application/json")

				return request
			},

			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupHistoryRecords(gomock.Any(), gomock.Cond(func(sqlFilter model.SQLFilter) bool {
					if !strings.Contains(
						sqlFilter.SQLString,
						"actor ILIKE ? OR email ILIKE ? OR action ILIKE ? OR target ILIKE ?",
					) {
						return false
					}

					for _, v := range sqlFilter.Params {
						switch p := v.(type) {
						case string:
							if p == "%UpdateTag%" {
								return true
							}
						case []any:
							for _, inner := range p {
								if s, ok := inner.(string); ok && s == "%UpdateTag%" {
									return true
								}
							}
						}
					}

					return false
				}), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					[]model.AssetGroupHistory{
						{ID: 2, CreatedAt: time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC), Actor: "UUID2", Email: null.StringFrom("user2@domain.com"), Action: model.AssetGroupHistoryActionUpdateTag, AssetGroupTagId: 1},
					},
					1,
					nil)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `
					{
							"count": 1,
							"limit": 50,
							"skip": 0,
							"data": {
								"records": [
								{
									"action": "UpdateTag",
									"actor": "UUID2",
									"asset_group_tag_id": 1,
									"created_at": "2025-06-11T00:00:00Z",
									"email": "user2@domain.com",
									"environment_id": null,
									"id": 2,
									"note": null,
									"target": ""
								}
							]
						}
					}
				`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "success - query by email sort by created_at",
			buildRequest: func(name string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/asset-group-tags-history",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				request.Body = io.NopCloser(strings.NewReader(`{"query": "user1@domain.com"}`))

				request.Header.Set("Content-Type", "application/json")

				return request
			},

			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupHistoryRecords(gomock.Any(), gomock.Cond(func(sqlFilter model.SQLFilter) bool {
					if !strings.Contains(
						sqlFilter.SQLString,
						"actor ILIKE ? OR email ILIKE ? OR action ILIKE ? OR target ILIKE ?",
					) {
						return false
					}

					for _, v := range sqlFilter.Params {
						switch p := v.(type) {
						case string:
							if p == "%user1@domain.com%" {
								return true
							}
						case []any:
							for _, inner := range p {
								if s, ok := inner.(string); ok && s == "%user1@domain.com%" {
									return true
								}
							}
						}
					}

					return false
				}), model.Sort{{Column: "created_at", Direction: model.DescendingSortDirection}}, gomock.Any(), gomock.Any()).Return(
					[]model.AssetGroupHistory{
						{ID: 4, CreatedAt: time.Date(2025, 6, 12, 2, 0, 0, 0, time.UTC), Actor: "UUID1", Email: null.StringFrom("user1@domain.com"), Action: model.AssetGroupHistoryActionDeleteSelector},
						{ID: 3, CreatedAt: time.Date(2025, 6, 12, 0, 0, 0, 0, time.UTC), Actor: "UUID1", Email: null.StringFrom("user1@domain.com"), Action: model.AssetGroupHistoryActionCreateSelector},
						{ID: 2, CreatedAt: time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC), Actor: "UUID1", Email: null.StringFrom("user1@domain.com"), Action: model.AssetGroupHistoryActionUpdateTag},
						{ID: 1, CreatedAt: time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC), Actor: "UUID1", Email: null.StringFrom("user1@domain.com"), Action: model.AssetGroupHistoryActionCreateTag},
					},
					4,
					nil)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `
					{
							"count": 4,
							"limit": 50,
							"skip": 0,
							"data": {
								"records": [
								{
									"action": "DeleteSelector",
									"actor": "UUID1",
									"asset_group_tag_id": 0,
									"created_at": "2025-06-12T02:00:00Z",
									"email": "user1@domain.com",
									"environment_id": null,
									"id": 4,
									"note": null,
									"target": ""
								},
								{
									"action": "CreateSelector",
									"actor": "UUID1",
									"asset_group_tag_id": 0,
									"created_at": "2025-06-12T00:00:00Z",
									"email": "user1@domain.com",
									"environment_id": null,
									"id": 3,
									"note": null,
									"target": ""
								},
								{
									"action": "UpdateTag",
									"actor": "UUID1",
									"asset_group_tag_id": 0,
									"created_at": "2025-06-11T00:00:00Z",
									"email": "user1@domain.com",
									"environment_id": null,
									"id": 2,
									"note": null,
									"target": ""
								},
								{
									"action": "CreateTag",
									"actor": "UUID1",
									"asset_group_tag_id": 0,
									"created_at": "2025-06-10T00:00:00Z",
									"email": "user1@domain.com",
									"environment_id": null,
									"id": 1,
									"note": null,
									"target": ""
								}
							]
						}
					}
				`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks_db.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest(t.Name())
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				DB:         mocks.mockDatabase,
				Authorizer: auth.NewAuthorizer(mocks.mockDatabase),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/asset-group-tags-history", resources.SearchAssetGroupTagHistory).Methods("POST")

			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_CertifyMembers(t *testing.T) {
	t.Parallel()
	var (
		endpoint                           = "/api/v2/asset-group-tags/members/certification"
		mockAssetGroupSelectorNodeExpanded = model.AssetGroupSelectorNodeExpanded{
			AssetGroupSelectorNode: model.AssetGroupSelectorNode{
				SelectorId:        1,
				NodeId:            graph.ID(1),
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
				NodePrimaryKind:   "",
				NodeEnvironmentId: "",
				NodeObjectId:      "",
				NodeName:          "",
			},
			AssetGroupTagId: 1,
			Position:        1,
		}
		mockAssetGroupSelectorNodeExpandedAlreadyCertified = model.AssetGroupSelectorNodeExpanded{
			AssetGroupSelectorNode: model.AssetGroupSelectorNode{
				SelectorId:        1,
				NodeId:            graph.ID(1),
				Certified:         model.AssetGroupCertificationManual,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
				NodePrimaryKind:   "",
				NodeEnvironmentId: "",
				NodeObjectId:      "",
				NodeName:          "",
			},
			AssetGroupTagId: 1,
			Position:        1,
		}
	)

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	type mock struct {
		mockDatabase *mocks_db.MockDatabase
	}

	type expected struct {
		responseBody string
		responseCode int
	}

	type apiError struct {
		Context    any    `json:"context"`
		Message    string `json:"message"`
		HttpStatus int    `json:"http_status"`
		RequestId  int    `json:"request_id"`
		Timestamp  any    `json:"timestamp"`
	}

	type apiResponse struct {
		Errors []apiError `json:"errors"`
	}

	type testData struct {
		name       string
		request    *http.Request
		setupMocks func(t *testing.T, mock *mock)
		expected   expected
	}

	tt := []testData{
		{
			name:       "cannot decode request body error",
			request:    httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(`{"member_ids":`)),
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: api.ErrorResponsePayloadUnmarshalError,
			}},
		{
			name:       "no user error",
			request:    httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(`{"member_ids": [1,2,3], "action": 5}`)),
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: api.ErrorResponseUnknownUser,
			}},
		{
			name:       "invalid certification type",
			request:    httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(`{"member_ids": [1,2,3], "action": 5}`)),
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: api.ErrorResponseAssetGroupCertTypeInvalid,
			}},

		{
			name:       "no member IDs",
			request:    httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: api.ErrorResponseAssetGroupMemberIDsRequired,
			}},

		{
			name:    "get asset group tag selector nodes db error",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1,2,3], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{}, errors.New("db error"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: api.ErrorResponseDetailsInternalServerError,
			}},

		{
			name:    "get asset group tag selector nodes db transaction error",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1,2,3], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{mockAssetGroupSelectorNodeExpanded}, nil)
				mock.mockDatabase.EXPECT().BeginTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: api.ErrorResponseDetailsInternalServerError,
			}},

		{
			name:    "get asset group tag selector nodes db update certification error",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1,2,3], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{
					mockAssetGroupSelectorNodeExpanded}, nil)
				mock.mockDatabase.EXPECT().BeginTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
					fakeTx := &gorm.DB{}
					return fn(fakeTx)
				})
				mock.mockDatabase.EXPECT().UpdateCertificationBySelectorNode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: api.ErrorResponseDetailsInternalServerError,
			}},
		{
			name:    "get asset group tag selector nodes multiple db update certification error",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1,2,3], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{
					mockAssetGroupSelectorNodeExpanded, mockAssetGroupSelectorNodeExpanded, mockAssetGroupSelectorNodeExpanded}, nil)
				mock.mockDatabase.EXPECT().BeginTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
					fakeTx := &gorm.DB{}
					return fn(fakeTx)
				})
				mock.mockDatabase.EXPECT().UpdateCertificationBySelectorNode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Return(errors.New("error 1")).Return(errors.New("error 2"))

			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: api.ErrorResponseDetailsInternalServerError,
			}},
		{
			name:    "get asset group tag selector nodes db create history record error",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1,2,3], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{
					mockAssetGroupSelectorNodeExpanded}, nil)
				mock.mockDatabase.EXPECT().BeginTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
					fakeTx := &gorm.DB{}
					return fn(fakeTx)
				})
				mock.mockDatabase.EXPECT().UpdateCertificationBySelectorNode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				mock.mockDatabase.EXPECT().CreateAssetGroupHistoryRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: api.ErrorResponseDetailsInternalServerError,
			}},

		{
			name:    "success - manual cert action",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{
					mockAssetGroupSelectorNodeExpanded}, nil)
				mock.mockDatabase.EXPECT().BeginTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
					fakeTx := &gorm.DB{}
					return fn(fakeTx)
				})
				mock.mockDatabase.EXPECT().UpdateCertificationBySelectorNode(gomock.Any(), mockAssetGroupSelectorNodeExpanded.SelectorId, model.AssetGroupCertificationManual, gomock.Any(), mockAssetGroupSelectorNodeExpanded.NodeId)
				mock.mockDatabase.EXPECT().CreateAssetGroupHistoryRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: "",
			}},
		{
			name:    "success - revoke cert action",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1], "action": %d}`, model.AssetGroupCertificationRevoked))),
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{
					mockAssetGroupSelectorNodeExpanded}, nil)
				mock.mockDatabase.EXPECT().BeginTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
					fakeTx := &gorm.DB{}
					return fn(fakeTx)
				})
				mock.mockDatabase.EXPECT().UpdateCertificationBySelectorNode(gomock.Any(), mockAssetGroupSelectorNodeExpanded.SelectorId, model.AssetGroupCertificationRevoked, gomock.Any(), mockAssetGroupSelectorNodeExpanded.NodeId)
				mock.mockDatabase.EXPECT().CreateAssetGroupHistoryRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: "",
			}},
		{
			name:    "success - update 3 entries, but only create one history record",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{
					mockAssetGroupSelectorNodeExpanded, mockAssetGroupSelectorNodeExpanded, mockAssetGroupSelectorNodeExpanded}, nil)
				mock.mockDatabase.EXPECT().BeginTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
					fakeTx := &gorm.DB{}
					return fn(fakeTx)
				})
				mock.mockDatabase.EXPECT().UpdateCertificationBySelectorNode(gomock.Any(), mockAssetGroupSelectorNodeExpanded.SelectorId, model.AssetGroupCertificationManual, gomock.Any(), mockAssetGroupSelectorNodeExpanded.NodeId).Times(3)
				mock.mockDatabase.EXPECT().CreateAssetGroupHistoryRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: "",
			}},
		{
			name:    "success - no updates",
			request: httptest.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPost, endpoint, strings.NewReader(fmt.Sprintf(`{"member_ids": [1], "action": %d}`, model.AssetGroupCertificationManual))),
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetAssetGroupSelectorNodeExpandedIgnoreAutoCertify(gomock.Any(), gomock.Any()).Return([]model.AssetGroupSelectorNodeExpanded{
					mockAssetGroupSelectorNodeExpandedAlreadyCertified}, nil)
				mock.mockDatabase.EXPECT().BeginTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(tx *gorm.DB) error) error {
					fakeTx := &gorm.DB{}
					return fn(fakeTx)
				})
				mock.mockDatabase.EXPECT().UpdateCertificationBySelectorNode(gomock.Any(), mockAssetGroupSelectorNodeExpanded.SelectorId, model.AssetGroupCertificationManual, gomock.Any(), mockAssetGroupSelectorNodeExpanded.NodeId).Times(0)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: "",
			}},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mocks := &mock{
				mockDatabase: mocks_db.NewMockDatabase(ctrl),
			}
			testCase.request.Header.Set("Content-Type", "application/json")
			testCase.setupMocks(t, mocks)
			resources := v2.Resources{
				DB: mocks.mockDatabase,
			}
			response := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc(endpoint, resources.CertifyMembers).Methods(testCase.request.Method)
			router.ServeHTTP(response, testCase.request)

			status, _, body := test.ProcessResponse(t, response)
			require.Equal(t, testCase.expected.responseCode, status)
			if body != "" {
				var response apiResponse
				json.Unmarshal([]byte(body), &response)
				require.Equal(t, testCase.expected.responseBody, response.Errors[0].Message)
			}

		})
	}
}

func TestResources_CreateUpdateCertificationBySelectorNodeInputs(t *testing.T) {
	t.Parallel()

	type testInput struct {
		nodes         []model.AssetGroupSelectorNodeExpanded
		certification model.AssetGroupCertification
	}

	type testData struct {
		name  string
		input testInput
		want  []v2.UpdateCertificationBySelectorNodeInput
	}

	var (
		selectorId1                         = 1
		nodeId1                             = graph.ID(1)
		assetGroupTag1                      = 1
		selectorId2                         = 2
		nodeId2                             = graph.ID(2)
		assetGroupTag2                      = 2
		selectorId3                         = 3
		nodeId3                             = graph.ID(3)
		assetGroupTag3                      = 3
		highestPosition                     = 1
		mediumPosition                      = 2
		lowestPosition                      = 3
		userEmailAddress                    = null.StringFrom("test@testy.com")
		assetGroupSelectorNode1HighPriority = model.AssetGroupSelectorNodeExpanded{AssetGroupSelectorNode: model.AssetGroupSelectorNode{
			SelectorId:        selectorId1,
			NodeId:            nodeId1,
			Certified:         model.AssetGroupCertificationPending,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			NodePrimaryKind:   "",
			NodeEnvironmentId: "",
			NodeObjectId:      "",
			NodeName:          "",
		},
			AssetGroupTagId: assetGroupTag1,
			Position:        highestPosition,
		}
		assetGroupSelectorNode1MedPriority = model.AssetGroupSelectorNodeExpanded{AssetGroupSelectorNode: model.AssetGroupSelectorNode{
			SelectorId:        selectorId1,
			NodeId:            nodeId1,
			Certified:         model.AssetGroupCertificationPending,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			NodePrimaryKind:   "",
			NodeEnvironmentId: "",
			NodeObjectId:      "",
			NodeName:          "",
		},
			AssetGroupTagId: assetGroupTag1,
			Position:        mediumPosition,
		}
		assetGroupSelectorNode1LowPriority = model.AssetGroupSelectorNodeExpanded{AssetGroupSelectorNode: model.AssetGroupSelectorNode{
			SelectorId:        selectorId1,
			NodeId:            nodeId1,
			Certified:         model.AssetGroupCertificationPending,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			NodePrimaryKind:   "",
			NodeEnvironmentId: "",
			NodeObjectId:      "",
			NodeName:          "",
		},
			AssetGroupTagId: assetGroupTag1,
			Position:        lowestPosition,
		}
		assetGroupSelectorNode1AlreadyCertified = model.AssetGroupSelectorNodeExpanded{AssetGroupSelectorNode: model.AssetGroupSelectorNode{
			SelectorId:        selectorId1,
			NodeId:            nodeId1,
			Certified:         model.AssetGroupCertificationManual,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			NodePrimaryKind:   "",
			NodeEnvironmentId: "",
			NodeObjectId:      "",
			NodeName:          "",
		},
			AssetGroupTagId: assetGroupTag1,
			Position:        highestPosition,
		}
		assetGroupSelectorNode1AlreadyRevoked = model.AssetGroupSelectorNodeExpanded{AssetGroupSelectorNode: model.AssetGroupSelectorNode{
			SelectorId:        selectorId1,
			NodeId:            nodeId1,
			Certified:         model.AssetGroupCertificationRevoked,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			NodePrimaryKind:   "",
			NodeEnvironmentId: "",
			NodeObjectId:      "",
			NodeName:          "",
		},
			AssetGroupTagId: assetGroupTag1,
			Position:        highestPosition,
		}
		assetGroupSelectorNode2 = model.AssetGroupSelectorNodeExpanded{AssetGroupSelectorNode: model.AssetGroupSelectorNode{
			SelectorId:        selectorId2,
			NodeId:            nodeId2,
			Certified:         model.AssetGroupCertificationPending,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			NodePrimaryKind:   "",
			NodeEnvironmentId: "",
			NodeObjectId:      "",
			NodeName:          "",
		},
			AssetGroupTagId: assetGroupTag2,
			Position:        highestPosition,
		}
		assetGroupSelectorNode3 = model.AssetGroupSelectorNodeExpanded{AssetGroupSelectorNode: model.AssetGroupSelectorNode{
			SelectorId:        selectorId3,
			NodeId:            nodeId3,
			Certified:         model.AssetGroupCertificationPending,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			NodePrimaryKind:   "",
			NodeEnvironmentId: "",
			NodeObjectId:      "",
			NodeName:          "",
		},
			AssetGroupTagId: assetGroupTag3,
			Position:        mediumPosition,
		}
	)

	tests := []testData{
		{name: "one node certified", input: testInput{nodes: []model.AssetGroupSelectorNodeExpanded{assetGroupSelectorNode1HighPriority}, certification: model.AssetGroupCertificationManual}, want: []v2.UpdateCertificationBySelectorNodeInput{{
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationManual,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         userEmailAddress,
		},
		}},
		{name: "one node decertified", input: testInput{nodes: []model.AssetGroupSelectorNodeExpanded{assetGroupSelectorNode1HighPriority}, certification: model.AssetGroupCertificationRevoked}, want: []v2.UpdateCertificationBySelectorNodeInput{{
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationRevoked,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         userEmailAddress,
		},
		}},
		{name: "three nodes certified", input: testInput{nodes: []model.AssetGroupSelectorNodeExpanded{assetGroupSelectorNode1HighPriority, assetGroupSelectorNode2, assetGroupSelectorNode3}, certification: model.AssetGroupCertificationManual}, want: []v2.UpdateCertificationBySelectorNodeInput{{
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationManual,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         userEmailAddress,
		}, {
			AssetGroupTagId:     assetGroupTag2,
			SelectorId:          selectorId2,
			CertificationStatus: model.AssetGroupCertificationManual,
			NodeId:              nodeId2,
			NodeName:            "",
			CertifiedBy:         userEmailAddress,
		}, {
			AssetGroupTagId:     assetGroupTag3,
			SelectorId:          selectorId3,
			CertificationStatus: model.AssetGroupCertificationManual,
			NodeId:              nodeId3,
			NodeName:            "",
			CertifiedBy:         userEmailAddress,
		},
		}},
		{name: "same node, only highest priority certified", input: testInput{nodes: []model.AssetGroupSelectorNodeExpanded{assetGroupSelectorNode1HighPriority, assetGroupSelectorNode1MedPriority, assetGroupSelectorNode1LowPriority}, certification: model.AssetGroupCertificationManual}, want: []v2.UpdateCertificationBySelectorNodeInput{{
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationManual,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         userEmailAddress,
		}, {
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationPending,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         null.String{},
		}, {
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationPending,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         null.String{},
		},
		}},
		{name: "same node, skip already certified node in between other nodes", input: testInput{nodes: []model.AssetGroupSelectorNodeExpanded{assetGroupSelectorNode1HighPriority, assetGroupSelectorNode1AlreadyCertified, assetGroupSelectorNode1MedPriority, assetGroupSelectorNode1LowPriority}, certification: model.AssetGroupCertificationManual}, want: []v2.UpdateCertificationBySelectorNodeInput{{
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationManual,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         userEmailAddress,
		}, {
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationPending,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         null.String{},
		}, {
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationPending,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         null.String{},
		},
		}},
		{name: "same node, skip already certified node if first node, and do not update lower priority nodes with same node ID", input: testInput{nodes: []model.AssetGroupSelectorNodeExpanded{assetGroupSelectorNode1AlreadyCertified, assetGroupSelectorNode1MedPriority, assetGroupSelectorNode1LowPriority}, certification: model.AssetGroupCertificationManual}, want: []v2.UpdateCertificationBySelectorNodeInput{{
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationPending,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         null.String{},
		}, {
			AssetGroupTagId:     assetGroupTag1,
			SelectorId:          selectorId1,
			CertificationStatus: model.AssetGroupCertificationPending,
			NodeId:              nodeId1,
			NodeName:            "",
			CertifiedBy:         null.String{},
		},
		}},
		{name: "skip already decertified node if request is to revoke certification", input: testInput{nodes: []model.AssetGroupSelectorNodeExpanded{assetGroupSelectorNode1AlreadyRevoked}, certification: model.AssetGroupCertificationRevoked}, want: []v2.UpdateCertificationBySelectorNodeInput{}},
		{name: "do not skip already decertified node if request is to certify", input: testInput{nodes: []model.AssetGroupSelectorNodeExpanded{assetGroupSelectorNode1AlreadyRevoked}, certification: model.AssetGroupCertificationManual}, want: []v2.UpdateCertificationBySelectorNodeInput{
			{
				AssetGroupTagId:     assetGroupTag1,
				SelectorId:          selectorId1,
				CertificationStatus: model.AssetGroupCertificationManual,
				NodeId:              nodeId1,
				NodeName:            "",
				CertifiedBy:         userEmailAddress,
			},
		}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := v2.CreateInputsForUpdateCertificationBySelectorNode(testCase.input.nodes, testCase.input.certification, userEmailAddress)
			require.Equal(t, testCase.want, got)
		})

	}
}
