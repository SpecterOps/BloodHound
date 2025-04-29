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

package tools_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api/tools"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestToolContainer_GetScheduledAnalysisConfiguration_Errors(t *testing.T) {
	t.Run("returns error on invalid rrule", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

		endpoint := "/analysis/schedule"
		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "abc123",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, response.Body.String(), "invalid rrule specified: wrong")
		}
	})

	t.Run("returns error on rrule with count", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

		endpoint := "/analysis/schedule"
		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "FREQ=DAILY;INTERVAL=1;COUNT=3",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, response.Body.String(), "invalid rrule specified: count/until not supported")
		}
	})

	t.Run("returns error on rrule with until", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

		endpoint := "/analysis/schedule"
		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "FREQ=DAILY;INTERVAL=1;UNTIL=20240930T000000Z",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, response.Body.String(), "invalid rrule specified: count/until not supported")
		}
	})

	t.Run("returns error on rrule without dtstart", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

		endpoint := "/analysis/schedule"
		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "RRULE:FREQ=DAILY;INTERVAL=1",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, response.Body.String(), "invalid rrule specified: dtstart is required")
		}
	})
}
