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

func TestStore_CreateGraphSchemaExtension_Integration(t *testing.T) {
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

			createdExtension, err := store.CreateGraphSchemaExtension(ctx, test.extension)
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

			persistedExtension, err := store.GetExtension(ctx, createdExtension.ID)
			require.NoError(t, err)
			assert.Equal(t, createdExtension, persistedExtension)
		})
	}
}

func TestStore_GetExtensions_Integration(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
	)

	_, err := pool.Exec(ctx, `DELETE FROM schema_extensions WHERE is_builtin = true`)
	require.NoError(t, err)

	type extensionData struct {
		name        string
		displayName string
		version     string
		isBuiltin   bool
		namespace   string
	}

	var extensions []services.Extension
	for _, extension := range []extensionData{
		{name: "ExtensionAlpha", displayName: "Alpha Extension", version: "1.0.0", isBuiltin: true, namespace: "ALP"},
		{name: "ExtensionBeta", displayName: "Beta Extension", version: "1.1.0", isBuiltin: false, namespace: "BET"},
		{name: "ExtensionGamma", displayName: "Gamma Extension", version: "2.0.0", isBuiltin: true, namespace: "GAM"},
	} {
		var extensionID int32
		require.NoError(t, pool.QueryRow(ctx, `
			INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			extension.name, extension.displayName, extension.version, extension.isBuiltin, extension.namespace,
		).Scan(&extensionID))

		extensions = append(extensions, services.Extension{
			ID:          extensionID,
			Name:        extension.name,
			DisplayName: extension.displayName,
			Version:     extension.version,
			IsBuiltin:   extension.isBuiltin,
			Namespace:   extension.namespace,
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
				actual, count, err := store.GetExtensions(ctx, input.filters, input.sort, input.skip, input.limit)
				want := test.want[index]

				if want.err != "" {
					assert.EqualError(t, err, want.err)
					assert.Nil(t, actual)
					assert.Zero(t, count)
					continue
				}

				require.NoError(t, err)
				assert.Equal(t, want.extensions, actual)
				assert.Equal(t, want.count, count)
			}
		})
	}
}
