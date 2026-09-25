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

//go:build integration

package appdb_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_GetExtension_Integration(t *testing.T) {
	type testCase struct {
		name    string
		setup   func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData
		wantErr error
		assert  func(t *testing.T, extension services.Extension, data testSetupData)
	}

	tests := []testCase{
		{
			name:  "success_-_returns_extension_fields",
			setup: seedNodeKind,
			assert: func(t *testing.T, extension services.Extension, data testSetupData) {
				assert.Equal(t, services.Extension{
					ID:          data.extensionID,
					Name:        "TestExtension",
					DisplayName: "Test Extension",
					Namespace:   "TST",
					IsBuiltin:   false,
					Version:     "1.0.0",
				}, extension)
			},
		},
		{
			name: "error_-_returns_ErrExtensionNotFound",
			setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
				return testSetupData{extensionID: int32(999999999)}
			},
			wantErr: services.ErrExtensionNotFound,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				ctx         = context.Background()
				data        = testCase.setup(t, ctx, pool)
			)

			if extension, err := store.GetExtension(ctx, data.extensionID); testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			} else {
				require.NoError(t, err)
				testCase.assert(t, extension, data)
			}
		})
	}
}
