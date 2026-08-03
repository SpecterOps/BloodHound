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

package hybrid

import (
	"context"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	analysisAD "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
)

func TestHybridAttackPaths(t *testing.T) {
	t.Run("SyncedEdgesCreatedAndLinkExistingNodes", func(t *testing.T) {
		// ADUser.ObjectID matches AZUser.OnPremID, AZUser.OnPremSyncEnabled is true
		// SyncedToEntraUser and SyncedToADUser edges should be created and link the two nodes
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := integration.RandomObjectID(t)
				azUserOnPremID := adUserObjectID
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, true, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := post.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, true)
			},
		)
	})

	t.Run("SyncedEdgesNotCreated", func(t *testing.T) {

		// ADUser.ObjectID do NOT match as AZUser.OnPremID is null, AZUser.OnPremSyncEnabled is false
		// SyncedToEntraUser and SyncedToADUser edges should NOT be created
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := integration.RandomObjectID(t)
				azUserOnPremID := ""
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, false, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := post.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})

	t.Run("OnPremSyncEnabled False", func(t *testing.T) {
		// ADUser.ObjectID matches AZUser.OnPremID, AZUser.OnPremSyncEnabled is false
		// SyncedToEntraUser and SyncedToADUser edges should NOT be created
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := integration.RandomObjectID(t)
				azUserOnPremID := adUserObjectID
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, false, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := post.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})

	t.Run("SyncedEdgesNotCreatedWithoutMatchingADUser", func(t *testing.T) {
		// ADUser does not exist. AZUser has OnPremID and OnPremSyncEnabled=true
		// No ADUser node should be created. SyncedToADUser and SyncedToEntraUser edges should not be created.
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := ""
				azUserOnPremID := integration.RandomObjectID(t)
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, true, false, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := post.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})

	t.Run("SyncedEdgesNotCreatedForUnknownADEntity", func(t *testing.T) {
		// ADUser does not exist, but the objectid from a selected AZUser exists in the graph. Selected AZUser has OnPremID and
		// OnPremSyncEnabled=true
		// Hybrid post-processing only links existing AD user nodes, so no synced edges should be created.
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := ""
				azUserOnPremID := integration.RandomObjectID(t)
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, true, false, true)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := post.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})

	t.Run("SyncedEdgesNotCreatedWhenADObjectIDDoesNotMatchOnPremID", func(t *testing.T) {
		// ADUser.ObjectID does NOT match AZUser.OnPremID, AZUser.OnPremSyncEnabled is true
		// SyncedToEntraUser and SyncedToADUser edges should not be created.
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := integration.RandomObjectID(t)
				azUserOnPremID := integration.RandomObjectID(t)
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, true, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := post.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})
}

