// Copyright 2025 Specter Ops, Inc.
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

package database_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context    context.Context
	BHDatabase *database.BloodhoundDB
}

// setupIntegrationTestSuite initializes and returns a test suite containing
// all necessary dependencies for integration tests, including a connected
// graph database instance and a configured graph service.
func setupIntegrationTestSuite(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	// #region Setup for dbs

	gormDB, err := database.OpenDatabase(connConf.URL())
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())

	err = db.Migrate(ctx)
	require.NoError(t, err)

	// #endregion

	return IntegrationTestSuite{
		Context:    ctx,
		BHDatabase: db,
	}
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

func teardownIntegrationTestSuite(t *testing.T, suite *IntegrationTestSuite) {
	t.Helper()

	if suite.BHDatabase != nil {
		suite.BHDatabase.Close(suite.Context)
	}
}
