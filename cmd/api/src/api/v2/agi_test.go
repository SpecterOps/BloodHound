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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/test/must"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/ctx"
	datapipeMocks "github.com/specterops/bloodhound/src/daemons/datapipe/mocks"
	dbmocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	queriesMocks "github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateAssetGroupRequest_AuditData(t *testing.T) {
	var (
		req = v2.CreateAssetGroupRequest{
			Name: "GoodRequest",
			Tag:  "Test",
		}
		data = req.AuditData()
	)
	require.Equal(t, req.Name, data["asset_group_name"])
	require.Equal(t, req.Tag, data["asset_group_tag"])
}

func TestUpdateAssetGroupRequest_AuditData(t *testing.T) {
	var (
		req = v2.UpdateAssetGroupRequest{
			Name: "GoodRequest",
		}
		data = req.AuditData()
	)
	require.Equal(t, req.Name, data["name"])
}

func TestResources_ListAssetGroups(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbmocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
		ag1       = model.AssetGroup{Name: "ag1"}
		ag2       = model.AssetGroup{Name: "ag2"}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListAssetGroups).
		Run([]apitest.Case{
			apitest.NewSortingErrorCase(),
			apitest.NewColumnNotFilterableCase(),
			apitest.NewInvalidFilterPredicateCase("id"),
			apitest.NewFilterPredicateMismatch("name", "gte:0"),
			{
				Name: "DatabaseError",
				Setup: func() {
					mockDB.EXPECT().
						GetAllAssetGroups(gomock.Any(), gomock.Any()).
						Return(model.AssetGroups{}, errors.New("database error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, api.ErrorResponseDetailsInternalServerError)
				},
			},
			{
				Name: "SuccessDataTest",
				Setup: func() {
					mockDB.EXPECT().
						GetAllAssetGroups(gomock.Any(), gomock.Any()).
						Return(model.AssetGroups{ag1, ag2}, nil)
				},
				Test: func(output apitest.Output) {
					groups := v2.ListAssetGroupsResponse{}
					apitest.UnmarshalData(output, &groups)
					apitest.Equal(output, ag2, groups.AssetGroups[1])
					apitest.Equal(output, ag1, groups.AssetGroups[0])
				},
			},
			{
				Name: "SuccessSorted",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "sort_by", "name")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAllAssetGroups("name", gomock.Any()).
						Return(model.AssetGroups{ag1, ag2}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessSortedDesc",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "sort_by", "-name")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAllAssetGroups("name desc", gomock.Any()).
						Return(model.AssetGroups{ag2, ag1}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessFiltered",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "name", "eq:ag1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAllAssetGroups("", model.SQLFilter{SQLString: "name = ?", Params: []any{"ag1"}}).
						Return(model.AssetGroups{ag1}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_GetAssetGroup(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbmocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	requestTemplate := test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("https://example.com/api/v2/asset-groups/{asset_group_id}")

	// Error where AG ID is not a valid int
	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "test",
		}).
		OnHandlerFunc(resources.GetAssetGroup).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	// DB fails
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, fmt.Errorf("explosions"))

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		OnHandlerFunc(resources.GetAssetGroup).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// Happy path
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		OnHandlerFunc(resources.GetAssetGroup).
		Require().
		ResponseStatusCode(http.StatusOK)
}

func TestResources_GetAssetGroupMemberCount_IDMalformed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req, err := http.NewRequest("GET", "/api/v2/asset-groups/1/customselectors", nil)
	require.Nil(t, err)

	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	resources := v2.Resources{DB: mockDB}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.GetAssetGroupCustomMemberCount)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsIDMalformed)
}

func TestResources_GetAssetGroupMemberCount_DBError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req, err := http.NewRequest("GET", "/api/v2/asset-groups/1/customselectors", nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupID: "1"})
	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(model.AssetGroup{}, fmt.Errorf("test error"))

	resources := v2.Resources{DB: mockDB}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.GetAssetGroupCustomMemberCount)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
}

