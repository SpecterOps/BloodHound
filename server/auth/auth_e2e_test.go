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
	"net/http"
	"testing"

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
