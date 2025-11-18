// Copyright 2023 Specter Ops, Inc.
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
// +build integration

package analysis_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchRDPEnsureNoDescent(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RDPB.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		groupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDPB.Computer.ID, groupExpansions, false, false)
			require.Nil(t, err)

			// We should expect all groups that have the RIL incoming privilege to the computer
			require.Equal(t, 1, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPB.RDPDomainUsersGroup.ID.Uint64()))

			return nil
		}))
	})
}

func TestFetchCanRDPEntityBitmapForComputer(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RDP.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		groupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		// Enforced URA validation
		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, true, false)
			require.Nil(t, err)

			// We should expect all entities that have the RIL incoming privilege to the computer
			require.Equal(t, 7, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DillonUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.IrshadUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.UliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.EliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupA.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupB.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.RohanUser.ID.Uint64()))

			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupC.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupD.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupE.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.RDPDomainUsersGroup.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.AlyxUser.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.AndyUser.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.JohnUser.ID.Uint64()))

			return nil
		}))

		// Unenforced URA validation. result set should only include first degree members of `Remote Desktop Users` group
		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, false, false)
			require.Nil(t, err)

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.IrshadUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.UliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.EliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupA.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupB.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupC.ID.Uint64()))

			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupD.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupE.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.RDPDomainUsersGroup.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.AlyxUser.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DillonUser.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.AndyUser.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.RohanUser.ID.Uint64()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.JohnUser.ID.Uint64()))

			return nil
		}))

		// Create a RemoteInteractiveLogonRight relationship from the RDP local group to the computer to test our most common case
		require.Nil(t, db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
			_, err := tx.CreateRelationshipByIDs(harness.RDP.RDPLocalGroup.ID, harness.RDP.Computer.ID, ad.RemoteInteractiveLogonRight, graph.NewProperties())
			return err
		}))

		// Recalculate group expansions
		groupExpansions, err = adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		// result set should only include first degree members of `Remote Desktop Users` group.
		test.RequireNilErr(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, true, false)
			require.Nil(t, err)

			require.Equal(t, 6, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupC.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.IrshadUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.UliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupB.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.EliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupA.ID.Uint64()))

			return nil
		}))
	})
}

func TestFetchCanRDPEntityBitmapForComputerWithCitrix(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RDPHarnessWithCitrix.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		groupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		// the Remote Desktop Users group does not have an RIL(Remote Interactive Login) edge to the computer.
		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDPHarnessWithCitrix.Computer.ID, groupExpansions, true, true)
			require.Nil(t, err)

			// We should expect the intersection of members of `Direct Access Users`, with entities that have the RIL privilege to the computer
			require.Equal(t, 4, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.UliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.IrshadUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.DillonUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.RohanUser.ID.Uint64()))

			return nil
		}))

		// When citrix is enabled but URA is not enforced, we should expect the cross product of Remote Desktop Users and Direct Access Users
		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDPHarnessWithCitrix.Computer.ID, groupExpansions, false, true)
			require.Nil(t, err)

			require.Equal(t, 5, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.IrshadUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.UliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.DomainGroupC.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.RohanUser.ID.Uint64()))

			// This group does not have the RIL privilege, but should get a CanRDP edge
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.DomainGroupG.ID.Uint64()))
			return nil
		}))

		// Create a RemoteInteractiveLogonRight relationship from the RDP local group to the computer to test our most common case
		require.Nil(t, db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
			_, err := tx.CreateRelationshipByIDs(harness.RDPHarnessWithCitrix.RDPLocalGroup.ID, harness.RDPHarnessWithCitrix.Computer.ID, ad.RemoteInteractiveLogonRight, graph.NewProperties())
			return err
		}))

		// Recalculate group expansions
		groupExpansions, err = adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDPHarnessWithCitrix.Computer.ID, groupExpansions, true, true)
			require.Nil(t, err)

			// We should expect the cross product of members of `Direct Access Users,` `Remote Desktop Users`, and entities with RIL privileges to
			// the computer
			require.Equal(t, 5, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.DomainGroupC.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.IrshadUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.UliUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.RohanUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.DomainGroupG.ID.Uint64()))
			return nil
		}))
	})
}

func TestFetchCanBackupEntityBitmapForComputerSkipURA(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.BackupOperators.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		groupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			backupEnabledBitmap, err := adAnalysis.FetchCanBackupEntityBitmapForComputer(tx, harness.BackupOperators.Computer.ID, groupExpansions, false)
			require.Nil(t, err)

			expected := []uint64{
				harness.BackupOperators.GroupA.ID.Uint64(),
				harness.BackupOperators.GroupB.ID.Uint64(),
				harness.BackupOperators.Eli.ID.Uint64(),
				harness.BackupOperators.GroupC.ID.Uint64(),
				harness.BackupOperators.Uli.ID.Uint64(),
				harness.BackupOperators.Irshad.ID.Uint64(),
			}

			assert.ElementsMatch(t, expected, backupEnabledBitmap.Slice())
			return nil
		}))
	})
}

