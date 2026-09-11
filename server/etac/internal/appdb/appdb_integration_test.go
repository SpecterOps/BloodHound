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
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/etac/internal/appdb"
	"github.com/specterops/bloodhound/server/etac/internal/services"
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
	for entry := range strings.FieldsSeq(cfg.Database.Connection) {
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

func TestStore_GetEnvironmentTargetedAccessControlForUser_Integration(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
		userID      = uuid.Must(uuid.NewV4())
	)

	_, err := pool.Exec(ctx,
		"INSERT INTO users (id, principal_name, all_environments, created_at, updated_at) VALUES ($1, $2, false, now(), now())",
		userID.String(), "etac-user")
	require.NoError(t, err)

	for _, environmentID := range []string{"env-a", "env-b"} {
		_, err = pool.Exec(ctx,
			"INSERT INTO environment_targeted_access_control (user_id, environment_id) VALUES ($1, $2)",
			userID.String(), environmentID)
		require.NoError(t, err)
	}

	list, err := store.GetEnvironmentTargetedAccessControlForUser(ctx, userID)
	require.NoError(t, err)

	assert.ElementsMatch(t, []services.EnvironmentTargetedAccessControl{
		{UserID: userID.String(), EnvironmentID: "env-a"},
		{UserID: userID.String(), EnvironmentID: "env-b"},
	}, list)
}

func TestStore_GetEnvironmentTargetedAccessControlForUser_Integration_Empty(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
		userID      = uuid.Must(uuid.NewV4())
	)

	_, err := pool.Exec(ctx,
		"INSERT INTO users (id, principal_name, all_environments, created_at, updated_at) VALUES ($1, $2, false, now(), now())",
		userID.String(), "etac-user-empty")
	require.NoError(t, err)

	list, err := store.GetEnvironmentTargetedAccessControlForUser(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, list)
}
