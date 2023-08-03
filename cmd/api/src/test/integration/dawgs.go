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
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/test"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func LoadConfiguration(testCtrl test.Controller) config.Configuration {
	cfg, err := utils.LoadIntegrationTestConfig()

	if err != nil {
		testCtrl.Fatalf("Failed loading integration test config: %v", err)
	}

	return cfg
}

func OpenPostgresqlGDB(testCtrl test.Controller) graph.Database {
	graphDatabase, err := dawgs.Open(neo4j.DriverName, LoadConfiguration(testCtrl).Database.PostgreSQLConnectionString())
	require.Nilf(testCtrl, err, "Failed connecting to graph database: %v", err)

	return graphDatabase
}

func OpenNeo4jGraphDB(testCtrl test.Controller) graph.Database {
	graphDatabase, err := dawgs.Open(neo4j.DriverName, LoadConfiguration(testCtrl).Neo4J.Neo4jConnectionString())
	require.Nilf(testCtrl, err, "Failed connecting to graph database: %v", err)

	return graphDatabase
}
