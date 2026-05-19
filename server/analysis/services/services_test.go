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

package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/specterops/bloodhound/server/analysis/services"
	"github.com/specterops/bloodhound/server/analysis/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetRequest(t *testing.T) {
	var ctx = context.Background()

	t.Run("returns the analysis request on success", func(t *testing.T) {
		var (
			expected = services.RequestedAnalysis{
				RequestedBy:           "test-user",
				RequestType:           services.RequestedAnalysisTypeAnalysis,
				RequestedAt:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				DeleteAllGraph:        true,
				DeleteSourcelessGraph: false,
				DeleteSourceKinds:     []string{"AZBase"},
				DeleteRelationships:   []string{"HasSession"},
			}
			databaseMock = mocks.NewMockDatabase(t)
			svc          = services.NewService(databaseMock)
		)

		databaseMock.EXPECT().GetAnalysisRequest(ctx).Return(expected, nil)

		result, err := svc.GetRequest(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("maps ErrNotFound to ErrNoPendingRequest", func(t *testing.T) {
		var (
			databaseMock = mocks.NewMockDatabase(t)
			svc          = services.NewService(databaseMock)
		)

		databaseMock.EXPECT().GetAnalysisRequest(ctx).Return(services.RequestedAnalysis{}, services.ErrNotFound)

		_, err := svc.GetRequest(ctx)
		assert.ErrorIs(t, err, services.ErrNoPendingRequest)
	})

	t.Run("propagates unexpected database errors", func(t *testing.T) {
		var (
			unexpectedErr = errors.New("connection refused")
			databaseMock  = mocks.NewMockDatabase(t)
			svc           = services.NewService(databaseMock)
		)

		databaseMock.EXPECT().GetAnalysisRequest(ctx).Return(services.RequestedAnalysis{}, unexpectedErr)

		_, err := svc.GetRequest(ctx)
		assert.ErrorIs(t, err, unexpectedErr)
	})
}

func TestService_CreateRequest(t *testing.T) {
	var (
		ctx       = context.Background()
		requester = "test-user"
	)

	t.Run("returns created=true and the new request on success", func(t *testing.T) {
		var (
			expected = services.RequestedAnalysis{
				RequestedBy: requester,
				RequestType: services.RequestedAnalysisTypeAnalysis,
				RequestedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			databaseMock = mocks.NewMockDatabase(t)
			svc          = services.NewService(databaseMock)
		)

		databaseMock.EXPECT().CreateAnalysisRequest(ctx, requester).Return(expected, true, nil)

		current, created, err := svc.CreateRequest(ctx, requester)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, expected, current)
	})

	t.Run("returns created=false and the existing request when one is already pending", func(t *testing.T) {
		var (
			existing = services.RequestedAnalysis{
				RequestedBy: "other-user",
				RequestType: services.RequestedAnalysisTypeAnalysis,
				RequestedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			databaseMock = mocks.NewMockDatabase(t)
			svc          = services.NewService(databaseMock)
		)

		databaseMock.EXPECT().CreateAnalysisRequest(ctx, requester).Return(existing, false, nil)

		current, created, err := svc.CreateRequest(ctx, requester)
		require.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, existing, current)
	})

	t.Run("propagates database errors", func(t *testing.T) {
		var (
			expectedErr  = errors.New("db unavailable")
			databaseMock = mocks.NewMockDatabase(t)
			svc          = services.NewService(databaseMock)
		)

		databaseMock.EXPECT().CreateAnalysisRequest(ctx, requester).Return(services.RequestedAnalysis{}, false, expectedErr)

		_, _, err := svc.CreateRequest(ctx, requester)
		assert.ErrorIs(t, err, expectedErr)
	})
}
