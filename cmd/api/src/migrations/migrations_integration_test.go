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
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
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

// newADNode returns an unsaved AD-kinded node with the given name and type kind,
// optionally polluted with extra kinds that the migration is expected to strip.
func newADNode(name string, typeKind graph.Kind, extraKinds ...graph.Kind) *graph.Node {
	return newPlatformNode(name, ad.Entity, typeKind, extraKinds...)
}

// newAzureNode returns an unsaved Azure-kinded node with the given name and type kind,
// optionally polluted with extra kinds that the migration is expected to strip.
func newAzureNode(name string, typeKind graph.Kind, extraKinds ...graph.Kind) *graph.Node {
	return newPlatformNode(name, azure.Entity, typeKind, extraKinds...)
}

func newPlatformNode(name string, baseKind, typeKind graph.Kind, extraKinds ...graph.Kind) *graph.Node {
	kinds := graph.Kinds{baseKind, typeKind}
	kinds = append(kinds, extraKinds...)
	return &graph.Node{
		Kinds: kinds,
		Properties: graph.AsProperties(graph.PropertyMap{
			common.Name:     name,
			common.ObjectID: name,
		}),
	}
}

func TestVersion_920_Migration(t *testing.T) {
	t.Run("strips non-ad.Entity source kinds from AD nodes", func(t *testing.T) {
		suite := setupIntegrationTestSuite(t)
		t.Cleanup(func() { suite.teardownIntegrationTestSuite(t) })

		var (
			adUserPolluted     = newADNode("PollutedUser", ad.User, azure.Entity, graph.StringKind("CustomKind"))
			adComputerPolluted = newADNode("PollutedComputer", ad.Computer, graph.StringKind("CustomKind"))
			adGroupClean       = newADNode("CleanGroup", ad.Group)
			azNodeUntouched    = &graph.Node{
				Kinds: graph.Kinds{azure.Entity, azure.User},
				Properties: graph.AsProperties(graph.PropertyMap{
					common.Name:     "AZUser",
					common.ObjectID: "AZUser",
				}),
			}
			ogNodeUntouched = &graph.Node{
				Kinds: graph.Kinds{graph.StringKind("CustomKind")},
				Properties: graph.AsProperties(graph.PropertyMap{
					common.Name:     "OGNode",
					common.ObjectID: "OGNode",
				}),
			}
		)

		suite.createNodes(t, adUserPolluted, adComputerPolluted, adGroupClean, azNodeUntouched, ogNodeUntouched)

		// ad.Entity and azure.Entity are auto-registered as source kinds by the
		// SQL migrations; only CustomKind needs to be registered explicitly.
		require.NoError(t, suite.bhDatabase.RegisterSourceKind(suite.context)(graph.StringKind("CustomKind")))

		err := migrations.Version_920_Migration(suite.bhDatabase)(suite.context, suite.graphDB)
		require.NoError(t, err)

		err = suite.graphDB.ReadTransaction(suite.context, func(tx graph.Transaction) error {
			adUser, err := ops.FetchNode(tx, adUserPolluted.ID)
			require.NoError(t, err)
			require.True(t, adUser.Kinds.ContainsOneOf(ad.Entity), "ad.Entity must be preserved")
			require.True(t, adUser.Kinds.ContainsOneOf(ad.User), "ad.User type kind must be preserved")
			require.False(t, adUser.Kinds.ContainsOneOf(azure.Entity), "azure.Entity must be stripped")
			require.False(t, adUser.Kinds.ContainsOneOf(graph.StringKind("CustomKind")), "CustomKind must be stripped")

			adComputer, err := ops.FetchNode(tx, adComputerPolluted.ID)
			require.NoError(t, err)
			require.True(t, adComputer.Kinds.ContainsOneOf(ad.Entity))
			require.True(t, adComputer.Kinds.ContainsOneOf(ad.Computer))
			require.False(t, adComputer.Kinds.ContainsOneOf(graph.StringKind("CustomKind")))

			adGroup, err := ops.FetchNode(tx, adGroupClean.ID)
			require.NoError(t, err)
			require.True(t, adGroup.Kinds.ContainsOneOf(ad.Entity))
			require.True(t, adGroup.Kinds.ContainsOneOf(ad.Group))

			azNode, err := ops.FetchNode(tx, azNodeUntouched.ID)
			require.NoError(t, err)
			require.True(t, azNode.Kinds.ContainsOneOf(azure.Entity), "azure.Entity must be preserved on clean azure nodes")
			require.True(t, azNode.Kinds.ContainsOneOf(azure.User), "azure type kinds must be preserved")

			ogNode, err := ops.FetchNode(tx, ogNodeUntouched.ID)
			require.NoError(t, err)
			require.True(t, ogNode.Kinds.ContainsOneOf(graph.StringKind("CustomKind")), "OpenGraph nodes must not be touched")

			return nil
		})
		require.NoError(t, err)
	})

	t.Run("strips non-azure.Entity source kinds from Azure nodes", func(t *testing.T) {
		suite := setupIntegrationTestSuite(t)
		t.Cleanup(func() { suite.teardownIntegrationTestSuite(t) })

		var (
			azUserPolluted  = newAzureNode("PollutedAZUser", azure.User, ad.Entity, graph.StringKind("CustomKind"))
			azGroupPolluted = newAzureNode("PollutedAZGroup", azure.Group, graph.StringKind("CustomKind"))
			azTenantClean   = newAzureNode("CleanTenant", azure.Tenant)
			adNodeUntouched = newADNode("ADUser", ad.User)
			ogNodeUntouched = &graph.Node{
				Kinds: graph.Kinds{graph.StringKind("CustomKind")},
				Properties: graph.AsProperties(graph.PropertyMap{
					common.Name:     "OGNode",
					common.ObjectID: "OGNode",
				}),
			}
		)

		suite.createNodes(t, azUserPolluted, azGroupPolluted, azTenantClean, adNodeUntouched, ogNodeUntouched)

		require.NoError(t, suite.bhDatabase.RegisterSourceKind(suite.context)(graph.StringKind("CustomKind")))

		suite.graphDB.ReadTransaction(suite.context, func(tx graph.Transaction) error {
			azUser, err := ops.FetchNode(tx, azUserPolluted.ID)
			require.NoError(t, err)
			fmt.Println(azUser)
			return nil
		})

		err := migrations.Version_920_Migration(suite.bhDatabase)(suite.context, suite.graphDB)
		require.NoError(t, err)

		err = suite.graphDB.ReadTransaction(suite.context, func(tx graph.Transaction) error {
			azUser, err := ops.FetchNode(tx, azUserPolluted.ID)
			require.NoError(t, err)
			require.True(t, azUser.Kinds.ContainsOneOf(azure.Entity), "azure.Entity must be preserved")
			require.True(t, azUser.Kinds.ContainsOneOf(azure.User), "azure.User type kind must be preserved")
			require.False(t, azUser.Kinds.ContainsOneOf(ad.Entity), "ad.Entity must be stripped")
			require.False(t, azUser.Kinds.ContainsOneOf(graph.StringKind("CustomKind")), "CustomKind must be stripped")

			azGroup, err := ops.FetchNode(tx, azGroupPolluted.ID)
			require.NoError(t, err)
			require.True(t, azGroup.Kinds.ContainsOneOf(azure.Entity))
			require.True(t, azGroup.Kinds.ContainsOneOf(azure.Group))
			require.False(t, azGroup.Kinds.ContainsOneOf(graph.StringKind("CustomKind")))

			azTenant, err := ops.FetchNode(tx, azTenantClean.ID)
			require.NoError(t, err)
			require.True(t, azTenant.Kinds.ContainsOneOf(azure.Entity))
			require.True(t, azTenant.Kinds.ContainsOneOf(azure.Tenant))

			adNode, err := ops.FetchNode(tx, adNodeUntouched.ID)
			require.NoError(t, err)
			require.True(t, adNode.Kinds.ContainsOneOf(ad.Entity), "ad.Entity must be preserved on clean AD nodes")
			require.True(t, adNode.Kinds.ContainsOneOf(ad.User), "AD type kinds must be preserved")

			ogNode, err := ops.FetchNode(tx, ogNodeUntouched.ID)
			require.NoError(t, err)
			require.True(t, ogNode.Kinds.ContainsOneOf(graph.StringKind("CustomKind")), "OpenGraph nodes must not be touched")

			return nil
		})
		require.NoError(t, err)
	})

	t.Run("is a no-op when SourceKindsData is nil", func(t *testing.T) {
		suite := setupIntegrationTestSuite(t)
		t.Cleanup(func() { suite.teardownIntegrationTestSuite(t) })

		require.NoError(t, suite.bhDatabase.RegisterSourceKind(suite.context)(graph.StringKind("CustomKind")))
		require.NoError(t, suite.graphDB.RefreshKinds(suite.context))

		adUserPolluted := newADNode("PollutedUser", ad.User, azure.Entity, graph.StringKind("CustomKind"))
		suite.createNodes(t, adUserPolluted)

		err := migrations.Version_920_Migration(nil)(suite.context, suite.graphDB)
		require.NoError(t, err)

		err = suite.graphDB.ReadTransaction(suite.context, func(tx graph.Transaction) error {
			adUser, err := ops.FetchNode(tx, adUserPolluted.ID)
			require.NoError(t, err)
			require.True(t, adUser.Kinds.ContainsOneOf(azure.Entity), "azure.Entity should remain when migration is skipped")
			require.True(t, adUser.Kinds.ContainsOneOf(graph.StringKind("CustomKind")), "CustomKind should remain when migration is skipped")
			return nil
		})
		require.NoError(t, err)
	})
}
