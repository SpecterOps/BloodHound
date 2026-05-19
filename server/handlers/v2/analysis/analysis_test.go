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

package analysis_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	analysishandlers "github.com/specterops/bloodhound/server/handlers/v2/analysis"
	analysisMocks "github.com/specterops/bloodhound/server/handlers/v2/analysis/mocks"
	"github.com/specterops/bloodhound/server/models"
	analysisservice "github.com/specterops/bloodhound/server/services/analysis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlers_GetRequest(t *testing.T) {
	newRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, "/api/v2/analysis/status", nil)
		require.NoError(t, err)
		return req
	}

	t.Run("returns 200 with the analysis request view on success", func(t *testing.T) {
		var (
			expected = models.RequestedAnalysis{
				RequestedBy:         "test-user",
				RequestType:         models.RequestedAnalysisTypeAnalysis,
				RequestedAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				DeleteSourceKinds:   []string{"AZBase"},
				DeleteRelationships: []string{"HasSession"},
			}
			analysisMock = analysisMocks.NewMockAnalysis(t)
			handlers     = analysishandlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newRequest(t)
		)

		analysisMock.EXPECT().GetRequest(mock.Anything).Return(expected, nil)

		handlers.GetRequest(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var envelope struct {
			Data analysishandlers.RequestedAnalysisView `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
		assert.Equal(t, expected.RequestedBy, envelope.Data.RequestedBy)
		assert.Equal(t, expected.RequestType, envelope.Data.RequestType)
		assert.Equal(t, expected.DeleteSourceKinds, envelope.Data.DeleteSourceKinds)
		assert.Equal(t, expected.DeleteRelationships, envelope.Data.DeleteRelationships)
	})

	t.Run("returns 200 with a zero-value view when no request is pending", func(t *testing.T) {
		var (
			analysisMock = analysisMocks.NewMockAnalysis(t)
			handlers     = analysishandlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newRequest(t)
		)

		analysisMock.EXPECT().GetRequest(mock.Anything).Return(models.RequestedAnalysis{}, analysisservice.ErrNoPendingRequest)

		handlers.GetRequest(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope struct {
			Data analysishandlers.RequestedAnalysisView `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
		assert.Empty(t, envelope.Data.RequestedBy)
	})

	t.Run("returns 500 on unexpected service errors", func(t *testing.T) {
		var (
			unexpectedErr = errors.New("unexpected database failure")
			analysisMock  = analysisMocks.NewMockAnalysis(t)
			handlers      = analysishandlers.NewHandlersContainer(analysisMock)
			recorder      = httptest.NewRecorder()
			request       = newRequest(t)
		)

		analysisMock.EXPECT().GetRequest(mock.Anything).Return(models.RequestedAnalysis{}, unexpectedErr)

		handlers.GetRequest(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}

func TestHandlers_CreateRequest(t *testing.T) {
	newRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodPut, "/api/v2/analysis", nil)
		require.NoError(t, err)
		return req
	}

	t.Run("returns 202 Accepted on success", func(t *testing.T) {
		var (
			analysisMock = analysisMocks.NewMockAnalysis(t)
			handlers     = analysishandlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newRequest(t)
		)

		// No auth context present — handler falls back to "unknown-user".
		analysisMock.EXPECT().CreateRequest(mock.Anything, "unknown-user").Return(models.RequestedAnalysis{}, true, nil)

		handlers.CreateRequest(recorder, request)

		assert.Equal(t, http.StatusAccepted, recorder.Code)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		var (
			expectedErr  = errors.New("db unavailable")
			analysisMock = analysisMocks.NewMockAnalysis(t)
			handlers     = analysishandlers.NewHandlersContainer(analysisMock)
			recorder     = httptest.NewRecorder()
			request      = newRequest(t)
		)

		analysisMock.EXPECT().CreateRequest(mock.Anything, "unknown-user").Return(models.RequestedAnalysis{}, false, expectedErr)

		handlers.CreateRequest(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}
