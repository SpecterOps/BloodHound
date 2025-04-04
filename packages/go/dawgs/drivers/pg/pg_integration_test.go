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

//go:build integration
// +build integration

package pg_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
)

func Test_ResetDB(t *testing.T) {
	// We don't need the reference to the DB but this will ensure that the canonical DB wipe method is called
	integration.SetupDB(t)

	ctx, done := context.WithCancel(context.Background())
	defer done()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.Nil(t, err)

	pool, err := pg.NewPool(cfg.Database.PostgreSQLConnectionString())
	require.Nil(t, err)

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		ConnectionString: cfg.Database.PostgreSQLConnectionString(),
		Pool:             pool,
	})
	require.Nil(t, err)

	require.Nil(t, graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema()))

	require.Nil(t, graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		user1, _ := tx.CreateNode(graph.AsProperties(map[string]any{
			"name": "user 1",
		}), ad.User)

		user2, _ := tx.CreateNode(graph.AsProperties(map[string]any{
			"name": "user 2",
		}), ad.User)

		group1, _ := tx.CreateNode(graph.AsProperties(map[string]any{
			"name": "group 1",
		}), ad.Group)

		computer1, _ := tx.CreateNode(graph.AsProperties(map[string]any{
			"name": "computer 1",
		}), ad.Computer)

		tx.CreateRelationshipByIDs(user1.ID, group1.ID, ad.MemberOf, nil)
		tx.CreateRelationshipByIDs(group1.ID, computer1.ID, ad.GenericAll, nil)
		tx.CreateRelationshipByIDs(computer1.ID, user2.ID, ad.HasSession, nil)

		return nil
	}))
}

func TestPG(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.Nil(t, err)

	pool, err := pg.NewPool(cfg.Database.PostgreSQLConnectionString())
	require.Nil(t, err)

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		ConnectionString: cfg.Database.PostgreSQLConnectionString(),
		Pool:             pool,
	})
	require.Nil(t, err)

	test.RequireNilErr(t, graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema()))
	test.RequireNilErr(t, graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Query("match p = (s:User)-[*..]->(:Computer) return p", nil).Error()
	}))
}

// TestInt64SequenceIDs is a test that validates against a regression in any logic that originally used graph IDs.
// Before this test and its related changes, graph IDs were stored as uint32 values. Scale of graphs necessitated
// a change to an int64 ID space, however, the nuance of this refactor revealed several bugs and assumptions in
// query formatting as well as business logic.
//
// This test represents validating that an instance can insert a node or an edge that has an ID greater than the maximum
// value of an int32 graph ID.
func TestInt64SequenceIDs(t *testing.T) {
	var (
		// We don't need the reference to the DB but this will ensure that the canonical DB wipe method is called
		_                = integration.SetupDB(t)
		graphTestContext = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	)

	if pg.IsPostgreSQLGraph(graphTestContext.Graph.Database) {
		// Move the ID space out past an int32's max value
		require.Nil(t, graphTestContext.Graph.Database.Run(context.Background(), "alter sequence node_id_seq restart with 2147483648;", make(map[string]any)))
		require.Nil(t, graphTestContext.Graph.Database.Run(context.Background(), "alter sequence edge_id_seq restart with 2147483648;", make(map[string]any)))
	}

	var (
		// Create two users, chuck and steve where chuck has GenericAll on steve
		chuckNode = graphTestContext.NewNode(graph.AsProperties(map[string]any{"name": "chuck"}), ad.User, ad.Entity)
		steveNode = graphTestContext.NewNode(graph.AsProperties(map[string]any{"name": "steve"}), ad.User, ad.Entity)
		_         = graphTestContext.NewRelationship(chuckNode, steveNode, ad.GenericAll)
	)
}
