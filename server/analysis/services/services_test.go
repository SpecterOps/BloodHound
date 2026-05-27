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
	var (
		ctx           = context.Background()
		unexpectedErr = errors.New("connection refused")
		expected      = services.RequestedAnalysis{
			RequestedBy:           "test-user",
			RequestType:           services.RequestedAnalysisTypeAnalysis,
			RequestedAt:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DeleteAllGraph:        true,
			DeleteSourcelessGraph: false,
			DeleteSourceKinds:     []string{"AZBase"},
			DeleteRelationships:   []string{"HasSession"},
		}
	)

	tests := []struct {
		name       string
		dbResult   services.RequestedAnalysis
		dbErr      error
		wantResult services.RequestedAnalysis
		wantErr    error
	}{
		{
			name:       "returns the analysis request on success",
			dbResult:   expected,
			wantResult: expected,
		},
		{
			name:    "maps ErrNotFound to ErrNoPendingRequest",
			dbErr:   services.ErrNotFound,
			wantErr: services.ErrNoPendingRequest,
		},
		{
			name:    "propagates unexpected database errors",
			dbErr:   unexpectedErr,
			wantErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			databaseMock.EXPECT().GetAnalysisRequest(ctx).Return(tt.dbResult, tt.dbErr)

			result, err := svc.GetRequest(ctx)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestService_CreateRequest(t *testing.T) {
	var (
		ctx         = context.Background()
		requester   = "test-user"
		requestedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		expectedErr = errors.New("db unavailable")
		created     = services.RequestedAnalysis{
			RequestedBy: requester,
			RequestType: services.RequestedAnalysisTypeAnalysis,
			RequestedAt: requestedAt,
		}
		existing = services.RequestedAnalysis{
			RequestedBy: "other-user",
			RequestType: services.RequestedAnalysisTypeAnalysis,
			RequestedAt: requestedAt,
		}
	)

	tests := []struct {
		name        string
		dbResult    services.RequestedAnalysis
		dbCreated   bool
		dbErr       error
		wantResult  services.RequestedAnalysis
		wantCreated bool
		wantErr     error
	}{
		{
			name:        "returns created=true and the new request on success",
			dbResult:    created,
			dbCreated:   true,
			wantResult:  created,
			wantCreated: true,
		},
		{
			name:       "returns created=false and the existing request when one is already pending",
			dbResult:   existing,
			wantResult: existing,
		},
		{
			name:    "propagates database errors",
			dbErr:   expectedErr,
			wantErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			databaseMock.EXPECT().CreateAnalysisRequest(ctx, requester).Return(tt.dbResult, tt.dbCreated, tt.dbErr)

			current, gotCreated, err := svc.CreateRequest(ctx, requester)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCreated, gotCreated)
				assert.Equal(t, tt.wantResult, current)
			}
		})
	}
}
