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

package migrations_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func TestVersion_730_Migration(t *testing.T) {
	t.Run("Migration_v730 Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.Version730_Migration.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			err := migrations.Version_730_Migration(context.Background(), db)
			require.Nil(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				computers, err := ops.FetchNodes(tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)))

				require.Nil(t, err)

				for _, computer := range computers {
					if computer.ID == harness.Version730_Migration.Computer1.ID {
						smbSigning, err := computer.Properties.Get(ad.SMBSigning.String()).Bool()
						require.Nil(t, err)
						require.True(t, smbSigning)
					} else {
						_, err := computer.Properties.Get(ad.SMBSigning.String()).Bool()
						require.Error(t, err)
						require.True(t, errors.Is(err, graph.ErrPropertyNotFound))
					}
				}

				return nil
			})
		})
	})
}

func TestVersion_900_Migration(t *testing.T) {
	t.Run("Migration_v9.0.0 Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.Version900_Migration_Harness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			err := migrations.Version_900_Migration(context.Background(), db)
			require.Nil(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				computers, err := ops.FetchNodes(tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)))

				require.Nil(t, err)

				for _, computer := range computers {
					environmentID, err := computer.Properties.Get(graphschema.EnvironmentIDKey).String()
					require.Nil(t, err)
					require.Equal(t, "ENV-001", environmentID)

					deletedProperty, err := computer.Properties.Get("environment_id").String()
					require.Error(t, err)
					require.True(t, errors.Is(err, graph.ErrPropertyNotFound))
					require.Empty(t, deletedProperty)
				}

				return nil
			})
		})
	})
}

func TestVersion_910_Migration(t *testing.T) {
	t.Run("Migration_v9.1.0 Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.Version910_Migration_Harness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			err := migrations.Version_910_Migration(context.Background(), db)
			require.Nil(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				nodes, err := ops.FetchNodes(tx.Nodes())
				require.Nil(t, err)

				for _, node := range nodes {
					switch node.ID {
					case harness.Version910_Migration_Harness.ADNode.ID:
						require.True(t, node.Kinds.ContainsOneOf(ad.Group))
					case harness.Version910_Migration_Harness.AZNode.ID:
						require.False(t, node.Kinds.ContainsOneOf(ad.Group))
					case harness.Version910_Migration_Harness.OGNode.ID:
						require.False(t, node.Kinds.ContainsOneOf(ad.Group))
					}
				}
				return nil
			})
		})
	})
}

