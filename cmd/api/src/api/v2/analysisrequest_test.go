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

package v2_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	v2 "github.com/specterops/bloodhound/src/api/v2"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/test"
	"go.uber.org/mock/gomock"
)

func TestResources_GetAnalysisRequest(t *testing.T) {
	const (
		url = "api/v2/analysis/status"
	)

	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	t.Run("success getting analysis", func(t *testing.T) {
		analysisRequest := model.AnalysisRequest{
			RequestedAt: time.Now(),
			RequestedBy: "test",
			RequestType: model.AnalysisRequestType("test-type"),
		}

		mockDB.EXPECT().GetAnalysisRequest(gomock.Any()).Return(analysisRequest, nil)

		test.Request(t).
			WithMethod(http.MethodGet).
			WithURL(url).
			OnHandlerFunc(resources.GetAnalysisRequest).
			Require().
			ResponseJSONBody(analysisRequest).
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("error getting analysis", func(t *testing.T) {
		mockDB.EXPECT().GetAnalysisRequest(gomock.Any()).Return(model.AnalysisRequest{}, fmt.Errorf("an error"))

		test.Request(t).
			WithMethod(http.MethodGet).
			WithURL(url).
			OnHandlerFunc(resources.GetAnalysisRequest).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}
