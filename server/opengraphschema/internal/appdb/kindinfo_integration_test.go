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
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/appdb"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_GetKindInfos_Integration(t *testing.T) {
	type testCase struct {
		name   string
		setup  func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData
		assert func(t *testing.T, infos []services.KindInfo, data testSetupData)
	}

	tests := []testCase{
		{
			name: "success_-_returns_ordered_infos_with_name",
			setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
				data := seedNodeKind(t, ctx, pool)
				seedKindInfos(t, ctx, pool, data.kindID, data.nodeKindID)
				return data
			},
			assert: func(t *testing.T, infos []services.KindInfo, data testSetupData) {
				require.Len(t, infos, 2)

				assert.Equal(t, "overview", infos[0].InfoKey)
				assert.Equal(t, int32(0), infos[0].Position)
				assert.Equal(t, "details", infos[1].InfoKey)
				assert.Equal(t, int32(1), infos[1].Position)

				for _, info := range infos {
					assert.Equal(t, "TestNodeKind", info.Name)
					require.NotNil(t, info.NodeKindID)
					assert.Equal(t, data.nodeKindID, *info.NodeKindID)
				}
				assert.JSONEq(t, `{"markdown":{"content":"overview md"}}`, string(infos[0].Content))
			},
		},
		{
			name:  "success_-_returns_empty_when_no_infos",
			setup: seedNodeKind,
			assert: func(t *testing.T, infos []services.KindInfo, data testSetupData) {
				require.NotNil(t, infos)
				assert.Len(t, infos, 0)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				ctx         = context.Background()
				data        = testCase.setup(t, ctx, pool)
			)

			infos, err := store.GetKindInfos(ctx, "TestNodeKind")
			require.NoError(t, err)
			testCase.assert(t, infos, data)
		})
	}
}

// seedKindInfos inserts two schema_kind_info rows for the supplied kind/node kind,
// deliberately ordering the INSERT by position 1 then 0 to prove GetKindInfos
// re-orders by position. Cleanup is handled by seedNodeKind's cascade delete.
func seedKindInfos(t *testing.T, ctx context.Context, pool *pgxpool.Pool, kindID, nodeKindID int32) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO schema_kind_info (kind_id, node_kind_id, info_key, title, position, content)
		VALUES ($1, $2, 'details', 'Details', 1, '{"markdown":{"content":"details md"}}'),
		       ($1, $2, 'overview', 'Overview', 0, '{"markdown":{"content":"overview md"}}')`,
		kindID, nodeKindID)
	require.NoError(t, err)
}

func TestStore_CreateKindInfo_Integration(t *testing.T) {

	type testSetupData struct {
		kindID       int32
		nodeKindID   int32
		input        services.KindInfoInput
		wantKindInfo services.KindInfo
		wantError    error
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData
		assert   func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, created services.KindInfo, err error)
		teardown func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData)
	}

	var (
		ctx   = context.Background()
		input = services.KindInfoInput{
			InfoKey:  "overview",
			Title:    "Overview",
			Position: 0,
			Content:  json.RawMessage(`{"markdown":{"content":"Overview content"}}`),
		}
		teardown = func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData) {
			t.Helper()
			_, err := pool.Exec(ctx, `DELETE FROM schema_kind_info WHERE kind_id = $1`, data.kindID)
			require.NoError(t, err)
		}
		tests = []testCase{
			{
				name: "success_-_creates_kind_info",
				setup: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seed := seedNodeKind(t, ctx, pool)
					return testSetupData{
						kindID:     seed.kindID,
						nodeKindID: seed.nodeKindID,
						input:      input,
						wantKindInfo: services.KindInfo{
							KindID:     seed.kindID,
							NodeKindID: &seed.nodeKindID,
							Name:       "TestNodeKind",
							InfoKey:    input.InfoKey,
							Title:      input.Title,
							Position:   input.Position,
							Content:    input.Content,
						},
					}
				},
				assert: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, created services.KindInfo, err error) {
					t.Helper()

					require.NoError(t, err)
					assert.NotZero(t, created.ID)
					assert.Equal(t, data.wantKindInfo.KindID, created.KindID)
					assert.Equal(t, data.wantKindInfo.NodeKindID, created.NodeKindID)
					assert.Equal(t, data.wantKindInfo.RelationshipKindID, created.RelationshipKindID)
					assert.Equal(t, data.wantKindInfo.Name, created.Name)
					assert.Equal(t, data.wantKindInfo.InfoKey, created.InfoKey)
					assert.Equal(t, data.wantKindInfo.Title, created.Title)
					assert.Equal(t, data.wantKindInfo.Position, created.Position)
					assert.JSONEq(t, string(data.wantKindInfo.Content), string(created.Content))
					assert.False(t, created.CreatedAt.IsZero())
					assert.False(t, created.UpdatedAt.IsZero())
				},
				teardown: teardown,
			},
			{
				name: "error_-_duplicate_info_key",
				setup: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData {
					t.Helper()

					var (
						seed = seedNodeKind(t, ctx, pool)
						data = testSetupData{
							kindID:     seed.kindID,
							nodeKindID: seed.nodeKindID,
							input:      input,
							wantError:  services.ErrKindInfoDuplicateInfoKey,
						}
					)

					_, err := store.CreateKindInfo(ctx, data.kindID, &data.nodeKindID, nil, input)
					require.NoError(t, err)
					data.input.Position = 1
					return data
				},
				assert: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, created services.KindInfo, err error) {
					t.Helper()

					assert.Equal(t, data.wantKindInfo, created)
					assert.ErrorIs(t, err, data.wantError)

					var rowCount int
					require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM schema_kind_info WHERE kind_id = $1`, data.kindID).Scan(&rowCount))
					assert.Equal(t, 1, rowCount)
				},
				teardown: teardown,
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var (
				store, pool = setupStore(t)
				data        = test.setup(t, ctx, store, pool)
			)
			t.Cleanup(func() { test.teardown(t, ctx, pool, data) })

			created, err := store.CreateKindInfo(ctx, data.kindID, &data.nodeKindID, nil, data.input)
			test.assert(t, ctx, pool, data, created, err)
		})
	}
}