func TestVersion_930_Migration(t *testing.T) {
	t.Run("backfills custom_node_kinds for a schemaless node kind that isn't in custom_node_kinds or schema_node_kinds", func(t *testing.T) {
		suite := setupIntegrationTestSuite(t)
		t.Cleanup(func() { suite.teardownIntegrationTestSuite(t) })

		schemalessNode := &graph.Node{
			Kinds: graph.Kinds{graph.StringKind("SchemalessKind")},
			Properties: graph.AsProperties(graph.PropertyMap{
				common.Name:     "SchemalessNode",
				common.ObjectID: "SchemalessNode",
			}),
		}
		suite.createNodes(t, schemalessNode)

		err := migrations.Version_930_Migration(suite.bhDatabase)(suite.context, suite.graphDB)
		require.NoError(t, err)

		customNodeKinds, err := suite.bhDatabase.GetCustomNodeKinds(suite.context, nil)
		require.NoError(t, err)

		var kindNames []string
		for _, customNodeKind := range customNodeKinds {
			kindNames = append(kindNames, customNodeKind.KindName)
		}
		require.Contains(t, kindNames, "SchemalessKind")
	})

	t.Run("does not backfill a kind already present in custom_node_kinds", func(t *testing.T) {
		suite := setupIntegrationTestSuite(t)
		t.Cleanup(func() { suite.teardownIntegrationTestSuite(t) })

		_, err := suite.bhDatabase.CreateCustomNodeKinds(suite.context, model.CustomNodeKinds{
			{
				KindName: "PreExistingCustomNodeKind",
				Config: model.CustomNodeKindConfig{
					Icon: graphschema.DisplayNodeIcon{
						Type:  graphschema.DisplayNodeTypeFontAwesome,
						Color: "#FF0000",
					},
				},
			},
		})
		require.NoError(t, err)

		preExistingNode := &graph.Node{
			Kinds: graph.Kinds{graph.StringKind("PreExistingCustomNodeKind")},
			Properties: graph.AsProperties(graph.PropertyMap{
				common.Name:     "PreExistingNode",
				common.ObjectID: "PreExistingNode",
			}),
		}
		suite.createNodes(t, preExistingNode)

		err = migrations.Version_930_Migration(suite.bhDatabase)(suite.context, suite.graphDB)
		require.NoError(t, err)

		customNodeKinds, err := suite.bhDatabase.GetCustomNodeKinds(suite.context, nil)
		require.NoError(t, err)

		var preExistingCount int
		for _, customNodeKind := range customNodeKinds {
			if customNodeKind.KindName == "PreExistingCustomNodeKind" {
				preExistingCount++
			}
		}
		require.Equal(t, 1, preExistingCount, "PreExistingCustomNodeKind must appear exactly once in custom_node_kinds")
	})

	t.Run("does not backfill a kind already present in schema_node_kinds", func(t *testing.T) {
		suite := setupIntegrationTestSuite(t)
		t.Cleanup(func() { suite.teardownIntegrationTestSuite(t) })

		schemaExtension, err := suite.bhDatabase.CreateGraphSchemaExtension(suite.context, "test-ext", "Test Extension", "1.0.0", "test")
		require.NoError(t, err)

		_, err = suite.bhDatabase.CreateGraphSchemaNodeKind(suite.context, "SchemaNodeKind", schemaExtension.ID, "Schema Kind", "", false, "fa-circle", "#FFFFFF")
		require.NoError(t, err)

		// RefreshKinds so that dawgs can resolve the new kind ID when creating the graph node.
		require.NoError(t, suite.graphDB.RefreshKinds(suite.context))

		schemaNode := &graph.Node{
			Kinds: graph.Kinds{graph.StringKind("SchemaNodeKind")},
			Properties: graph.AsProperties(graph.PropertyMap{
				common.Name:     "SchemaNode",
				common.ObjectID: "SchemaNode",
			}),
		}
		suite.createNodes(t, schemaNode)

		err = migrations.Version_930_Migration(suite.bhDatabase)(suite.context, suite.graphDB)
		require.NoError(t, err)

		customNodeKinds, err := suite.bhDatabase.GetCustomNodeKinds(suite.context, nil)
		require.NoError(t, err)

		for _, customNodeKind := range customNodeKinds {
			require.NotEqual(t, "SchemaNodeKind", customNodeKind.KindName, "schema_node_kinds entries must not be duplicated in custom_node_kinds")
		}
	})

	t.Run("does not backfill a kind that has no nodes in the graph", func(t *testing.T) {
		suite := setupIntegrationTestSuite(t)
		t.Cleanup(func() { suite.teardownIntegrationTestSuite(t) })

		// Create a node to register the kind in the kinds table, then immediately delete it.
		ephemeralNode := &graph.Node{
			Kinds: graph.Kinds{graph.StringKind("EphemeralKind")},
			Properties: graph.AsProperties(graph.PropertyMap{
				common.Name:     "EphemeralNode",
				common.ObjectID: "EphemeralNode",
			}),
		}
		suite.createNodes(t, ephemeralNode)

		err := suite.graphDB.BatchOperation(suite.context, func(batch graph.Batch) error {
			return batch.DeleteNode(ephemeralNode.ID)
		})
		require.NoError(t, err)

		err = migrations.Version_930_Migration(suite.bhDatabase)(suite.context, suite.graphDB)
		require.NoError(t, err)

		customNodeKinds, err := suite.bhDatabase.GetCustomNodeKinds(suite.context, nil)
		require.NoError(t, err)

		for _, customNodeKind := range customNodeKinds {
			require.NotEqual(t, "EphemeralKind", customNodeKind.KindName, "kinds with no nodes in the graph must not be backfilled")
		}
	})

	t.Run("returns an error when backfillData is nil", func(t *testing.T) {
		suite := setupIntegrationTestSuite(t)
		t.Cleanup(func() { suite.teardownIntegrationTestSuite(t) })

		err := migrations.Version_930_Migration(nil)(suite.context, suite.graphDB)
		require.Error(t, err)
	})
}
