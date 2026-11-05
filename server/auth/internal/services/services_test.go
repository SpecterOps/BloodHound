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

	"github.com/specterops/bloodhound/server/auth/internal/services"
	"github.com/specterops/bloodhound/server/auth/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService_PanicsOnNilDatabase(t *testing.T) {
	assert.PanicsWithValue(t, "auth: service requires a non-nil Database", func() {
		services.NewService(nil)
	})
}

func TestService_Logout(t *testing.T) {
	var (
		ctx           = context.Background()
		sessionID     = int64(42)
		unexpectedErr = errors.New("db unavailable")
	)

	tests := []struct {
		name    string
		dbErr   error
		wantErr error
	}{
		{
			name: "returns nil on success",
		},
		{
			name:    "propagates database errors",
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

			databaseMock.EXPECT().Logout(ctx, sessionID).Return(tt.dbErr)

			err := svc.Logout(ctx, sessionID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
