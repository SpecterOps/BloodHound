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
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

				if _, err := PostHybrid(context.Background(), db, false); err != nil {
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

				if _, err := PostHybrid(context.Background(), db, false); err != nil {
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

				if _, err := PostHybrid(context.Background(), db, false); err != nil {
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

				if _, err := PostHybrid(context.Background(), db, false); err != nil {
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

				if _, err := PostHybrid(context.Background(), db, false); err != nil {
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

				if _, err := PostHybrid(context.Background(), db, false); err != nil {
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
				if _, err := PostHybrid(context.Background(), db, true); err != nil {
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
				if _, err := PostHybrid(context.Background(), db, true); err != nil {
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
				unsyncedGroup := testContext.NewAzureGroup("Unsynced Group", integration.RandomObjectID(t), integration.RandomObjectID(t))
				testContext.NewRelationship(azUser, unsyncedGroup, azure.AddMembers)
				azUserObjectID = getObjectID(t, azUser)
				adGroupObjectID = getObjectID(t, adGroup)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				if _, err := PostHybrid(context.Background(), db, true); err != nil {
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
				if _, err := PostHybrid(context.Background(), db, true); err != nil {
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
				if _, err := PostHybrid(context.Background(), db, true); err != nil {
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
				if _, err := PostHybrid(context.Background(), db, true); err != nil {
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
				if _, err := PostHybrid(context.Background(), db, true); err != nil {
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
			if _, err := PostHybrid(context.Background(), db, true); err != nil {
				t.Fatalf("failed post processing for AddEntraDSGroupMember edge: %v", err)
			}

			// Grab the created AddEntraDSGroupMember edge and reconstruct its composition
			var edges []*graph.Relationship
			err := db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				var err error
				edges, err = ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), azure.AddEntraDSGroupMember)
				}))
				return err
			})
			require.NoError(t, err)
			require.Len(t, edges, 1)

			composition, err := GetAddEntraDSGroupMemberEdgeComposition(context.Background(), db, edges[0])
			require.NoError(t, err)

			nodes := composition.AllNodes()
			require.NotZero(t, composition.Len())
			// The composition should include every object involved in the three composing paths
			assert.True(t, nodes.Contains(azUser), "composition should contain the AZUser")
			assert.True(t, nodes.Contains(adUser), "composition should contain the synced on-prem User")
			assert.True(t, nodes.Contains(azGroup), "composition should contain the AZGroup")
			assert.True(t, nodes.Contains(adGroup), "composition should contain the synced on-prem Group")
		},
	)
}

func TestPostHybridIgnoresIncompleteEntraNodes(t *testing.T) {
	var (
		testContext                     = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		azUserObjectID, adGroupObjectID string
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			azUser, _, _, adGroup := setupEntraDSGroupMemberHarness(t, testContext, azure.AddMembers, true, true)
			azUserObjectID = getObjectID(t, azUser)
			adGroupObjectID = getObjectID(t, adGroup)

			tenant := testContext.NewAzureTenant(integration.RandomObjectID(t))
			incompleteUser := testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name: "Incomplete User",
			}), azure.Entity, azure.User)
			incompleteGroup := testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name: entraDSAdminGroupNamePrefix + "EXAMPLE.COM",
			}), azure.Entity, azure.Group)
			testContext.NewRelationship(tenant, incompleteUser, azure.Contains)
			testContext.NewRelationship(tenant, incompleteGroup, azure.Contains)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, err := PostHybrid(context.Background(), db, true)
			require.NoError(t, err)
			verifyAddEntraDSGroupMemberEdge(t, db, azUserObjectID, adGroupObjectID, true)
		},
	)
}