func TestResources_GetAssetGroupMemberCount_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	assetGroup := model.AssetGroup{
		Name:        "test group",
		Tag:         "test tag",
		SystemGroup: false,
		Selectors: model.AssetGroupSelectors{
			model.AssetGroupSelector{
				AssetGroupID:   1,
				Name:           "custom selector",
				Selector:       "selector",
				SystemSelector: false,
				Serial:         model.Serial{},
			},
			model.AssetGroupSelector{
				AssetGroupID:   1,
				Name:           "custom selector 2",
				Selector:       "selector2",
				SystemSelector: false,
				Serial:         model.Serial{},
			},
			model.AssetGroupSelector{
				AssetGroupID:   1,
				Name:           "system selector",
				Selector:       "selector3",
				SystemSelector: true,
				Serial:         model.Serial{},
			},
		},
		Collections: model.AssetGroupCollections{
			model.AssetGroupCollection{
				Entries: model.AssetGroupCollectionEntries{
					model.AssetGroupCollectionEntry{
						ObjectID:  "a",
						NodeLabel: "a",
						BigSerial: model.BigSerial{ID: 1},
					},
					model.AssetGroupCollectionEntry{
						ObjectID:  "b",
						NodeLabel: "b",
						BigSerial: model.BigSerial{ID: 2},
					},
				},
				BigSerial: model.BigSerial{ID: 1},
			},
			model.AssetGroupCollection{
				Entries:   nil,
				BigSerial: model.BigSerial{ID: 2},
			},
		},
	}

	req, err := http.NewRequest("GET", "/api/v2/asset-groups/1/customselectors", nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupID: "1"})
	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(assetGroup, nil)

	resources := v2.Resources{DB: mockDB}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.GetAssetGroupCustomMemberCount)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)

	var result = api.ResponseWrapper{}
	err = json.Unmarshal(response.Body.Bytes(), &result)
	require.Nil(t, err)

	require.Len(t, result.Data, 1)
	require.Equal(t, float64(2), result.Data.(map[string]any)["custom_member_count"].(float64))
}

func TestResources_UpdateAssetGroup(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbmocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	requestTemplate := test.Request(t).
		WithMethod(http.MethodPut).
		WithURL("https://example.com/api/v2/asset-groups/{asset_group_id}")

	// Error where AG ID is not a valid int
	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "test",
		}).
		OnHandlerFunc(resources.UpdateAssetGroup).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	// Error where no body is provided
	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		OnHandlerFunc(resources.UpdateAssetGroup).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	// GetAssetGroup DB fails
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, fmt.Errorf("exploded"))

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		WithBody(v2.UpdateAssetGroupRequest{}).
		OnHandlerFunc(resources.UpdateAssetGroup).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// UpdateAssetGroup DB fails
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)
	mockDB.EXPECT().UpdateAssetGroup(gomock.Any(), model.AssetGroup{}).Return(fmt.Errorf("exploded"))

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		WithBody(v2.UpdateAssetGroupRequest{}).
		OnHandlerFunc(resources.UpdateAssetGroup).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// Success
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)
	mockDB.EXPECT().UpdateAssetGroup(gomock.Any(), model.AssetGroup{}).Return(nil)

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		WithBody(v2.UpdateAssetGroupRequest{}).
		OnHandlerFunc(resources.UpdateAssetGroup).
		Require().
		ResponseStatusCode(http.StatusOK)
}

