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
	"github.com/specterops/bloodhound/packages/go/params"
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

func TestStore_CreateExtension_Integration(t *testing.T) {
	var (
		ctx   = context.Background()
		tests = []struct {
			name      string
			extension services.Extension
			wantErr   error
		}{
			{
				name: "success_-_creates_non_builtin_extension",
				extension: services.Extension{
					Name:        "CreatedExtension",
					DisplayName: "Created Extension",
					Version:     "2.0.0",
					IsBuiltin:   true,
					Namespace:   "NEW",
				},
			},
			{
				name: "error_-_rejects_duplicate_name",
				extension: services.Extension{
					Name:        "TestExtension",
					DisplayName: "Duplicate Name",
					Version:     "2.0.0",
					Namespace:   "NEW",
				},
				wantErr: services.ErrDuplicateExtensionName,
			},
			{
				name: "error_-_rejects_duplicate_namespace",
				extension: services.Extension{
					Name:        "AnotherExtension",
					DisplayName: "Duplicate Namespace",
					Version:     "2.0.0",
					Namespace:   "TST",
				},
				wantErr: services.ErrDuplicateExtensionNamespace,
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store, pool := setupStore(t)
			seedExtension(t, ctx, pool)

			createdExtension, err := store.CreateExtension(ctx, test.extension)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				assert.Equal(t, services.Extension{}, createdExtension)
				return
			}

			require.NoError(t, err)
			assert.NotZero(t, createdExtension.ID)
			assert.Equal(t, test.extension.Name, createdExtension.Name)
			assert.Equal(t, test.extension.DisplayName, createdExtension.DisplayName)
			assert.Equal(t, test.extension.Version, createdExtension.Version)
			assert.Equal(t, test.extension.Namespace, createdExtension.Namespace)
			assert.False(t, createdExtension.IsBuiltin)
			assert.False(t, createdExtension.CreatedAt.IsZero())
			assert.False(t, createdExtension.UpdatedAt.IsZero())

			persistedExtension, err := store.GetExtension(ctx, createdExtension.ID)
			require.NoError(t, err)
			assert.Equal(t, createdExtension, persistedExtension)
		})
	}
}

func TestStore_UpdateExtension_Integration(t *testing.T) {
	var ctx = context.Background()

	t.Run("success_-_updates_and_persists_extension_fields_ignores_builtin", func(t *testing.T) {
		store, pool := setupStore(t)
		extensionID := seedExtension(t, ctx, pool)
		_, err := pool.Exec(ctx, `UPDATE schema_extensions SET is_builtin = true WHERE id = $1`, extensionID)
		require.NoError(t, err)

		input := services.Extension{
			ID:          extensionID,
			Name:        "UpdatedExtension",
			DisplayName: "Updated Extension",
			Version:     "2.0.0",
			Namespace:   "UPD",
			IsBuiltin:   false,
		}

		updatedExtension, err := store.UpdateExtension(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, services.Extension{
			ID:          extensionID,
			Name:        input.Name,
			DisplayName: input.DisplayName,
			Version:     input.Version,
			Namespace:   input.Namespace,
			IsBuiltin:   true,
			CreatedAt:   updatedExtension.CreatedAt,
			UpdatedAt:   updatedExtension.UpdatedAt,
		}, updatedExtension)
		assert.False(t, updatedExtension.CreatedAt.IsZero())
		assert.False(t, updatedExtension.UpdatedAt.IsZero())

		persistedExtension, err := store.GetExtension(ctx, extensionID)
		require.NoError(t, err)
		assert.Equal(t, updatedExtension, persistedExtension)
	})

	t.Run("error_-_returns_extension_not_found", func(t *testing.T) {
		store, _ := setupStore(t)
		updatedExtension, err := store.UpdateExtension(ctx, services.Extension{
			ID:          0,
			Name:        "MissingExtension",
			DisplayName: "Missing Extension",
			Version:     "1.0.0",
			Namespace:   "MIS",
		})
		require.ErrorIs(t, err, services.ErrExtensionNotFound)
		assert.Equal(t, services.Extension{}, updatedExtension)
	})

	t.Run("error_-_rejects_duplicate_name", func(t *testing.T) {
		store, pool := setupStore(t)
		extensionID := seedExtension(t, ctx, pool)
		var duplicateID int32
		require.NoError(t, pool.QueryRow(ctx, `
			INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
			VALUES ('OtherExtension', 'Other Extension', '1.0.0', false, 'OTH')
			RETURNING id`).Scan(&duplicateID))
		t.Cleanup(func() {
			_, err := pool.Exec(ctx, `DELETE FROM schema_extensions WHERE id = $1`, duplicateID)
			require.NoError(t, err)
		})

		updatedExtension, err := store.UpdateExtension(ctx, services.Extension{
			ID:          extensionID,
			Name:        "OtherExtension",
			DisplayName: "Updated Extension",
			Version:     "2.0.0",
			Namespace:   "UPD",
		})
		require.ErrorIs(t, err, services.ErrDuplicateExtensionName)
		assert.Equal(t, services.Extension{}, updatedExtension)
	})

	t.Run("error_-_rejects_duplicate_namespace", func(t *testing.T) {
		store, pool := setupStore(t)
		extensionID := seedExtension(t, ctx, pool)
		var duplicateID int32
		require.NoError(t, pool.QueryRow(ctx, `
			INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
			VALUES ('OtherExtension', 'Other Extension', '1.0.0', false, 'OTH')
			RETURNING id`).Scan(&duplicateID))
		t.Cleanup(func() {
			_, err := pool.Exec(ctx, `DELETE FROM schema_extensions WHERE id = $1`, duplicateID)
			require.NoError(t, err)
		})

		updatedExtension, err := store.UpdateExtension(ctx, services.Extension{
			ID:          extensionID,
			Name:        "UpdatedExtension",
			DisplayName: "Updated Extension",
			Version:     "2.0.0",
			Namespace:   "OTH",
		})
		require.ErrorIs(t, err, services.ErrDuplicateExtensionNamespace)
		assert.Equal(t, services.Extension{}, updatedExtension)
	})
}