func TestManageEntraDSSyncEdges(t *testing.T) {
	testCases := []struct {
		name               string
		options            manageEntraDSSyncHarnessOptions
		expectCorrelation  bool
		expectManageSync   bool
		expectManageFilter bool
	}{
		{
			name:              "BothEdgesCreatedForCorrelatedDomainAndEnabledFilter",
			options:           validManageEntraDSSyncOptions(),
			expectCorrelation: true, expectManageSync: true, expectManageFilter: true,
		},
		{
			name: "BroadEdgeIgnoresCurrentSynchronizationBoundary",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.filteredSyncEnabled = false
				options.syncScope = "CloudOnly"
				return options
			}(),
			expectCorrelation: true, expectManageSync: true, expectManageFilter: false,
		},
		{
			name: "FilterEdgeRequiresFilteredSyncEnabled",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.filteredSyncEnabled = false
				return options
			}(),
			expectCorrelation: true, expectManageSync: true, expectManageFilter: false,
		},
		{
			name: "FilterEdgeRequiresSyncScopeAll",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.syncScope = "CloudOnly"
				return options
			}(),
			expectCorrelation: true, expectManageSync: true, expectManageFilter: false,
		},
		{
			name: "FilterEdgeRequiresKnownApplication",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.applicationID = integration.RandomObjectID(t)
				return options
			}(),
			expectCorrelation: true, expectManageSync: true, expectManageFilter: false,
		},
		{
			name: "FilterEdgeRequiresSameTenant",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.sameTenant = false
				return options
			}(),
			expectCorrelation: true, expectManageSync: true, expectManageFilter: false,
		},
		{
			name: "CorrelationRequiresAADDCAdministratorsName",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.adminGroupName = "NOT AAD DC ADMINISTRATORS@SPECTER.DEV"
				return options
			}(),
		},
		{
			name: "CorrelationRequiresSynchronizedAADDCAdministrators",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.syncAdminGroup = false
				return options
			}(),
		},
		{
			name: "CorrelationRequiresMatchingDomainName",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.matchingDomainName = false
				return options
			}(),
		},
		{
			name: "CorrelationRequiresMatchingAdminGroupDomainSID",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.matchingAdminGroupDomainSID = false
				return options
			}(),
		},
		{
			name: "DomainUsersRequiresRID513InCorrelatedDomain",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.matchingDomainUsersSID = false
				return options
			}(),
			expectCorrelation: true,
		},
		{
			name: "BroadSyncRequiresDomainUsersContainment",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.containDomainUsers = false
				return options
			}(),
			expectCorrelation: true, expectManageSync: false, expectManageFilter: true,
		},
		{
			name: "MissingManagerDoesNotSuppressCorrelationOrFilterEdge",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.manageDomainService = false
				return options
			}(),
			expectCorrelation: true, expectManageSync: false, expectManageFilter: true,
		},
		{
			name: "MissingRunsAsDoesNotSuppressCorrelationOrManagerEdge",
			options: func() manageEntraDSSyncHarnessOptions {
				options := validManageEntraDSSyncOptions()
				options.includeRunsAs = false
				return options
			}(),
			expectCorrelation: true, expectManageSync: true, expectManageFilter: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
			var syncHarness manageEntraDSSyncHarness

			testContext.DatabaseTestWithSetup(
				func(harness *integration.HarnessDetails) error {
					syncHarness = setupManageEntraDSSyncHarness(t, testContext, testCase.options)
					return nil
				},
				func(harness integration.HarnessDetails, db graph.Database) {
					_, err := PostHybrid(context.Background(), db, true)
					require.NoError(t, err)
					verifyManageEntraDSSyncEdges(t, db, syncHarness, testCase.expectCorrelation, testCase.expectManageSync, testCase.expectManageFilter)
				},
			)
		})
	}

	t.Run("AmbiguousDomainNameFailsClosed", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var syncHarness manageEntraDSSyncHarness

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				syncHarness = setupManageEntraDSSyncHarness(t, testContext, validManageEntraDSSyncOptions())
				testContext.NewNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     "specter.dev",
					common.ObjectID: integration.RandomDomainSID(),
					ad.DomainSID:    integration.RandomDomainSID(),
				}), ad.Entity, ad.Domain)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				_, err := PostHybrid(context.Background(), db, true)
				require.NoError(t, err)
				verifyManageEntraDSSyncEdges(t, db, syncHarness, false, false, false)
			},
		)
	})

	t.Run("SelfReferentialRunsAsIsIgnored", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		var syncHarness manageEntraDSSyncHarness

		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				syncHarness = setupManageEntraDSSyncHarness(t, testContext, validManageEntraDSSyncOptions())
				mergedApplication := testContext.NewNode(graph.AsProperties(graph.PropertyMap{
					common.ObjectID: integration.RandomObjectID(t),
					azure.TenantID:  integration.RandomObjectID(t),
				}), azure.Entity, azure.App, azure.ServicePrincipal)
				testContext.NewRelationship(mergedApplication, mergedApplication, azure.RunsAs)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				_, err := PostHybrid(context.Background(), db, true)
				require.NoError(t, err)
				verifyManageEntraDSSyncEdges(t, db, syncHarness, true, true, true)
			},
		)
	})
}

