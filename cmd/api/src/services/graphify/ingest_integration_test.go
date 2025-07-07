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

//go:build serial_integration
// +build serial_integration

package graphify_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

// verify that nodes and edges are created or updated based on existing graph state
func Test_IngestRelationships(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	t.Run("Create rel by exact name match. Source and target node names both resolve to nodes with objectids.", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.IngestRelationships.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "computer a", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "computer b", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleRel{RelType: graph.StringKind("related_to")},
				)
				rels := []ein.IngestibleRelationship{ingestibleRel}

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())
					err := graphify.IngestRelationships(timestampedBatch, graph.EmptyKind, rels)
					require.Nil(t, err)
					return nil
				})

				require.Nil(t, err)

				// verify an edge was created
				err = db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
					count, err := tx.Relationships().Filter(
						query.And(
							query.Equals(query.StartID(), harness.IngestRelationships.Node1.ID),
							query.Equals(query.EndID(), harness.IngestRelationships.Node2.ID),
							query.Kind(query.Relationship(), graph.StringKind("related_to")),
						),
					).Count()

					require.Equal(t, int64(1), count)
					require.Nil(t, err)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Update rel by exact name match. Source and target node names both resolve to nodes with objectids.", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.IngestRelationships.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "computer a", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "computer b", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleRel{RelType: graph.StringKind("existing_edge_kind"), RelProps: map[string]any{"hello": "world"}},
				)
				rels := []ein.IngestibleRelationship{ingestibleRel}

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())
					err := graphify.IngestRelationships(timestampedBatch, graph.EmptyKind, rels)
					require.Nil(t, err)
					return nil
				})

				require.Nil(t, err)

				// verify an edge was updated
				err = db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
					rels := []*graph.Relationship{}
					err := tx.Relationships().Filter(
						query.And(
							query.Equals(query.StartID(), harness.IngestRelationships.Node1.ID),
							query.Equals(query.EndID(), harness.IngestRelationships.Node2.ID),
							query.Kind(query.Relationship(), graph.StringKind("existing_edge_kind")),
						),
					).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
						for rel := range cursor.Chan() {
							rels = append(rels, rel)
						}

						return nil
					})
					require.Nil(t, err)
					// was there only one exact match?
					require.Len(t, rels, 1)
					// were properties merged?
					property, _ := rels[0].Properties.Get("hello").String()
					require.Equal(t, "world", property)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Create rel using source/target nodes that specify objectid.", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.IngestRelationships.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "1234"},
					ein.IngestibleEndpoint{Value: "5678"},
					ein.IngestibleRel{RelType: graph.StringKind("related_to")},
				)
				rels := []ein.IngestibleRelationship{ingestibleRel}

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())
					err := graphify.IngestRelationships(timestampedBatch, graph.EmptyKind, rels)
					require.Nil(t, err)
					return nil
				})

				require.Nil(t, err)

				// verify an edge was created
				err = db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
					count, err := tx.Relationships().Filter(
						query.And(
							query.Equals(query.StartID(), harness.IngestRelationships.Node1.ID),
							query.Equals(query.EndID(), harness.IngestRelationships.Node2.ID),
							query.Kind(query.Relationship(), graph.StringKind("related_to")),
						),
					).Count()

					require.Equal(t, int64(1), count)
					require.Nil(t, err)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Update rel using source/target nodes that specify objectid.", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.IngestRelationships.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "1234"},
					ein.IngestibleEndpoint{Value: "5678"},
					ein.IngestibleRel{RelType: graph.StringKind("existing_edge_kind"), RelProps: map[string]any{"hello": "world"}},
				)
				rels := []ein.IngestibleRelationship{ingestibleRel}

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())
					err := graphify.IngestRelationships(timestampedBatch, graph.EmptyKind, rels)
					require.Nil(t, err)
					return nil
				})

				require.Nil(t, err)

				// verify an edge was updated
				err = db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
					rels := []*graph.Relationship{}
					err := tx.Relationships().Filter(
						query.And(
							query.Equals(query.StartID(), harness.IngestRelationships.Node1.ID),
							query.Equals(query.EndID(), harness.IngestRelationships.Node2.ID),
							query.Kind(query.Relationship(), graph.StringKind("existing_edge_kind")),
						),
					).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
						for rel := range cursor.Chan() {
							rels = append(rels, rel)
						}

						return nil
					})
					require.Nil(t, err)
					// was there only one exact match?
					require.Len(t, rels, 1)
					// were properties merged?
					property, _ := rels[0].Properties.Get("hello").String()
					require.Equal(t, "world", property)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Create rel. Source/target nodes' objectid's dont exist. Both nodes get created and rel gets created.", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.IngestRelationships.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "0001"},
					ein.IngestibleEndpoint{Value: "0002"},
					ein.IngestibleRel{RelType: graph.StringKind("related_to")},
				)
				rels := []ein.IngestibleRelationship{ingestibleRel}

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())
					err := graphify.IngestRelationships(timestampedBatch, graph.EmptyKind, rels)
					require.Nil(t, err)
					return nil
				})

				require.Nil(t, err)

				err = db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
					nodeIDs := map[string]graph.ID{}
					// verify start and end nodes were created
					_ = tx.Nodes().Filter(
						query.Or(
							query.Equals(query.Property(query.Node(), "objectid"), "0001"),
							query.Equals(query.Property(query.Node(), "objectid"), "0002"),
						)).
						OrderBy(query.Order(query.Property(query.Node(), "objectid"), query.Ascending())).Fetch(
						func(cursor graph.Cursor[*graph.Node]) error {
							for node := range cursor.Chan() {
								objectid, _ := node.Properties.Get("objectid").String()
								nodeIDs[objectid] = node.ID
							}
							return nil
						},
					)

					require.Len(t, nodeIDs, 2)

					// verify rel created
					count, err := tx.Relationships().Filter(
						query.And(
							query.Equals(query.StartID(), nodeIDs["0001"]),
							query.Equals(query.EndID(), nodeIDs["0002"]),
							query.Kind(query.Relationship(), graph.StringKind("related_to")),
						),
					).Count()

					require.Equal(t, int64(1), count)
					require.Nil(t, err)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Dont create rel. Source/target nodes' have names that don't resolve to objectids. Neither node gets created and rel creation is skipped.", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.IngestRelationships.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "bubba", MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "lubba", MatchBy: ein.MatchByName},
					ein.IngestibleRel{RelType: graph.StringKind("related_to")},
				)
				rels := []ein.IngestibleRelationship{ingestibleRel}

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())
					err := graphify.IngestRelationships(timestampedBatch, graph.EmptyKind, rels)
					require.ErrorContains(t, err, "skipping invalid relationship")
					return nil
				})

				require.Nil(t, err)

				err = db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
					// verify start and end nodes were not created
					numNodes, _ := tx.Nodes().Filter(query.Or(
						query.Equals(query.Property(query.Node(), "name"), "bubba"),
						query.Equals(query.Property(query.Node(), "name"), "lubba"))).Count()
					require.Zero(t, numNodes)

					// verify rel not created
					numRels, _ := tx.Relationships().Filter(query.Kind(query.Relationship(), graph.StringKind("related_to"))).Count()
					require.Zero(t, numRels)

					return nil
				})

				require.Nil(t, err)

			})
	})
}

