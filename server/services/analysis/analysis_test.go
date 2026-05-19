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
	"context"
	"errors"
	"testing"
	"time"

	appdbAnalysis "github.com/specterops/bloodhound/server/appdb/analysis"
	"github.com/specterops/bloodhound/server/models"
	"github.com/specterops/bloodhound/server/services/analysis"
	analysisMocks "github.com/specterops/bloodhound/server/services/analysis/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetRequest(t *testing.T) {
	var ctx = context.Background()

	t.Run("returns the analysis request on success", func(t *testing.T) {
		var (
			expected = models.RequestedAnalysis{
				RequestedBy:           "test-user",
				RequestType:           models.RequestedAnalysisTypeAnalysis,
				RequestedAt:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				DeleteAllGraph:        true,
				DeleteSourcelessGraph: false,
				DeleteSourceKinds:     []string{"AZBase"},
				DeleteRelationships:   []string{"HasSession"},
			}
			databaseMock = analysisMocks.NewMockDatabase(t)
			svc          = analysis.NewService(databaseMock)
		)

		databaseMock.EXPECT().GetAnalysisRequest(ctx).Return(expected, nil)

		result, err := svc.GetRequest(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("maps ErrNotFound to ErrNoPendingRequest", func(t *testing.T) {
		var (
			databaseMock = analysisMocks.NewMockDatabase(t)
			svc          = analysis.NewService(databaseMock)
		)

		databaseMock.EXPECT().GetAnalysisRequest(ctx).Return(models.RequestedAnalysis{}, appdbAnalysis.ErrNotFound)

		_, err := svc.GetRequest(ctx)
		assert.ErrorIs(t, err, analysis.ErrNoPendingRequest)
	})

	t.Run("propagates unexpected database errors", func(t *testing.T) {
		var (
			unexpectedErr = errors.New("connection refused")
			databaseMock  = analysisMocks.NewMockDatabase(t)
			svc           = analysis.NewService(databaseMock)
		)

		databaseMock.EXPECT().GetAnalysisRequest(ctx).Return(models.RequestedAnalysis{}, unexpectedErr)

		_, err := svc.GetRequest(ctx)
		assert.ErrorIs(t, err, unexpectedErr)
	})
}
