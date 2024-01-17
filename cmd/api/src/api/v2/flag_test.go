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

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/utils/test"
)

func TestResources_GetFlags(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)

	defer mockCtrl.Finish()

	mockDB.EXPECT().GetAllFlags().Return([]appcfg.FeatureFlag{}, nil)

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/features").
		OnHandlerFunc(resources.GetFlags).
		Require().
		ResponseStatusCode(http.StatusOK).
		ResponseJSONBody(v2.ListFlagsResponse{
			Data: []appcfg.FeatureFlag{},
		})

	mockDB.EXPECT().GetAllFlags().Return(nil, errors.Error("db error"))

	test.Request(t).
		WithMethod(http.MethodGet).
		WithURL("/api/v2/features").
		OnHandlerFunc(resources.GetFlags).
		Require().
		ResponseStatusCode(http.StatusInternalServerError)
}

func TestResources_ToggleFlag(t *testing.T) {
	const (
		featureID    = int32(1)
		featureIDStr = "1"
	)

	var (
		mockCtrl     = gomock.NewController(t)
		mockDB       = mocks.NewMockDatabase(mockCtrl)
		resources    = v2.Resources{DB: mockDB}
		requestSetup = test.Request(t).
				WithMethod(http.MethodGet).
				WithURL("/api/v2/features/%s/toggle", featureIDStr).
				WithURLPathVars(map[string]string{api.URIPathVariableFeatureID: featureIDStr}).
				OnHandlerFunc(resources.ToggleFlag)
	)

	defer mockCtrl.Finish()

	mockDB.EXPECT().GetFlag(featureID).Return(appcfg.FeatureFlag{
		UserUpdatable: false,
	}, nil)

	requestSetup.Require().ResponseStatusCode(http.StatusForbidden)

	mockDB.EXPECT().GetFlag(featureID).Return(appcfg.FeatureFlag{
		UserUpdatable: true,
	}, nil)
	mockDB.EXPECT().SetFlag(gomock.Any()).Return(nil)

	requestSetup.Require().
		ResponseStatusCode(http.StatusOK).
		ResponseJSONBody(v2.ToggleFlagResponse{
			Enabled: true,
		})
}
