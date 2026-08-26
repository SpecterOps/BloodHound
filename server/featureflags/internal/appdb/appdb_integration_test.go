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
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/featureflags/internal/appdb"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStore(t *testing.T) (*appdb.Store, *pgxpool.Pool) {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)
	cfg.Database.Connection = connConf.URL()

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)

	err = db.Migrate(ctx)
	require.NoError(t, err)

	t.Cleanup(func() { db.Close(ctx) })

	return appdb.NewStore(db.Pool()), db.Pool()
}

func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(cfg.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

func TestStore_GetFlagByKey_Integration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		seed       func(t *testing.T, pool *pgxpool.Pool)
		key        string
		wantErr    error
		assertFlag func(t *testing.T, flag services.FeatureFlag)
	}{
		{
			name: "returns the flag for the supplied key",
			seed: func(t *testing.T, pool *pgxpool.Pool) {
				t.Helper()
				_, err := pool.Exec(ctx,
					"INSERT INTO feature_flags (id, key, name, description, enabled, user_updatable, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, now(), now())",
					int32(9001), "integration_test_flag", "Integration Test Flag", "a flag", true, false)
				require.NoError(t, err)
			},
			key: "integration_test_flag",
			assertFlag: func(t *testing.T, flag services.FeatureFlag) {
				t.Helper()
				assert.Equal(t, "integration_test_flag", flag.Key)
				assert.Equal(t, "Integration Test Flag", flag.Name)
				assert.True(t, flag.Enabled)
			},
		},
		{
			name:    "returns ErrNotFound for an unknown key",
			key:     "does_not_exist_flag",
			wantErr: services.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := setupStore(t)
			if tt.seed != nil {
				tt.seed(t, pool)
			}

			flag, err := store.GetFlagByKey(ctx, tt.key)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				tt.assertFlag(t, flag)
			}
		})
	}
}

// seedFlag inserts a feature_flags row using the table sequence to allocate an id and
// returns the assigned id. The Store API does not expose flag creation.
func seedFlag(t *testing.T, ctx context.Context, pool *pgxpool.Pool, key, name string, enabled, userUpdatable bool) int32 {
	t.Helper()

	var id int32
	err := pool.QueryRow(ctx,
		`INSERT INTO feature_flags (id, key, name, description, enabled, user_updatable, created_at, updated_at)
		   VALUES (nextval('feature_flags_id_seq'), $1, $2, $3, $4, $5, now(), now())
		   RETURNING id`,
		key, name, "seeded by integration test", enabled, userUpdatable,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// authedContext attaches a bhctx.Context carrying the supplied user as the
// auth owner. SetFlag's audit-log path requires this when UserUpdatable is true.
func authedContext(t *testing.T) context.Context {
	t.Helper()

	userID, err := uuid.NewV4()
	require.NoError(t, err)

	return bhctx.Set(context.Background(), &bhctx.Context{
		RequestID: "integration-request",
		RequestIP: "127.0.0.1",
		AuthCtx: auth.Context{
			Owner: model.User{
				Unique:        model.Unique{ID: userID},
				PrincipalName: "integration-user",
			},
		},
	})
}

func TestStore_GetFlagByID_Integration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		seed       func(t *testing.T, pool *pgxpool.Pool) int32
		id         func(seededID int32) int32
		wantErr    error
		assertFlag func(t *testing.T, seededID int32, flag services.FeatureFlag)
	}{
		{
			name: "returns the flag for the supplied id",
			seed: func(t *testing.T, pool *pgxpool.Pool) int32 {
				t.Helper()
				return seedFlag(t, ctx, pool, "by_id_flag", "ID Flag", true, false)
			},
			id: func(seededID int32) int32 { return seededID },
			assertFlag: func(t *testing.T, seededID int32, flag services.FeatureFlag) {
				t.Helper()
				assert.Equal(t, seededID, flag.ID)
				assert.Equal(t, "by_id_flag", flag.Key)
				assert.True(t, flag.Enabled)
			},
		},
		{
			name:    "returns ErrNotFound for an unknown id",
			id:      func(seededID int32) int32 { return 999999 },
			wantErr: services.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				seededID    int32
			)

			if tt.seed != nil {
				seededID = tt.seed(t, pool)
			}

			flag, err := store.GetFlagByID(ctx, tt.id(seededID))
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				tt.assertFlag(t, seededID, flag)
			}
		})
	}
}

