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

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/auth/internal/appdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// userSessionTimestamps holds the timestamps Logout is expected to change, alongside
// whether expires_at has passed according to the database clock.
type userSessionTimestamps struct {
	ExpiresAt time.Time
	UpdatedAt time.Time
	Expired   bool
}

// seedUser inserts a user row directly via the pool and returns its generated id.
func seedUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, principalName string) uuid.UUID {
	t.Helper()

	id, err := uuid.NewV4()
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO users (id, principal_name, first_name, last_name, email_address, last_login, is_disabled, all_environments, eula_accepted, support_account, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, now(), false, true, false, false, now(), now())`,
		id.String(), principalName, principalName+"-first", principalName+"-last", principalName+"@example.com",
	)
	require.NoError(t, err)

	return id
}

// seedActiveUserSession inserts a session for userID that expires an hour from now and was last
// updated an hour ago, so a successful Logout visibly moves both timestamps. Returns the session id.
func seedActiveUserSession(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) int64 {
	t.Helper()

	var sessionID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO `+appdb.TableUserSessions+` (user_id, auth_provider_type, auth_provider_id, expires_at, created_at, updated_at, flags)
		 VALUES ($1, 0, 0, now() + interval '1 hour', now() - interval '1 hour', now() - interval '1 hour', '{}')
		 RETURNING id`,
		userID.String(),
	).Scan(&sessionID)
	require.NoError(t, err)

	return sessionID
}

// readUserSessionTimestamps reads a session's timestamps, comparing expires_at against the
// database clock so the result does not depend on clock skew between the test and postgres.
func readUserSessionTimestamps(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sessionID int64) userSessionTimestamps {
	t.Helper()

	var timestamps userSessionTimestamps
	err := pool.QueryRow(ctx,
		`SELECT expires_at, updated_at, expires_at <= current_timestamp FROM `+appdb.TableUserSessions+` WHERE id = $1`,
		sessionID,
	).Scan(&timestamps.ExpiresAt, &timestamps.UpdatedAt, &timestamps.Expired)
	require.NoError(t, err)

	return timestamps
}

func TestStore_Logout_Integration(t *testing.T) {
	t.Run("expires only the target session", func(t *testing.T) {
		var (
			ctx          = context.Background()
			store, pool  = setupStoreAndPool(t)
			userID       = seedUser(t, ctx, pool, "logout-user")
			targetID     = seedActiveUserSession(t, ctx, pool, userID)
			otherID      = seedActiveUserSession(t, ctx, pool, userID)
			targetBefore = readUserSessionTimestamps(t, ctx, pool, targetID)
			otherBefore  = readUserSessionTimestamps(t, ctx, pool, otherID)
			targetAfter  userSessionTimestamps
			otherAfter   userSessionTimestamps
		)

		require.False(t, targetBefore.Expired, "seeded session must start active")

		require.NoError(t, store.Logout(ctx, targetID))

		targetAfter = readUserSessionTimestamps(t, ctx, pool, targetID)
		assert.True(t, targetAfter.Expired, "target session must be expired")
		assert.True(t, targetAfter.UpdatedAt.After(targetBefore.UpdatedAt), "target session updated_at must advance")
		assert.True(t, targetAfter.ExpiresAt.Equal(targetAfter.UpdatedAt), "expires_at and updated_at must be set from the same timestamp")

		otherAfter = readUserSessionTimestamps(t, ctx, pool, otherID)
		assert.False(t, otherAfter.Expired, "other sessions must remain active")
		assert.True(t, otherAfter.ExpiresAt.Equal(otherBefore.ExpiresAt), "other session expires_at must be unchanged")
		assert.True(t, otherAfter.UpdatedAt.Equal(otherBefore.UpdatedAt), "other session updated_at must be unchanged")
	})
}
