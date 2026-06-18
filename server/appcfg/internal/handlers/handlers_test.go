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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/handlers"
	"github.com/specterops/bloodhound/server/appcfg/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers_GetDatapipeStatus(t *testing.T) {
	newRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, "/api/v2/datapipe/status", nil)
		require.NoError(t, err)
		return req
	}

	var (
		unexpectedErr     = errors.New("unexpected database failure")
		nextScheduledTime = time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
		expected          = services.DatapipeStatus{
			Status:                  services.DatapipeStatusIdle,
			UpdatedAt:               time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC),
			LastCompleteAnalysisAt:  null.TimeFrom(time.Date(2026, 6, 18, 9, 30, 0, 0, time.UTC)),
			LastAnalysisRunAt:       null.TimeFrom(time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)),
			NextScheduledAnalysisAt: null.TimeFrom(nextScheduledTime),
		}
	)

	tests := []struct {
		name       string
		svcResult  services.DatapipeStatus
		svcErr     error
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:       "returns 200 with the datapipe status view on success",
			svcResult:  expected,
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.DatapipeStatusView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, services.DatapipeStatusIdle, envelope.Data.Status)
				assert.Equal(t, expected.UpdatedAt, envelope.Data.UpdatedAt)
				assert.Equal(t, expected.LastCompleteAnalysisAt, envelope.Data.LastCompleteAnalysisAt)
				assert.Equal(t, expected.LastAnalysisRunAt, envelope.Data.LastAnalysisRunAt)
				assert.True(t, envelope.Data.NextScheduledAnalysisAt.Valid)
				assert.Equal(t, nextScheduledTime, envelope.Data.NextScheduledAnalysisAt.Time)
			},
		},
		{
			name:       "returns 404 when datapipe status not found",
			svcErr:     services.ErrNotFound,
			wantStatus: http.StatusNotFound,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Errors []struct {
						Message string `json:"message"`
					} `json:"errors"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.NotEmpty(t, envelope.Errors)
			},
		},
		{
			name:       "returns 500 on unexpected database error",
			svcErr:     unexpectedErr,
			wantStatus: http.StatusInternalServerError,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Errors []struct {
						Message string `json:"message"`
					} `json:"errors"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.NotEmpty(t, envelope.Errors)
			},
		},
		{
			name:       "returns 500 on context deadline exceeded",
			svcErr:     context.DeadlineExceeded,
			wantStatus: http.StatusInternalServerError,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Errors []struct {
						Message string `json:"message"`
					} `json:"errors"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.NotEmpty(t, envelope.Errors)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mockSvc = mocks.NewMockService(t)
				h       = handlers.NewHandlers(mockSvc)
				rr      = httptest.NewRecorder()
				req     = newRequest(t)
			)

			mockSvc.EXPECT().
				GetDatapipeStatus(req.Context()).
				Return(tt.svcResult, tt.svcErr).
				Once()

			h.GetDatapipeStatus(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, rr.Body.Bytes())
			}
		})
	}
}