func TestResources_CreateAssetGroup(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbmocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	requestTemplate := test.Request(t).
		WithMethod(http.MethodPost).
		WithURL("http://example.com/api/v2/asset-groups")

	// Error where no body is provided
	requestTemplate.
		OnHandlerFunc(resources.CreateAssetGroup).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	// Create DB Query fails
	mockDB.EXPECT().CreateAssetGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AssetGroup{}, fmt.Errorf("exploded"))

	requestTemplate.
		WithBody(v2.CreateAssetGroupRequest{}).
		OnHandlerFunc(resources.CreateAssetGroup).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// Success
	mockDB.EXPECT().CreateAssetGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AssetGroup{}, nil)

	requestTemplate.
		WithContext(&ctx.Context{
			Host: &url.URL{},
		}).
		WithBody(v2.CreateAssetGroupRequest{}).
		OnHandlerFunc(resources.CreateAssetGroup).
		Require().
		ResponseStatusCode(http.StatusCreated)
}

func TestResources_UpdateAssetGroupSelectors_GetAssetGroupError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := []model.AssetGroupSelectorSpec{
		{
			SelectorName:   "test",
			EntityObjectID: "1",
		},
	}

	req, err := http.NewRequest("POST", "/api/v2/asset-groups/1/selectors", must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupID: "1"})

	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(model.AssetGroup{}, fmt.Errorf("test error"))
	handlers := v2.Resources{DB: mockDB}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateAssetGroupSelectors)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
}

func TestResources_UpdateAssetGroupSelectors_PayloadError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := "INVALID PAYLOAD"

	req, err := http.NewRequest("POST", "/api/v2/asset-groups/1/selectors", must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupID: "1"})

	assetGroup := model.AssetGroup{
		Name:        "test group",
		Tag:         "test tag",
		SystemGroup: false,
	}

	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(assetGroup, nil)
	handlers := v2.Resources{DB: mockDB}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateAssetGroupSelectors)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponsePayloadUnmarshalError)
}

func TestResources_UpdateAssetGroupSelectors_SuccessT0(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := []model.AssetGroupSelectorSpec{
		{
			SelectorName:   "test",
			EntityObjectID: "1",
			Action:         model.SelectorSpecActionAdd,
		},
		{
			SelectorName:   "test3",
			EntityObjectID: "3",
			Action:         model.SelectorSpecActionAdd,
		},
		{
			SelectorName:   "test2",
			EntityObjectID: "2",
			Action:         model.SelectorSpecActionRemove,
		},
	}

	req, err := http.NewRequest("POST", "/api/v2/asset-groups/1/selectors", must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	bheCtx := ctx.Context{
		RequestID: "requestID",
		AuthCtx: auth.Context{
			Owner:   model.User{},
			Session: model.UserSession{},
		},
	}
	req = req.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bheCtx.WithRequestID("requestID")))
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupID: "1"})

	assetGroup := model.AssetGroup{
		Name:        model.TierZeroAssetGroupName,
		Tag:         model.TierZeroAssetGroupTag,
		SystemGroup: true,
	}

	expectedResult := model.UpdatedAssetGroupSelectors{
		Added: model.AssetGroupSelectors{
			{
				AssetGroupID: assetGroup.ID,
				Name:         payload[0].SelectorName,
				Selector:     payload[0].EntityObjectID,
			},
		},

		Removed: model.AssetGroupSelectors{
			{
				AssetGroupID: assetGroup.ID,
				Name:         payload[1].SelectorName,
				Selector:     payload[1].EntityObjectID,
			},
		},
	}

	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockGraph := queriesMocks.NewMockGraph(mockCtrl)

	mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(assetGroup, nil)
	mockDB.EXPECT().UpdateAssetGroupSelectors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedResult, nil)
	mockGraph.EXPECT().UpdateSelectorTags(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	mockTasker := datapipeMocks.NewMockTasker(mockCtrl)
	// MockTasker should receive a call to RequestAnalysis() since this is a Tier Zero Asset group.
	// Analysis must be run upon updating a T0 AG
	mockTasker.EXPECT().RequestAnalysis()

	handlers := v2.Resources{
		DB:           mockDB,
		TaskNotifier: mockTasker,
		GraphQuery:   mockGraph,
	}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateAssetGroupSelectors)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusCreated, response.Code)

	resp := api.ResponseWrapper{}
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	require.Nil(t, err)

	dataJSON, err := json.Marshal(resp.Data)
	require.Nil(t, err)

	data := make(map[string][]model.AssetGroupSelector, 0)
	err = json.Unmarshal(dataJSON, &data)
	require.Nil(t, err)

	require.Equal(t, expectedResult.Added[0].Name, data["added_selectors"][0].Name)
	require.Equal(t, expectedResult.Removed[0].Name, data["removed_selectors"][0].Name)
}