func TestSyncedToEntraDSEdges(t *testing.T) {
	t.Run("EdgesCreatedForMatchingADUserAndGroup", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		expectedEdges := []expectedSyncedToEntraDSEdge{}

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				tenantID := integration.RandomObjectID(t)
				tenant := testContext.NewAzureTenant(tenantID)

				azUserObjectID := integration.RandomObjectID(t)
				azGroupObjectID := integration.RandomObjectID(t)
				azUser := testContext.NewAzureUser("AZ User", "azuser@specter.dev", "", azUserObjectID, "", tenantID, false)
				azGroup := testContext.NewAzureGroup("AZ Group", azGroupObjectID, tenantID)
				testContext.NewRelationship(tenant, azUser, azure.Contains)
				testContext.NewRelationship(tenant, azGroup, azure.Contains)

				adUserObjectID := integration.RandomObjectID(t)
				adGroupObjectID := integration.RandomObjectID(t)
				testContext.NewCustomActiveDirectoryUser(graph.AsProperties(graph.PropertyMap{
					common.Name:     "ad_user",
					common.ObjectID: adUserObjectID,
					ad.DomainSID:    integration.RandomDomainSID(),
					ad.AADObjectID:  strings.ToLower(azUserObjectID),
				}))
				testContext.NewNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     "ad_group",
					common.ObjectID: adGroupObjectID,
					ad.DomainSID:    integration.RandomDomainSID(),
					ad.AADObjectID:  strings.ToLower(azGroupObjectID),
				}), ad.Entity, ad.Group)

				expectedEdges = []expectedSyncedToEntraDSEdge{
					{
						startObjectID: azUserObjectID,
						startKind:     azure.User,
						endObjectID:   adUserObjectID,
						endKind:       ad.User,
						kind:          azure.SyncedToEntraDSUser,
					},
					{
						startObjectID: azGroupObjectID,
						startKind:     azure.Group,
						endObjectID:   adGroupObjectID,
						endKind:       ad.Group,
						kind:          azure.SyncedToEntraDSGroup,
					},
				}

				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for Entra DS sync edges: %v", err)
				}

				verifySyncedToEntraDSEdges(t, db, expectedEdges)
			},
		)
	})

	t.Run("EdgesNotCreatedAcrossMismatchedObjectTypes", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				tenantID := integration.RandomObjectID(t)
				tenant := testContext.NewAzureTenant(tenantID)

				azUserObjectID := integration.RandomObjectID(t)
				azGroupObjectID := integration.RandomObjectID(t)
				azUser := testContext.NewAzureUser("AZ User", "azuser@specter.dev", "", azUserObjectID, "", tenantID, false)
				azGroup := testContext.NewAzureGroup("AZ Group", azGroupObjectID, tenantID)
				testContext.NewRelationship(tenant, azUser, azure.Contains)
				testContext.NewRelationship(tenant, azGroup, azure.Contains)

				testContext.NewCustomActiveDirectoryUser(graph.AsProperties(graph.PropertyMap{
					common.Name:     "ad_user",
					common.ObjectID: integration.RandomObjectID(t),
					ad.DomainSID:    integration.RandomDomainSID(),
					ad.AADObjectID:  azGroupObjectID,
				}))
				testContext.NewNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     "ad_group",
					common.ObjectID: integration.RandomObjectID(t),
					ad.DomainSID:    integration.RandomDomainSID(),
					ad.AADObjectID:  azUserObjectID,
				}), ad.Entity, ad.Group)

				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for Entra DS sync edges: %v", err)
				}

				verifySyncedToEntraDSEdges(t, db, nil)
			},
		)
	})
}

type expectedSyncedToEntraDSEdge struct {
	startObjectID string
	startKind     graph.Kind
	endObjectID   string
	endKind       graph.Kind
	kind          graph.Kind
}

func verifySyncedToEntraDSEdges(t *testing.T, db graph.Database, expectedEdges []expectedSyncedToEntraDSEdge) {
	t.Helper()

	expectedByObjectIDs := map[string]expectedSyncedToEntraDSEdge{}
	for _, expectedEdge := range expectedEdges {
		expectedByObjectIDs[expectedEdge.startObjectID+"|"+expectedEdge.endObjectID] = expectedEdge
	}

	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		edges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.KindIn(query.Relationship(), azure.SyncedToEntraDSUser, azure.SyncedToEntraDSGroup)
		}))
		assert.Nil(t, err)
		assert.Len(t, edges, len(expectedEdges))

		for _, edge := range edges {
			start, end, err := ops.FetchRelationshipNodes(tx, edge)
			assert.Nil(t, err)

			startObjectID, err := start.Properties.Get(common.ObjectID.String()).String()
			assert.Nil(t, err)

			endObjectID, err := end.Properties.Get(common.ObjectID.String()).String()
			assert.Nil(t, err)

			expectedEdge, ok := expectedByObjectIDs[startObjectID+"|"+endObjectID]
			assert.True(t, ok)
			assert.True(t, start.Kinds.ContainsOneOf(expectedEdge.startKind))
			assert.True(t, end.Kinds.ContainsOneOf(expectedEdge.endKind))
			assert.True(t, edge.Kind.Is(expectedEdge.kind))

			delete(expectedByObjectIDs, startObjectID+"|"+endObjectID)
		}

		assert.Empty(t, expectedByObjectIDs)

		return nil
	})
}