func TestStore_DeleteExtension_Integration(t *testing.T) {
	var ctx = context.Background()

	t.Run("error_-_returns_extension_not_found", func(t *testing.T) {
		store, _ := setupStore(t)

		err := store.DeleteExtension(ctx, int32(999999999))
		require.ErrorIs(t, err, services.ErrExtensionNotFound)
	})

	t.Run("error_-_rejects_builtin_extension_without_deleting_it", func(t *testing.T) {
		store, pool := setupStore(t)
		extensionID := seedExtension(t, ctx, pool)
		_, err := pool.Exec(ctx, `UPDATE schema_extensions SET is_builtin = true WHERE id = $1`, extensionID)
		require.NoError(t, err)

		err = store.DeleteExtension(ctx, extensionID)
		require.ErrorIs(t, err, services.ErrGraphExtensionBuiltIn)

		var extensionCount int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_extensions WHERE id = $1`, extensionID).Scan(&extensionCount))
		assert.Equal(t, 1, extensionCount)
	})

	t.Run("success_-_deletes_extension_without_node_kinds", func(t *testing.T) {
		store, pool := setupStore(t)
		extensionID := seedExtension(t, ctx, pool)

		require.NoError(t, store.DeleteExtension(ctx, extensionID))

		var extensionCount int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_extensions WHERE id = $1`, extensionID).Scan(&extensionCount))
		assert.Zero(t, extensionCount)
	})

	t.Run("success_-_deletes_extension_and_stubs_non_display_node_kinds", func(t *testing.T) {
		store, pool := setupStore(t)
		extensionID := seedExtension(t, ctx, pool)

		// seed 2 node kinds (one's a display kind, one isn't)
		type seededNodeKind struct {
			kindID        int32
			name          string
			isDisplayKind bool
		}
		var nodeKinds []seededNodeKind
		for _, nodeKind := range []struct {
			name          string
			isDisplayKind bool
		}{
			{name: "DeleteDisplayKind", isDisplayKind: true},
			{name: "DeleteNonDisplayKind", isDisplayKind: false},
		} {
			var kindID int32
			require.NoError(t, pool.QueryRow(ctx, `INSERT INTO kind (name) VALUES ($1) RETURNING id`, nodeKind.name).Scan(&kindID))
			nodeKinds = append(nodeKinds, seededNodeKind{kindID: kindID, name: nodeKind.name, isDisplayKind: nodeKind.isDisplayKind})

			_, err := pool.Exec(ctx, `
				INSERT INTO schema_node_kinds (schema_extension_id, kind_id, display_name, description, is_display_kind, icon, icon_color)
				VALUES ($1, $2, $3, 'test node kind', $4, 'user', '#fff')`,
				extensionID, kindID, nodeKind.name, nodeKind.isDisplayKind)
			require.NoError(t, err)
		}
		t.Cleanup(func() {
			_, err := pool.Exec(ctx, `DELETE FROM schema_node_kinds WHERE schema_extension_id = $1`, extensionID)
			require.NoError(t, err)
			for _, nodeKind := range nodeKinds {
				_, err := pool.Exec(ctx, `DELETE FROM custom_node_kinds WHERE kind_id = $1`, nodeKind.kindID)
				require.NoError(t, err)
				_, err = pool.Exec(ctx, `DELETE FROM kind WHERE id = $1`, nodeKind.kindID)
				require.NoError(t, err)
			}
		})

		// execute delete
		require.NoError(t, store.DeleteExtension(ctx, extensionID))

		// extension and node kinds deleted
		var extensionCount int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_extensions WHERE id = $1`, extensionID).Scan(&extensionCount))
		assert.Zero(t, extensionCount)

		var schemaNodeKindCount int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_node_kinds WHERE schema_extension_id = $1`, extensionID).Scan(&schemaNodeKindCount))
		assert.Zero(t, schemaNodeKindCount)

		// stub node kind created
		for _, nodeKind := range nodeKinds {
			var customNodeKindCount int
			require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM custom_node_kinds WHERE kind_id = $1`, nodeKind.kindID).Scan(&customNodeKindCount))
			if !nodeKind.isDisplayKind {
				assert.Equal(t, 1, customNodeKindCount)

				var config []byte
				require.NoError(t, pool.QueryRow(ctx, `SELECT config FROM custom_node_kinds WHERE kind_id = $1`, nodeKind.kindID).Scan(&config))
				assert.JSONEq(t, `{"icon":{"type":"font-awesome","name":"question","color":"#FFFFFF"}}`, string(config))
			} else {
				assert.Zero(t, customNodeKindCount)
			}
		}
	})
}

func TestStore_ListExtensions_Integration(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
	)

	normalizeExtensionTimestamps := func(extensions []services.Extension) []services.Extension {
		normalized := make([]services.Extension, len(extensions))
		copy(normalized, extensions)
		for i := range normalized {
			normalized[i].CreatedAt = normalized[i].CreatedAt.UTC()
			normalized[i].UpdatedAt = normalized[i].UpdatedAt.UTC()
		}
		return normalized
	}

	_, err := pool.Exec(ctx, `DELETE FROM schema_extensions WHERE is_builtin = true`)
	require.NoError(t, err)

	type extensionData struct {
		name        string
		displayName string
		version     string
		isBuiltin   bool
		namespace   string
		createdAt   time.Time
		updatedAt   time.Time
	}

	var extensions []services.Extension
	for _, extension := range []extensionData{
		{name: "ExtensionAlpha", displayName: "Alpha Extension", version: "1.0.0", isBuiltin: true, namespace: "ALP", createdAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), updatedAt: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)},
		{name: "ExtensionBeta", displayName: "Beta Extension", version: "1.1.0", isBuiltin: false, namespace: "BET", createdAt: time.Date(2026, time.January, 3, 0, 0, 0, 0, time.UTC), updatedAt: time.Date(2026, time.January, 4, 0, 0, 0, 0, time.UTC)},
		{name: "ExtensionGamma", displayName: "Gamma Extension", version: "2.0.0", isBuiltin: true, namespace: "GAM", createdAt: time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC), updatedAt: time.Date(2026, time.January, 6, 0, 0, 0, 0, time.UTC)},
	} {
		var extensionID int32
		require.NoError(t, pool.QueryRow(ctx, `
			INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			extension.name, extension.displayName, extension.version, extension.isBuiltin, extension.namespace, extension.createdAt, extension.updatedAt,
		).Scan(&extensionID))

		extensions = append(extensions, services.Extension{
			ID:          extensionID,
			Name:        extension.name,
			DisplayName: extension.displayName,
			Version:     extension.version,
			IsBuiltin:   extension.isBuiltin,
			Namespace:   extension.namespace,
			CreatedAt:   extension.createdAt,
			UpdatedAt:   extension.updatedAt,
		})
	}

	type query struct {
		filters params.Filters
		sort    params.SortItems
		skip    int
		limit   int
	}

	type expected struct {
		extensions []services.Extension
		count      int
		err        string
	}

	type testCase struct {
		name    string
		queries []query
		want    []expected
	}

	tests := []testCase{
		{
			name:    "success_-_returns_all_extensions_in_default_order",
			queries: []query{{}},
			want:    []expected{{extensions: extensions, count: 3}},
		},
		{
			name: "success_-_filters_by_name",
			queries: []query{
				{filters: params.Filters{"name": {{Operator: params.Equals, Value: "ExtensionAlpha"}}}},
				{filters: params.Filters{"name": {{Operator: params.Equals, Value: "MissingExtension"}}}},
			},
			want: []expected{
				{extensions: []services.Extension{extensions[0]}, count: 1},
				{extensions: []services.Extension{}, count: 0},
			},
		},
		{
			name: "success_-_filters_by_display_name",
			queries: []query{
				{filters: params.Filters{"display_name": {{Operator: params.Equals, Value: "Beta Extension"}}}},
			},
			want: []expected{{extensions: []services.Extension{extensions[1]}, count: 1}},
		},
		{
			name: "success_-_filters_by_builtin_status",
			queries: []query{
				{filters: params.Filters{"is_builtin": {{Operator: params.Equals, Value: "true"}}}},
			},
			want: []expected{{extensions: []services.Extension{extensions[0], extensions[2]}, count: 2}},
		},
		{
			name: "success_-_combines_or_and_filters",
			queries: []query{
				{
					filters: params.Filters{
						"name": {
							{Operator: params.ApproximatelyEquals, Value: "Alpha"},
							{Operator: params.Equals, Value: "ExtensionGamma", SetOperator: params.FilterOr},
						},
						"is_builtin": {{Operator: params.Equals, Value: "true"}},
					},
				},
			},
			want: []expected{{extensions: []services.Extension{extensions[0], extensions[2]}, count: 2}},
		},
		{
			name: "success_-_sorts_by_display_name_in_both_directions",
			queries: []query{
				{sort: params.SortItems{{Field: "display_name", Direction: params.Ascending}}},
				{sort: params.SortItems{{Field: "display_name", Direction: params.Descending}}},
			},
			want: []expected{
				{extensions: []services.Extension{extensions[0], extensions[1], extensions[2]}, count: 3},
				{extensions: []services.Extension{extensions[2], extensions[1], extensions[0]}, count: 3},
			},
		},
		{
			name: "success_-_paginates_filtered_results_and_returns_full_count",
			queries: []query{
				{
					filters: params.Filters{"is_builtin": {{Operator: params.Equals, Value: "true"}}},
					sort:    params.SortItems{{Field: "name", Direction: params.Ascending}},
					skip:    1,
					limit:   1,
				},
			},
			want: []expected{{extensions: []services.Extension{extensions[2]}, count: 2}},
		},
		{
			name: "error_-_rejects_unsupported_fields_and_sort_direction",
			queries: []query{
				{filters: params.Filters{"version": {{Operator: params.Equals, Value: "1.0.0"}}}},
				{sort: params.SortItems{{Field: "version", Direction: params.Ascending}}},
				{sort: params.SortItems{{Field: "name", Direction: "sideways"}}},
			},
			want: []expected{
				{err: `filter references unknown field "version"`},
				{err: `sort references unknown field "version"`},
				{err: `invalid sort direction "sideways"`},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for index, input := range test.queries {
				actual, count, err := store.ListExtensions(ctx, input.filters, input.sort, input.skip, input.limit)
				want := test.want[index]

				if want.err != "" {
					assert.EqualError(t, err, want.err)
					assert.Nil(t, actual)
					assert.Zero(t, count)
					continue
				}

				require.NoError(t, err)
				assert.Equal(t, normalizeExtensionTimestamps(want.extensions), normalizeExtensionTimestamps(actual))
				assert.Equal(t, want.count, count)
			}
		})
	}
}
