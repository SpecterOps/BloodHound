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

package datapipe_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEinNodeProperties(t *testing.T) {
	var (
		nowUTC     = time.Now().UTC()
		objectID   = "objectid"
		properties = map[string]any{
			datapipe.ReconcileProperty:      false,
			common.Name.String():            "name",
			common.OperatingSystem.String(): "temple",
			ad.DistinguishedName.String():   "distinguished-name",
		}
		normalizedProperties = datapipe.NormalizeEinNodeProperties(properties, objectID, nowUTC)
	)

	assert.Nil(t, normalizedProperties[datapipe.ReconcileProperty])
	assert.NotNil(t, normalizedProperties[common.LastSeen.String()])
	assert.Equal(t, "OBJECTID", normalizedProperties[common.ObjectID.String()])
	assert.Equal(t, "NAME", normalizedProperties[common.Name.String()])
	assert.Equal(t, "DISTINGUISHED-NAME", normalizedProperties[ad.DistinguishedName.String()])
	assert.Equal(t, "TEMPLE", normalizedProperties[common.OperatingSystem.String()])
}

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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
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
					update, err := datapipe.ResolveRelationshipByName(batch, ingestibleRel)
					require.Nil(t, err)
					require.Empty(t, update)

					return nil
				})

				require.Nil(t, err)

			})
	})
}
