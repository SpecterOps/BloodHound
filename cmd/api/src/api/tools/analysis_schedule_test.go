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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/tools"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestToolContainer_GetScheduledAnalysisConfiguration(t *testing.T) {
	endpoint := "/analysis/schedule"

	t.Run("success getting scheduled analysis", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			mockDB   = mocks.NewMockDatabase(ctrl)
			handlers = tools.NewToolContainer(mockDB)
		)

		defer ctrl.Finish()

		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})

		scheduledAnalysisValue, _ := types.NewJSONBObject(map[string]any{
			"enabled": true,
			"rrule":   "FREQ=DAILY;INTERVAL=1;DTSTART=20230101T100000Z",
		})

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{
			Key:   appcfg.ScheduledAnalysis,
			Value: scheduledAnalysisValue,
		}, nil)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpoint, nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.GetScheduledAnalysisConfiguration).Methods(http.MethodGet)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)
			require.Contains(t, response.Body.String(), `"enabled":true`)
			require.Contains(t, response.Body.String(), `"rrule":"FREQ=DAILY;INTERVAL=1;DTSTART=20230101T100000Z"`)
		}
	})

	t.Run("returns error when GetConfigurationParameter fails", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			mockDB   = mocks.NewMockDatabase(ctrl)
			handlers = tools.NewToolContainer(mockDB)
		)

		defer ctrl.Finish()

		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{}, fmt.Errorf("database connection lost"))

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpoint, nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.GetScheduledAnalysisConfiguration).Methods(http.MethodGet)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusInternalServerError, response.Code)
			require.Contains(t, response.Body.String(), "error retrieving configuration data")
		}
	})
}

func TestToolContainer_SetScheduledAnalysisConfiguration(t *testing.T) {
	endpoint := "/analysis/schedule"
	validRRule := "FREQ=DAILY;INTERVAL=1;DTSTART=20230101T100000Z"

	t.Run("success enabling scheduled analysis", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			mockDB   = mocks.NewMockDatabase(ctrl)
			handlers = tools.NewToolContainer(mockDB)
		)

		defer ctrl.Finish()

		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   validRRule,
		}

		jsonValue, _ := types.NewJSONBObject(appcfg.ScheduledAnalysisParameter{
			Enabled: true,
			RRule:   validRRule,
		})

		parameterConfig := appcfg.Parameter{
			Key:   appcfg.ScheduledAnalysis,
			Value: jsonValue,
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		mockDB.EXPECT().SetConfigurationParameter(gomock.Any(), parameterConfig).Return(nil)
		mockDB.EXPECT().SetNextScheduledAnalysisStartTime(gomock.Any(), gomock.Any()).Return(nil)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)
		}
	})

	t.Run("success -- disable scheduled analysis", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			mockDB   = mocks.NewMockDatabase(ctrl)
			handlers = tools.NewToolContainer(mockDB)
		)

		defer ctrl.Finish()

		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: false,
		}

		jsonValue, _ := types.NewJSONBObject(appcfg.ScheduledAnalysisParameter{
			Enabled: false})

		parameterConfig := appcfg.Parameter{
			Key:   appcfg.ScheduledAnalysis,
			Value: jsonValue,
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		mockDB.EXPECT().SetConfigurationParameter(gomock.Any(), parameterConfig).Return(nil)
		mockDB.EXPECT().SetNextScheduledAnalysisStartTime(gomock.Any(), time.Time{}).Return(nil)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)
		}
	})

	t.Run("returns error on invalid rrule", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

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
			require.Contains(t, response.Body.String(), "invalid rrule specified")
		}
	})

	t.Run("returns error on rrule with count", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

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
			require.Contains(t, response.Body.String(), "invalid rrule specified: count not supported")
		}
	})

	t.Run("returns error on rrule with until", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

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
			require.Contains(t, response.Body.String(), "invalid rrule specified: until not supported")
		}
	})

	t.Run("returns error when SetConfigurationParameter fails", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			mockDB   = mocks.NewMockDatabase(ctrl)
			handlers = tools.NewToolContainer(mockDB)
		)

		defer ctrl.Finish()

		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "FREQ=DAILY;INTERVAL=1;DTSTART=20230101T100000Z",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		mockDB.EXPECT().SetConfigurationParameter(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database connection lost"))

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusInternalServerError, response.Code)
			require.Contains(t, response.Body.String(), "error setting analysis schedule")
		}
	})

	t.Run("returns error when SetNextScheduledAnalysisStartTime fails", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			mockDB   = mocks.NewMockDatabase(ctrl)
			handlers = tools.NewToolContainer(mockDB)
		)

		defer ctrl.Finish()

		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "FREQ=DAILY;INTERVAL=1;DTSTART=20230101T100000Z",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		mockDB.EXPECT().SetConfigurationParameter(gomock.Any(), gomock.Any()).Return(nil)
		mockDB.EXPECT().SetNextScheduledAnalysisStartTime(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database connection lost"))

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusInternalServerError, response.Code)
			require.Contains(t, response.Body.String(), "scheduled analysis updated, but there was an error updating the next scheduled analysis start time")
		}
	})

	t.Run("returns error on rrule without dtstart", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

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
