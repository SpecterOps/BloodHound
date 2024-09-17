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

	analysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestFetchRDPEnsureNoDescent(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RDPB.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		groupExpansions, err := analysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := analysis.FetchRemoteDesktopUsersBitmapForComputer(tx, harness.RDPB.Computer.ID, groupExpansions, false)
			require.Nil(t, err)

			// We should expect all groups that have the RIL incoming privilege to the computer
			require.Equal(t, 1, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPB.RDPDomainUsersGroup.ID.Uint32()))

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
		groupExpansions, err := analysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		// Enforced URA validation
		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := analysis.FetchRemoteDesktopUsersBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, true)
			require.Nil(t, err)

			// We should expect all entities that have the RIL incoming privilege to the computer
			require.Equal(t, 7, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DillonUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.IrshadUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.UliUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.EliUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupA.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupB.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.RohanUser.ID.Uint32()))

			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupC.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupD.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupE.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.RDPDomainUsersGroup.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.AlyxUser.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.AndyUser.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.JohnUser.ID.Uint32()))

			return nil
		}))

		// Unenforced URA validation. result set should only include first degree members of `Remote Desktop Users` group
		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := analysis.FetchRemoteDesktopUsersBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, false)
			require.Nil(t, err)

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.IrshadUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.UliUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.EliUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupA.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupB.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupC.ID.Uint32()))

			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupD.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupE.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.RDPDomainUsersGroup.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.AlyxUser.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DillonUser.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.AndyUser.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.RohanUser.ID.Uint32()))
			require.False(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.JohnUser.ID.Uint32()))

			return nil
		}))

		// Create a RemoteInteractiveLogonPrivilege relationship from the RDP local group to the computer to test our most common case
		require.Nil(t, db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
			_, err := tx.CreateRelationshipByIDs(harness.RDP.RDPLocalGroup.ID, harness.RDP.Computer.ID, ad.RemoteInteractiveLogonPrivilege, graph.NewProperties())
			return err
		}))

		// Recalculate group expansions
		groupExpansions, err = analysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		// result set should only include first degree members of `Remote Desktop Users` group.
		test.RequireNilErr(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := analysis.FetchRemoteDesktopUsersBitmapForComputer(tx, harness.RDP.Computer.ID, groupExpansions, true)
			require.Nil(t, err)

			require.Equal(t, 6, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupC.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.IrshadUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.UliUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupB.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.EliUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDP.DomainGroupA.ID.Uint32()))

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
		groupExpansions, err := analysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		// the Remote Desktop Users group does not have an RIL(Remote Interactive Login) edge to the computer.
		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := analysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDPHarnessWithCitrix.Computer.ID, groupExpansions, true, true)
			require.Nil(t, err)

			// We should expect the intersection of members of `Direct Access Users`, with entities that have the RIL privilege to the computer
			require.Equal(t, 4, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.UliUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.IrshadUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.DillonUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.RohanUser.ID.Uint32()))

			return nil
		}))

		// Create a RemoteInteractiveLogonPrivilege relationship from the RDP local group to the computer to test our most common case
		require.Nil(t, db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
			_, err := tx.CreateRelationshipByIDs(harness.RDPHarnessWithCitrix.RDPLocalGroup.ID, harness.RDPHarnessWithCitrix.Computer.ID, ad.RemoteInteractiveLogonPrivilege, graph.NewProperties())
			return err
		}))

		// Recalculate group expansions
		groupExpansions, err = analysis.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)

		require.Nil(t, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			rdpEnabledEntityIDBitmap, err := analysis.FetchCanRDPEntityBitmapForComputer(tx, harness.RDPHarnessWithCitrix.Computer.ID, groupExpansions, true, true)
			require.Nil(t, err)

			// We should expect the intersection of members of `Direct Access Users,` with entities that are first degree members of the `Remote Desktop Users` group
			require.Equal(t, 3, int(rdpEnabledEntityIDBitmap.Cardinality()))

			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.DomainGroupC.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.IrshadUser.ID.Uint32()))
			require.True(t, rdpEnabledEntityIDBitmap.Contains(harness.RDPHarnessWithCitrix.UliUser.ID.Uint32()))

			return nil
		}))
	})
}
