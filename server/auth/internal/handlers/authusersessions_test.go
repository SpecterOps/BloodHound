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
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/server/auth/internal/handlers"
	"github.com/specterops/bloodhound/server/auth/internal/handlers/mocks"
	"github.com/stretchr/testify/assert"
)

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
