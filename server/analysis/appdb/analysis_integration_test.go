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

//go:build slow_integration

package appdb_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/analysis/appdb"
	"github.com/specterops/bloodhound/server/analysis/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupStore spins up an isolated postgres database via pgtestdb, applies all
// migrations, and returns an analysis Store backed by the resulting pgx pool.
func setupStore(t *testing.T) *appdb.Store {
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

	t.Cleanup(func() { bhDB.Close(ctx) })

	return appdb.NewStore(bhDB.Pool())
}

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

func TestStore_CreateAnalysisRequest_Integration(t *testing.T) {
	var (
		ctx   = context.Background()
		store = setupStore(t)
	)

	t.Run("first call creates the request and reports created=true", func(t *testing.T) {
		current, created, err := store.CreateAnalysisRequest(ctx, "first-user")
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, "first-user", current.RequestedBy)
		assert.Equal(t, service.RequestedAnalysisTypeAnalysis, current.RequestType)
	})

	t.Run("second call is a no-op and reports created=false with the original requester", func(t *testing.T) {
		current, created, err := store.CreateAnalysisRequest(ctx, "second-user")
		require.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, "first-user", current.RequestedBy)
	})

	t.Run("GetAnalysisRequest returns the persisted request", func(t *testing.T) {
		current, err := store.GetAnalysisRequest(ctx)
		require.NoError(t, err)
		assert.Equal(t, "first-user", current.RequestedBy)
	})
}

func TestStore_CreateAnalysisRequest_ConcurrentCallsDoNotDuplicate(t *testing.T) {
	const concurrentCallers = 16

	var (
		ctx      = context.Background()
		store    = setupStore(t)
		results  = make(chan bool, concurrentCallers)
		errs     = make(chan error, concurrentCallers)
		startGun sync.WaitGroup
		done     sync.WaitGroup
	)

	startGun.Add(1)
	for callerIndex := 0; callerIndex < concurrentCallers; callerIndex++ {
		done.Add(1)
		go func(index int) {
			defer done.Done()
			startGun.Wait()
			_, created, err := store.CreateAnalysisRequest(ctx, fmt.Sprintf("user-%d", index))
			results <- created
			errs <- err
		}(callerIndex)
	}

	startGun.Done()
	done.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	createdCount := 0
	for created := range results {
		if created {
			createdCount++
		}
	}
	assert.Equal(t, 1, createdCount, "exactly one concurrent caller should report created=true")
}