func TestResources_UpdateAssetGroupSelectors_SuccessOwned(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := []model.AssetGroupSelectorSpec{
		{
			SelectorName:   "test",
			EntityObjectID: "1",
			Action:         model.SelectorSpecActionAdd,
		},
		{
			SelectorName:   "test2",
			EntityObjectID: "2",
			Action:         model.SelectorSpecActionRemove,
		},
	}

	req, err := http.NewRequest("POST", "/api/v2/asset-groups/1/selectors", must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	bheCtx := ctx.Context{
		RequestID: "requestID",
		AuthCtx: auth.Context{
			Owner:   model.User{},
			Session: model.UserSession{},
		},
	}
	req = req.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bheCtx.WithRequestID("requestID")))
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableAssetGroupID: "1"})

	assetGroup := model.AssetGroup{
		Name:        model.OwnedAssetGroupName,
		Tag:         model.OwnedAssetGroupTag,
		SystemGroup: true,
	}

	expectedResult := model.UpdatedAssetGroupSelectors{
		Added: model.AssetGroupSelectors{
			model.AssetGroupSelector{
				AssetGroupID: assetGroup.ID,
				Name:         payload[0].SelectorName,
				Selector:     payload[0].EntityObjectID,
			},
		},
		Removed: model.AssetGroupSelectors{
			model.AssetGroupSelector{
				AssetGroupID: assetGroup.ID,
				Name:         payload[1].SelectorName,
				Selector:     payload[1].EntityObjectID,
			},
		},
	}

	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockGraph := queriesMocks.NewMockGraph(mockCtrl)

	mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(assetGroup, nil)
	mockDB.EXPECT().UpdateAssetGroupSelectors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedResult, nil)

	mockGraph.EXPECT().UpdateSelectorTags(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	mockTasker := datapipeMocks.NewMockTasker(mockCtrl)
	// NOTE MockTasker should NOT receive a call to RequestAnalysis() since this is not a Tier Zero Asset group.
	// Analysis should not be re-run when a non T0 AG is updated

	handlers := v2.Resources{
		DB:           mockDB,
		TaskNotifier: mockTasker,
		GraphQuery:   mockGraph,
	}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateAssetGroupSelectors)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusCreated, response.Code)

	resp := api.ResponseWrapper{}
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	require.Nil(t, err)

	dataJSON, err := json.Marshal(resp.Data)
	require.Nil(t, err)

	data := make(map[string][]model.AssetGroupSelector, 0)
	err = json.Unmarshal(dataJSON, &data)
	require.Nil(t, err)

	require.Equal(t, expectedResult.Added[0].Name, data["added_selectors"][0].Name)
	require.Equal(t, expectedResult.Removed[0].Name, data["removed_selectors"][0].Name)
}

func TestResources_DeleteAssetGroup(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbmocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	requestTemplate := test.Request(t).
		WithMethod(http.MethodDelete).
		WithURL("https://example.com/api/v2/asset-groups/{asset_group_id}")

	// Error where AG ID is not a valid int
	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "test",
		}).
		OnHandlerFunc(resources.DeleteAssetGroup).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	// GetAssetGroup DB fails
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, fmt.Errorf("exploded"))

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		OnHandlerFunc(resources.DeleteAssetGroup).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// DeleteAssetGroup DB fails
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)
	mockDB.EXPECT().DeleteAssetGroup(gomock.Any(), model.AssetGroup{}).Return(fmt.Errorf("exploded"))

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		OnHandlerFunc(resources.DeleteAssetGroup).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// Success
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)
	mockDB.EXPECT().DeleteAssetGroup(gomock.Any(), model.AssetGroup{}).Return(nil)

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		OnHandlerFunc(resources.DeleteAssetGroup).
		Require().
		ResponseStatusCode(http.StatusOK)
}

