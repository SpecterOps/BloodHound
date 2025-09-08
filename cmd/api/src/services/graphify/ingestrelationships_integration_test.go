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

package graphify

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

// verify that nodes and edges are created or updated based on existing graph state
func Test_IngestRelationships(t *testing.T) {
	t.Run("Create rel by exact name match. Source and target node names both resolve to nodes with objectids.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

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
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					err := IngestRelationships(ingestContext, graph.EmptyKind, rels)
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
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

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
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					err := IngestRelationships(ingestContext, graph.EmptyKind, rels)
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
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

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
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					err := IngestRelationships(ingestContext, graph.EmptyKind, rels)
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
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

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
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					err := IngestRelationships(ingestContext, graph.EmptyKind, rels)
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
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

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
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					err := IngestRelationships(ingestContext, graph.EmptyKind, rels)
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
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

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
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					err := IngestRelationships(ingestContext, graph.EmptyKind, rels)
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

func Test_ResolveRelationships(t *testing.T) {
	var (
		NAME_NOT_EXISTS        = "bippity boppity"       // simulates a name that does not exist
		NAME_MULTIPLE_MATCH    = "same name"             // simulates a name that will be matched by multiple nodes
		NAME_RESOLVED_BY_KINDS = "namey name kindy kind" // simulates a name that will be matched by multiple nodes but will resolve by optional kind filter
	)

	t.Run("Exact match (happy path). Source and target node names both resolve unambiguously to nodes with objectids.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "name a", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "name b", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))
					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.Nil(t, err)
					require.Len(t, updates, 1)

					startNode, endNode := updates[0].Start, updates[0].End

					startObjectID, _ := startNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, startObjectID)
					require.True(t, startNode.Kinds.ContainsOneOf(ad.Computer))

					endObjectID, _ := endNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, endObjectID)
					require.True(t, endNode.Kinds.ContainsOneOf(ad.Computer))

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Only source matches. Target node is unmatched by name - update should be skipped.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "name a", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: NAME_NOT_EXISTS, MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Only target matches.	Source node is unmatched by name - update should be skipped.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: NAME_NOT_EXISTS, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "name b", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Neither matches. No node resolves to source or target — update should be skipped.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: NAME_NOT_EXISTS, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: NAME_NOT_EXISTS, MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Multiple matches for source — ambiguity, update should be skipped.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: NAME_MULTIPLE_MATCH, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "name b", MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Multiple matches for target — ambiguity, update should be skipped.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "name a", MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: NAME_MULTIPLE_MATCH, MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Kind filters for endpoints are nil. Resolved by name only", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "bob", Kind: graph.StringKind("KindA"), MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "bobby", Kind: graph.StringKind("KindB"), MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.Nil(t, err)
					require.Len(t, updates, 1)

					startNode, endNode := updates[0].Start, updates[0].End

					startObjectID, _ := startNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, startObjectID)
					require.True(t, startNode.Kinds.ContainsOneOf(graph.StringKind("KindA")))

					endObjectID, _ := endNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, endObjectID)
					require.True(t, endNode.Kinds.ContainsOneOf(graph.StringKind("KindB")))

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Mixed resolution strategy. Source uses MatchByName, Target uses MatchByID.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "bob", Kind: graph.StringKind("KindA"), MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "5678"},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.Nil(t, err)
					require.Len(t, updates, 1)

					startNode, endNode := updates[0].Start, updates[0].End

					startObjectID, _ := startNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, startObjectID)
					require.True(t, startNode.Kinds.ContainsOneOf(graph.StringKind("KindA")))

					endObjectID, _ := endNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, endObjectID)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Mixed resolution strategy. Target uses MatchByName, Source uses MatchByID.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "1234"},
					ein.IngestibleEndpoint{Value: "bobby", MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.Nil(t, err)
					require.Len(t, updates, 1)

					startNode, endNode := updates[0].Start, updates[0].End

					startObjectID, _ := startNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, startObjectID)

					endObjectID, _ := endNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, endObjectID)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Source name matches 2 nodes with the same name but different kinds. Resolution should honor optional kind filter.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: NAME_RESOLVED_BY_KINDS, MatchBy: ein.MatchByName, Kind: ad.User},
					ein.IngestibleEndpoint{Value: "name b", MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.Nil(t, err)
					require.Len(t, updates, 1)

					startNode, endNode := updates[0].Start, updates[0].End

					startOID, _ := startNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, startOID)

					endOID, _ := endNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, endOID)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Both nodes match but with mismatched kinds. Filtered out due to kind mismatch — update should be skipped.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "name a", MatchBy: ein.MatchByName, Kind: ad.User},
					ein.IngestibleEndpoint{Value: "name b", MatchBy: ein.MatchByName, Kind: ad.User},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Empty or null Source or Target values — update should be skipped.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("ID match fallback. Both source/target use MatchByID.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				ingestibleRel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "1234"},
					ein.IngestibleEndpoint{Value: "5678"},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.Nil(t, err)
					require.Len(t, updates, 1)

					startNode, endNode := updates[0].Start, updates[0].End

					startOID, _ := startNode.Properties.Get(string(common.ObjectID)).String()
					require.Equal(t, "1234", startOID)

					endOID, _ := endNode.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, "5678", endOID)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Batch input, multiple updates returned.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				harness.GenericIngest.Setup(testContext)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				rels := []ein.IngestibleRelationship{
					{
						Source: ein.IngestibleEndpoint{
							Value:   "bob",
							MatchBy: ein.MatchByName,
							Kind:    graph.StringKind("KindA"),
						},
						Target: ein.IngestibleEndpoint{
							Value:   "server-01",
							MatchBy: ein.MatchByName,
							Kind:    graph.StringKind("Device"),
						},
						RelType: graph.StringKind("AdminTo"),
						RelProps: map[string]any{
							"hello": "world",
						},
					},
					{
						Source: ein.IngestibleEndpoint{
							Value:   "bobby",
							MatchBy: ein.MatchByName,
							Kind:    graph.StringKind("KindB"),
						},
						Target: ein.IngestibleEndpoint{
							Value:   "dc-01",
							MatchBy: ein.MatchByName,
							Kind:    graph.StringKind("DomainController"),
						},
						RelType: graph.StringKind("HasSession"),
						RelProps: map[string]any{
							"isItWednesday": true,
						},
					},
				}

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					ingestContext := NewIngestContext(testContext.Context(), WithBatchUpdater(batch))

					updates, err := resolveRelationships(ingestContext, rels, graph.EmptyKind)
					require.Nil(t, err)
					require.Len(t, updates, 2)

					update1, update2 := updates[0], updates[1]

					startOID, _ := update1.Start.Properties.Get(string(common.ObjectID)).String()
					require.Equal(t, "1234", startOID)
					endOID, _ := update1.Start.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, "0001", endOID)
					require.True(t, update1.Relationship.Kind.Is(graph.StringKind("AdminTo")))
					relProp1, _ := update1.Relationship.Properties.Get("hello").String()
					require.Equal(t, "world", relProp1)

					startOID, _ = update2.Start.Properties.Get(string(common.ObjectID)).String()
					require.Equal(t, "5678", startOID)
					endOID, _ = update2.Start.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, "0002", endOID)
					require.True(t, update2.Relationship.Kind.Is(graph.StringKind("HasSession")))
					relProp2, _ := update2.Relationship.Properties.Get("isItWednesday").Bool()
					require.Equal(t, true, relProp2)

					return nil
				})

				require.Nil(t, err)

			})
	})
}