func TestFilterContainedDomainUsersEmptyTraversal(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	var domain, domainUsers *graph.Node

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domain = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name:     "SPECTER.DEV",
				common.ObjectID: integration.RandomDomainSID(),
			}), ad.Entity, ad.Domain)
			domainUsers = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name:     "DOMAIN USERS@SPECTER.DEV",
				common.ObjectID: integration.RandomDomainSID(),
			}), ad.Entity, ad.Group)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			var containedDomainUsers []graph.ID
			err := db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				var err error
				containedDomainUsers, err = filterContainedDomainUsers(tx, domain, []graph.ID{domainUsers.ID})
				return err
			})
			require.NoError(t, err)
			assert.Empty(t, containedDomainUsers)
		},
	)
}

func TestGetManageEntraDSSyncEdgeComposition(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	var syncHarness manageEntraDSSyncHarness
	var container *graph.Node
	var unrelatedDomainService *graph.Node
	var unrelatedDomain *graph.Node

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			options := validManageEntraDSSyncOptions()
			options.containDomainUsers = false
			syncHarness = setupManageEntraDSSyncHarness(t, testContext, options)
			container = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name:     "USERS@SPECTER.DEV",
				common.ObjectID: integration.RandomObjectID(t),
			}), ad.Entity, ad.Container)
			testContext.NewRelationship(syncHarness.domain, container, ad.Contains)
			testContext.NewRelationship(container, syncHarness.domainUsers, ad.Contains)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, err := PostHybrid(context.Background(), db, true)
			require.NoError(t, err)

			unrelatedDomainService = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name:     "UNRELATED.SPECTER.DEV",
				common.ObjectID: integration.RandomObjectID(t),
			}), azure.Entity, azure.EntraDS)
			unrelatedDomain = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name:     "UNRELATED.SPECTER.DEV",
				common.ObjectID: integration.RandomDomainSID(),
				ad.DomainSID:    integration.RandomDomainSID(),
			}), ad.Entity, ad.Domain)
			testContext.NewRelationship(syncHarness.manager, unrelatedDomainService, azure.ManageEntraDS)
			testContext.NewRelationship(unrelatedDomainService, unrelatedDomain, azure.EntraDSFor)

			var edge *graph.Relationship
			err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				edges, err := ops.FetchRelationships(tx.Relationships().Filter(query.Kind(query.Relationship(), azure.ManageEntraDSSync)))
				require.NoError(t, err)
				require.Len(t, edges, 1)
				edge = edges[0]
				return nil
			})
			require.NoError(t, err)

			composition, err := GetManageEntraDSSyncEdgeComposition(context.Background(), db, edge)
			require.NoError(t, err)
			nodes := composition.AllNodes()
			assert.True(t, nodes.Contains(syncHarness.manager))
			assert.True(t, nodes.Contains(syncHarness.domainService))
			assert.True(t, nodes.Contains(syncHarness.domain))
			assert.True(t, nodes.Contains(container))
			assert.True(t, nodes.Contains(syncHarness.domainUsers))
			assert.False(t, nodes.Contains(syncHarness.azAdminGroup))
			assert.False(t, nodes.Contains(syncHarness.adAdminGroup))
			assert.False(t, nodes.Contains(unrelatedDomainService))
			assert.False(t, nodes.Contains(unrelatedDomain))
		},
	)
}

