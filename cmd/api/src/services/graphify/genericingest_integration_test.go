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
	"time"

	"github.com/specterops/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func Test_ResolveRelationships(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		NAME_NOT_EXISTS        = "bippity boppity"       // simulates a name that does not exist
		NAME_MULTIPLE_MATCH    = "same name"             // simulates a name that will be matched by multiple nodes
		NAME_RESOLVED_BY_KINDS = "namey name kindy kind" // simulates a name that will be matched by multiple nodes but will resolve by optional kind filter
	)

	t.Run("Exact match (happy path). Source and target node names both resolve unambiguously to nodes with objectids.", func(t *testing.T) {
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Only target matches.	Source node is unmatched by name - update should be skipped.", func(t *testing.T) {
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Neither matches. No node resolves to source or target — update should be skipped.", func(t *testing.T) {
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Multiple matches for source — ambiguity, update should be skipped.", func(t *testing.T) {
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Multiple matches for target — ambiguity, update should be skipped.", func(t *testing.T) {
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Kind filters for endpoints are nil. Resolved by name only", func(t *testing.T) {
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Empty or null Source or Target values — update should be skipped.", func(t *testing.T) {
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
					require.ErrorContains(t, err, "skipping invalid relationship")
					require.Empty(t, updates)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("ID match fallback. Both source/target use MatchByID.", func(t *testing.T) {
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, []ein.IngestibleRelationship{ingestibleRel}, graph.EmptyKind)
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
					timestampedBatch := NewTimestampedBatch(batch, time.Now().UTC())
					updates, err := resolveRelationships(timestampedBatch, rels, graph.EmptyKind)
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
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	generateKey := func(name, kind string) endpointKey {
		return endpointKey{
			Name: name,
			Kind: kind,
		}
	}
	t.Run("Single match. One node with name and kind found, and valid objectid returned.", func(t *testing.T) {
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