func TestResources_DeleteAssetGroupSelector(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbmocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	requestTemplate := test.Request(t).
		WithMethod(http.MethodDelete).
		WithURL("https://example.com/api/v2/asset-groups/{asset_group_id}/selectors/{asset_group_selector_id}")

	// Error where AG ID is not a valid int
	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "test",
		}).
		OnHandlerFunc(resources.DeleteAssetGroupSelector).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	// GetAssetGroup DB fails
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, fmt.Errorf("exploded"))

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id": "1234",
		}).
		OnHandlerFunc(resources.DeleteAssetGroupSelector).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// Error where AG Selector ID is not a valid int
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id":          "1234",
			"asset_group_selector_id": "test",
		}).
		OnHandlerFunc(resources.DeleteAssetGroupSelector).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	// GetAssetGroupSelector DB fails
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)
	mockDB.EXPECT().GetAssetGroupSelector(int32(1234)).Return(model.AssetGroupSelector{}, fmt.Errorf("exploded"))

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id":          "1234",
			"asset_group_selector_id": "1234",
		}).
		OnHandlerFunc(resources.DeleteAssetGroupSelector).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// Is System Selector should fail
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)
	mockDB.EXPECT().GetAssetGroupSelector(int32(1234)).Return(model.AssetGroupSelector{SystemSelector: true}, nil)

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id":          "1234",
			"asset_group_selector_id": "1234",
		}).
		OnHandlerFunc(resources.DeleteAssetGroupSelector).
		Require().
		ResponseStatusCode(http.StatusConflict)

	// DeleteAssetGroupSelector DB fails
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)
	mockDB.EXPECT().GetAssetGroupSelector(int32(1234)).Return(model.AssetGroupSelector{}, nil)
	mockDB.EXPECT().DeleteAssetGroupSelector(gomock.Any(), model.AssetGroupSelector{}).Return(fmt.Errorf("exploded"))

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id":          "1234",
			"asset_group_selector_id": "1234",
		}).
		OnHandlerFunc(resources.DeleteAssetGroupSelector).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	// Success
	mockDB.EXPECT().GetAssetGroup(int32(1234)).Return(model.AssetGroup{}, nil)
	mockDB.EXPECT().GetAssetGroupSelector(int32(1234)).Return(model.AssetGroupSelector{}, nil)
	mockDB.EXPECT().DeleteAssetGroupSelector(gomock.Any(), model.AssetGroupSelector{}).Return(nil)

	requestTemplate.
		WithURLPathVars(map[string]string{
			"asset_group_id":          "1234",
			"asset_group_selector_id": "1234",
		}).
		OnHandlerFunc(resources.DeleteAssetGroupSelector).
		Require().
		ResponseStatusCode(http.StatusOK)
}

