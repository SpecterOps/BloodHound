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
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/extensions/internal/appdb"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal SQL expected by the Store, compared via pgxmock.QueryMatcherEqual (whitespace
// normalised). Column order, aliases, JOIN and WHERE shape are load-bearing.
const expectedGetNodeKindSQL = `SELECT nk.id, nk.schema_extension_id, nk.kind_id, k.name, nk.display_name, nk.description, nk.is_display_kind, nk.icon, nk.icon_color, nk.created_at, nk.updated_at, nk.deleted_at FROM schema_node_kinds AS nk JOIN kind AS k ON nk.kind_id = k.id WHERE nk.id = $1`

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func nodeKindColumns() []string {
	return []string{
		"id", "schema_extension_id", "kind_id", "name", "display_name", "description",
		"is_display_kind", "icon", "icon_color", "created_at", "updated_at", "deleted_at",
	}
}

func TestStore_GetNodeKind(t *testing.T) {
	var (
		ctx        = context.Background()
		nodeKindID = int32(42)
		createdAt  = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		updatedAt  = time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
		dbErr      = errors.New("db error")
	)

	tests := []struct {
		name         string
		expectations func(pool pgxmock.PgxPoolIface)
		wantNodeKind services.NodeKind
		wantErr      error
	}{
		{
			name: "success_-_maps_all_columns",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetNodeKindSQL).WithArgs(nodeKindID).WillReturnRows(
					pool.NewRows(nodeKindColumns()).AddRow(
						int32(42), int32(7), int32(99), "User", "User", "a user",
						true, "user", "#fff", createdAt, updatedAt, nil,
					),
				)
			},
			wantNodeKind: services.NodeKind{
				ID: 42, SchemaExtensionID: 7, KindID: 99, Name: "User", DisplayName: "User",
				Description: "a user", IsDisplayKind: true, Icon: "user", Color: "#fff",
				CreatedAt: createdAt, UpdatedAt: updatedAt, DeletedAt: nil,
			},
		},
		{
			name: "error_-_maps_no_rows_to_ErrNodeKindNotFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetNodeKindSQL).WithArgs(nodeKindID).WillReturnRows(
					pool.NewRows(nodeKindColumns()),
				)
			},
			wantNodeKind: services.NodeKind{},
			wantErr:      services.ErrNodeKindNotFound,
		},
		{
			name: "error_-_propagates_database_error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetNodeKindSQL).WithArgs(nodeKindID).WillReturnError(dbErr)
			},
			wantNodeKind: services.NodeKind{},
			wantErr:      dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			nodeKind, err := store.GetNodeKind(ctx, nodeKindID)
			assert.Equal(t, tt.wantNodeKind, nodeKind)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
