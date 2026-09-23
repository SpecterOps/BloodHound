// Copyright 2023 Specter Ops, Inc.
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

package v2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/teambition/rrule-go"
	"go.uber.org/mock/gomock"
)

func Test_SetApplicationConfiguration(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}

		appConfigRequest = appcfg.AppConfigUpdateRequest{
			Key: string(appcfg.PasswordExpirationWindow),
			Value: map[string]any{
				"setting": "setting",
			},
		}
	)
	defer mockCtrl.Finish()

	t.Run("No payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	})

	t.Run("Invalid Parameters", func(t *testing.T) {
		invalidRequest := appcfg.AppConfigUpdateRequest{
			Key: "invalidKey",
			Value: map[string]any{
				"someKey": "someValue",
			},
		}
		reqBody, _ := json.Marshal(invalidRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	})

	t.Run("Error from DB", func(t *testing.T) {
		appConfigRequest = appcfg.AppConfigUpdateRequest{
			Key: string(appcfg.ReconciliationKey),
			Value: map[string]any{
				"enabled": true,
			},
		}
		mockDB.EXPECT().
			SetConfigurationParameter(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("database error"))

		reqBody, _ := json.Marshal(appConfigRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusInternalServerError)
		}
	})

	t.Run("Success", func(t *testing.T) {
		appConfigRequest = appcfg.AppConfigUpdateRequest{
			Key: string(appcfg.ReconciliationKey),
			Value: map[string]any{
				"enabled": true,
			},
		}

		mockDB.EXPECT().
			SetConfigurationParameter(gomock.Any(), gomock.Any()).
			Return(nil)

		reqBody, _ := json.Marshal(appConfigRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})

	t.Run("Scheduled Analysis updates next scheduled analysis start time", func(t *testing.T) {
		var (
			futureTime = time.Now().Add(48 * time.Hour).Format("20060102T150405Z")
			validRRule = fmt.Sprintf("FREQ=DAILY;INTERVAL=1;DTSTART=%s", futureTime)

			scheduledAnalysisRequest = appcfg.AppConfigUpdateRequest{
				Key: string(appcfg.ScheduledAnalysis),
				Value: map[string]any{
					"enabled": true,
					"rrule":   validRRule,
				},
			}

			expectedParameter = appcfg.Parameter{
				Key:   appcfg.ScheduledAnalysis,
				Value: must.NewJSONBObject(map[string]any{"enabled": true, "rrule": validRRule}),
			}
		)

		rule, err := rrule.StrToRRule(validRRule)
		if err != nil {
			t.Fatalf("failed to parse rrule: %v", err)
		}
		expectedNextAnalysis := null.TimeFrom(rule.After(time.Now(), true))

		mockDB.EXPECT().
			SetConfigurationParameter(gomock.Any(), expectedParameter).
			Return(nil)

		mockDB.EXPECT().
			SetNextScheduledAnalysisStartTime(gomock.Any(), expectedNextAnalysis).
			Return(nil)

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})

	t.Run("Disabling Scheduled Analysis removes the next scheduled analysis start time", func(t *testing.T) {
		var (
			scheduledAnalysisRequest = appcfg.AppConfigUpdateRequest{
				Key: string(appcfg.ScheduledAnalysis),
				Value: map[string]any{
					"enabled": false,
					"rrule":   "",
				},
			}

			expectedParameter = appcfg.Parameter{
				Key:   appcfg.ScheduledAnalysis,
				Value: must.NewJSONBObject(map[string]any{"enabled": false, "rrule": ""}),
			}
		)

		mockDB.EXPECT().
			SetConfigurationParameter(gomock.Any(), expectedParameter).
			Return(nil)

		mockDB.EXPECT().
			SetNextScheduledAnalysisStartTime(gomock.Any(), null.Time{}).
			Return(nil)

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/config", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		resources.SetApplicationConfiguration(rec, req)

		if status := rec.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}
