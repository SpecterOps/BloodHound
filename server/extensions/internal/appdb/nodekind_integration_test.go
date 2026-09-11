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
// Store backed by the resulting pgx pool.
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
