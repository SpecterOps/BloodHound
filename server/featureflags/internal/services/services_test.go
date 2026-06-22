// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/specterops/bloodhound/server/featureflags/internal/services"
	"github.com/specterops/bloodhound/server/featureflags/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeFlagDatabase is a minimal in-memory implementation of the Database port
// used to drive the Service use cases without a real connection pool.
type fakeFlagDatabase struct {
	flag services.FeatureFlag
	err  error
}

func (f fakeFlagDatabase) GetFlagByKey(_ context.Context, _ string) (services.FeatureFlag, error) {
	return f.flag, f.err
}

func TestNewService(t *testing.T) {
	mockDb := mocks.NewMockDatabase(t)
	assert.NotNil(t, services.NewService(mockDb))
}

func TestService_GetFlagByKey(t *testing.T) {
	var (
		ctx     = context.Background()
		want    = services.FeatureFlag{ID: 7, Key: services.FeatureOpenHoundSupport, Enabled: true}
		notFErr = services.ErrNotFound
	)

	t.Run("returns the flag from the database", func(t *testing.T) {
		svc := services.NewService(fakeFlagDatabase{flag: want})

		got, err := svc.GetFlagByKey(ctx, services.FeatureOpenHoundSupport)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("propagates the database error", func(t *testing.T) {
		svc := services.NewService(fakeFlagDatabase{err: notFErr})

		_, err := svc.GetFlagByKey(ctx, services.FeatureOpenHoundSupport)

		assert.ErrorIs(t, err, notFErr)
	})
}

func TestService_IsEnabled(t *testing.T) {
	var (
		ctx   = context.Background()
		dbErr = errors.New("connection refused")
	)

	tests := []struct {
		name    string
		db      fakeFlagDatabase
		want    bool
		wantErr error
	}{
		{
			name: "true when the flag is enabled",
			db:   fakeFlagDatabase{flag: services.FeatureFlag{Key: services.FeatureOpenHoundSupport, Enabled: true}},
			want: true,
		},
		{
			name: "false when the flag is disabled",
			db:   fakeFlagDatabase{flag: services.FeatureFlag{Key: services.FeatureOpenHoundSupport, Enabled: false}},
			want: false,
		},
		{
			name:    "propagates database errors",
			db:      fakeFlagDatabase{err: dbErr},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := services.NewService(tt.db)

			got, err := svc.IsEnabled(ctx, services.FeatureOpenHoundSupport)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.False(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
