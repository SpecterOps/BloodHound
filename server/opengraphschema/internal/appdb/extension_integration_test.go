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
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
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

func TestStore_ListExtensions_Integration(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
		tests       = []struct {
			name  string
			setup func(t *testing.T)
			want  []services.Extension
		}{
			{name: "empty", want: []services.Extension{}},
			{name: "orders by name", setup: func(t *testing.T) {
				_, err := pool.Exec(ctx, `INSERT INTO schema_extensions (id, name, display_name, version, is_builtin, namespace) VALUES
     (10003, 'zeta', 'Zeta', 'v3', false, 'Z'),
     (10002, 'alpha2', 'Alpha 2', 'v2', false, 'A2'),
     (10001, 'alpha1', 'Alpha 1', 'v1', true, 'A1')`)
				require.NoError(t, err)
			}, want: []services.Extension{
				{ID: 10001, Name: "alpha1", DisplayName: "Alpha 1", Version: "v1", IsBuiltin: true, Namespace: "A1"},
				{ID: 10002, Name: "alpha2", DisplayName: "Alpha 2", Version: "v2", Namespace: "A2"},
				{ID: 10003, Name: "zeta", DisplayName: "Zeta", Version: "v3", Namespace: "Z"},
			}},
		}
	)
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := pool.Exec(ctx, "DELETE FROM schema_extensions")
			require.NoError(t, err)
			if testCase.setup != nil {
				testCase.setup(t)
			}
			extensions, err := store.ListExtensions(ctx)
			require.NoError(t, err)
			assert.Equal(t, testCase.want, extensions)
		})
	}
}
