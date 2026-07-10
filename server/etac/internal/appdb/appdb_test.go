// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/gofrs/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/etac/internal/appdb"
	"github.com/specterops/bloodhound/server/etac/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedSelectSQL is compared via pgxmock.QueryMatcherEqual, which
// whitespace-normalises both sides; column order, table name and the WHERE
// predicate are load-bearing.
const expectedSelectSQL = `SELECT user_id, environment_id FROM environment_targeted_access_control WHERE user_id = $1`

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func TestStore_GetEnvironmentTargetedAccessControlForUser(t *testing.T) {
	var (
		ctx    = context.Background()
		userID = uuid.Must(uuid.NewV4())
		dbErr  = errors.New("connection refused")
	)

	t.Run("returns the access-control rows on success", func(t *testing.T) {
		store, pool := newTestStore(t)
		pool.ExpectQuery(expectedSelectSQL).WithArgs(userID.String()).WillReturnRows(
			pool.NewRows([]string{"user_id", "environment_id"}).
				AddRow(userID.String(), "env-1").
				AddRow(userID.String(), "env-2"),
		)

		list, err := store.GetEnvironmentTargetedAccessControlForUser(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, []services.EnvironmentTargetedAccessControl{
			{UserID: userID.String(), EnvironmentID: "env-1"},
			{UserID: userID.String(), EnvironmentID: "env-2"},
		}, list)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("returns an empty slice when the user has no rows", func(t *testing.T) {
		store, pool := newTestStore(t)
		pool.ExpectQuery(expectedSelectSQL).WithArgs(userID.String()).WillReturnRows(
			pool.NewRows([]string{"user_id", "environment_id"}),
		)

		list, err := store.GetEnvironmentTargetedAccessControlForUser(ctx, userID)

		require.NoError(t, err)
		assert.Empty(t, list)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("propagates database errors", func(t *testing.T) {
		store, pool := newTestStore(t)
		pool.ExpectQuery(expectedSelectSQL).WithArgs(userID.String()).WillReturnError(dbErr)

		_, err := store.GetEnvironmentTargetedAccessControlForUser(ctx, userID)

		assert.ErrorIs(t, err, dbErr)
		require.NoError(t, pool.ExpectationsWereMet())
	})
}