func TestAddEntraDSGroupMemberEdge(t *testing.T) {
	t.Run("EdgeCreatedViaAZAddMembers", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var azUserObjectID, adGroupObjectID string

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				azUser, _, _, adGroup := setupEntraDSGroupMemberHarness(t, testContext, azure.AddMembers, true, true)
				azUserObjectID = getObjectID(t, azUser)
				adGroupObjectID = getObjectID(t, adGroup)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for AddEntraDSGroupMember edge: %v", err)
				}
				verifyAddEntraDSGroupMemberEdge(t, db, azUserObjectID, adGroupObjectID, true)
			},
		)
	})

	t.Run("EdgeCreatedViaAZOwns", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var azUserObjectID, adGroupObjectID string

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				azUser, _, _, adGroup := setupEntraDSGroupMemberHarness(t, testContext, azure.Owns, true, true)
				azUserObjectID = getObjectID(t, azUser)
				adGroupObjectID = getObjectID(t, adGroup)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for AddEntraDSGroupMember edge: %v", err)
				}
				verifyAddEntraDSGroupMemberEdge(t, db, azUserObjectID, adGroupObjectID, true)
			},
		)
	})

	t.Run("EdgeNotCreatedWithoutControlEdge", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				// User and group both synced, but no AZOwns/AZAddMembers control edge
				setupEntraDSGroupMemberHarness(t, testContext, graph.StringKind(""), true, true)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for AddEntraDSGroupMember edge: %v", err)
				}
				verifyAddEntraDSGroupMemberEdge(t, db, "", "", false)
			},
		)
	})

	t.Run("EdgeNotCreatedWhenGroupNotSynced", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				// User synced and has control over the group, but the AZGroup has no synced on-prem counterpart
				setupEntraDSGroupMemberHarness(t, testContext, azure.AddMembers, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for AddEntraDSGroupMember edge: %v", err)
				}
				verifyAddEntraDSGroupMemberEdge(t, db, "", "", false)
			},
		)
	})

	t.Run("EdgeNotCreatedWhenUserNotSynced", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				// Group synced and user has control over the group, but the AZUser has no synced on-prem counterpart
				setupEntraDSGroupMemberHarness(t, testContext, azure.AddMembers, false, true)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for AddEntraDSGroupMember edge: %v", err)
				}
				verifyAddEntraDSGroupMemberEdge(t, db, "", "", false)
			},
		)
	})
}

func TestGetAddEntraDSGroupMemberEdgeComposition(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	var azUser, adUser, azGroup, adGroup *graph.Node

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			azUser, adUser, azGroup, adGroup = setupEntraDSGroupMemberHarness(t, testContext, azure.AddMembers, true, true)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			if _, err := PostHybrid(context.Background(), db); err != nil {
				t.Fatalf("failed post processing for AddEntraDSGroupMember edge: %v", err)
			}

			// Grab the created AddEntraDSGroupMember edge and reconstruct its composition
			var edge *graph.Relationship
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				edges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), azure.AddEntraDSGroupMember)
				}))
				assert.Nil(t, err)
				assert.Len(t, edges, 1)
				edge = edges[0]
				return nil
			})

			composition, err := analysisAD.GetAddEntraDSGroupMemberEdgeComposition(context.Background(), db, edge)
			assert.Nil(t, err)

			nodes := composition.AllNodes()
			// The composition should include every object involved in the three composing paths
			assert.True(t, nodes.Contains(azUser), "composition should contain the AZUser")
			assert.True(t, nodes.Contains(adUser), "composition should contain the synced on-prem User")
			assert.True(t, nodes.Contains(azGroup), "composition should contain the AZGroup")
			assert.True(t, nodes.Contains(adGroup), "composition should contain the synced on-prem Group")
		},
	)
}

