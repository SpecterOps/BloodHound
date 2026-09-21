// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package tools_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api/tools"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/stretchr/testify/require"
)

const (
	createDatabaseSwitchTableSQL = `create table if not exists database_switch (driver text not null, primary key(driver));`
	resolveApplicationName       = "resolve_graph_driver_integration_test"
)

func TestResolveGraphDriverIntegration(t *testing.T) {
	t.Run("returns configured driver when no driver is stored", func(t *testing.T) {
		var (
			ctx = context.Background()
			cfg = newGraphDriverTestConfig(t)
		)

		driverName, err := tools.ResolveGraphDriver(ctx, cfg)

		require.NoError(t, err)
		require.Equal(t, cfg.GraphDriver, driverName)
	})

	t.Run("returns stored driver", func(t *testing.T) {
		var (
			ctx          = context.Background()
			cfg          = newGraphDriverTestConfig(t)
			storedDriver = "pg"
		)

		connection, err := pgx.Connect(ctx, cfg.Database.PostgreSQLConnectionString())
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, connection.Close(ctx))
		})

		_, err = connection.Exec(ctx, createDatabaseSwitchTableSQL)
		require.NoError(t, err)
		_, err = connection.Exec(ctx, `insert into database_switch (driver) values ($1);`, storedDriver)
		require.NoError(t, err)

		driverName, err := tools.ResolveGraphDriver(ctx, cfg)

		require.NoError(t, err)
		require.Equal(t, storedDriver, driverName)
	})

	t.Run("returns table creation error", func(t *testing.T) {
		var (
			ctx = context.Background()
			cfg = newGraphDriverTestConfig(t)
		)
		cfg.Database.Connection = withConnectionParameter(t, cfg.Database.Connection, "search_path", "missing_schema")

		driverName, err := tools.ResolveGraphDriver(ctx, cfg)

		require.Error(t, err)
		require.Empty(t, driverName)
	})

	t.Run("returns driver query error", func(t *testing.T) {
		var (
			ctx = context.Background()
			cfg = newGraphDriverTestConfig(t)
		)

		connection, err := pgx.Connect(ctx, cfg.Database.PostgreSQLConnectionString())
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, connection.Close(ctx))
		})

		_, err = connection.Exec(ctx, `create table database_switch (not_driver text);`)
		require.NoError(t, err)

		driverName, err := tools.ResolveGraphDriver(ctx, cfg)

		require.Error(t, err)
		require.Empty(t, driverName)
	})
}

func TestResolveGraphDriverIgnoresConcurrentTableCreation(t *testing.T) {
	var (
		ctx, cancel      = context.WithTimeout(context.Background(), 15*time.Second)
		cfg              = newGraphDriverTestConfig(t)
		resolveError     error
		resolveWaitGroup sync.WaitGroup
	)
	t.Cleanup(cancel)

	creatingConnection, err := pgx.Connect(ctx, cfg.Database.PostgreSQLConnectionString())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, creatingConnection.Close(context.Background()))
	})

	creatingTransaction, err := creatingConnection.Begin(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = creatingTransaction.Rollback(context.Background())
	})

	_, err = creatingTransaction.Exec(ctx, createDatabaseSwitchTableSQL)
	require.NoError(t, err)

	resolveCfg := cfg
	resolveCfg.Database.Connection = withConnectionParameter(t, cfg.Database.Connection, "application_name", resolveApplicationName)
	resolveWaitGroup.Go(func() {
		_, resolveError = tools.ResolveGraphDriver(ctx, resolveCfg)
	})

	waitForBlockedTableCreation(t, ctx, cfg)
	require.NoError(t, creatingTransaction.Commit(ctx))
	resolveWaitGroup.Wait()
	require.NoError(t, resolveError)
}

func waitForBlockedTableCreation(t *testing.T, ctx context.Context, cfg config.Configuration) {
	t.Helper()

	observerConnection, err := pgx.Connect(ctx, cfg.Database.PostgreSQLConnectionString())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, observerConnection.Close(context.Background()))
	})

	require.Eventually(t, func() bool {
		var waitEventType string
		err := observerConnection.QueryRow(ctx, `
			select wait_event_type
			from pg_stat_activity
			where application_name = $1
			  and state = 'active'
			  and lower(query) like 'create table if not exists database_switch%';`, resolveApplicationName).Scan(&waitEventType)

		return err == nil && waitEventType == "Lock"
	}, time.Second, 10*time.Millisecond, "ResolveGraphDriver did not block behind the uncommitted CREATE TABLE")
}

func newGraphDriverTestConfig(t *testing.T) config.Configuration {
	t.Helper()

	var (
		connectionConfig = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
		cfg, err         = config.NewDefaultConnectionConfiguration(connectionConfig.URL())
	)
	require.NoError(t, err)
	cfg.GraphDriver = "neo4j"

	return cfg
}

func withConnectionParameter(t *testing.T, connectionString, key, value string) string {
	t.Helper()

	parsedURL, err := url.Parse(connectionString)
	require.NoError(t, err)

	query := parsedURL.Query()
	query.Set(key, value)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}

// getPostgresConfig reads key/value pairs from the integration test configuration and returns a pgtestdb configuration.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	var (
		integrationConfig, err = utils.LoadIntegrationTestConfig()
		environmentMap         = make(map[string]string)
	)
	require.NoError(t, err)

	for entry := range strings.FieldsSeq(integrationConfig.Database.Connection) {
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
