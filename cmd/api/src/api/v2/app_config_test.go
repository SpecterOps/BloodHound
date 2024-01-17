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

	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/specterops/bloodhound/errors"
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
		GetAllConfigurationParameters().
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
		GetAllConfigurationParameters().
		Return(nil, errors.Error("db error"))

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
		GetConfigurationParameter(appcfg.PasswordExpirationWindow).
		Return(appcfg.Parameter{}, errors.Error("db error"))

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
		GetConfigurationParameter(appcfg.PasswordExpirationWindow).
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
