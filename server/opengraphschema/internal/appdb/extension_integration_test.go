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
	"time"

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
			name: "success_-_returns_extension_fields",
			setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
				var data = seedNodeKind(t, ctx, pool)
				_, err := pool.Exec(ctx, "UPDATE schema_extensions SET created_at = $1, updated_at = $2 WHERE id = $3",
					time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC), data.extensionID)
				require.NoError(t, err)
				return data
			},
			assert: func(t *testing.T, extension services.Extension, data testSetupData) {
				assert.Equal(t, data.extensionID, extension.ID)
				assert.Equal(t, "TestExtension", extension.Name)
				assert.Equal(t, "Test Extension", extension.DisplayName)
				assert.Equal(t, "TST", extension.Namespace)
				assert.False(t, extension.IsBuiltin)
				assert.Equal(t, "1.0.0", extension.Version)
				assert.WithinDuration(t, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), extension.CreatedAt, 0)
				assert.WithinDuration(t, time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC), extension.UpdatedAt, 0)
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
				_, err := pool.Exec(ctx, `INSERT INTO schema_extensions (id, name, display_name, version, is_builtin, namespace, created_at, updated_at) VALUES
     (10003, 'zeta', 'Zeta', 'v3', false, 'Z', '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z'),
     (10002, 'alpha2', 'Alpha 2', 'v2', false, 'A2', '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z'),
     (10001, 'alpha1', 'Alpha 1', 'v1', true, 'A1', '2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z')`)
				require.NoError(t, err)
			}, want: []services.Extension{
				{ID: 10001, Name: "alpha1", DisplayName: "Alpha 1", Version: "v1", IsBuiltin: true, Namespace: "A1", CreatedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)},
				{ID: 10002, Name: "alpha2", DisplayName: "Alpha 2", Version: "v2", Namespace: "A2", CreatedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)},
				{ID: 10003, Name: "zeta", DisplayName: "Zeta", Version: "v3", Namespace: "Z", CreatedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)},
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
			require.Len(t, extensions, len(testCase.want))
			assert.NotNil(t, extensions)
			for index, expected := range testCase.want {
				var actual = extensions[index]
				assert.Equal(t, expected.ID, actual.ID)
				assert.Equal(t, expected.Name, actual.Name)
				assert.Equal(t, expected.DisplayName, actual.DisplayName)
				assert.Equal(t, expected.Version, actual.Version)
				assert.Equal(t, expected.IsBuiltin, actual.IsBuiltin)
				assert.Equal(t, expected.Namespace, actual.Namespace)
				assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, 0)
				assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, 0)
			}
		})
	}
}
