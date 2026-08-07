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
	"github.com/specterops/bloodhound/server/assetgrouptags/internal/appdb"
	"github.com/specterops/bloodhound/server/assetgrouptags/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getPostgresConfig reads the integration test connection details from the
// configured environment and returns a pgtestdb.Config suitable for spinning up
// isolated databases. Supports both TCP and unix-socket host values.
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

func TestStore_GetAssetGroupTagByID_Integration(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
		tagID       int
	)

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO asset_group_tags (name, type, description, created_by, created_at, updated_by, updated_at, position, require_certify)
		VALUES ('Custom Tag', 2, 'Custom', 'SYSTEM', current_timestamp, 'SYSTEM', current_timestamp, 2, FALSE)
		RETURNING id
	`).Scan(&tagID))

	t.Run("success_-_returns_existing_tag", func(t *testing.T) {
		tag, err := store.GetAssetGroupTagByID(ctx, tagID)
		require.NoError(t, err)
		assert.Equal(t, services.AssetGroupTag{ID: tagID}, tag)
	})

	t.Run("error_-_returns_not_found_for_missing_tag", func(t *testing.T) {
		_, err := store.GetAssetGroupTagByID(ctx, 999999)
		assert.ErrorIs(t, err, services.ErrAssetGroupTagNotFound)
	})
}

func TestStore_GetTierZeroTag_Integration(t *testing.T) {
	var (
		ctx      = context.Background()
		store, _ = setupStore(t)
	)

	// The baseline migration seeds a Tier Zero tag (type=1, position=1).
	tag, err := store.GetTierZeroTag(ctx)
	require.NoError(t, err)
	assert.NotZero(t, tag.ID)
}
