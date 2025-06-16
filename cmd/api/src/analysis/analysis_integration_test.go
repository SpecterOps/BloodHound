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

	"github.com/specterops/bloodhound/analysis"
	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchRDPEnsurenodesInPathcent(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RDPB.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		groupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchRemoteDesktopUsersBitmapForComputer(tx, harness.RDPB.Computer.ID, groupExpansions, false)
			require.Nil(t, err)

			// We should expect all groups that have the RIL incoming privilege to the computer
			require.Equal(t, 1, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPB.RDPDomainUsersGroup.ID.Uint64()))

			return nil
		}))
	})
}

func TestFetchRemoteDesktopUsersBitmapForComputer(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RDP.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		groupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		// Enforced URA validation
		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchRemoteDesktopUsersBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, true)
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
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchRemoteDesktopUsersBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, false)
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
			rdpEnabledEntityIDBitmap, err := adAnalysis.FetchRemoteDesktopUsersBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, true)
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

func TestFetchRDPEntityBitmapForComputer(t *testing.T) {
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

			// We should expect the intersection of members of `Direct Access Users,` with entities that are first degree members of the `Remote Desktop Users` group
			require.Equal(t, 3, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.DomainGroupC.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.IrshadUser.ID.Uint64()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.UliUser.ID.Uint64()))

			return nil
		}))
	})
}

func TestFetchACEInheritancePath(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ACEInheritedFrom.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		// First positive case
		startId := harness.ACEInheritedFrom.OU2.ID
		endId := harness.ACEInheritedFrom.Computer1.ID

		edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, startId, endId, ad.Contains)
		test.RequireNilErr(t, err)

		pathSet, err := adAnalysis.FetchACEInheritancePath(testContext.Context(), db, edge)
		test.RequireNilErr(t, err)

		nodesInPath := pathSet.AllNodes()

		assert.True(t, nodesInPath.Contains(harness.ACEInheritedFrom.Domain1))
		assert.True(t, nodesInPath.Contains(harness.ACEInheritedFrom.OU1))
		assert.True(t, nodesInPath.Contains(harness.ACEInheritedFrom.OU2))
		assert.True(t, nodesInPath.Contains(harness.ACEInheritedFrom.Computer1))
		assert.Len(t, nodesInPath, 4)

		// Second positive case
		startId = harness.ACEInheritedFrom.OU6.ID
		endId = harness.ACEInheritedFrom.Computer4.ID

		edge, err = analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, startId, endId, ad.Contains)
		test.RequireNilErr(t, err)

		pathSet, err = adAnalysis.FetchACEInheritancePath(testContext.Context(), db, edge)
		test.RequireNilErr(t, err)

		nodesInPath = pathSet.AllNodes()

		assert.True(t, nodesInPath.Contains(harness.ACEInheritedFrom.OU4))
		assert.True(t, nodesInPath.Contains(harness.ACEInheritedFrom.OU6))
		assert.True(t, nodesInPath.Contains(harness.ACEInheritedFrom.Computer4))
		assert.Len(t, nodesInPath, 3)
	})
}
