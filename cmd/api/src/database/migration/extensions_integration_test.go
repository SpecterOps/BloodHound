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

package migration_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context    context.Context
	BHDatabase *database.BloodhoundDB
	DB         *gorm.DB
}

func setupIntegrationTestSuite(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
		gormDB   *gorm.DB
		db       *database.BloodhoundDB
		err      error
	)

	// #region Setup for dbs

	gormDB, err = database.OpenDatabase(connConf.URL())
	require.NoError(t, err)

	db = database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())

	err = db.Migrate(ctx)
	require.NoError(t, err)

	err = db.PopulateExtensionData(ctx)
	require.NoError(t, err)

	// #endregion

	return IntegrationTestSuite{
		Context:    ctx,
		BHDatabase: db,
		DB:         gormDB,
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

func (s *IntegrationTestSuite) teardownIntegrationTestSuite(t *testing.T) {
	t.Helper()

	if s.BHDatabase != nil {
		s.BHDatabase.Close(s.Context)
	}
}

func TestExtensions_GetOnStartExtensionData(t *testing.T) {
	var (
		testSuite = setupIntegrationTestSuite(t)
	)

	defer testSuite.teardownIntegrationTestSuite(t)

	err := testSuite.BHDatabase.PopulateExtensionData(testSuite.Context)
	require.NoError(t, err)

	// Validate Both Schema Extensions Exist
	extensions, totalRecords, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)

	require.NoError(t, err)

	require.Equal(t, 2, totalRecords)

	for _, extension := range extensions {
		require.True(t, extension.IsBuiltin, "All extensions should be marked as built-in")
		// Validate Schema Environments Exist
		schemaEnvironments, err := testSuite.BHDatabase.GetEnvironmentsByExtensionId(testSuite.Context, extension.ID)
		require.NoError(t, err)

		// There should only be one schema environment per built-in extension
		require.Len(t, schemaEnvironments, 1)
		schemaEnvironment := schemaEnvironments[0]

		// Validate Source Kinds Exist
		sourceKind, err := testSuite.BHDatabase.GetSourceKindByID(testSuite.Context, int(schemaEnvironment.SourceKindId))
		require.NoError(t, err)
		require.NotNil(t, sourceKind)
		validateSourceKind(t, extension.Name, sourceKind.Name.String())

		// Validate Environment Kinds Exist
		environmentKind, err := testSuite.BHDatabase.GetKindsByIDs(testSuite.Context, schemaEnvironment.EnvironmentKindId)
		require.NoError(t, err)
		validateEnvironmentKind(t, extension.Name, environmentKind[0].Name)
	}

}

func validateSourceKind(t *testing.T, extensionName, sourceKindName string) {
	t.Helper()
	switch extensionName {
	case "AD":
		require.Equal(t, "Base", sourceKindName)
	case "AZ":
		require.Equal(t, "AZBase", sourceKindName)
	default:
		t.Fatalf("Invalid extension name %s", extensionName)
	}
}

func validateEnvironmentKind(t *testing.T, extensionName, environmentKindName string) {
	t.Helper()
	switch extensionName {
	case "AD":
		require.Equal(t, "Domain", environmentKindName)
	case "AZ":
		require.Equal(t, "AZTenant", environmentKindName)
	default:
		t.Fatalf("Invalid extension name %s", extensionName)
	}
}