func TestResources_ListAssetGroupCollections(t *testing.T) {
	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = dbmocks.NewMockDatabase(mockCtrl)
		resources   = v2.Resources{DB: mockDB}
		collections = model.AssetGroupCollections{
			model.AssetGroupCollection{
				Entries: model.AssetGroupCollectionEntries{
					model.AssetGroupCollectionEntry{
						ObjectID:  "a",
						NodeLabel: "b",
						BigSerial: model.BigSerial{ID: 1},
					},
					model.AssetGroupCollectionEntry{
						ObjectID:  "c",
						NodeLabel: "d",
						BigSerial: model.BigSerial{ID: 2},
					},
				},
				BigSerial: model.BigSerial{ID: 3},
			},
			model.AssetGroupCollection{
				Entries:   nil,
				BigSerial: model.BigSerial{ID: 4},
			},
		}
		assetGroup = model.AssetGroup{
			Name:        "test group",
			Tag:         "test tag",
			SystemGroup: false,
			Selectors: model.AssetGroupSelectors{
				model.AssetGroupSelector{
					AssetGroupID:   1,
					Name:           "test selector",
					Selector:       "selector",
					SystemSelector: false,
					Serial:         model.Serial{},
				},
			},
			Collections: collections,
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListAssetGroupCollections).
		WithCommonRequest(func(input *apitest.Input) {
			apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
		}).
		Run([]apitest.Case{
			apitest.NewSortingErrorCase(),
			apitest.NewColumnNotFilterableCase(),
			apitest.NewInvalidFilterPredicateCase("id"),
			{
				Name: "InvalidAssetGroupID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "invalid")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "DatabaseGetAssetGroupError",
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(model.AssetGroup{}, errors.New("GetAssetGroup fail"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, api.ErrorResponseDetailsInternalServerError)
				},
			},
			{
				Name: "DatabaseGetAssetGroupCollectionsError",
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockDB.EXPECT().
						GetAssetGroupCollections(gomock.Any(), "", gomock.Any()).
						Return(nil, errors.New("GetAssetGroupCollectionsError"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "SuccessDataTest",
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockDB.EXPECT().
						GetAssetGroupCollections(gomock.Any(), "", gomock.Any()).
						Return(collections, nil)
				},
				Test: func(output apitest.Output) {
					result := model.AssetGroupCollections{}
					apitest.UnmarshalData(output, &result)

					require.Equal(t, len(collections), len(result))

					require.Equal(t, collections[0].Entries[0].ObjectID, result[0].Entries[0].ObjectID)
					require.Equal(t, collections[0].Entries[0].NodeLabel, result[0].Entries[0].NodeLabel)

					require.Equal(t, collections[0].Entries[1].ObjectID, result[0].Entries[1].ObjectID)
					require.Equal(t, collections[0].Entries[1].NodeLabel, result[0].Entries[1].NodeLabel)

					require.Equal(t, 0, len(result[1].Entries))
				},
			},
			{
				Name: "SuccessSorted",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "sort_by", "created_at")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockDB.EXPECT().
						GetAssetGroupCollections(gomock.Any(), "created_at", gomock.Any()).
						Return(collections, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessSortedDesc",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "sort_by", "-created_at")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockDB.EXPECT().
						GetAssetGroupCollections(gomock.Any(), "created_at desc", gomock.Any()).
						Return(collections, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
			{
				Name: "SuccessFiltered",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "id", "eq:1")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(assetGroup, nil)
					mockDB.EXPECT().
						GetAssetGroupCollections(gomock.Any(), "",
							model.SQLFilter{SQLString: "id = ?", Params: []any{"1"}}).
						Return(collections, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}

func TestResources_ListAssetGroupMembers(t *testing.T) {
	var (
		mockCtrl   = gomock.NewController(t)
		mockGraph  = queriesMocks.NewMockGraph(mockCtrl)
		mockDB     = dbmocks.NewMockDatabase(mockCtrl)
		resources  = v2.Resources{DB: mockDB, GraphQuery: mockGraph}
		collection = model.AssetGroupCollection{
			Entries: model.AssetGroupCollectionEntries{
				model.AssetGroupCollectionEntry{
					ObjectID:  "a",
					NodeLabel: "b",
					BigSerial: model.BigSerial{ID: 1},
				},
				model.AssetGroupCollectionEntry{
					ObjectID:  "c",
					NodeLabel: "d",
					BigSerial: model.BigSerial{ID: 2},
				},
			},
			BigSerial: model.BigSerial{ID: 3},
		}

		assetGroup = model.AssetGroup{
			Name:        "test group",
			Tag:         "test tag",
			SystemGroup: false,
			Selectors: model.AssetGroupSelectors{
				model.AssetGroupSelector{
					AssetGroupID:   1,
					Name:           "a",
					Selector:       "a",
					SystemSelector: false,
					Serial:         model.Serial{},
				},
			},
			Collections: model.AssetGroupCollections{collection},
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListAssetGroupMembers).
		Run([]apitest.Case{
			{
				Name: "InvalidAssetGroupID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "invalid")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "DatabaseGetAssetGroupError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(model.AssetGroup{}, errors.New("GetAssetGroup fail"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, api.ErrorResponseDetailsInternalServerError)
				},
			},
			{
				Name: "GraphDatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{}, fmt.Errorf("GetAssetGroupNodes fail"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Node missing base entity kind",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{
							1: &graph.Node{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
								},
							},
							2: &graph.Node{
								ID:    2,
								Kinds: graph.Kinds{azure.Entity, azure.Tenant},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", azure.TenantID.String(): "b"},
								},
							},
							3: &graph.Node{
								ID:    3,
								Kinds: graph.Kinds{ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "c", common.Name.String(): "c", ad.DomainSID.String(): "c"},
								},
							},
							4: &graph.Node{
								ID:    4,
								Kinds: graph.Kinds{azure.Tenant},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "d", common.Name.String(): "d", azure.TenantID.String(): "d"},
								},
							},
						}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					result := api.ListAssetGroupMembersResponse{}
					apitest.UnmarshalData(output, &result)
					apitest.BodyContains(output, `"custom_member":true`)
					apitest.Equal(output, 2, len(result.Members))
				},
			},
			{
				Name: "SuccessDataTest",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{
							1: &graph.Node{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
								},
							},
							2: &graph.Node{
								ID:    2,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", ad.DomainSID.String(): "b"},
								},
							}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					result := api.ListAssetGroupMembersResponse{}
					apitest.UnmarshalData(output, &result)
					apitest.BodyContains(output, `"custom_member":true`)
				},
			},
			{
				Name: "InvalidSkip",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
					apitest.AddQueryParam(input, model.PaginationQueryParameterSkip, "1000000")
					apitest.AddQueryParam(input, model.PaginationQueryParameterLimit, "4")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{
							1: &graph.Node{
								ID:    1,
								Kinds: graph.Kinds{ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
								},
							},
							2: &graph.Node{
								ID:    2,
								Kinds: graph.Kinds{ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", ad.DomainSID.String(): "b"},
								},
							},
							3: &graph.Node{
								ID:    3,
								Kinds: graph.Kinds{ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "c", common.Name.String(): "c", ad.DomainSID.String(): "c"},
								},
							}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "invalid skip")
				},
			},
			{
				Name: "SuccessPaginated",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
					apitest.AddQueryParam(input, model.PaginationQueryParameterSkip, "1")
					apitest.AddQueryParam(input, model.PaginationQueryParameterLimit, "4")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{
							1: &graph.Node{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
								},
							},
							2: &graph.Node{
								ID:    2,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", ad.DomainSID.String(): "b"},
								},
							},
							3: &graph.Node{
								ID:    3,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "c", common.Name.String(): "c", ad.DomainSID.String(): "c"},
								},
							}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)

					result := api.ListAssetGroupMembersResponse{}
					apitest.UnmarshalData(output, &result)
					apitest.Equal(output, 2, len(result.Members))
				},
			},
			{
				Name: "SuccessSorted",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
					apitest.AddQueryParam(input, api.QueryParameterSortBy, "-object_id")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{
							1: &graph.Node{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
								},
							},
							2: &graph.Node{
								ID:    2,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", ad.DomainSID.String(): "b"},
								},
							}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					result := api.ListAssetGroupMembersResponse{}
					apitest.UnmarshalData(output, &result)
					apitest.Equal(output, "b", result.Members[0].ObjectID)
				},
			},
			{
				Name: "SuccessFiltered",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "object_id", "eq:a")
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{
							1: &graph.Node{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
								},
							},
							2: &graph.Node{
								ID:    2,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", ad.DomainSID.String(): "b"},
								},
							}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)

					result := api.ListAssetGroupMembersResponse{}
					apitest.UnmarshalData(output, &result)

					require.Equal(t, 1, len(result.Members))
					require.Equal(t, "a", result.Members[0].ObjectID)
				},
			},
		})
}

