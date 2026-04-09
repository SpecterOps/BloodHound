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
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type IntegrationTestSuite struct {
	Context    context.Context
	BHDatabase *database.BloodhoundDB
	DB         *gorm.DB
}

// setupIntegrationTestSuite initializes and returns a test suite containing
// all necessary dependencies for integration tests, including a connected
// graph database instance and a configured graph service. The base GORM db
// can be used for scenarios where tests require additional data
// that cannot be inserted via public database.BloodhoundDB methods
// (ex: insert a built-in OpenGraph Extension).
func setupIntegrationTestSuite(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
		cfg      = config.Configuration{}
	)

	// #region Setup for dbs

	gormDB, err := database.OpenDatabase(connConf.URL())
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver(), cfg)

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

func teardownIntegrationTestSuite(t *testing.T, suite *IntegrationTestSuite) {
	t.Helper()

	if suite.BHDatabase != nil {
		suite.BHDatabase.Close(suite.Context)
	}
}

func TestTransaction(t *testing.T) {
	t.Run("Success: operations commit together", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Get initial flag state
		flag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, "opengraph_search")
		require.NoError(t, err)
		originalEnabled := flag.Enabled

		// Update flag in a transaction
		err = testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			flag.Enabled = !originalEnabled
			return tx.SetFlag(testSuite.Context, flag)
		})
		require.NoError(t, err)

		// Verify the flag was updated
		updatedFlag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, "opengraph_search")
		require.NoError(t, err)
		require.Equal(t, !originalEnabled, updatedFlag.Enabled)
	})

	t.Run("Rollback: error causes operations to rollback", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Get initial flag state
		flag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, "opengraph_search")
		require.NoError(t, err)
		originalEnabled := flag.Enabled

		// Update flag then return error - should rollback
		expectedErr := fmt.Errorf("intentional error to trigger rollback")
		err = testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			flag.Enabled = !originalEnabled
			if err := tx.SetFlag(testSuite.Context, flag); err != nil {
				return err
			}
			return expectedErr
		})
		require.ErrorIs(t, err, expectedErr)

		// Verify the flag was NOT updated (rolled back)
		unchangedFlag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, "opengraph_search")
		require.NoError(t, err)
		require.Equal(t, originalEnabled, unchangedFlag.Enabled)
	})

	t.Run("Success: nested method calls work within transaction", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Verify we can call multiple different methods in a transaction
		err := testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			// Call GetAllFlags - read operation
			flags, err := tx.GetAllFlags(testSuite.Context)
			if err != nil {
				return err
			}
			require.NotEmpty(t, flags)

			// Call GetFlagByKey - another read operation
			_, err = tx.GetFlagByKey(testSuite.Context, "opengraph_search")
			return err
		})
		require.NoError(t, err)
	})

	t.Run("Success: transaction with isolation level option", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Get initial flag state
		flag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, "opengraph_search")
		require.NoError(t, err)
		originalEnabled := flag.Enabled

		// Update flag in a transaction with serializable isolation
		err = testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			flag.Enabled = !originalEnabled
			return tx.SetFlag(testSuite.Context, flag)
		}, &sql.TxOptions{Isolation: sql.LevelSerializable})
		require.NoError(t, err)

		// Verify the flag was updated
		updatedFlag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, "opengraph_search")
		require.NoError(t, err)
		require.Equal(t, !originalEnabled, updatedFlag.Enabled)
	})

	t.Run("Success: read-only transaction", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Read-only transaction should work for read operations
		err := testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			_, err := tx.GetAllFlags(testSuite.Context)
			return err
		}, &sql.TxOptions{ReadOnly: true})
		require.NoError(t, err)
	})

	t.Run("Fail: write in read-only transaction", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Get a flag to modify
		flag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, "opengraph_search")
		require.NoError(t, err)

		// Attempting to write in a read-only transaction should fail
		err = testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			flag.Enabled = !flag.Enabled
			return tx.SetFlag(testSuite.Context, flag)
		}, &sql.TxOptions{ReadOnly: true})
		require.Error(t, err)
	})
}

func TestGetNodeKindDisplayLabel(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
	require.NoError(t, err)
	// Create 2 OG display kinds
	nodeKindDis, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, "TestKindDis", extension.ID, "", "", true, "", "")
	require.NoError(t, err)
	nodeKindNoDis, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, "TestKindNoDis", extension.ID, "", "", false, "", "")
	require.NoError(t, err)

	primaryNodeKinds, err := testSuite.BHDatabase.GetValidDisplayKinds(testSuite.Context)
	require.NoError(t, err)

	assert.Equal(t, ad.Entity.String(), graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity)), "should return base kind if no other valid kinds are present")
	assert.Equal(t, ad.User.String(), graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.User)), "should return valid AD kind when base and kind are present")
	assert.Equal(t, ad.Group.String(), graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.Group, ad.LocalGroup)), "should return valid kind other than LocalGroup if one is present")
	assert.Equal(t, ad.LocalGroup.String(), graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.LocalGroup)), "should return LocalGroup if no other valid kinds are present")
	assert.Equal(t, azure.Group.String(), graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), azure.Entity, azure.Group)), "should return valid Azure kind when base and kind are present")
	assert.Equal(t, analysis.NodeKindUnknown, graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), graph.StringKind("unsupported"))), "should return Unknown when only an unsupported kind is present")
	assert.Equal(t, ad.Entity.String(), graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, graph.StringKind("unsupported"))), "should return valid kind if one is present even if an unsupported kind is also present")
	assert.Equal(t, analysis.NodeKindUnknown, graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties())), "should return Unknown if no node has no kinds on it")
	assert.Equal(t, nodeKindDis.Name, graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), nodeKindDis.ToKind())), "should return open graph kind if it is a display kind")
	assert.Equal(t, analysis.NodeKindUnknown, graphschema.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), nodeKindNoDis.ToKind())), "should return unknown kind if open graph kind is not a display kind and has no base")
}
