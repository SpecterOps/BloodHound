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
