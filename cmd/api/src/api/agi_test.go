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

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	datapipeMocks "github.com/specterops/bloodhound/src/daemons/datapipe/mocks"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAssetGroupMembers_SortBy(t *testing.T) {
	input := api.AssetGroupMembers{
		api.AssetGroupMember{
			AssetGroupID:    2,
			ObjectID:        "b",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name2",
			CustomMember:    true,
		},
		api.AssetGroupMember{
			AssetGroupID:    1,
			ObjectID:        "a",
			PrimaryKind:     ad.Computer.String(),
			Kinds:           []string{"Base", "Computer"},
			EnvironmentID:   "domainsid",
			EnvironmentKind: "Domain",
			Name:            "name1",
			CustomMember:    false,
		},
		api.AssetGroupMember{
			AssetGroupID:    3,
			ObjectID:        "c",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid2",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name3",
			CustomMember:    false,
		},
	}

	_, err := input.SortBy([]string{""})
	require.NotNil(t, err)
	require.Equal(t, api.ErrorResponseEmptySortParameter, err.Error())

	_, err = input.SortBy([]string{"foobar"})
	require.Equal(t, api.ErrorResponseDetailsNotSortable, err.Error())

	output, err := input.SortBy([]string{"-asset_group_id"})
	require.Nil(t, err)
	require.Equal(t, 3, output[0].AssetGroupID)

	output, err = input.SortBy([]string{"primary_kind"})
	require.Nil(t, err)
	require.Equal(t, azure.Group.String(), output[0].PrimaryKind)

	output, err = input.SortBy([]string{"environment_id"})
	require.Nil(t, err)
	require.Equal(t, "domainsid", output[0].EnvironmentID)

	output, err = input.SortBy([]string{"environment_kind"})
	require.Nil(t, err)
	require.Equal(t, azure.Tenant.String(), output[0].EnvironmentKind)

	output, err = input.SortBy([]string{"name"})
	require.Nil(t, err)
	require.Equal(t, "name1", output[0].Name)
}