func TestStore_GetAllFlags_Integration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		seed        func(t *testing.T, pool *pgxpool.Pool) int32
		assertFlags func(t *testing.T, seededID int32, flags []services.FeatureFlag)
	}{
		{
			name: "includes seeded and migration-provided flags",
			seed: func(t *testing.T, pool *pgxpool.Pool) int32 {
				t.Helper()
				return seedFlag(t, ctx, pool, "get_all_flag", "Get All Flag", false, true)
			},
			assertFlags: func(t *testing.T, seededID int32, flags []services.FeatureFlag) {
				t.Helper()
				require.NotEmpty(t, flags, "migrations should populate baseline flags")

				found := false
				for _, flag := range flags {
					if flag.ID == seededID {
						found = true
						assert.Equal(t, "get_all_flag", flag.Key)
						assert.False(t, flag.Enabled)
						assert.True(t, flag.UserUpdatable)
					}
				}
				assert.True(t, found, "seeded flag should appear in the result set")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := setupStore(t)
			seededID := tt.seed(t, pool)

			flags, err := store.GetAllFlags(ctx)
			require.NoError(t, err)
			tt.assertFlags(t, seededID, flags)
		})
	}
}

func TestStore_SetFlag_Integration(t *testing.T) {
	tests := []struct {
		name          string
		makeCtx       func(t *testing.T) context.Context
		seedKey       string
		seedName      string
		userUpdatable bool
		verify        func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool, id int32, setErr error)
	}{
		{
			name:          "flips a non-user-updatable flag without writing an audit log",
			makeCtx:       func(t *testing.T) context.Context { return context.Background() },
			seedKey:       "set_flag_locked",
			seedName:      "Locked",
			userUpdatable: false,
			verify: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool, id int32, setErr error) {
				t.Helper()
				require.NoError(t, setErr)

				got, err := store.GetFlagByID(ctx, id)
				require.NoError(t, err)
				assert.True(t, got.Enabled)

				var auditCount int
				require.NoError(t, pool.QueryRow(ctx,
					"SELECT COUNT(*) FROM audit_logs WHERE action = $1",
					string(model.AuditLogActionToggleEarlyAccessFeatureFlag),
				).Scan(&auditCount))
				assert.Zero(t, auditCount, "no audit log should be written for non-user-updatable flags")
			},
		},
		{
			name:          "flips a user-updatable flag and writes an audit log",
			makeCtx:       authedContext,
			seedKey:       "set_flag_unlocked",
			seedName:      "Unlocked",
			userUpdatable: true,
			verify: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool, id int32, setErr error) {
				t.Helper()
				require.NoError(t, setErr)

				got, err := store.GetFlagByID(ctx, id)
				require.NoError(t, err)
				assert.True(t, got.Enabled)

				var auditCount int
				require.NoError(t, pool.QueryRow(ctx,
					"SELECT COUNT(*) FROM audit_logs WHERE action = $1 AND actor_name = $2",
					string(model.AuditLogActionToggleEarlyAccessFeatureFlag),
					"integration-user",
				).Scan(&auditCount))
				assert.Equal(t, 1, auditCount, "exactly one audit log should be written")
			},
		},
		{
			name:          "returns an error when no authenticated user is on the context for a user-updatable flag",
			makeCtx:       func(t *testing.T) context.Context { return context.Background() },
			seedKey:       "set_flag_anon",
			seedName:      "Anon",
			userUpdatable: true,
			verify: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool, id int32, setErr error) {
				t.Helper()
				require.Error(t, setErr)
				assert.Contains(t, setErr.Error(), "no authenticated user on context")

				// Transaction should have rolled back, leaving the flag disabled.
				got, err := store.GetFlagByID(ctx, id)
				require.NoError(t, err)
				assert.False(t, got.Enabled, "rolled-back update must not be visible")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				ctx         = tt.makeCtx(t)
				store, pool = setupStore(t)
				id          = seedFlag(t, ctx, pool, tt.seedKey, tt.seedName, false, tt.userUpdatable)
			)

			updated, err := store.GetFlagByID(ctx, id)
			require.NoError(t, err)
			updated.Enabled = true

			setErr := store.SetFlag(ctx, updated)
			tt.verify(t, ctx, store, pool, id, setErr)
		})
	}
}