func TestSyncEntraDSUsersEdge(t *testing.T) {
	t.Run("DomainServiceEdgeCreatedAcrossSyncScopes", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var domainService, domainUsers *graph.Node

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				domainService, _, _, _, _, domainUsers = setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     entraDSScopedSyncApplicationID,
					adminGroupName:    entraDSAdminGroupNamePrefix + "SPECTER.DEV",
					sameTenant:        true,
					syncAdminGroup:    true,
					matchingDomainSID: true,
					filteredSync:      "Disabled",
					syncScope:         "CloudOnly",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, getObjectID(t, domainUsers), map[string]graph.Kind{
					getObjectID(t, domainService): azure.DomainService,
				})
			},
		)
	})

	t.Run("ServicePrincipalEdgeCreatedWhenFilteredSyncEnabledAndScopeAll", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var domainService, servicePrincipal, domainUsers *graph.Node

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				domainService, _, servicePrincipal, _, _, domainUsers = setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     entraDSScopedSyncApplicationID,
					adminGroupName:    entraDSAdminGroupNamePrefix + "SPECTER.DEV",
					sameTenant:        true,
					syncAdminGroup:    true,
					matchingDomainSID: true,
					filteredSync:      "Enabled",
					syncScope:         "All",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, getObjectID(t, domainUsers), map[string]graph.Kind{
					getObjectID(t, domainService):    azure.DomainService,
					getObjectID(t, servicePrincipal): azure.ServicePrincipal,
				})
			},
		)
	})

	t.Run("ServicePrincipalEdgeNotCreatedWhenFilteredSyncDisabled", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var domainService, domainUsers *graph.Node

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				domainService, _, _, _, _, domainUsers = setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     entraDSScopedSyncApplicationID,
					adminGroupName:    entraDSAdminGroupNamePrefix + "SPECTER.DEV",
					sameTenant:        true,
					syncAdminGroup:    true,
					matchingDomainSID: true,
					filteredSync:      "Disabled",
					syncScope:         "All",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, getObjectID(t, domainUsers), map[string]graph.Kind{
					getObjectID(t, domainService): azure.DomainService,
				})
			},
		)
	})

	t.Run("ServicePrincipalEdgeNotCreatedWhenSyncScopeCloudOnly", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var domainService, domainUsers *graph.Node

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				domainService, _, _, _, _, domainUsers = setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     entraDSScopedSyncApplicationID,
					adminGroupName:    entraDSAdminGroupNamePrefix + "SPECTER.DEV",
					sameTenant:        true,
					syncAdminGroup:    true,
					matchingDomainSID: true,
					filteredSync:      "Enabled",
					syncScope:         "CloudOnly",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, getObjectID(t, domainUsers), map[string]graph.Kind{
					getObjectID(t, domainService): azure.DomainService,
				})
			},
		)
	})

	t.Run("ServicePrincipalEdgeNotCreatedForWrongApplication", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var domainService, domainUsers *graph.Node

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				domainService, _, _, _, _, domainUsers = setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     integration.RandomObjectID(t),
					adminGroupName:    entraDSAdminGroupNamePrefix + "SPECTER.DEV",
					sameTenant:        true,
					syncAdminGroup:    true,
					matchingDomainSID: true,
					filteredSync:      "Enabled",
					syncScope:         "All",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, getObjectID(t, domainUsers), map[string]graph.Kind{
					getObjectID(t, domainService): azure.DomainService,
				})
			},
		)
	})

	t.Run("EdgeNotCreatedForUnrelatedGroup", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     entraDSScopedSyncApplicationID,
					adminGroupName:    "NOT AAD DC ADMINISTRATORS@SPECTER.DEV",
					sameTenant:        true,
					syncAdminGroup:    true,
					matchingDomainSID: true,
					filteredSync:      "Enabled",
					syncScope:         "All",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, "", nil)
			},
		)
	})

	t.Run("EdgeNotCreatedWhenAdminGroupIsNotSynced", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     entraDSScopedSyncApplicationID,
					adminGroupName:    entraDSAdminGroupNamePrefix + "SPECTER.DEV",
					sameTenant:        true,
					syncAdminGroup:    false,
					matchingDomainSID: true,
					filteredSync:      "Enabled",
					syncScope:         "All",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, "", nil)
			},
		)
	})

	t.Run("EdgeNotCreatedForDifferentDomainSID", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     entraDSScopedSyncApplicationID,
					adminGroupName:    entraDSAdminGroupNamePrefix + "SPECTER.DEV",
					sameTenant:        true,
					syncAdminGroup:    true,
					matchingDomainSID: false,
					filteredSync:      "Enabled",
					syncScope:         "All",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, "", nil)
			},
		)
	})

	t.Run("ServicePrincipalEdgeNotCreatedAcrossTenants", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var domainService, domainUsers *graph.Node

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				domainService, _, _, _, _, domainUsers = setupSyncEntraDSUsersHarness(t, testContext, syncEntraDSUsersHarnessOptions{
					applicationID:     entraDSScopedSyncApplicationID,
					adminGroupName:    entraDSAdminGroupNamePrefix + "SPECTER.DEV",
					sameTenant:        false,
					syncAdminGroup:    true,
					matchingDomainSID: true,
					filteredSync:      "Enabled",
					syncScope:         "All",
				})
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for SyncEntraDSUsers edge: %v", err)
				}

				verifySyncEntraDSUsersEdges(t, db, getObjectID(t, domainUsers), map[string]graph.Kind{
					getObjectID(t, domainService): azure.DomainService,
				})
			},
		)
	})
}