func TestFetchCanBackupEntityBitmapForComputerWithURA(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.BackupOperators.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		groupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			backupEnabledBitmap, err := adAnalysis.FetchCanBackupEntityBitmapForComputer(tx, harness.BackupOperators.Computer.ID, groupExpansions, true)
			require.Nil(t, err)

			expected := []uint64{
				harness.BackupOperators.GroupA.ID.Uint64(),
				harness.BackupOperators.GroupB.ID.Uint64(),
				harness.BackupOperators.Dillon.ID.Uint64(),
				harness.BackupOperators.Eli.ID.Uint64(),
				harness.BackupOperators.Irshad.ID.Uint64(),
				harness.BackupOperators.Rohan.ID.Uint64(),
				harness.BackupOperators.Uli.ID.Uint64(),
			}

			assert.ElementsMatch(t, expected, backupEnabledBitmap.Slice())
			assert.NotContains(t, backupEnabledBitmap.Slice(), harness.BackupOperators.Alex.ID.Uint64())
			assert.NotContains(t, backupEnabledBitmap.Slice(), harness.BackupOperators.Andy.ID.Uint64())
			assert.NotContains(t, backupEnabledBitmap.Slice(), harness.BackupOperators.John.ID.Uint64())
			assert.NotContains(t, backupEnabledBitmap.Slice(), harness.BackupOperators.GroupC.ID.Uint64())
			return nil
		}))
	})
}

func TestFetchACLInheritancePath(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ACLInheritanceHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		// Positive case 1
		startId := harness.ACLInheritanceHarness.User1.ID
		endId := harness.ACLInheritanceHarness.Computer1.ID

		edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, startId, endId, ad.GenericAll)
		test.RequireNilErr(t, err)

		pathSet, err := adAnalysis.FetchACLInheritancePath(testContext.Context(), db, edge)
		test.RequireNilErr(t, err)

		nodesInPath := pathSet.AllNodes()

		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.Domain1))
		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.OU1))
		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.OU2))
		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.User1))
		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.Computer1))
		assert.Len(t, nodesInPath, 5)

		// Positive case 2
		startId = harness.ACLInheritanceHarness.Group2.ID
		endId = harness.ACLInheritanceHarness.Computer2.ID

		edge, err = analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, startId, endId, ad.ReadLAPSPassword)
		test.RequireNilErr(t, err)

		pathSet, err = adAnalysis.FetchACLInheritancePath(testContext.Context(), db, edge)
		test.RequireNilErr(t, err)

		nodesInPath = pathSet.AllNodes()

		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.OU4))
		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.Container1))
		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.Group2))
		assert.True(t, nodesInPath.Contains(harness.ACLInheritanceHarness.Computer2))
		assert.Len(t, nodesInPath, 4)

		// Negative cases should all return empty result sets
		startId = harness.ACLInheritanceHarness.User2.ID
		endId = harness.ACLInheritanceHarness.Computer1.ID

		edge, err = analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, startId, endId, ad.GenericAll)
		test.RequireNilErr(t, err)

		pathSet, err = adAnalysis.FetchACLInheritancePath(testContext.Context(), db, edge)
		test.RequireNilErr(t, err)

		assert.Len(t, pathSet.AllNodes(), 0)

		// Negative case 2
		startId = harness.ACLInheritanceHarness.User3.ID
		endId = harness.ACLInheritanceHarness.Computer1.ID

		edge, err = analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, startId, endId, ad.GenericAll)
		test.RequireNilErr(t, err)

		pathSet, err = adAnalysis.FetchACLInheritancePath(testContext.Context(), db, edge)
		test.RequireNilErr(t, err)

		assert.Len(t, pathSet.AllNodes(), 0)

		// Negative case 3
		startId = harness.ACLInheritanceHarness.Group1.ID
		endId = harness.ACLInheritanceHarness.OU2.ID

		edge, err = analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, startId, endId, ad.GenericWrite)
		test.RequireNilErr(t, err)

		pathSet, err = adAnalysis.FetchACLInheritancePath(testContext.Context(), db, edge)
		test.RequireNilErr(t, err)

		assert.Len(t, pathSet.AllNodes(), 0)

		// Negative case 4
		startId = harness.ACLInheritanceHarness.Group3.ID
		endId = harness.ACLInheritanceHarness.Computer3.ID

		edge, err = analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, startId, endId, ad.ReadLAPSPassword)
		test.RequireNilErr(t, err)

		pathSet, err = adAnalysis.FetchACLInheritancePath(testContext.Context(), db, edge)
		test.RequireNilErr(t, err)

		assert.Len(t, pathSet.AllNodes(), 0)
	})
}
