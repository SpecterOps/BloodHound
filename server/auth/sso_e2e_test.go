// Copyright 2026 Specter Ops, Inc.
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

//go:build integration

package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/auth"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertSSOProviderNames checks response ordering and hydrated provider details.
func assertSSOProviderNames(t *testing.T, response *http.Response, expectedNames []string) {
	t.Helper()

	var (
		envelope struct {
			Data []struct {
				Name    string         `json:"name"`
				Details map[string]any `json:"details"`
			} `json:"data"`
		}
		names []string
	)

	require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))

	for _, provider := range envelope.Data {
		names = append(names, provider.Name)
		assert.NotNil(t, provider.Details)
	}

	assert.Equal(t, expectedNames, names)
}

func TestE2E_ListAuthProviders(t *testing.T) {
	var (
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInstance *router.Router, databaseInstance *database.BloodhoundDB) {
			auth.Register(routerInstance, databaseInstance.Pool())
		})
		target = harness.Server.URL + "/api/v2/sso-providers"
	)

	// Populate through the legacy write APIs to check compatibility with persisted records.
	_, err := harness.DB.CreateOIDCProvider(ctx, "Zulu OIDC", "https://issuer.example.com", "client-id", model.SSOProviderConfig{})
	require.NoError(t, err)

	_, err = harness.DB.CreateSAMLIdentityProvider(ctx, model.SAMLProvider{Name: "Alpha SAML", IssuerURI: "https://saml.example.com"}, model.SSOProviderConfig{})
	require.NoError(t, err)

	flag, err := harness.DB.GetFlagByKey(ctx, "oidc_support")
	require.NoError(t, err)

	flag.Enabled = true
	require.NoError(t, harness.DB.SetFlag(ctx, flag))

	testCases := []struct {
		name       string
		query      string
		setup      func(t *testing.T)
		wantStatus int
		assertBody func(t *testing.T, response *http.Response)
	}{
		{
			name:       "Unauthenticated requests receive providers in name order",
			query:      "sort_by=name",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response *http.Response) {
				assertSSOProviderNames(t, response, []string{"Alpha SAML", "Zulu OIDC"})
			},
		},
		{
			name:       "Descending sort reverses provider name order",
			query:      "sort_by=-name",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response *http.Response) {
				assertSSOProviderNames(t, response, []string{"Zulu OIDC", "Alpha SAML"})
			},
		},
		{
			name:       "Name filter returns the matching provider",
			query:      "name=eq:Zulu%20OIDC",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response *http.Response) {
				assertSSOProviderNames(t, response, []string{"Zulu OIDC"})
			},
		},
		{
			name:       "Numeric type filter returns OIDC providers",
			query:      "type=eq:2",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response *http.Response) {
				assertSSOProviderNames(t, response, []string{"Zulu OIDC"})
			},
		},
		{
			name:       "Date-only timestamp filter preserves legacy behavior",
			query:      "created_at=gte:2026-01-01&sort_by=name",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response *http.Response) {
				assertSSOProviderNames(t, response, []string{"Alpha SAML", "Zulu OIDC"})
			},
		},
		{
			name:       "Null creation timestamp filter is accepted",
			query:      "created_at=eq:null",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response *http.Response) {
				assertSSOProviderNames(t, response, nil)
			},
		},
		{
			name:       "Non-null update timestamp filter is accepted",
			query:      "updated_at=neq:null&sort_by=name",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response *http.Response) {
				assertSSOProviderNames(t, response, []string{"Alpha SAML", "Zulu OIDC"})
			},
		},
		{
			name:       "Non-sortable column returns 400",
			query:      "sort_by=details",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Unsupported filter operator returns 400",
			query:      "name=gt:x",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Unknown filter field returns 400",
			query:      "details=eq:x",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "OIDC feature flag disabled returns only SAML providers",
			setup: func(t *testing.T) {
				disabledFlag := flag
				disabledFlag.Enabled = false
				require.NoError(t, harness.DB.SetFlag(ctx, disabledFlag))
				t.Cleanup(func() { require.NoError(t, harness.DB.SetFlag(ctx, flag)) })
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response *http.Response) {
				var envelope struct {
					Data []struct {
						Type string `json:"type"`
					} `json:"data"`
				}

				require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))
				require.Len(t, envelope.Data, 1)
				assert.Equal(t, "SAML", envelope.Data[0].Type)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(t)
			}

			requestTarget := target
			if testCase.query != "" {
				requestTarget += "?" + testCase.query
			}

			response, err := harness.Server.Client().Get(requestTarget)
			require.NoError(t, err)
			defer response.Body.Close()

			assert.Equal(t, testCase.wantStatus, response.StatusCode)
			if testCase.assertBody != nil {
				testCase.assertBody(t, response)
			}
		})
	}
}