// verifies that files that bypassed validation controls due to being uploaded as zips receive validation attention in the datapipe,
// and that invalid files are not ingested into the graph
func Test_ReadFileForIngest(t *testing.T) {

	var (
		testContext     = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		ingestSchema, _ = upload.LoadIngestSchema()
		validReader     = bytes.NewReader([]byte(`{"graph":{"nodes":[{"id": "1234", "kinds": ["kindA","kindB"],"properties":{"true": true,"hello":"world"}}]}}`))
		// invalidReader simulates reading a file that doesn't pass jsonschema validation against the nodes schema.
		// ReadFileForIngest() should kick out, ingesting no graph data
		invalidReader = bytes.NewReader([]byte(`{"graph":{"nodes": [{"id":1234}]}}`))
		readOptions   = graphify.ReadOptions{
			IngestSchema: ingestSchema,
			FileType:     model.FileTypeZip,
			ADCSEnabled:  false,
		}
	)

	t.Run("happy path. a file uploaded as a zip passes validation and is written to the graph", func(t *testing.T) {
		testContext.BatchTest(func(harness integration.HarnessDetails, batch graph.Batch) {
			timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())
			err := graphify.ReadFileForIngest(timestampedBatch, validReader, readOptions)
			require.Nil(t, err)

		}, func(details integration.HarnessDetails, tx graph.Transaction) {

			err := tx.Nodes().
				Filter(query.Equals(query.Property(query.Node(), "objectid"), "1234")).
				Fetch(func(cursor graph.Cursor[*graph.Node]) error {
					numNodes := 0
					for node := range cursor.Chan() {
						// assert kinds were added correctly
						require.Contains(t, node.Kinds, graph.StringKind("kindA"))
						require.Contains(t, node.Kinds, graph.StringKind("kindB"))

						// assert properties were saved correctly
						booleanProperty, _ := node.Properties.Get("true").Bool()
						require.Equal(t, true, booleanProperty)
						stringProperty, _ := node.Properties.Get("hello").String()
						require.Equal(t, "world", stringProperty)

						numNodes++
					}

					// assert 1 node was ingested
					require.Equal(t, 1, numNodes)
					return nil
				})

			require.Nil(t, err)
		})
	})

	t.Run("failure path. a file uploaded as a zip fails validation and nothing is written to the graph", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error { return nil }, func(harness integration.HarnessDetails, db graph.Database) {
			_ = db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
				timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())
				err := graphify.ReadFileForIngest(timestampedBatch, invalidReader, readOptions)
				require.NotNil(t, err)
				var report upload.ValidationReport
				if errors.As(err, &report) {
					// verify nodes[0] caused a validation error
					require.Len(t, report.ValidationErrors, 1)
				}
				return nil
			})

			// assert that zero nodes exist
			_ = db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
				numNodes, err := tx.Nodes().Count()
				require.Nil(t, err)
				require.Equal(t, int64(0), numNodes)
				return nil
			})

		})
	})
}
