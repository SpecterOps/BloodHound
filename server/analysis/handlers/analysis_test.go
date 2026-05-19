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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/analysis/handlers"
	"github.com/specterops/bloodhound/server/analysis/handlers/mocks"
	"github.com/specterops/bloodhound/server/analysis/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newAuthenticatedRequest returns an *http.Request whose context carries a
// bhctx.Context with the supplied user wired in as the auth Owner. This mirrors
// what the auth middleware does for real requests.
func newAuthenticatedRequest(t *testing.T, method, target string, userID uuid.UUID) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, target, nil)
	require.NoError(t, err)
	bhCtx := &bhctx.Context{
		AuthCtx: auth.Context{Owner: model.User{Unique: model.Unique{ID: userID}}},
	}
	return bhctx.SetRequestContext(req, bhCtx)
}

func TestHandlers_GetRequest(t *testing.T) {
	newRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, "/api/v2/analysis/status", nil)
		require.NoError(t, err)
		return req
	}

	t.Run("returns 200 with the analysis request view on success", func(t *testing.T) {
		var (
			expected = service.RequestedAnalysis{
				RequestedBy:         "test-user",
				RequestType:         service.RequestedAnalysisTypeAnalysis,
				RequestedAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				DeleteSourceKinds:   []string{"AZBase"},
				DeleteRelationships: []string{"HasSession"},
			}
			analysisMock = mocks.NewMockAnalysis(t)
			handlerSet   = handlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newRequest(t)
		)

		analysisMock.EXPECT().GetRequest(mock.Anything).Return(expected, nil)

		handlerSet.GetRequest(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var envelope struct {
			Data handlers.RequestedAnalysisView `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
		assert.Equal(t, expected.RequestedBy, envelope.Data.RequestedBy)
		assert.Equal(t, expected.RequestType, envelope.Data.RequestType)
		assert.Equal(t, expected.DeleteSourceKinds, envelope.Data.DeleteSourceKinds)
		assert.Equal(t, expected.DeleteRelationships, envelope.Data.DeleteRelationships)
	})

	t.Run("returns 204 No Content when no request is pending", func(t *testing.T) {
		var (
			analysisMock = mocks.NewMockAnalysis(t)
			handlerSet   = handlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newRequest(t)
		)

		analysisMock.EXPECT().GetRequest(mock.Anything).Return(service.RequestedAnalysis{}, service.ErrNoPendingRequest)

		handlerSet.GetRequest(recorder, request)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Empty(t, recorder.Body.Bytes())
	})

	t.Run("returns 500 on unexpected service errors", func(t *testing.T) {
		var (
			unexpectedErr = errors.New("unexpected database failure")
			analysisMock  = mocks.NewMockAnalysis(t)
			handlerSet    = handlers.NewHandlersContainer(analysisMock)
			recorder      = httptest.NewRecorder()
			request       = newRequest(t)
		)

		analysisMock.EXPECT().GetRequest(mock.Anything).Return(service.RequestedAnalysis{}, unexpectedErr)

		handlerSet.GetRequest(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}

func TestHandlers_CreateRequest(t *testing.T) {
	var (
		userID        = uuid.Must(uuid.NewV4())
		userIDString  = userID.String()
		requestedAt   = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		createdResult = service.RequestedAnalysis{
			RequestedBy: userIDString,
			RequestType: service.RequestedAnalysisTypeAnalysis,
			RequestedAt: requestedAt,
		}
		existingResult = service.RequestedAnalysis{
			RequestedBy: "other-user",
			RequestType: service.RequestedAnalysisTypeAnalysis,
			RequestedAt: requestedAt,
		}
	)

	decodeBody := func(t *testing.T, body []byte) handlers.RequestedAnalysisView {
		t.Helper()
		var envelope struct {
			Data handlers.RequestedAnalysisView `json:"data"`
		}
		require.NoError(t, json.Unmarshal(body, &envelope))
		return envelope.Data
	}

	t.Run("returns 202 Accepted with the new request body when this call accepted it", func(t *testing.T) {
		var (
			analysisMock = mocks.NewMockAnalysis(t)
			handlerSet   = handlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newAuthenticatedRequest(t, http.MethodPut, "/api/v2/analysis", userID)
		)

		analysisMock.EXPECT().CreateRequest(mock.Anything, userIDString).Return(createdResult, true, nil)

		handlerSet.CreateRequest(recorder, request)

		assert.Equal(t, http.StatusAccepted, recorder.Code)
		assert.Equal(t, userIDString, decodeBody(t, recorder.Body.Bytes()).RequestedBy)
	})

	t.Run("returns 200 OK with the existing request body when one was already pending", func(t *testing.T) {
		var (
			analysisMock = mocks.NewMockAnalysis(t)
			handlerSet   = handlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newAuthenticatedRequest(t, http.MethodPut, "/api/v2/analysis", userID)
		)

		analysisMock.EXPECT().CreateRequest(mock.Anything, userIDString).Return(existingResult, false, nil)

		handlerSet.CreateRequest(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, existingResult.RequestedBy, decodeBody(t, recorder.Body.Bytes()).RequestedBy)
	})

	t.Run("returns 401 Unauthorized when no authenticated user is present", func(t *testing.T) {
		var (
			analysisMock = mocks.NewMockAnalysis(t)
			handlerSet   = handlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
		)

		req, err := http.NewRequest(http.MethodPut, "/api/v2/analysis", nil)
		require.NoError(t, err)

		handlerSet.CreateRequest(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "authentication is required")
		// The service must not be invoked when authentication fails.
		analysisMock.AssertNotCalled(t, "CreateRequest", mock.Anything, mock.Anything)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		var (
			expectedErr  = errors.New("db unavailable")
			analysisMock = mocks.NewMockAnalysis(t)
			handlerSet   = handlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newAuthenticatedRequest(t, http.MethodPut, "/api/v2/analysis", userID)
		)

		analysisMock.EXPECT().CreateRequest(mock.Anything, userIDString).Return(service.RequestedAnalysis{}, false, expectedErr)

		handlerSet.CreateRequest(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}
