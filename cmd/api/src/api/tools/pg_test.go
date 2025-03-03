// Copyright 2024 Specter Ops, Inc.
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

//go:build disabled
// +build disabled

package tools_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	pg_query "github.com/specterops/bloodhound/dawgs/drivers/pg/query"
	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/api/tools"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSwitchPostgreSQL(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		graphDB  = graph_mocks.NewMockDatabase(mockCtrl)
		request  = httptest.NewRequest(http.MethodPut, "/graph-db/switch/pg", nil)
		recorder = httptest.NewRecorder()
		ctx      = request.Context()
		migrator = setupTestMigrator(t, ctx, graphDB)
	)

	// lookup creates the database_switch table if needed
	driver, err := tools.LookupGraphDriver(migrator.ServerCtx, migrator.Cfg)
	require.Nil(t, err)

	if driver != neo4j.DriverName {
		err = tools.SetGraphDriver(migrator.ServerCtx, migrator.Cfg, neo4j.DriverName)
		require.Nil(t, err)
	}

	migrator.SwitchPostgreSQL(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)

	driver, err = tools.LookupGraphDriver(migrator.ServerCtx, migrator.Cfg)
	require.Nil(t, err)
	require.Equal(t, pg.DriverName, driver)
}

func TestSwitchNeo4j(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		graphDB  = graph_mocks.NewMockDatabase(mockCtrl)
		request  = httptest.NewRequest(http.MethodPut, "/graph-db/switch/neo4j", nil)
		recorder = httptest.NewRecorder()
		ctx      = request.Context()
		migrator = setupTestMigrator(t, ctx, graphDB)
	)

	driver, err := tools.LookupGraphDriver(migrator.ServerCtx, migrator.Cfg)
	require.Nil(t, err)

	if driver != pg.DriverName {
		err = tools.SetGraphDriver(migrator.ServerCtx, migrator.Cfg, pg.DriverName)
		require.Nil(t, err)
	}

	migrator.SwitchNeo4j(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)

	driver, err = tools.LookupGraphDriver(migrator.ServerCtx, migrator.Cfg)
	require.Nil(t, err)
	require.Equal(t, neo4j.DriverName, driver)
}

func TestPGMigrator(t *testing.T) {
	var (
		schema      = graphschema.DefaultGraphSchema()
		testContext = integration.NewGraphTestContext(t, schema)
	)

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.DBMigrateHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, neo4jDB graph.Database) {
		var (
			migrator        = setupTestMigrator(t, testContext.Context(), neo4jDB)
			testID          = harness.DBMigrateHarness.TestID.String()
			sourceNodeKinds graph.Kinds
			sourceEdgeKinds graph.Kinds
			sourceNodes     []*graph.Node
			sourceEdges     []*graph.Relationship
		)

		pgDB, err := migrator.OpenPostgresGraphConnection()
		require.Nil(t, err)

		// clear out nodes to avoid conflict when running the test multiple times
		err = pgDB.WriteTransaction(testContext.Context(), func(tx graph.Transaction) error {
			return tx.Nodes().Delete()
		})
		require.Nil(t, err)

		err = migrator.StartMigrationToPG()
		require.Nil(t, err)

		// wait until migration status returns to "idle"
		for {
			if migrator.State == tools.StateMigrating {
				time.Sleep(time.Second / 10)
			} else if migrator.State == tools.StateIdle {
				break
			} else {
				t.Fatalf("Encountered invalid migration status: %s", migrator.State)
			}
		}

		// query nodes/relationships in neo4j
		err = neo4jDB.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
			sourceNodes, err = ops.FetchNodes(tx.Nodes())
			require.Nil(t, err)

			sourceEdges, err = ops.FetchRelationships(tx.Relationships())
			require.Nil(t, err)

			return nil
		})
		require.Nil(t, err)

		// grab source kinds
		// NOTE: the call to db.labels() in our migrator returns all possible node kinds in neo4j, while db.relationshipTypes()
		// returns just those edge kinds that have an associated edge in the db, so that is the behavior we are testing here
		sourceNodeKinds = schema.DefaultGraph.Nodes

		for _, edge := range sourceEdges {
			if !slices.Contains(sourceEdgeKinds, edge.Kind) {
				sourceEdgeKinds = append(sourceEdgeKinds, edge.Kind)
			}
		}

		// confirm that all the data from neo4j made it to pg
		err = pgDB.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {

			// check nodes
			for _, sourceNode := range sourceNodes {
				id, err := sourceNode.Properties.Get(testID).String()
				require.Nil(t, err)

				if targetNode, err := tx.Nodes().Filterf(func() graph.Criteria {
					return query.Equals(query.NodeProperty(testID), id)
				}).First(); err != nil {
					t.Fatalf("Could not find migrated node with '%s' == %s", testID, id)
				} else {
					require.Equal(t, sourceNode.Kinds, targetNode.Kinds)
					require.Equal(t, sourceNode.Properties.Get(common.Name.String()), targetNode.Properties.Get(common.Name.String()))
					require.Equal(t, sourceNode.Properties.Get(common.ObjectID.String()), targetNode.Properties.Get(common.ObjectID.String()))
				}
			}

			// check edges
			for _, sourceEdge := range sourceEdges {
				id, err := sourceEdge.Properties.Get(testID).String()
				require.Nil(t, err)

				if targetRel, err := tx.Relationships().Filterf(func() graph.Criteria {
					return query.Equals(query.RelationshipProperty(testID), id)
				}).First(); err != nil {
					t.Fatalf("Could not find migrated relationship with '%s' == %s", testID, id)
				} else {
					require.Equal(t, sourceEdge.Kind, targetRel.Kind)
				}
			}

			// check kinds
			targetKinds, err := pg_query.On(tx).SelectKinds()
			require.Nil(t, err)

			for _, kind := range append(sourceNodeKinds, sourceEdgeKinds...) {
				require.NotNil(t, targetKinds[kind])
			}

			return nil
		})
		require.Nil(t, err)
	})
}

func setupTestMigrator(t *testing.T, ctx context.Context, graphDB graph.Database) *tools.PGMigrator {
	var (
		schema   = graphschema.DefaultGraphSchema()
		dbSwitch = graph.NewDatabaseSwitch(ctx, graphDB)
	)

	cfg, err := utils.LoadIntegrationTestConfig()
	require.Nil(t, err)

	return tools.NewPGMigrator(ctx, cfg, schema, dbSwitch)
}
