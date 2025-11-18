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
//go:build slow_integration

package graphify_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

// testNodeData represents a test node for ingestion
type testNodeData struct {
	ObjectID   string
	SourceKind graph.Kind
	Labels     graph.Kinds
	Properties map[string]any
}

// createTestNode creates a standard test node
func createTestNode() testNodeData {
	return testNodeData{
		ObjectID:   "ABC123",
		SourceKind: graph.StringKind("hellobase"),
		Labels:     graph.Kinds{graph.StringKind("labelA"), graph.StringKind("labelB")},
		Properties: map[string]any{
			"hello": "world",
			"1":     2,
		},
	}
}

// ingestTestNode performs node ingestion with optional changelog
func ingestTestNode(t *testing.T, suite IntegrationTestSuite, nodeData testNodeData, withChangelog bool) {
	t.Helper()

	ctx := suite.Context
	propertyBag := graph.NewProperties().SetAll(nodeData.Properties)

	err := suite.GraphDB.BatchOperation(ctx, func(batch graph.Batch) error {
		var ingestCtx *graphify.IngestContext

		if withChangelog {
			ingestCtx = graphify.NewIngestContext(ctx,
				graphify.WithBatchUpdater(batch),
				graphify.WithChangeManager(suite.Changelog))
		} else {
			ingestCtx = graphify.NewIngestContext(ctx,
				graphify.WithBatchUpdater(batch))
		}

		return graphify.IngestNode(ingestCtx, nodeData.SourceKind,
			ein.IngestibleNode{
				ObjectID:    nodeData.ObjectID,
				Labels:      nodeData.Labels,
				PropertyMap: propertyBag.Map,
			})
	})

	require.NoError(t, err)
}

// verifyNodeExists verifies that exactly one node exists with the specific ObjectID and expected properties
func verifyNodeExists(t *testing.T, suite IntegrationTestSuite, nodeData testNodeData) {
	t.Helper()

	ctx := suite.Context
	// Filter for the specific node by ObjectID
	nodeFilter := query.And(
		query.Not(query.Kind(query.Node(), common.MigrationData)),
		query.Equals(query.NodeProperty(common.ObjectID.String()), nodeData.ObjectID),
	)

	err := suite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		count, err := tx.Nodes().Filter(nodeFilter).Count()
		require.NoError(t, err)
		require.Equal(t, int64(1), count, "Expected exactly one node with ObjectID %s", nodeData.ObjectID)

		return tx.Nodes().Filter(nodeFilter).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for node := range cursor.Chan() {
				// Verify ObjectID
				objectid, _ := node.Properties.Get(common.ObjectID.String()).String()
				require.Equal(t, nodeData.ObjectID, objectid)

				// Verify LastSeen was added
				lastseen, _ := node.Properties.Get(common.LastSeen.String()).String()
				require.NotEmpty(t, lastseen)

				// Verify kinds (should include source kind + labels)
				expectedKinds := nodeData.Labels.Add(nodeData.SourceKind)
				for _, actualKind := range node.Kinds {
					require.Contains(t, expectedKinds, actualKind)
				}

				// Verify original properties are preserved
				for key, expectedValue := range nodeData.Properties {
					actualValue := node.Properties.Get(key)
					require.NotNil(t, actualValue, "Property %s should exist", key)
					require.EqualValues(t, expectedValue, actualValue.Any())
				}
			}
			return nil
		})
	})

	require.NoError(t, err)
}

func TestIngestNode(t *testing.T) {
	// Use the shared setup function from the main integration test file
	testSuite := setupIntegrationTestSuite(t, "fixtures")
	defer teardownIntegrationTestSuite(t, &testSuite)

	// Initialize changelog for tests that need it
	cl := changelog.NewChangelog(testSuite.GraphDB, testSuite.BHDatabase, changelog.DefaultOptions())
	cl.Start(testSuite.Context)
	testSuite.Changelog = cl
	defer cl.Stop(context.Background())

	t.Run("ingested node gets created", func(t *testing.T) {
		nodeData := createTestNode()
		ingestTestNode(t, testSuite, nodeData, false)
		verifyNodeExists(t, testSuite, nodeData)
	})

	t.Run("ingested node gets created with changelog on", func(t *testing.T) {
		nodeData := createTestNode()
		nodeData.ObjectID = "ABC124" // Use different ID to avoid conflicts
		ingestTestNode(t, testSuite, nodeData, true)
		verifyNodeExists(t, testSuite, nodeData)
	})

	t.Run("ingested node gets deduped with changelog on", func(t *testing.T) {
		nodeData := createTestNode()
		nodeData.ObjectID = "ABC125" // Use different ID to avoid conflicts

		// Ingest the same node twice with changelog enabled
		ctx := testSuite.Context
		propertyBag := graph.NewProperties().SetAll(nodeData.Properties)

		err := testSuite.GraphDB.BatchOperation(ctx, func(batch graph.Batch) error {
			ingestCtx := graphify.NewIngestContext(ctx,
				graphify.WithBatchUpdater(batch),
				graphify.WithChangeManager(testSuite.Changelog))

			node := ein.IngestibleNode{
				ObjectID:    nodeData.ObjectID,
				Labels:      nodeData.Labels,
				PropertyMap: propertyBag.Map,
			}

			// Ingest same node twice - second should be deduplicated
			err := graphify.IngestNode(ingestCtx, nodeData.SourceKind, node)
			require.NoError(t, err)
			err = graphify.IngestNode(ingestCtx, nodeData.SourceKind, node)
			require.NoError(t, err)

			return nil
		})

		require.NoError(t, err)

		// Verify only one node exists (deduplication worked)
		verifyNodeExists(t, testSuite, nodeData)
	})
}
