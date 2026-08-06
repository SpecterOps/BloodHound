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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/appdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func setupStore(t *testing.T) (*appdb.Store, *pgxpool.Pool) {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	bhDB := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, bhDB.Migrate(ctx))

	t.Cleanup(func() {
		bhDB.Close(ctx)
	})

	return appdb.NewStore(dbPool), dbPool
}

// seedEnvironment inserts an extension, environment kind and source kind, and a
// schema_environment linking them, returning the environment kind name.
func seedEnvironment(t *testing.T, pool *pgxpool.Pool, extName, kindName string, isBuiltin bool) string {
	t.Helper()
	ctx := context.Background()

	var extID int
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_extensions (name, namespace, display_name, version, is_builtin)
		VALUES ($1, $2, $3, '1.0.0', $4) RETURNING id
	`, extName, extName, extName+" Display", isBuiltin).Scan(&extID))

	var envKindID, sourceKindID int
	require.NoError(t, pool.QueryRow(ctx, `INSERT INTO kind (name) VALUES ($1) RETURNING id`, kindName).Scan(&envKindID))
	require.NoError(t, pool.QueryRow(ctx, `INSERT INTO kind (name) VALUES ($1) RETURNING id`, kindName+"_source").Scan(&sourceKindID))

	_, err := pool.Exec(ctx, `
		INSERT INTO schema_environments (schema_extension_id, environment_kind_id, source_kind_id)
		VALUES ($1, $2, $3)
	`, extID, envKindID, sourceKindID)
	require.NoError(t, err)

	return kindName
}

func TestStore_GetEnvironmentsFiltered_Integration(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
		builtinKind = seedEnvironment(t, pool, "BuiltinExt", "IntgBuiltinEnv", true)
		customKind  = seedEnvironment(t, pool, "CustomExt", "IntgCustomEnv", false)
	)

	t.Run("returns_all_environments", func(t *testing.T) {
		environments, err := store.GetEnvironmentsFiltered(ctx, false)
		require.NoError(t, err)

		var names []string
		for _, environment := range environments {
			names = append(names, environment.EnvironmentKindName)
		}
		assert.Contains(t, names, builtinKind)
		assert.Contains(t, names, customKind)
	})

	t.Run("returns_only_builtin_environments", func(t *testing.T) {
		environments, err := store.GetEnvironmentsFiltered(ctx, true)
		require.NoError(t, err)

		var names []string
		for _, environment := range environments {
			names = append(names, environment.EnvironmentKindName)
		}
		assert.Contains(t, names, builtinKind)
		assert.NotContains(t, names, customKind)
	})
}
