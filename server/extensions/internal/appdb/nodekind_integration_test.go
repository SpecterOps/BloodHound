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
	"github.com/specterops/bloodhound/server/extensions/internal/appdb"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getPostgresConfig reads the integration test connection details from the
// configured environment and returns a pgtestdb.Config suitable for spinning
// up isolated databases. Supports both TCP and unix-socket host values.
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

// setupStore spins up an isolated postgres database via pgtestdb, applies the
// relational migrations and populates the built-in extension data, then returns a
// Store backed by the resulting pgx pool. No graph/dawgs setup is required because
// the extensions store reads only relational schema tables.
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
	require.NoError(t, bhDB.PopulateExtensionData(ctx))
	t.Cleanup(func() { bhDB.Close(ctx) })

	return appdb.NewStore(dbPool), dbPool
}

// testSetupData carries the ids produced by seedNodeKind so table-driven cases can
// reference the seeded node kind (and its kind) without re-querying.
type testSetupData struct {
	kindID     int32
	nodeKindID int32
}

// seedNodeKind inserts a schema_extensions row, a kind row and a schema_node_kinds row
// wiring them together, registers a cascade-delete cleanup and returns the seeded ids.
func seedNodeKind(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
	t.Helper()

	var extensionID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
		VALUES ('TestExtension', 'Test Extension', '1.0.0', false, 'TST')
		RETURNING id`).Scan(&extensionID))

	var kindID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO kind (name) VALUES ('TestNodeKind') RETURNING id`).Scan(&kindID))

	var nodeKindID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_node_kinds (schema_extension_id, kind_id, display_name, description, is_display_kind, icon, icon_color)
		VALUES ($1, $2, 'Test Node Kind', 'a test node kind', true, 'user', '#fff')
		RETURNING id`, extensionID, kindID).Scan(&nodeKindID))

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM schema_extensions WHERE name = 'TestExtension'`)
		_, _ = pool.Exec(ctx, `DELETE FROM kind WHERE name = 'TestNodeKind'`)
	})

	return testSetupData{kindID: kindID, nodeKindID: nodeKindID}
}

func TestStore_GetNodeKind_Integration(t *testing.T) {
	type testCase struct {
		name    string
		setup   func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData
		wantErr error
		assert  func(t *testing.T, nodeKind services.NodeKind, data testSetupData)
	}

	tests := []testCase{
		{
			name:  "success_-_returns_node_kind_fields",
			setup: seedNodeKind,
			assert: func(t *testing.T, nodeKind services.NodeKind, data testSetupData) {
				assert.Equal(t, data.nodeKindID, nodeKind.ID)
				assert.Equal(t, data.kindID, nodeKind.KindID)
				assert.Equal(t, "TestNodeKind", nodeKind.Name)
				assert.Equal(t, "Test Node Kind", nodeKind.DisplayName)
				assert.Equal(t, "a test node kind", nodeKind.Description)
				assert.True(t, nodeKind.IsDisplayKind)
				assert.Equal(t, "user", nodeKind.Icon)
				assert.Equal(t, "#fff", nodeKind.Color)
			},
		},
		{
			name: "error_-_returns_ErrNodeKindNotFound",
			setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
				return testSetupData{nodeKindID: int32(999999999)}
			},
			wantErr: services.ErrNodeKindNotFound,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				ctx         = context.Background()
				data        = testCase.setup(t, ctx, pool)
			)

			nodeKind, err := store.GetNodeKind(ctx, data.nodeKindID)
			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}
			require.NoError(t, err)
			testCase.assert(t, nodeKind, data)
		})
	}
}

func TestStore_GetKindInfosByNodeKindID_Integration(t *testing.T) {
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

			infos, err := store.GetKindInfosByNodeKindID(ctx, data.nodeKindID)
			require.NoError(t, err)
			testCase.assert(t, infos, data)
		})
	}
}

// seedKindInfos inserts two schema_kind_info rows for the supplied kind/node kind,
// deliberately ordering the INSERT by position 1 then 0 to prove GetKindInfosByNodeKindID
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