func TestGetManageEntraDSSyncEdgeCompositionWithoutContainment(t *testing.T) {
	var (
		testContext = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		edge        *graph.Relationship
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			options := validManageEntraDSSyncOptions()
			options.containDomainUsers = false
			syncHarness := setupManageEntraDSSyncHarness(t, testContext, options)
			testContext.NewRelationship(syncHarness.domainService, syncHarness.domain, azure.EntraDSFor)
			edge = testContext.NewRelationship(syncHarness.manager, syncHarness.domainUsers, azure.ManageEntraDSSync)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			composition, err := GetManageEntraDSSyncEdgeComposition(context.Background(), db, edge)
			require.NoError(t, err)
			assert.Zero(t, composition.Len())
		},
	)
}

type manageEntraDSSyncHarnessOptions struct {
	applicationID               string
	adminGroupName              string
	manageDomainService         bool
	includeRunsAs               bool
	sameTenant                  bool
	syncAdminGroup              bool
	matchingDomainName          bool
	matchingAdminGroupDomainSID bool
	matchingDomainUsersSID      bool
	containDomainUsers          bool
	filteredSyncEnabled         bool
	syncScope                   string
}

type manageEntraDSSyncHarness struct {
	domainService, application, servicePrincipal, manager, azAdminGroup, adAdminGroup, domain, domainUsers *graph.Node
}

func validManageEntraDSSyncOptions() manageEntraDSSyncHarnessOptions {
	return manageEntraDSSyncHarnessOptions{
		applicationID:               entraDSScopedSyncApplicationID,
		adminGroupName:              entraDSAdminGroupNamePrefix + "SPECTER.DEV",
		manageDomainService:         true,
		includeRunsAs:               true,
		sameTenant:                  true,
		syncAdminGroup:              true,
		matchingDomainName:          true,
		matchingAdminGroupDomainSID: true,
		matchingDomainUsersSID:      true,
		containDomainUsers:          true,
		filteredSyncEnabled:         true,
		syncScope:                   "All",
	}
}

func setupManageEntraDSSyncHarness(t *testing.T, testContext *integration.GraphTestContext, options manageEntraDSSyncHarnessOptions) manageEntraDSSyncHarness {
	t.Helper()

	var (
		tenantID                 = integration.RandomObjectID(t)
		servicePrincipalTenantID = tenantID
		domainSID                = integration.RandomDomainSID()
		adminGroupDomainSID      = domainSID
		domainUsersDomainSID     = domainSID
		domainName               = "SPECTER.DEV"
		domainServiceDomainName  = " specter.dev "
	)

	if !options.sameTenant {
		servicePrincipalTenantID = integration.RandomObjectID(t)
	}
	if !options.matchingDomainName {
		domainServiceDomainName = "other.example"
	}
	if !options.matchingAdminGroupDomainSID {
		adminGroupDomainSID = integration.RandomDomainSID()
	}
	if !options.matchingDomainUsersSID {
		domainUsersDomainSID = integration.RandomDomainSID()
	}

	tenant := testContext.NewAzureTenant(tenantID)
	servicePrincipalTenant := tenant
	if !options.sameTenant {
		servicePrincipalTenant = testContext.NewAzureTenant(servicePrincipalTenantID)
	}

	domainService := testContext.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:               "Managed Domain",
		common.ObjectID:           integration.RandomObjectID(t),
		azure.TenantID:            tenantID,
		azure.DomainName:          domainServiceDomainName,
		azure.FilteredSyncEnabled: options.filteredSyncEnabled,
		azure.SyncScope:           options.syncScope,
	}), azure.Entity, azure.EntraDS)
	application := testContext.NewAzureApplication("Domain Controller Services", options.applicationID, servicePrincipalTenantID)
	servicePrincipal := testContext.NewAzureServicePrincipal("Domain Controller Services", integration.RandomObjectID(t), servicePrincipalTenantID)
	manager := testContext.NewAzureGroup("Managed Domain Manager", integration.RandomObjectID(t), tenantID)
	azAdminGroupObjectID := integration.RandomObjectID(t)
	azAdminGroup := testContext.NewAzureGroup(options.adminGroupName, azAdminGroupObjectID, tenantID)
	if options.includeRunsAs {
		testContext.NewRelationship(application, servicePrincipal, azure.RunsAs)
	}
	testContext.NewRelationship(servicePrincipalTenant, servicePrincipal, azure.Contains)
	testContext.NewRelationship(tenant, manager, azure.Contains)
	testContext.NewRelationship(tenant, azAdminGroup, azure.Contains)
	if options.manageDomainService {
		testContext.NewRelationship(manager, domainService, azure.ManageEntraDS)
	}

	adminGroupAADObjectID := integration.RandomObjectID(t)
	if options.syncAdminGroup {
		adminGroupAADObjectID = azAdminGroupObjectID
	}

	adAdminGroup := testContext.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     "AAD DC ADMINISTRATORS",
		common.ObjectID: adminGroupDomainSID + "-1104",
		ad.DomainSID:    adminGroupDomainSID,
		ad.AADObjectID:  adminGroupAADObjectID,
	}), ad.Entity, ad.Group)
	domain := testContext.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     domainName,
		common.ObjectID: domainSID,
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.Domain)
	domainUsers := testContext.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     "DOMAIN USERS",
		common.ObjectID: domainUsersDomainSID + domainUsersObjectIDSuffix,
		ad.DomainSID:    domainUsersDomainSID,
	}), ad.Entity, ad.Group)
	if options.containDomainUsers {
		testContext.NewRelationship(domain, domainUsers, ad.Contains)
	}

	return manageEntraDSSyncHarness{
		domainService: domainService, application: application, servicePrincipal: servicePrincipal, manager: manager,
		azAdminGroup: azAdminGroup, adAdminGroup: adAdminGroup, domain: domain, domainUsers: domainUsers,
	}
}