func TestAssetGroupMembers_Filter_Equals(t *testing.T) {
	input := api.AssetGroupMembers{
		api.AssetGroupMember{
			AssetGroupID:    2,
			ObjectID:        "b",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name2",
			CustomMember:    true,
		},
		api.AssetGroupMember{
			AssetGroupID:    1,
			ObjectID:        "a",
			PrimaryKind:     ad.Computer.String(),
			Kinds:           []string{"Base", "Computer"},
			EnvironmentID:   "domainsid",
			EnvironmentKind: "Domain",
			Name:            "name1",
			CustomMember:    false,
		},
		api.AssetGroupMember{
			AssetGroupID:    3,
			ObjectID:        "c",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid2",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name3",
			CustomMember:    false,
		},
	}

	_, err := input.Filter(model.QueryParameterFilterMap{
		"badcolumn": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "badcolumn",
				Operator: model.Equals,
				Value:    "1",
			},
		},
	})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), model.ErrorResponseDetailsColumnNotFilterable)

	_, err = input.Filter(model.QueryParameterFilterMap{
		"object_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "object_id",
				Operator: model.GreaterThan, // invalid operator
				Value:    "a",
			},
		},
	})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), model.ErrorResponseDetailsFilterPredicateNotSupported)

	// filter on object_id
	output, err := input.Filter(model.QueryParameterFilterMap{
		"object_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "object_id",
				Operator: model.Equals,
				Value:    "a",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 1, output[0].AssetGroupID)

	// filter on name
	output, err = input.Filter(model.QueryParameterFilterMap{
		"name": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "name",
				Operator: model.Equals,
				Value:    "name3",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 3, output[0].AssetGroupID)

	// filter on custom_member
	output, err = input.Filter(model.QueryParameterFilterMap{
		"custom_member": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "custom_member",
				Operator: model.Equals,
				Value:    "false",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on environment_id
	output, err = input.Filter(model.QueryParameterFilterMap{
		"environment_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "environment_id",
				Operator: model.Equals,
				Value:    "tenantid",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 2, output[0].AssetGroupID)

	// filter on environment_kind
	output, err = input.Filter(model.QueryParameterFilterMap{
		"environment_kind": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "environment_kind",
				Operator: model.Equals,
				Value:    "AZTenant",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on primary_kind
	output, err = input.Filter(model.QueryParameterFilterMap{
		"primary_kind": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "primary_kind",
				Operator: model.Equals,
				Value:    "Computer",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 1, output[0].AssetGroupID)
}

func TestAssetGroupMembers_Filter_NotEquals(t *testing.T) {
	input := api.AssetGroupMembers{
		api.AssetGroupMember{
			AssetGroupID:    2,
			ObjectID:        "b",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name2",
			CustomMember:    true,
		},
		api.AssetGroupMember{
			AssetGroupID:    1,
			ObjectID:        "a",
			PrimaryKind:     ad.Computer.String(),
			Kinds:           []string{"Base", "Computer"},
			EnvironmentID:   "domainsid",
			EnvironmentKind: "Domain",
			Name:            "name1",
			CustomMember:    false,
		},
		api.AssetGroupMember{
			AssetGroupID:    3,
			ObjectID:        "c",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid2",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name3",
			CustomMember:    false,
		},
	}

	// filter on object_id
	output, err := input.Filter(model.QueryParameterFilterMap{
		"object_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "object_id",
				Operator: model.NotEquals,
				Value:    "b",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on name
	output, err = input.Filter(model.QueryParameterFilterMap{
		"name": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "name",
				Operator: model.NotEquals,
				Value:    "name3",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on custom_member
	output, err = input.Filter(model.QueryParameterFilterMap{
		"custom_member": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "custom_member",
				Operator: model.NotEquals,
				Value:    "false",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 2, output[0].AssetGroupID)

	// filter on environment_id
	output, err = input.Filter(model.QueryParameterFilterMap{
		"environment_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "environment_id",
				Operator: model.NotEquals,
				Value:    "tenantid",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on environment_kind
	output, err = input.Filter(model.QueryParameterFilterMap{
		"environment_kind": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "environment_kind",
				Operator: model.NotEquals,
				Value:    "AZTenant",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 1, output[0].AssetGroupID)

	// filter on primary_kind
	output, err = input.Filter(model.QueryParameterFilterMap{
		"primary_kind": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "primary_kind",
				Operator: model.NotEquals,
				Value:    "Computer",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))
}

func TestAssetGroupMembers_BuildFilteringConditional_Error(t *testing.T) {
	input := api.AssetGroupMembers{}
	columns := input.GetFilterableColumns()

	// invalid predicates for all columns
	for _, column := range columns {
		_, err := input.BuildFilteringConditional(column, model.GreaterThan, "1")
		require.Error(t, err)
		require.Contains(t, err.Error(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}

	// invalid column
	_, err := input.BuildFilteringConditional("badcolumn", model.GreaterThan, "1")
	require.Error(t, err)
	require.Contains(t, err.Error(), api.ErrorResponseDetailsColumnNotFilterable)

	// invalid values
	_, err = input.BuildFilteringConditional("custom_member", model.Equals, "1234")
	require.Error(t, err)
	require.Contains(t, err.Error(), api.ErrorResponseDetailsBadQueryParameterFilters)

	_, err = input.BuildFilteringConditional("asset_group_id", model.Equals, "abcd")
	require.Error(t, err)
	require.Contains(t, err.Error(), api.ErrorResponseDetailsBadQueryParameterFilters)
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

	mockDB := dbMocks.NewMockDatabase(mockCtrl)
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

	mockDB := dbMocks.NewMockDatabase(mockCtrl)
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

	expectedResult := map[string]model.AssetGroupSelectors{
		"added_selectors": {
			model.AssetGroupSelector{
				AssetGroupID: assetGroup.ID,
				Name:         payload[0].SelectorName,
				Selector:     payload[0].EntityObjectID,
			},
		},
		"removed_selectors": {
			model.AssetGroupSelector{
				AssetGroupID: assetGroup.ID,
				Name:         payload[1].SelectorName,
				Selector:     payload[1].EntityObjectID,
			},
		},
	}

	mockDB := dbMocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(assetGroup, nil)
	mockDB.EXPECT().UpdateAssetGroupSelectors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedResult, nil)

	mockTasker := datapipeMocks.NewMockTasker(mockCtrl)
	// MockTasker should receive a call to RequestAnalysis() since this is a Tier Zero Asset group.
	// Analysis must be run upon updating a T0 AG
	mockTasker.EXPECT().RequestAnalysis()

	handlers := v2.Resources{
		DB:           mockDB,
		TaskNotifier: mockTasker,
	}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateAssetGroupSelectors)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)

	resp := api.ResponseWrapper{}
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	require.Nil(t, err)

	dataJSON, err := json.Marshal(resp.Data)
	require.Nil(t, err)

	data := make(map[string][]model.AssetGroupSelector, 0)
	err = json.Unmarshal(dataJSON, &data)
	require.Nil(t, err)

	require.Equal(t, expectedResult["added_selectors"][0].Name, data["added_selectors"][0].Name)
	require.Equal(t, expectedResult["removed_selectors"][0].Name, data["removed_selectors"][0].Name)
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

	expectedResult := map[string]model.AssetGroupSelectors{
		"added_selectors": {
			model.AssetGroupSelector{
				AssetGroupID: assetGroup.ID,
				Name:         payload[0].SelectorName,
				Selector:     payload[0].EntityObjectID,
			},
		},
		"removed_selectors": {
			model.AssetGroupSelector{
				AssetGroupID: assetGroup.ID,
				Name:         payload[1].SelectorName,
				Selector:     payload[1].EntityObjectID,
			},
		},
	}

	mockDB := dbMocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAssetGroup(gomock.Any()).Return(assetGroup, nil)
	mockDB.EXPECT().UpdateAssetGroupSelectors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedResult, nil)

	mockTasker := datapipeMocks.NewMockTasker(mockCtrl)
	// NOTE MockTasker should NOT receive a call to RequestAnalysis() since this is not a Tier Zero Asset group.
	// Analysis should not be re-run when a non T0 AG is updated

	handlers := v2.Resources{
		DB:           mockDB,
		TaskNotifier: mockTasker,
	}

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateAssetGroupSelectors)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)

	resp := api.ResponseWrapper{}
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	require.Nil(t, err)

	dataJSON, err := json.Marshal(resp.Data)
	require.Nil(t, err)

	data := make(map[string][]model.AssetGroupSelector, 0)
	err = json.Unmarshal(dataJSON, &data)
	require.Nil(t, err)

	require.Equal(t, expectedResult["added_selectors"][0].Name, data["added_selectors"][0].Name)
	require.Equal(t, expectedResult["removed_selectors"][0].Name, data["removed_selectors"][0].Name)
}
