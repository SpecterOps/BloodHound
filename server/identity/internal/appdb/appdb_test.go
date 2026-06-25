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
	"github.com/specterops/bloodhound/server/identity/internal/appdb"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedGetPermissionSQL is the literal SQL the Store issues for getPermission.
// It is compared via pgxmock.QueryMatcherEqual, which whitespace-normalises both
// sides, so token order, table name and parameter shape are what matter.
const expectedGetPermissionSQL = `SELECT * FROM permissions WHERE id = $1 LIMIT $2`

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func permissionRowColumns() []string {
	return []string{"authority", "name", "id", "created_at", "updated_at"}
}

func TestStore_GetPermission(t *testing.T) {
	var (
		ctx          = context.Background()
		dbErr        = errors.New("connection refused")
		permissionID = 5
		expected     = services.Permission{
			Authority: "clients",
			Name:      "ReadClients",
			ID:        5,
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		}
	)

	tests := []struct {
		name            string
		expectations    func(pool pgxmock.PgxPoolIface)
		wantResult      services.Permission
		wantErr         error
		wantErrContains string
	}{
		{
			name: "returns the permission on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					pool.NewRows(permissionRowColumns()).AddRow(
						expected.Authority,
						expected.Name,
						expected.ID,
						expected.CreatedAt,
						expected.UpdatedAt,
					),
				)
			},
			wantResult: expected,
		},
		{
			name: "maps CollectOneRow pgx.ErrNoRows to services.ErrNoPermissionFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				// Query succeeds but returns zero rows; CollectOneRow returns pgx.ErrNoRows
				pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					pool.NewRows(permissionRowColumns()),
				)
			},
			wantErr: services.ErrNoPermissionFound,
		},
		{
			name: "wraps CollectOneRow iteration error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				// The rows object carries a close error that pgx.CollectOneRow surfaces
				// via rows.Err() when Next() returns false.
				pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					pool.NewRows(permissionRowColumns()).CloseError(errors.New("forced iteration error")),
				)
			},
			wantErrContains: "finding permission:",
		},
		{
			name: "propagates other database errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			result, err := store.GetPermission(ctx, permissionID)
			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				assert.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
