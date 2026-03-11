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
package test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context context.Context
	GraphDB graph.Database
}

// getPostgresConfig reads key/value pairs from the default integration
// config file and creates a pgtestdb configuration object.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	config, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(config.Database.Connection) {
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

func teardownIntegrationTestSuite(t *testing.T, suite IntegrationTestSuite) {
	t.Helper()

	if suite.GraphDB != nil {
		if err := suite.GraphDB.Close(suite.Context); err != nil {
			t.Logf("Failed to close GraphDB: %v", err)
		}
	}
}

// setupIntegrationTestSuite initializes and returns a test suite containing
// all necessary dependencies for integration tests, including a connected
// graph database instance and a configured graph service.
func SetupIntegrationTestSuite(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	//#region Setup for dbs
	pool, err := pg.NewPool(connConf.URL())
	require.NoError(t, err)

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  pool,
	})
	require.NoError(t, err)

	err = graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema())
	require.NoError(t, err)

	suite := IntegrationTestSuite{
		Context: ctx,
		GraphDB: graphDB,
	}

	t.Cleanup(func() {
		teardownIntegrationTestSuite(t, suite)
	})

	return suite
}