func TestResources_ListAssetGroupMembersCount(t *testing.T) {
	var (
		mockCtrl   = gomock.NewController(t)
		mockGraph  = queriesMocks.NewMockGraph(mockCtrl)
		mockDB     = dbmocks.NewMockDatabase(mockCtrl)
		resources  = v2.Resources{DB: mockDB, GraphQuery: mockGraph}
		collection = model.AssetGroupCollection{
			Entries: model.AssetGroupCollectionEntries{
				model.AssetGroupCollectionEntry{
					ObjectID:  "a",
					NodeLabel: "b",
					BigSerial: model.BigSerial{ID: 1},
				},
				model.AssetGroupCollectionEntry{
					ObjectID:  "c",
					NodeLabel: "d",
					BigSerial: model.BigSerial{ID: 2},
				},
			},
			BigSerial: model.BigSerial{ID: 3},
		}

		assetGroup = model.AssetGroup{
			Name:        "test group",
			Tag:         "test tag",
			SystemGroup: false,
			Selectors: model.AssetGroupSelectors{
				model.AssetGroupSelector{
					AssetGroupID:   1,
					Name:           "a",
					Selector:       "a",
					SystemSelector: false,
					Serial:         model.Serial{},
				},
			},
			Collections: model.AssetGroupCollections{collection},
		}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListAssetGroupMemberCountsByKind).
		Run([]apitest.Case{
			{
				Name: "InvalidAssetGroupID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "invalid")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, api.ErrorResponseDetailsIDMalformed)
				},
			},
			{
				Name: "DatabaseGetAssetGroupError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(model.AssetGroup{}, errors.New("GetAssetGroup fail"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, api.ErrorResponseDetailsInternalServerError)
				},
			},
			{
				Name: "GraphDatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{}, fmt.Errorf("GetAssetGroupNodes fail"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "SuccessDataTest",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{
							1: &graph.Node{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
								},
							},
							2: &graph.Node{
								ID:    2,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", ad.DomainSID.String(): "b"},
								},
							}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					result := api.ListAssetGroupMemberCountsResponse{}
					apitest.UnmarshalData(output, &result)
					require.Equal(t, 2, result.TotalCount)
					require.Equal(t, 2, result.Counts[ad.Domain.String()])
				},
			},
			{
				Name: "SuccessFiltered",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "environment_kind", "eq:"+azure.Tenant.String())
					apitest.SetURLVar(input, api.URIPathVariableAssetGroupID, "1")
				},
				Setup: func() {
					mockDB.EXPECT().
						GetAssetGroup(gomock.Any()).
						Return(assetGroup, nil)
					mockGraph.EXPECT().
						GetAssetGroupNodes(gomock.Any(), gomock.Any()).
						Return(graph.NodeSet{
							1: &graph.Node{
								ID:    1,
								Kinds: graph.Kinds{ad.Entity, ad.Domain},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
								},
							},
							2: &graph.Node{
								ID:    2,
								Kinds: graph.Kinds{azure.Entity, azure.Tenant},
								Properties: &graph.Properties{
									Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", azure.TenantID.String(): "b"},
								},
							}}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					result := api.ListAssetGroupMemberCountsResponse{}
					apitest.UnmarshalData(output, &result)
					require.Equal(t, 1, result.TotalCount)
					require.Equal(t, 1, result.Counts[azure.Tenant.String()])
				},
			},
		})
}
