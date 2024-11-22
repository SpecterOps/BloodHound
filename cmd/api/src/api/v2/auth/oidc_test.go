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
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/test"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_CreateOIDCProvider(t *testing.T) {
	const (
		url = "/api/v2/sso/providers/oidc"
	)
	var (
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
	)
	defer mockCtrl.Finish()

	t.Run("successfully create a new OIDCProvider", func(t *testing.T) {
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "Bloodhound gang", "https://localhost/auth", "bloodhound").Return(model.OIDCProvider{
			ClientID: "bloodhound",
			Issuer:   "https://localhost/auth",
		}, nil)

		test.Request(t).
			WithMethod(http.MethodPost).
			WithURL(url).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Bloodhound gang",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusCreated)
	})

	t.Run("error parsing body request", func(t *testing.T) {
		test.Request(t).
			WithMethod(http.MethodPost).
			WithURL(url).
			WithBody("").
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error validating request field", func(t *testing.T) {
		test.Request(t).
			WithMethod(http.MethodPost).
			WithURL(url).
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
			WithMethod(http.MethodPost).
			WithURL(url).
			WithBody(request).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error creating oidc provider db entry", func(t *testing.T) {
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "test", "https://localhost/auth", "bloodhound").Return(model.OIDCProvider{}, fmt.Errorf("error"))

		test.Request(t).
			WithMethod(http.MethodPost).
			WithURL(url).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}