type syncEntraDSUsersHarnessOptions struct {
	applicationID     string
	adminGroupName    string
	sameTenant        bool
	syncAdminGroup    bool
	matchingDomainSID bool
	filteredSync      string
	syncScope         string
}

func setupSyncEntraDSUsersHarness(t *testing.T, testContext *integration.GraphTestContext, options syncEntraDSUsersHarnessOptions) (domainService, application, servicePrincipal, azAdminGroup, adAdminGroup, domainUsers *graph.Node) {
	t.Helper()

	var (
		adminGroupTenantID       = integration.RandomObjectID(t)
		servicePrincipalTenantID = adminGroupTenantID
		domainSID                = integration.RandomDomainSID()
		domainUsersDomainSID     = domainSID
		filteredSync             = options.filteredSync
		syncScope                = options.syncScope
	)

	if !options.sameTenant {
		servicePrincipalTenantID = integration.RandomObjectID(t)
	}
	if !options.matchingDomainSID {
		domainUsersDomainSID = integration.RandomDomainSID()
	}
	if filteredSync == "" {
		filteredSync = "Disabled"
	}
	if syncScope == "" {
		syncScope = "CloudOnly"
	}

	adminGroupTenant := testContext.NewAzureTenant(adminGroupTenantID)
	servicePrincipalTenant := adminGroupTenant
	if !options.sameTenant {
		servicePrincipalTenant = testContext.NewAzureTenant(servicePrincipalTenantID)
	}

	domainService = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:        "SPECTER.DEV",
		common.ObjectID:    integration.RandomObjectID(t),
		azure.TenantID:     adminGroupTenantID,
		azure.FilteredSync: filteredSync,
		azure.SyncScope:    syncScope,
	}), azure.Entity, azure.DomainService)
	application = testContext.NewAzureApplication("Domain Controller Services", options.applicationID, servicePrincipalTenantID)
	servicePrincipal = testContext.NewAzureServicePrincipal("Domain Controller Services", integration.RandomObjectID(t), servicePrincipalTenantID)
	azAdminGroupObjectID := integration.RandomObjectID(t)
	azAdminGroup = testContext.NewAzureGroup(options.adminGroupName, azAdminGroupObjectID, adminGroupTenantID)
	testContext.NewRelationship(application, servicePrincipal, azure.RunsAs)
	testContext.NewRelationship(servicePrincipalTenant, servicePrincipal, azure.Contains)
	testContext.NewRelationship(adminGroupTenant, azAdminGroup, azure.Contains)

	adminGroupAADObjectID := integration.RandomObjectID(t)
	if options.syncAdminGroup {
		adminGroupAADObjectID = azAdminGroupObjectID
	}

	adAdminGroup = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     "AAD DC ADMINISTRATORS",
		common.ObjectID: domainSID + "-1104",
		ad.DomainSID:    domainSID,
		ad.AADObjectID:  adminGroupAADObjectID,
	}), ad.Entity, ad.Group)
	domainUsers = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     "DOMAIN USERS",
		common.ObjectID: domainUsersDomainSID + domainUsersObjectIDSuffix,
		ad.DomainSID:    domainUsersDomainSID,
	}), ad.Entity, ad.Group)

	return domainService, application, servicePrincipal, azAdminGroup, adAdminGroup, domainUsers
}

