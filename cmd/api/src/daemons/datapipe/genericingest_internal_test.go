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

package datapipe

import (
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func Test_ResolveRelationshipByName(t *testing.T) {
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
					ein.IngestibleEndpoint{Value: "name a", MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: "name b", MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)

					startOID, _ := update.Start.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, startOID)

					endOID, _ := update.End.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, endOID)

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
					ein.IngestibleEndpoint{Value: "name a", MatchBy: ein.MatchByName},
					ein.IngestibleEndpoint{Value: NAME_NOT_EXISTS, MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)
					require.Empty(t, update)

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
					ein.IngestibleEndpoint{Value: "name b", MatchBy: ein.MatchByName},
					ein.IngestibleRel{},
				)

				err := db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)
					require.Empty(t, update)

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
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)
					require.Empty(t, update)

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
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)
					require.Empty(t, update)

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
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)
					require.Empty(t, update)

					return nil
				})

				require.Nil(t, err)

			})
	})

	t.Run("Nodes with the same name but different kinds. Resolution should honor optional kind filter.", func(t *testing.T) {
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
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)

					startOID, _ := update.Start.Properties.Get(string(common.ObjectID)).String()
					require.NotNil(t, startOID)

					endOID, _ := update.End.Properties.Get(string(common.ObjectID)).String()
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
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)
					require.Empty(t, update)

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
					update, err := resolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)
					require.Empty(t, update)

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
				require.Len(t, cache, 2) // cache has keys for 'User' and 'Base'

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
				require.Len(t, cache, 3) // Alice node has keys for 'User' and 'Base'. Bob just has GenericBase

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
