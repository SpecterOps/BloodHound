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

func TestResources_GetDatapipeStatus(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	t.Run("success get datapipe status", func(t *testing.T) {
		lastCompleteAnalysisAt := time.Now()
		lastAnalysisRunAt := time.Now().Add(-time.Minute)

		mockDB.EXPECT().GetDatapipeStatus(gomock.Any()).Return(model.DatapipeStatusWrapper{
			Status:                 "idle",
			LastCompleteAnalysisAt: lastCompleteAnalysisAt,
			LastAnalysisRunAt:      lastAnalysisRunAt,
		}, nil)

		test.Request(t).
			WithMethod(http.MethodGet).
			WithURL("api/v2/datapipe/status").
			OnHandlerFunc(resources.GetDatapipeStatus).
			Require().
			ResponseJSONBody(model.DatapipeStatusWrapper{
				Status:                 "idle",
				LastCompleteAnalysisAt: lastCompleteAnalysisAt,
				LastAnalysisRunAt:      lastAnalysisRunAt,
			}).
			ResponseStatusCode(200)
	})

	t.Run("error getting datapipe status", func(t *testing.T) {
		mockDB.EXPECT().GetDatapipeStatus(gomock.Any()).Return(model.DatapipeStatusWrapper{}, fmt.Errorf("an error"))

		test.Request(t).
			WithMethod(http.MethodGet).
			WithURL("api/v2/datapipe/status").
			OnHandlerFunc(resources.GetDatapipeStatus).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}
