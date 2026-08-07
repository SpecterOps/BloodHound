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

	"github.com/specterops/bloodhound/server/assetgrouptags/internal/services"
	"github.com/specterops/bloodhound/server/assetgrouptags/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ResolveTagIDsWithFallback(t *testing.T) {
	var (
		ctx           = context.Background()
		unexpectedErr = errors.New("connection refused")
	)

	tests := []struct {
		name       string
		tagID      string
		setupMock  func(databaseMock *mocks.MockDatabase)
		wantResult []int
		wantErr    error
	}{
		{
			name:  "success_-_falls_back_to_hygiene_and_tier_zero_when_empty",
			tagID: "",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetTierZeroTag(ctx).Return(services.AssetGroupTag{ID: 7}, nil)
			},
			wantResult: []int{services.TierHygienePlaceholderID, 7},
		},
		{
			name:       "success_-_passes_through_hygiene_placeholder_id",
			tagID:      "0",
			wantResult: []int{services.TierHygienePlaceholderID},
		},
		{
			name:  "success_-_returns_validated_tag_id",
			tagID: "42",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetAssetGroupTagByID(ctx, 42).Return(services.AssetGroupTag{ID: 42}, nil)
			},
			wantResult: []int{42},
		},
		{
			name:    "error_-_returns_error_when_tag_id_is_not_an_integer",
			tagID:   "not-a-number",
			wantErr: errStrconv,
		},
		{
			name:  "error_-_propagates_error_when_tag_id_lookup_fails",
			tagID: "42",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetAssetGroupTagByID(ctx, 42).Return(services.AssetGroupTag{}, services.ErrAssetGroupTagNotFound)
			},
			wantErr: services.ErrAssetGroupTagNotFound,
		},
		{
			name:  "error_-_propagates_error_when_tier_zero_lookup_fails",
			tagID: "",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetTierZeroTag(ctx).Return(services.AssetGroupTag{}, unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			if tt.setupMock != nil {
				tt.setupMock(databaseMock)
			}

			result, err := svc.ResolveTagIDsWithFallback(ctx, tt.tagID)
			if tt.wantErr != nil {
				if errors.Is(tt.wantErr, errStrconv) {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestService_GetTierZeroTag(t *testing.T) {
	var (
		ctx           = context.Background()
		unexpectedErr = errors.New("connection refused")
	)

	t.Run("success_-_returns_tier_zero_tag", func(t *testing.T) {
		var (
			databaseMock = mocks.NewMockDatabase(t)
			svc          = services.NewService(databaseMock)
		)

		databaseMock.EXPECT().GetTierZeroTag(ctx).Return(services.AssetGroupTag{ID: 3}, nil)

		result, err := svc.GetTierZeroTag(ctx)
		require.NoError(t, err)
		assert.Equal(t, services.AssetGroupTag{ID: 3}, result)
	})

	t.Run("error_-_propagates_database_error", func(t *testing.T) {
		var (
			databaseMock = mocks.NewMockDatabase(t)
			svc          = services.NewService(databaseMock)
		)

		databaseMock.EXPECT().GetTierZeroTag(ctx).Return(services.AssetGroupTag{}, unexpectedErr)

		result, err := svc.GetTierZeroTag(ctx)
		assert.ErrorIs(t, err, unexpectedErr)
		assert.Equal(t, services.AssetGroupTag{}, result)
	})
}

// errStrconv is a sentinel used to select the "any error" assertion branch for the
// non-integer tag id case, where the underlying error comes from strconv.Atoi.
var errStrconv = errors.New("strconv error sentinel")
