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
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func preferencesRequest(method, target string, owner uuid.UUID, body io.Reader) *http.Request {
	var request = httptest.NewRequest(method, "/api/v2/bloodhound-users/"+target+"/preferences", body)
	request = mux.SetURLVars(request, map[string]string{api.URIPathVariableUserID: target})
	request.Header.Set("Content-Type", "application/octet-stream")
	return request.WithContext(bhctx.Set(request.Context(), &bhctx.Context{
		AuthCtx: auth.Context{Owner: model.User{Unique: model.Unique{ID: owner}}},
	}))
}

func TestHandlers_GetUserPreferences(t *testing.T) {
	var userID = uuid.Must(uuid.NewV4())
	for _, testCase := range []struct {
		name         string
		target       string
		storage      []byte
		err          error
		status       int
		callsService bool
	}{
		{"binary", userID.String(), []byte{0, 255, 128, 13, 10}, nil, 200, true},
		{"empty", userID.String(), []byte{}, nil, 200, true},
		{"equivalent UUID", strings.ToUpper(userID.String()), []byte{1}, nil, 200, true},
		{"missing user", userID.String(), nil, services.ErrUserNotFound, 404, true},
		{"database failure", userID.String(), nil, errors.New("private database error"), 500, true},
		{"timeout", userID.String(), nil, context.DeadlineExceeded, 500, true},
		{"malformed", "bad-id", nil, nil, 400, false},
		{"another user", uuid.Must(uuid.NewV4()).String(), nil, nil, 403, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				service  = mocks.NewMockIdentity(t)
				handler  = handlers.NewHandlersContainer(service)
				response = httptest.NewRecorder()
				request  = preferencesRequest(http.MethodGet, testCase.target, userID, nil)
			)
			if testCase.callsService {
				service.EXPECT().GetUserPreferences(mock.Anything, userID).Return(testCase.storage, testCase.err).Once()
			}
			handler.GetUserPreferences(response, request)
			require.Equal(t, testCase.status, response.Code)
			require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
			if testCase.status == http.StatusOK {
				require.True(t, bytes.Equal(testCase.storage, response.Body.Bytes()))
				require.Equal(t, "application/octet-stream", response.Header().Get("Content-Type"))
			} else {
				require.Equal(t, "application/json", response.Header().Get("Content-Type"))
				require.NotContains(t, response.Body.String(), "private database error")
			}
		})
	}
}

type failedPreferencesReader struct{}

func (failedPreferencesReader) Read([]byte) (int, error) {
	return 0, errors.New("interrupted read")
}

func TestHandlers_UpsertUserPreferences(t *testing.T) {
	var userID = uuid.Must(uuid.NewV4())
	for _, testCase := range []struct {
		name          string
		contentType   string
		body          []byte
		readFailure   bool
		unknownLength bool
		err           error
		status        int
	}{
		{name: "binary", contentType: "application/octet-stream", body: []byte{0, 255, 128, 10}, status: 204},
		{name: "empty", contentType: "application/octet-stream", body: []byte{}, status: 204},
		{name: "media type parameters", contentType: "application/octet-stream; charset=binary", body: []byte{1}, status: 204},
		{name: "below limit", contentType: "application/octet-stream", body: make([]byte, services.MaxUserPreferencesBytes-1), status: 204},
		{name: "at limit", contentType: "application/octet-stream", body: make([]byte, services.MaxUserPreferencesBytes), status: 204},
		{name: "over limit", contentType: "application/octet-stream", body: make([]byte, services.MaxUserPreferencesBytes+1), status: 413},
		{name: "chunked over limit", contentType: "application/octet-stream", body: make([]byte, services.MaxUserPreferencesBytes+1), unknownLength: true, status: 413},
		{name: "missing content type", body: []byte{1}, status: 415},
		{name: "JSON content type", contentType: "application/json", body: []byte("{}"), status: 415},
		{name: "malformed content type", contentType: ";", status: 415},
		{name: "read failure", contentType: "application/octet-stream", readFailure: true, status: 400},
		{name: "missing user", contentType: "application/octet-stream", body: []byte{1}, err: services.ErrUserNotFound, status: 404},
		{name: "database limit", contentType: "application/octet-stream", body: []byte{1}, err: services.ErrUserPreferencesTooLarge, status: 413},
		{name: "database failure", contentType: "application/octet-stream", body: []byte{1}, err: errors.New("database unavailable"), status: 500},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				service  = mocks.NewMockIdentity(t)
				handler  = handlers.NewHandlersContainer(service)
				response = httptest.NewRecorder()
				request  = preferencesRequest(http.MethodPut, userID.String(), userID, bytes.NewReader(testCase.body))
			)
			request.Header.Set("Content-Type", testCase.contentType)
			if testCase.readFailure {
				request.Body = io.NopCloser(io.MultiReader(bytes.NewReader([]byte("partial body")), failedPreferencesReader{}))
			}
			if testCase.unknownLength {
				request.ContentLength = -1
				request.TransferEncoding = []string{"chunked"}
			}
			if testCase.status == 204 || testCase.err != nil {
				service.EXPECT().UpsertUserPreferences(mock.Anything, userID, testCase.body).Return(testCase.err).Once()
			}
			handler.UpsertUserPreferences(response, request)
			require.Equal(t, testCase.status, response.Code)
			require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
			if testCase.status == 204 {
				require.Empty(t, response.Body.Bytes())
			}
		})
	}
}

// Non-user principals must not be able to address user preferences, even when
// the route's permission middleware has already run.
func TestHandlers_UserPreferencesRejectNonUserOwner(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			var (
				userID   = uuid.Must(uuid.NewV4())
				service  = mocks.NewMockIdentity(t)
				handler  = handlers.NewHandlersContainer(service)
				request  = preferencesRequest(method, userID.String(), userID, bytes.NewReader([]byte{1}))
				response = httptest.NewRecorder()
			)
			bhctx.Get(request.Context()).AuthCtx.Owner = struct{}{}
			if method == http.MethodGet {
				handler.GetUserPreferences(response, request)
			} else {
				handler.UpsertUserPreferences(response, request)
			}
			require.Equal(t, http.StatusForbidden, response.Code)
		})
	}
}