func verifySyncEntraDSUsersEdges(t *testing.T, db graph.Database, expectedEndObjectID string, expectedStartKinds map[string]graph.Kind) {
	t.Helper()

	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		edges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), azure.SyncEntraDSUsers)
		}))
		assert.Nil(t, err)

		if len(expectedStartKinds) == 0 {
			assert.Empty(t, edges)
			return nil
		}

		assert.Len(t, edges, len(expectedStartKinds))
		seen := make(map[string]struct{}, len(expectedStartKinds))
		for _, edge := range edges {
			start, end, err := ops.FetchRelationshipNodes(tx, edge)
			assert.Nil(t, err)

			startObjectID := getObjectID(t, start)
			expectedStartKind, ok := expectedStartKinds[startObjectID]
			assert.True(t, ok)
			if !ok {
				continue
			}

			assert.True(t, start.Kinds.ContainsOneOf(expectedStartKind))
			assert.True(t, end.Kinds.ContainsOneOf(ad.Group))
			assert.Equal(t, expectedEndObjectID, getObjectID(t, end))
			seen[startObjectID] = struct{}{}
		}
		assert.Len(t, seen, len(expectedStartKinds))

		return nil
	})
}

// setupEntraDSGroupMemberHarness builds an AZUser and AZGroup under a tenant with an optional control edge
// (AZAddMembers / AZOwns) between them. When syncUser/syncGroup are true, matching on-prem AD User/Group nodes are
// created (via ad.AADObjectID) so the corresponding SyncedToEntraDS edges are produced by PostHybrid. Pass an empty
// kind as controlKind to omit the control edge entirely. Returns the AZUser, on-prem User, AZGroup, on-prem Group.
func setupEntraDSGroupMemberHarness(t *testing.T, testContext *integration.GraphTestContext, controlKind graph.Kind, syncUser, syncGroup bool) (azUser, adUser, azGroup, adGroup *graph.Node) {
	t.Helper()

	tenantID := integration.RandomObjectID(t)
	tenant := testContext.NewAzureTenant(tenantID)

	azUserObjectID := integration.RandomObjectID(t)
	azGroupObjectID := integration.RandomObjectID(t)
	azUser = testContext.NewAzureUser("AZ User", "azuser@specter.dev", "", azUserObjectID, "", tenantID, false)
	azGroup = testContext.NewAzureGroup("AZ Group", azGroupObjectID, tenantID)
	testContext.NewRelationship(tenant, azUser, azure.Contains)
	testContext.NewRelationship(tenant, azGroup, azure.Contains)

	if controlKind.String() != "" {
		testContext.NewRelationship(azUser, azGroup, controlKind)
	}

	if syncUser {
		adUser = testContext.NewCustomActiveDirectoryUser(graph.AsProperties(graph.PropertyMap{
			common.Name:     "ad_user",
			common.ObjectID: integration.RandomObjectID(t),
			ad.DomainSID:    integration.RandomDomainSID(),
			ad.AADObjectID:  strings.ToLower(azUserObjectID),
		}))
	}

	if syncGroup {
		adGroup = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
			common.Name:     "ad_group",
			common.ObjectID: integration.RandomObjectID(t),
			ad.DomainSID:    integration.RandomDomainSID(),
			ad.AADObjectID:  strings.ToLower(azGroupObjectID),
		}), ad.Entity, ad.Group)
	}

	return azUser, adUser, azGroup, adGroup
}

func getObjectID(t *testing.T, node *graph.Node) string {
	t.Helper()
	objectID, err := node.Properties.Get(common.ObjectID.String()).String()
	assert.Nil(t, err)
	return objectID
}

