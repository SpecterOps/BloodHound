// Copyright 2023 Specter Ops, Inc.
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

// package integration is a set of old integration testing tools
//
// Deprecated: integration package is deprecated, see most recent guidance on proper testing
package integration

import (
	"context"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/test"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
)

func LoadConfiguration(testCtrl test.Controller) config.Configuration {
	cfg, err := utils.LoadIntegrationTestConfig()

	if err != nil {
		testCtrl.Fatalf("Failed loading integration test config: %v", err)
	}

	return cfg
}

// OpenGraphDB opens a new graph Database
//
// Deprecated: see newer testing guidance on how to open DB connnections, this will be removed in the future as old tests are rewritten.
func OpenGraphDB(t *testing.T, schema graph.Schema) graph.Database {
	var (
		graphDatabase graph.Database
	)

	cfg, err := utils.LoadIntegrationTestConfig()
	test.RequireNilErrf(t, err, "Failed to Load Integration Test Config: %v", err)

	switch cfg.GraphDriver {
	case pg.DriverName:
		connConf := pgtestdb.Custom(t, GetPostgresConfig(cfg), pgtestdb.NoopMigrator{})
		pool, err := pg.NewPool(connConf.URL())
		test.RequireNilErrf(t, err, "Failed to create new pgx pool: %v", err)
		graphDatabase, err = dawgs.Open(context.Background(), cfg.GraphDriver, dawgs.Config{
			ConnectionString: connConf.URL(),
			Pool:             pool,
		})
		test.RequireNilErrf(t, err, "Failed connecting to graph database: %v", err)

	default:
		t.Fatalf("unsupported graph driver name %s", cfg.GraphDriver)
	}

	test.RequireNilErr(t, graphDatabase.AssertSchema(context.Background(), schema))

	return graphDatabase
}
