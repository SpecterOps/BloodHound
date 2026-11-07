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

package handlers_test

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/auth/internal/handlers"
	"github.com/specterops/bloodhound/server/auth/internal/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers_GetSelf_PreservesLegacyResponse(t *testing.T) {
	t.Parallel()

	var (
		timestamp = time.Date(2026, time.November, 6, 12, 30, 0, 0, time.UTC)
		metadata  = model.Basic{
			CreatedAt: timestamp,
			UpdatedAt: timestamp.Add(time.Hour),
			DeletedAt: sql.NullTime{Time: timestamp.Add(2 * time.Hour), Valid: true},
		}
		serial = model.Serial{ID: 42, Basic: metadata}
		userID = uuid.FromStringOrNil("c8602481-4dd1-43d3-9026-4dd8fc8e9691")
		tests  = []struct {
			name  string
			owner any
		}{
			{
				name: "authenticated user",
				owner: model.User{
					Unique:          model.Unique{ID: userID, Basic: metadata},
					PrincipalName:   "john.doe",
					FirstName:       null.StringFrom("John"),
					LastName:        null.StringFrom("Doe"),
					EmailAddress:    null.StringFrom("john.doe@example.com"),
					SSOProviderID:   null.Int32From(7),
					LastLogin:       timestamp,
					IsDisabled:      true,
					AllEnvironments: true,
					EULAAccepted:    true,
					SupportAccount:  true,
					AuthSecret: &model.AuthSecret{
						Serial:        serial,
						UserID:        userID,
						Digest:        "not exposed",
						TOTPSecret:    "not exposed",
						DigestMethod:  "argon2",
						ExpiresAt:     timestamp,
						TOTPActivated: true,
					},
					EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{{
						BigSerial:     model.BigSerial{ID: 64, Basic: metadata},
						UserID:        userID.String(),
						EnvironmentID: "example.com",
					}},
					Roles: model.Roles{{
						Serial:      serial,
						Name:        "User",
						Description: "Example role",
						Permissions: model.Permissions{{Serial: serial, Authority: "auth", Name: "read"}},
					}},
				},
			},
			{
				name:  "empty user",
				owner: model.User{},
			},
			{
				name: "empty arrays",
				owner: model.User{
					Roles:                            model.Roles{},
					EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{},
				},
			},
			{
				name: "nil and empty permission arrays",
				owner: model.User{
					Roles: model.Roles{{Permissions: nil}, {Permissions: model.Permissions{}}},
				},
			},
			{
				name: "valid zero and empty nullable values",
				owner: model.User{
					SSOProviderID: null.Int32From(0),
					FirstName:     null.StringFrom(""),
					LastName:      null.StringFrom(""),
					EmailAddress:  null.StringFrom(""),
				},
			},
			{
				name: "invalid nullable values do not expose their contents",
				owner: model.User{
					SSOProviderID: null.NewInt32(7, false),
					FirstName:     null.NewString("hidden first name", false),
					LastName:      null.NewString("hidden last name", false),
					EmailAddress:  null.NewString("hidden@example.com", false),
				},
			},
			{
				name: "unauthenticated request",
			},
		}
	)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				// No service expectations: GetSelf must only use the request context.
				handlerSet = handlers.NewHandlersContainer(mocks.NewMockAuth(t))
				request    = bhctx.SetRequestContext(httptest.NewRequest(http.MethodGet, "/api/v2/self", nil), &bhctx.Context{
					AuthCtx: auth.Context{Owner: testCase.owner},
				})
				recorder       = httptest.NewRecorder()
				legacyRecorder = httptest.NewRecorder()
			)

			api.WriteBasicResponse(request.Context(), testCase.owner, http.StatusOK, legacyRecorder)
			handlerSet.GetSelf(recorder, request)

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.Equal(t, legacyRecorder.Header(), recorder.Header())
			assert.JSONEq(t, legacyRecorder.Body.String(), recorder.Body.String())
		})
	}
}

func TestHandlers_GetSelf_WithoutRequestContext(t *testing.T) {
	t.Parallel()

	var (
		handlerSet = handlers.NewHandlersContainer(nil)
		request    = httptest.NewRequest(http.MethodGet, "/api/v2/self", nil)
		recorder   = httptest.NewRecorder()
	)

	// A nil service also makes any accidental dependency on the service fail.
	require.NotPanics(t, func() { handlerSet.GetSelf(recorder, request) })
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"data":null}`, recorder.Body.String())
}

func TestHandlers_Logout(t *testing.T) {
	t.Parallel()

	var (
		sessionID       = int64(42)
		host            = &url.URL{Scheme: "https", Host: "bloodhound.example.com"}
		wantRedirectURL = "https://bloodhound.example.com" + handlers.UserInterfacePath
		serviceErr      = errors.New("db unavailable")
	)

	tests := []struct {
		name      string
		sessionID int64
		svcErr    error
	}{
		{
			name:      "redirects to the UI on success",
			sessionID: sessionID,
		},
		{
			name:      "still redirects to the UI when ending the session fails",
			sessionID: sessionID,
			svcErr:    serviceErr,
		},
		{
			name:      "still redirects to the UI when the request has no session",
			sessionID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				authMock   = mocks.NewMockAuth(t)
				handlerSet = handlers.NewHandlersContainer(authMock)
				recorder   = httptest.NewRecorder()
				request    = newSessionRequest(t, tt.sessionID, host)
			)

			authMock.EXPECT().Logout(request.Context(), tt.sessionID).Return(tt.svcErr)

			handlerSet.Logout(recorder, request)

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.Equal(t, wantRedirectURL, recorder.Header().Get("Location"))
		})
	}
}