func verifyAddEntraDSGroupMemberEdge(t *testing.T, db graph.Database, expectedStartObjectID, expectedEndObjectID string, shouldExist bool) {
	t.Helper()

	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		edges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), azure.AddEntraDSGroupMember)
		}))
		assert.Nil(t, err)

		if !shouldExist {
			assert.Len(t, edges, 0)
			return nil
		}

		assert.Len(t, edges, 1)
		for _, edge := range edges {
			start, end, err := ops.FetchRelationshipNodes(tx, edge)
			assert.Nil(t, err)

			startObjectID, err := start.Properties.Get(common.ObjectID.String()).String()
			assert.Nil(t, err)

			endObjectID, err := end.Properties.Get(common.ObjectID.String()).String()
			assert.Nil(t, err)

			// AddEntraDSGroupMember is drawn from the AZUser to the on-prem Group
			assert.True(t, start.Kinds.ContainsOneOf(azure.User))
			assert.True(t, end.Kinds.ContainsOneOf(ad.Group))
			assert.Equal(t, expectedStartObjectID, startObjectID)
			assert.Equal(t, expectedEndObjectID, endObjectID)
		}

		return nil
	})
}

func verifyHybridPaths(t *testing.T, db graph.Database, harness integration.HarnessDetails, shouldHaveEdges bool) {
	expectedEdgeCount := 1
	if !shouldHaveEdges {
		expectedEdgeCount = 0
	}

	// Verify the SyncedToADUser edge
	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		// Pull the edges
		syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.SyncedToADUser)
		}))
		assert.Nil(t, err)
		assert.Len(t, syncedToADUserEdges, expectedEdgeCount)

		for _, edge := range syncedToADUserEdges {
			// Retrieve the nodes connected to the edge
			start, end, err := ops.FetchRelationshipNodes(tx, edge)
			assert.Nil(t, err)

			// Get ObjectID and OnPremID from the AZUser node
			startObjectProp := start.Properties.Get(common.ObjectID.String())
			startObjectID, err := startObjectProp.String()
			assert.Nil(t, err)

			startObjectOnPremIdProp := start.Properties.Get(azure.OnPremID.String())
			startObjectOnPremId, err := startObjectOnPremIdProp.String()
			assert.Nil(t, err)

			// Get the ObjectID from the ADUser node
			endObjectProp := end.Properties.Get(common.ObjectID.String())
			endObjectID, err := endObjectProp.String()
			assert.Nil(t, err)

			// Ensure we got the correct node types
			assert.True(t, end.Kinds.ContainsOneOf(ad.User))
			assert.True(t, start.Kinds.ContainsOneOf(azure.User))

			// Verify the AZUser is the first node, User is the second
			assert.Equal(t, harness.HybridAttackPaths.AZUserObjectID, startObjectID)
			assert.Equal(t, harness.HybridAttackPaths.ADUserObjectID, endObjectID)

			// Verify the AZUser OnPremID property matches the User's ObjectID
			assert.Equal(t, startObjectOnPremId, endObjectID)
		}

		return nil
	})

	// Verify the SyncedToEntraUser edge
	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		// Pull the edges
		syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), azure.SyncedToEntraUser)
		}))
		assert.Nil(t, err)
		assert.Len(t, syncedToADUserEdges, expectedEdgeCount)

		for _, edge := range syncedToADUserEdges {
			// Retrieve the nodes connected to the edge
			start, end, err := ops.FetchRelationshipNodes(tx, edge)
			assert.Nil(t, err)

			// Get the ObjectID from the ADUser node
			startObjectProp := start.Properties.Get(common.ObjectID.String())
			startObjectID, err := startObjectProp.String()
			assert.Nil(t, err)

			// Get ObjectID and OnPremID from the AZUser node
			endObjectProp := end.Properties.Get(common.ObjectID.String())
			endObjectID, err := endObjectProp.String()
			assert.Nil(t, err)

			endObjectOnPremIdProp := end.Properties.Get(azure.OnPremID.String())
			endObjectOnPremId, err := endObjectOnPremIdProp.String()
			assert.Nil(t, err)

			// Ensure we got the correct node types
			assert.True(t, start.Kinds.ContainsOneOf(ad.User))
			assert.True(t, end.Kinds.ContainsOneOf(azure.User))

			// Verify the User is the first node, AZUser is the second
			assert.Equal(t, harness.HybridAttackPaths.ADUserObjectID, startObjectID)
			assert.Equal(t, harness.HybridAttackPaths.AZUserObjectID, endObjectID)

			// Verify the User's ObjectID matches the AZUser's OnPremID property
			assert.Equal(t, endObjectOnPremId, startObjectID)
		}

		return nil
	})
}
