// Copyright 2024 Specter Ops, Inc.
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

package auth_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/api/v2/auth"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/test"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_CreateOIDCProvider(t *testing.T) {
	var (
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
		config            = model.SSOProviderConfig{}
	)
	defer mockCtrl.Finish()

	t.Run("successfully create a new OIDCProvider", func(t *testing.T) {
		mockDB.EXPECT().GetRole(gomock.Any(), int32(0)).Return(model.Role{}, nil)
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "Bloodhound gang", "https://localhost/auth", "bloodhound", config).Return(model.OIDCProvider{
			ClientID: "bloodhound",
			Issuer:   "https://localhost/auth",
		}, nil)

		test.Request(t).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Bloodhound gang",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
				Config:   &config,
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusCreated)
	})

	t.Run("error parsing body request", func(t *testing.T) {
		test.Request(t).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error validating request field", func(t *testing.T) {
		test.Request(t).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "1234:not:a:url",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error invalid Issuer", func(t *testing.T) {
		request := auth.UpsertOIDCProviderRequest{
			Issuer: "12345:bloodhound",
		}
		test.Request(t).
			WithBody(request).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error creating oidc provider db entry", func(t *testing.T) {
		mockDB.EXPECT().GetRole(gomock.Any(), int32(0)).Return(model.Role{}, nil)
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "test", "https://localhost/auth", "bloodhound", config).Return(model.OIDCProvider{}, fmt.Errorf("error"))

		test.Request(t).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
				Config:   &config,
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}

func TestManagementResource_UpdateOIDCProvider(t *testing.T) {
	var (
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
		baseProvider      = model.SSOProvider{
			Type: model.SessionAuthProviderOIDC,
			Name: "Gotham Net",
			OIDCProvider: &model.OIDCProvider{
				ClientID: "gotham-net",
				Issuer:   "https://gotham.net",
			},
		}
		urlParams = map[string]string{"sso_provider_id": "1"}
	)
	defer mockCtrl.Finish()

	t.Run("successfully update an OIDCProvider", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)
		mockDB.EXPECT().UpdateOIDCProvider(gomock.Any(), gomock.Any())

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Gotham Net 2",
				Issuer:   "https://gotham-2.net",
				ClientID: "gotham-net-2",
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("error not found while updating an unknown OIDCProvider", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{}, database.ErrNotFound)

		test.Request(t).
			WithURLPathVars(urlParams).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusNotFound)
	})

	t.Run("error parsing body request", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)

		test.Request(t).
			WithURLPathVars(urlParams).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error validating request field", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "1234:not:a:url",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error creating oidc provider db entry", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)
		mockDB.EXPECT().UpdateOIDCProvider(gomock.Any(), gomock.Any()).Return(model.OIDCProvider{}, fmt.Errorf("error"))

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}