func verifyManageEntraDSSyncEdges(t *testing.T, db graph.Database, syncHarness manageEntraDSSyncHarness, expectCorrelation, expectManageSync, expectManageFilter bool) {
	t.Helper()

	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		for _, expectation := range []struct {
			kind        graph.Kind
			start, end  *graph.Node
			shouldExist bool
		}{
			{azure.EntraDSFor, syncHarness.domainService, syncHarness.domain, expectCorrelation},
			{azure.ManageEntraDSSync, syncHarness.manager, syncHarness.domainUsers, expectManageSync},
			{azure.ManageEntraDSSyncFilter, syncHarness.servicePrincipal, syncHarness.domainUsers, expectManageFilter},
		} {
			edges, err := ops.FetchRelationships(tx.Relationships().Filter(query.Kind(query.Relationship(), expectation.kind)))
			require.NoError(t, err)
			if !expectation.shouldExist {
				assert.Empty(t, edges)
				continue
			}

			require.Len(t, edges, 1)
			assert.Equal(t, expectation.start.ID, edges[0].StartID)
			assert.Equal(t, expectation.end.ID, edges[0].EndID)
		}

		return nil
	})
}

// setupEntraDSGroupMemberHarness builds an AZUser and AZGroup under a tenant with an optional control edge
// (AZAddMembers / AZOwns) between them. When syncUser/syncGroup are true, matching on-prem AD User/Group nodes are
// created (via ad.AADObjectID) so the corresponding SyncedToEntraDS edges are produced by PostHybrid. Pass an empty
// kind as controlKind to omit the control edge entirely. Returns the AZUser, on-prem User, AZGroup, on-prem Group.
func setupEntraDSGroupMemberHarness(t *testing.T, testContext *integration.GraphTestContext, controlKind graph.Kind, syncUser, syncGroup bool) (*graph.Node, *graph.Node, *graph.Node, *graph.Node) {
	t.Helper()

	var (
		adUser  *graph.Node
		adGroup *graph.Node
	)

	tenantID := integration.RandomObjectID(t)
	tenant := testContext.NewAzureTenant(tenantID)

	azUserObjectID := integration.RandomObjectID(t)
	azGroupObjectID := integration.RandomObjectID(t)
	azUser := testContext.NewAzureUser("AZ User", "azuser@specter.dev", "", azUserObjectID, "", tenantID, false)
	azGroup := testContext.NewAzureGroup("AZ Group", azGroupObjectID, tenantID)
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
