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

package appdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/assetgrouptags/internal/appdb"
	"github.com/specterops/bloodhound/server/assetgrouptags/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()

	pool, err := pgxmock.NewPool()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, pool.ExpectationsWereMet())
		pool.Close()
	})

	return appdb.NewStore(pool), pool
}

func TestStore_GetAssetGroupTagByID(t *testing.T) {
	t.Parallel()

	unexpectedErr := errors.New("db unavailable")

	tests := []struct {
		name      string
		setupMock func(pool pgxmock.PgxPoolIface)
		wantTag   services.AssetGroupTag
		wantErr   error
	}{
		{
			name: "Success: returns the matching tag",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`SELECT id FROM asset_group_tags WHERE id = \$1 AND deleted_at IS NULL`).
					WithArgs(42).
					WillReturnRows(pool.NewRows([]string{"id"}).AddRow(42))
			},
			wantTag: services.AssetGroupTag{ID: 42},
		},
		{
			name: "Error: returns ErrAssetGroupTagNotFound when no rows",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`SELECT id FROM asset_group_tags WHERE id = \$1 AND deleted_at IS NULL`).
					WithArgs(42).
					WillReturnRows(pool.NewRows([]string{"id"}))
			},
			wantErr: services.ErrAssetGroupTagNotFound,
		},
		{
			name: "Error: propagates query error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`SELECT id FROM asset_group_tags WHERE id = \$1 AND deleted_at IS NULL`).
					WithArgs(42).
					WillReturnError(unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newTestStore(t)
			test.setupMock(pool)

			tag, err := store.GetAssetGroupTagByID(context.Background(), 42)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.wantTag, tag)
		})
	}
}

func TestStore_GetTierZeroTag(t *testing.T) {
	t.Parallel()

	unexpectedErr := errors.New("db unavailable")

	tests := []struct {
		name      string
		setupMock func(pool pgxmock.PgxPoolIface)
		wantTag   services.AssetGroupTag
		wantErr   error
	}{
		{
			name: "Success: returns first tier zero tag",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`SELECT id FROM asset_group_tags WHERE deleted_at IS NULL AND type = \$1 AND position = \$2 ORDER BY name ASC`).
					WithArgs(1, 1).
					WillReturnRows(pool.NewRows([]string{"id"}).AddRow(3).AddRow(9))
			},
			wantTag: services.AssetGroupTag{ID: 3},
		},
		{
			name: "Error: returns ErrTierZeroTagNotFound when empty",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`SELECT id FROM asset_group_tags WHERE deleted_at IS NULL AND type = \$1 AND position = \$2 ORDER BY name ASC`).
					WithArgs(1, 1).
					WillReturnRows(pool.NewRows([]string{"id"}))
			},
			wantErr: services.ErrTierZeroTagNotFound,
		},
		{
			name: "Error: propagates query error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`SELECT id FROM asset_group_tags WHERE deleted_at IS NULL AND type = \$1 AND position = \$2 ORDER BY name ASC`).
					WithArgs(1, 1).
					WillReturnError(unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newTestStore(t)
			test.setupMock(pool)

			tag, err := store.GetTierZeroTag(context.Background())
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.wantTag, tag)
		})
	}
}