func Test_ResolveAllEndpointsByName(t *testing.T) {
	generateKey := func(name, kind string) endpointKey {
		return endpointKey{
			Name: name,
			Kind: kind,
		}
	}
	t.Run("Single match. One node with name and kind found, and valid objectid returned.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.ResolveEndpointsByName.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {

			err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
				rel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "alice", Kind: ad.User, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{},
					ein.IngestibleRel{},
				)

				rels := []ein.IngestibleRelationship{rel} // simulate a "batch"

				cache, err := resolveAllEndpointsByName(batch, rels)
				require.Nil(t, err)
				require.Len(t, cache, 3) // cache has keys for 'User' and 'Base' and ""

				key := generateKey("ALICE", "User")
				require.Contains(t, cache, key)
				require.NotEmpty(t, cache[key])

				return nil
			})

			require.Nil(t, err)
		})
	})

	t.Run("No match. Lookup requests name/kind that do not exist in DB.	Empty result map returned.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.ResolveEndpointsByName.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {

			err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
				rel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "not alice", Kind: ad.User, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{},
					ein.IngestibleRel{},
				)

				rels := []ein.IngestibleRelationship{rel} // simulate a "batch"

				cache, err := resolveAllEndpointsByName(batch, rels)
				require.Nil(t, err)
				require.Len(t, cache, 0)

				return nil
			})

			require.Nil(t, err)
		})
	})

	t.Run("Ambiguous match.	Two nodes with same name + kind. Skipped from result map.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.ResolveEndpointsByName.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {

			err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
				rel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "SAME NAME", Kind: ad.Computer, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{},
					ein.IngestibleRel{},
				)

				rels := []ein.IngestibleRelationship{rel} // simulate a "batch"

				cache, err := resolveAllEndpointsByName(batch, rels)
				require.Nil(t, err)
				require.Len(t, cache, 0)

				return nil
			})

			require.Nil(t, err)
		})
	})

	t.Run("Multiple distinct matches for Alice n Bob. Both returned", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.ResolveEndpointsByName.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {

			err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
				rel := ein.NewIngestibleRelationship(
					ein.IngestibleEndpoint{Value: "alice", Kind: ad.User, MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "bob", Kind: graph.StringKind("GenericDevice"), MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				rels := []ein.IngestibleRelationship{rel} // simulate a "batch"

				cache, err := resolveAllEndpointsByName(batch, rels)
				require.Nil(t, err)
				require.Len(t, cache, 5) // Alice node has keys for 'User' and 'Base' and "". Bob just has GenericBase and ""

				aliceKey := generateKey("ALICE", "User")
				require.Contains(t, cache, aliceKey)
				require.NotEmpty(t, cache[aliceKey])

				bobKey := generateKey("BOB", "GenericDevice")
				require.Contains(t, cache, bobKey)
				require.NotEmpty(t, cache[bobKey])

				return nil
			})

			require.Nil(t, err)
		})
	})

	t.Run("Empty input.	Empty result map returned.", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.ResolveEndpointsByName.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {

			err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
				rels := []ein.IngestibleRelationship{} // simulate a "batch"

				cache, err := resolveAllEndpointsByName(batch, rels)
				require.Nil(t, err)
				require.Len(t, cache, 0)

				return nil
			})

			require.Nil(t, err)
		})
	})
}