func TestStore_UpdateKindInfo_Integration(t *testing.T) {

	type testSetupData struct {
		kindID       int32
		kindInfo     services.KindInfo
		wantKindInfo services.KindInfo
		wantError    error
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData
		assert   func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, updated services.KindInfo, err error)
		teardown func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData)
	}

	var (
		ctx   = context.Background()
		input = services.KindInfoInput{
			InfoKey:  "overview",
			Title:    "Overview",
			Position: 0,
			Content:  json.RawMessage(`{"markdown":{"content":"Original content"}}`),
		}
		teardown = func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData) {
			t.Helper()
			_, err := pool.Exec(ctx, `DELETE FROM schema_kind_info WHERE kind_id = $1`, data.kindID)
			require.NoError(t, err)
		}
		tests = []testCase{
			{
				name: "success_-_updates_kind_info_fields",
				setup: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seed := seedNodeKind(t, ctx, pool)
					created, err := store.CreateKindInfo(ctx, seed.kindID, &seed.nodeKindID, nil, input)
					require.NoError(t, err)

					created.InfoKey = "summary"
					created.Title = "Summary"
					created.Position = 1
					created.Content = json.RawMessage(`{"markdown":{"content":"Updated content"}}`)

					return testSetupData{kindID: seed.kindID, kindInfo: created, wantKindInfo: created}
				},
				assert: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, updated services.KindInfo, err error) {
					t.Helper()

					require.NoError(t, err)
					assert.Equal(t, data.wantKindInfo.ID, updated.ID)
					assert.Equal(t, data.wantKindInfo.KindID, updated.KindID)
					assert.Equal(t, data.wantKindInfo.NodeKindID, updated.NodeKindID)
					assert.Equal(t, data.wantKindInfo.RelationshipKindID, updated.RelationshipKindID)
					assert.Equal(t, data.wantKindInfo.Name, updated.Name)
					assert.Equal(t, data.wantKindInfo.InfoKey, updated.InfoKey)
					assert.Equal(t, data.wantKindInfo.Title, updated.Title)
					assert.Equal(t, data.wantKindInfo.Position, updated.Position)
					assert.JSONEq(t, string(data.wantKindInfo.Content), string(updated.Content))
					assert.Equal(t, data.wantKindInfo.CreatedAt, updated.CreatedAt)
					assert.False(t, updated.UpdatedAt.IsZero())
				},
				teardown: teardown,
			},
			{
				name: "error_-_kind_info_not_found",
				setup: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seed := seedNodeKind(t, ctx, pool)
					return testSetupData{
						kindID: seed.kindID,
						kindInfo: services.KindInfo{
							ID:         -1,
							KindID:     seed.kindID,
							NodeKindID: &seed.nodeKindID,
							InfoKey:    input.InfoKey,
							Title:      input.Title,
							Position:   input.Position,
							Content:    input.Content,
						},
						wantError: services.ErrKindInfoNotFound,
					}
				},
				assert: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, updated services.KindInfo, err error) {
					t.Helper()

					assert.Equal(t, data.wantKindInfo, updated)
					assert.ErrorIs(t, err, data.wantError)

					var rowCount int
					require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM schema_kind_info WHERE kind_id = $1`, data.kindID).Scan(&rowCount))
					assert.Zero(t, rowCount)
				},
				teardown: teardown,
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var (
				store, pool = setupStore(t)
				data        = test.setup(t, ctx, store, pool)
			)
			t.Cleanup(func() { test.teardown(t, ctx, pool, data) })

			updated, err := store.UpdateKindInfo(ctx, data.kindInfo)
			test.assert(t, ctx, pool, data, updated, err)
		})
	}
}

func TestStore_DeleteKindInfo_Integration(t *testing.T) {

	type testSetupData struct {
		kindID     int32
		kindInfoID int32
		wantCount  int
		wantError  error
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData
		assert   func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, err error)
		teardown func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData)
	}

	var (
		ctx   = context.Background()
		input = services.KindInfoInput{
			InfoKey:  "overview",
			Title:    "Overview",
			Position: 0,
			Content:  json.RawMessage(`{"markdown":{"content":"Overview content"}}`),
		}
		teardown = func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData) {
			t.Helper()
			_, err := pool.Exec(ctx, `DELETE FROM schema_kind_info WHERE kind_id = $1`, data.kindID)
			require.NoError(t, err)
		}
		tests = []testCase{
			{
				name: "success_-_deletes_kind_info",
				setup: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seed := seedNodeKind(t, ctx, pool)
					created, err := store.CreateKindInfo(ctx, seed.kindID, &seed.nodeKindID, nil, input)
					require.NoError(t, err)
					return testSetupData{kindID: seed.kindID, kindInfoID: created.ID}
				},
				assert: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, err error) {
					t.Helper()

					assert.ErrorIs(t, err, data.wantError)

					var rowCount int
					require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM schema_kind_info WHERE kind_id = $1`, data.kindID).Scan(&rowCount))
					assert.Equal(t, data.wantCount, rowCount)
				},
				teardown: teardown,
			},
			{
				name: "error_-_kind_info_not_found",
				setup: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seed := seedNodeKind(t, ctx, pool)
					_, err := store.CreateKindInfo(ctx, seed.kindID, &seed.nodeKindID, nil, input)
					require.NoError(t, err)
					return testSetupData{
						kindID:     seed.kindID,
						kindInfoID: -1,
						wantCount:  1,
						wantError:  services.ErrKindInfoNotFound,
					}
				},
				assert: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, err error) {
					t.Helper()

					assert.ErrorIs(t, err, data.wantError)

					var rowCount int
					require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM schema_kind_info WHERE kind_id = $1`, data.kindID).Scan(&rowCount))
					assert.Equal(t, data.wantCount, rowCount)
				},
				teardown: teardown,
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var (
				store, pool = setupStore(t)
				data        = test.setup(t, ctx, store, pool)
			)
			t.Cleanup(func() { test.teardown(t, ctx, pool, data) })

			err := store.DeleteKindInfo(ctx, data.kindInfoID)
			test.assert(t, ctx, pool, data, err)
		})
	}
}
