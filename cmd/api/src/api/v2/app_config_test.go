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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"go.uber.org/mock/gomock"
)

func Test_GetApplicationConfigurations(t *testing.T) {
	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = mocks.NewMockDatabase(mockCtrl)
		queryParser = model.NewQueryParameterFilterParser() // will fail without this
		resources   = v2.Resources{DB: mockDB, QueryParameterFilterParser: queryParser}

		expectedAppConfig = appcfg.Parameter{
			Key: appcfg.PasswordExpirationWindow,
			Value: must.NewJSONBObject(map[string]any{
				"setting": "setting",
			}),
			Serial: model.Serial{
				ID: 1,
			},
		}

		expectedAppConfigs = appcfg.Parameters{
			expectedAppConfig,
		}
	)
	defer mockCtrl.Finish()

	mockDB.EXPECT().
		GetAllConfigurationParameters(gomock.Any()).
		Return(expectedAppConfigs, nil)

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/config").
		OnHandlerFunc(resources.GetApplicationConfigurations).
		Require().
		ResponseStatusCode(http.StatusOK).
		ResponseJSONBody(v2.ListAppConfigParametersResponse{
			Data: expectedAppConfigs,
		})

	// Second call to GetAll should fail
	mockDB.EXPECT().
		GetAllConfigurationParameters(gomock.Any()).
		Return(nil, errors.New("db error"))

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/config").
		OnHandlerFunc(resources.GetApplicationConfigurations).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/config?parameter=eq:badParameter").
		OnHandlerFunc(resources.GetApplicationConfigurations).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/config?parameter=eqtz:badParameter").
		OnHandlerFunc(resources.GetApplicationConfigurations).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	mockDB.EXPECT().
		GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).
		Return(appcfg.Parameter{}, errors.New("db error"))

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/config?parameter=eq:%s", appcfg.PasswordExpirationWindow).
		OnHandlerFunc(resources.GetApplicationConfigurations).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/config?parameter=gt:%s", appcfg.PasswordExpirationWindow).
		OnHandlerFunc(resources.GetApplicationConfigurations).
		Require().
		ResponseStatusCode(http.StatusBadRequest)

	mockDB.EXPECT().
		GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).
		Return(expectedAppConfig, nil)

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/config?parameter=eq:%s", appcfg.PasswordExpirationWindow).
		OnHandlerFunc(resources.GetApplicationConfigurations).
		Require().
		ResponseStatusCode(http.StatusOK).
		ResponseJSONBody(v2.ListAppConfigParametersResponse{
			Data: expectedAppConfigs,
		})
}

func Test_SetApplicationConfiguration(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}

		appConfigRequest = appcfg.AppConfigUpdateRequest{
			Key: string(appcfg.PasswordExpirationWindow),
			Value: map[string]any{
				"setting": "setting",
			},
		}
	)
	defer mockCtrl.Finish()

	t.Run("No payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	})

	t.Run("Invalid Parameters", func(t *testing.T) {
		invalidRequest := appcfg.AppConfigUpdateRequest{
			Key: "invalidKey",
			Value: map[string]any{
				"someKey": "someValue",
			},
		}
		reqBody, _ := json.Marshal(invalidRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	})

	t.Run("Error from DB", func(t *testing.T) {
		appConfigRequest = appcfg.AppConfigUpdateRequest{
			Key: string(appcfg.ReconciliationKey),
			Value: map[string]any{
				"enabled": true,
			},
		}
		mockDB.EXPECT().
			SetConfigurationParameter(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("database error"))

		reqBody, _ := json.Marshal(appConfigRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusInternalServerError)
		}
	})

	t.Run("Success", func(t *testing.T) {
		appConfigRequest = appcfg.AppConfigUpdateRequest{
			Key: string(appcfg.ReconciliationKey),
			Value: map[string]any{
				"enabled": true,
			},
		}

		mockDB.EXPECT().
			SetConfigurationParameter(gomock.Any(), gomock.Any()).
			Return(nil)

		reqBody, _ := json.Marshal(appConfigRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}
