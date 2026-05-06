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

package migrations_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite carries only the fields needed to exercise
// Version_920_Migration: a context, a graph database handle, and a
// BloodhoundDB that satisfies database.SourceKindsData.
type IntegrationTestSuite struct {
	context    context.Context
	graphDB    graph.Database
	bhDatabase *database.BloodhoundDB
}

// setupIntegrationTestSuite stands up an isolated relational and graph
// database for a single migration test. The BloodhoundDB SQL migrations
// auto-register ad.Entity and azure.Entity as source kinds; tests that need
// additional source kinds can call bhDatabase.RegisterSourceKind directly.
func setupIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	pool, err := pg.NewPool(cfg.Database)
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	bhDatabase := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, bhDatabase.Migrate(ctx))
	require.NoError(t, bhDatabase.PopulateExtensionData(ctx))

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  pool,
	})
	require.NoError(t, err)

	err = graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema())
	require.NoError(t, err)

	return &IntegrationTestSuite{
		context:    ctx,
		graphDB:    graphDB,
		bhDatabase: bhDatabase,
	}
}

func (s *IntegrationTestSuite) teardownIntegrationTestSuite(t *testing.T) {
	t.Helper()

	if s.graphDB != nil {
		if err := s.graphDB.Close(s.context); err != nil {
			t.Logf("failed to close GraphDB: %v", err)
		}
	}
	if s.bhDatabase != nil {
		s.bhDatabase.Close(s.context)
	}
}

// createNodes writes each provided node to the graph and back-fills the node
// ID so callers can reference the persisted nodes in assertions.
func (s *IntegrationTestSuite) createNodes(t *testing.T, nodes ...*graph.Node) {
	t.Helper()

	for _, node := range nodes {
		err := s.graphDB.WriteTransaction(s.context, func(tx graph.Transaction) error {
			createdNode, err := tx.CreateNode(node.Properties, node.Kinds...)
			if err != nil {
				return err
			}
			node.ID = createdNode.ID
			return nil
		})
		require.NoError(t, err, "unexpected error occurred while creating nodes")
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
