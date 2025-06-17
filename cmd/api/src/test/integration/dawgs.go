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

package integration

import (
	"context"

	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/test"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/neo4j"
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

func OpenGraphDB(testCtrl test.Controller, schema graph.Schema) graph.Database {
	var (
		cfg           = LoadConfiguration(testCtrl)
		graphDatabase graph.Database
		err           error
	)

	switch cfg.GraphDriver {
	case pg.DriverName:
		pool, err := pg.NewPool(cfg.Database.PostgreSQLConnectionString())
		test.RequireNilErrf(testCtrl, err, "Failed to create new pgx pool: %v", err)
		graphDatabase, err = dawgs.Open(context.TODO(), cfg.GraphDriver, dawgs.Config{
			ConnectionString: cfg.Database.PostgreSQLConnectionString(),
			Pool:             pool,
		})
		test.RequireNilErrf(testCtrl, err, "Failed connecting to graph database: %v", err)

	case neo4j.DriverName:
		graphDatabase, err = dawgs.Open(context.TODO(), cfg.GraphDriver, dawgs.Config{
			ConnectionString: cfg.Neo4J.Neo4jConnectionString(),
		})
		test.RequireNilErrf(testCtrl, err, "Failed connecting to graph database: %v", err)

	default:
		testCtrl.Fatalf("unsupported graph driver name %s", cfg.GraphDriver)
	}

	test.RequireNilErr(testCtrl, graphDatabase.AssertSchema(context.Background(), schema))

	return graphDatabase
}
