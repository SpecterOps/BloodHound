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
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/auth"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogout(t *testing.T) {
	var (
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInst *router.Router, db *database.BloodhoundDB) {
			auth.Register(routerInst, db.Pool())
		})
		server          = harness.Server
		wantRedirectURL = server.URL + "/ui"
	)

	newLogoutRequest := func(t *testing.T, token string) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v2/logout", nil)
		require.NoError(t, err)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return req
	}

	t.Run("ends the session so its token is rejected afterwards", func(t *testing.T) {
		var (
			user = model.User{
				PrincipalName: "logout-user@example.com",
				EmailAddress:  null.StringFrom("logout-user@example.com"),
				EULAAccepted:  true,
			}
			token = servertest.MintJWT(t, ctx, harness.DB, harness.Auther, user)
		)

		resp, err := http.DefaultClient.Do(newLogoutRequest(t, token))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, wantRedirectURL, resp.Header.Get("Location"))

		reusedResp, err := http.DefaultClient.Do(newLogoutRequest(t, token))
		require.NoError(t, err)
		defer reusedResp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, reusedResp.StatusCode,
			"a token whose session was ended by logout must be rejected by the auth middleware")
	})

	t.Run("redirects to the UI when no authentication token is provided", func(t *testing.T) {
		resp, err := http.DefaultClient.Do(newLogoutRequest(t, ""))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, wantRedirectURL, resp.Header.Get("Location"))
	})

	t.Run("returns 401 Unauthorized when an invalid token is provided", func(t *testing.T) {
		resp, err := http.DefaultClient.Do(newLogoutRequest(t, "invalid-token-that-is-not-valid"))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, resp.Header.Get("Location"), "the request must be rejected before reaching the logout handler")
	})

	t.Run("returns 400 Bad Request when Bearer prefix is missing", func(t *testing.T) {
		var (
			user = model.User{
				PrincipalName: "logout-no-bearer-user@example.com",
				EmailAddress:  null.StringFrom("logout-no-bearer-user@example.com"),
				EULAAccepted:  true,
			}
			token = servertest.MintJWT(t, ctx, harness.DB, harness.Auther, user)
		)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v2/logout", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Empty(t, resp.Header.Get("Location"), "the request must be rejected before reaching the logout handler")
	})
}

// TestGetSelfSessionLifecycle covers auth middleware, slice registration and
// session revocation together. Detailed JSON compatibility belongs in handler tests.
func TestGetSelfSessionLifecycle(t *testing.T) {
	var (
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInstance *router.Router, databaseInstance *database.BloodhoundDB) {
			auth.Register(routerInstance, databaseInstance.Pool())
		})
		client = harness.Server.Client()
		target = harness.Server.URL + "/api/v2/self"
		user   = model.User{
			PrincipalName: "self-user@example.com",
			EmailAddress:  null.StringFrom("self-user@example.com"),
			FirstName:     null.StringFrom("Self"),
			// No roles or EULA acceptance: reading self must not require permissions.
		}
		token       = servertest.MintJWT(t, ctx, harness.DB, harness.Auther, user)
		logoutToken = servertest.MintJWT(t, ctx, harness.DB, harness.Auther, model.User{
			PrincipalName: "self-logout-user@example.com",
		})
	)

	client.Timeout = 10 * time.Second

	testCases := []struct {
		name       string
		token      string
		setup      func(t *testing.T)
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:       "returns the authenticated user without requiring permissions",
			token:      token,
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data *struct {
						ID            uuid.UUID `json:"id"`
						PrincipalName string    `json:"principal_name"`
						EmailAddress  string    `json:"email_address"`
						FirstName     string    `json:"first_name"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				require.NotNil(t, envelope.Data)
				require.NotEqual(t, uuid.Nil, envelope.Data.ID)

				// Check against persistence rather than trusting a nonempty response ID.
				persistedUser, err := harness.DB.GetUser(ctx, envelope.Data.ID)
				require.NoError(t, err)
				assert.Equal(t, user.PrincipalName, persistedUser.PrincipalName)
				assert.Equal(t, user.PrincipalName, envelope.Data.PrincipalName)
				assert.Equal(t, user.EmailAddress.String, envelope.Data.EmailAddress)
				assert.Equal(t, user.FirstName.String, envelope.Data.FirstName)
			},
		},
		{
			name:  "rejects the session after logout",
			token: logoutToken,
			setup: func(t *testing.T) {
				request, err := http.NewRequestWithContext(ctx, http.MethodPost, harness.Server.URL+"/api/v2/logout", nil)
				require.NoError(t, err)
				request.Header.Set("Authorization", "Bearer "+logoutToken)

				response, err := client.Do(request)
				require.NoError(t, err)
				defer response.Body.Close()

				require.Equal(t, http.StatusOK, response.StatusCode)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns a null payload when credentials are omitted",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				assert.JSONEq(t, `{"data":null}`, string(body))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(t)
			}

			request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
			require.NoError(t, err)
			if testCase.token != "" {
				request.Header.Set("Authorization", "Bearer "+testCase.token)
			}

			response, err := client.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, testCase.wantStatus, response.StatusCode, "body: %s", body)
			assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
			if testCase.assertBody != nil {
				testCase.assertBody(t, body)
			}
		})
	}
}
