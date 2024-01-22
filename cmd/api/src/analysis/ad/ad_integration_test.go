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

package ad_test

import (
	"context"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/src/test"
	"testing"

	"github.com/specterops/bloodhound/analysis"
	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestFetchEnforcedGPOs(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		// Check the first user
		var (
			enforcedGPOs, err = adAnalysis.FetchEnforcedGPOs(tx, harness.GPOEnforcement.UserC, 0, 0)
		)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, enforcedGPOs.Len())

		// Check the second user
		enforcedGPOs, err = adAnalysis.FetchEnforcedGPOs(tx, harness.GPOEnforcement.UserB, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, enforcedGPOs.Len())
	})
}

func TestFetchGPOAffectedContainerPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		containers, err := adAnalysis.FetchGPOAffectedContainerPaths(tx, harness.GPOEnforcement.GPOEnforced)

		test.RequireNilErr(t, err)
		nodes := containers.AllNodes().IDs()
		require.Equal(t, 6, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitD.ID)

		containers, err = adAnalysis.FetchGPOAffectedContainerPaths(tx, harness.GPOEnforcement.GPOUnenforced)
		test.RequireNilErr(t, err)
		nodes = containers.AllNodes().IDs()
		require.Equal(t, 5, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitD.ID)
	})
}

func TestCreateGPOAffectedIntermediariesListDelegateAffectedContainers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		containers, err := adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter)(tx, harness.GPOEnforcement.GPOEnforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 5, containers.Len())
		require.Equal(t, 4, containers.ContainingNodeKinds(ad.OU).Len())
		require.Equal(t, 1, containers.ContainingNodeKinds(ad.Domain).Len())

		containers, err = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter)(tx, harness.GPOEnforcement.GPOUnenforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 4, containers.Len())
		require.False(t, containers.Contains(harness.GPOEnforcement.OrganizationalUnitC))
		require.Equal(t, 3, containers.ContainingNodeKinds(ad.OU).Len())
		require.Equal(t, 1, containers.ContainingNodeKinds(ad.Domain).Len())
	})
}

func TestCreateGPOAffectedIntermediariesPathDelegateAffectedUsers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		users, err := adAnalysis.CreateGPOAffectedIntermediariesPathDelegate(ad.User)(tx, harness.GPOEnforcement.GPOEnforced)

		test.RequireNilErr(t, err)
		nodes := users.AllNodes().IDs()
		require.Equal(t, 10, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserD.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitC.ID)

		users, err = adAnalysis.CreateGPOAffectedIntermediariesPathDelegate(ad.User)(tx, harness.GPOEnforcement.GPOUnenforced)
		test.RequireNilErr(t, err)
		nodes = users.AllNodes().IDs()
		require.Equal(t, 8, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)
		require.NotContains(t, nodes, harness.GPOEnforcement.UserC.ID)
		require.NotContains(t, nodes, harness.GPOEnforcement.OrganizationalUnitC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserD.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserA.ID)
	})
}

func TestCreateGPOAffectedResultsListDelegateAffectedUsers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		users, err := adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter)(tx, harness.GPOEnforcement.GPOEnforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 4, users.Len())

		users, err = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter)(tx, harness.GPOEnforcement.GPOUnenforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 3, users.Len())
		require.Equal(t, 3, users.ContainingNodeKinds(ad.User).Len())
	})
}

func TestCreateGPOAffectedIntermediariesListDelegateTierZero(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		harness.GPOEnforcement.UserC.Properties.Set(common.SystemTags.String(), ad.AdminTierZero)
		harness.GPOEnforcement.UserD.Properties.Set(common.SystemTags.String(), ad.AdminTierZero)
		tx.UpdateNode(harness.GPOEnforcement.UserC)
		tx.UpdateNode(harness.GPOEnforcement.UserD)

		users, err := adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter)(tx, harness.GPOEnforcement.GPOEnforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, users.Len())

		users, err = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter)(tx, harness.GPOEnforcement.GPOUnenforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, users.Len())
	})
}

func TestFetchComputerSessionPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		sessions, err := adAnalysis.FetchComputerSessionPaths(tx, harness.Session.ComputerA)

		test.RequireNilErr(t, err)
		nodes := sessions.AllNodes().IDs()
		require.Equal(t, 2, len(nodes))
		require.Contains(t, nodes, harness.Session.ComputerA.ID)
		require.Contains(t, nodes, harness.Session.User.ID)
	})
}

func TestFetchComputerSessions(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		sessions, err := adAnalysis.FetchComputerSessions(tx, harness.Session.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, sessions.Len())
	})
}

func TestFetchGroupSessionPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.FetchGroupSessionPaths(tx, harness.Session.GroupA)

		test.RequireNilErr(t, err)
		nodes := computers.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))

		nestedComputers, err := adAnalysis.FetchGroupSessionPaths(tx, harness.Session.GroupC)

		test.RequireNilErr(t, err)
		nestedNodes := nestedComputers.AllNodes().IDs()
		require.Equal(t, 5, len(nestedNodes))
	})
}

func TestFetchGroupSessions(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.FetchGroupSessions(tx, harness.Session.GroupA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, computers.Len())
		require.Equal(t, 2, computers.ContainingNodeKinds(ad.Computer).Len())

		nestedComputers, err := adAnalysis.FetchGroupSessions(tx, harness.Session.GroupC, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, nestedComputers.Len())
		require.Equal(t, 2, nestedComputers.ContainingNodeKinds(ad.Computer).Len())
	})
}

func TestFetchUserSessionPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		sessions, err := adAnalysis.FetchUserSessionPaths(tx, harness.Session.User)

		test.RequireNilErr(t, err)
		nodes := sessions.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.Session.User.ID)
		require.Contains(t, nodes, harness.Session.ComputerA.ID)
		require.Contains(t, nodes, harness.Session.ComputerB.ID)
	})
}

func TestFetchUserSessions(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.FetchUserSessions(tx, harness.Session.User, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, computers.Len())
		require.Equal(t, 2, computers.ContainingNodeKinds(ad.Computer).Len())
	})
}

func TestCreateOutboundLocalGroupPathDelegateUser(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.UserB)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserB.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 3, len(nodes))
	})
}

func TestCreateOutboundLocalGroupListDelegateUser(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.UserB, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, computers.Len())
		require.Equal(t, harness.LocalGroupSQL.ComputerA.ID, computers.Slice()[0].ID)
	})
}

func TestCreateOutboundLocalGroupPathDelegateGroup(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.GroupA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))
	})
}

func TestCreateOutboundLocalGroupListDelegateGroup(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.GroupA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, computers.Len())
		require.Equal(t, harness.LocalGroupSQL.ComputerA.ID, computers.Slice()[0].ID)
	})
}

func TestCreateOutboundLocalGroupPathDelegateComputer(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerB.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerC.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupB.ID)
		require.Equal(t, 4, len(nodes))
	})
}

func TestCreateOutboundLocalGroupListDelegateComputer(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, computers.Len())
	})
}

func TestCreateInboundLocalGroupPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateInboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserB.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.UserA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 4, len(nodes))

		path, err = adAnalysis.CreateInboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerC)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupB.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerC.ID)
		require.Equal(t, 3, len(nodes))
	})
}

func TestCreateInboundLocalGroupListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, admins.Len())
		require.Equal(t, 2, admins.ContainingNodeKinds(ad.User).Len())

		admins, err = adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerC, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, admins.Len())
		require.Equal(t, harness.LocalGroupSQL.ComputerA.ID, admins.Slice()[0].ID)

		admins, err = adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerB, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, admins.Len())
		require.Equal(t, harness.LocalGroupSQL.ComputerA.ID, admins.Slice()[0].ID)
	})
}

func TestCreateSQLAdminPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateSQLAdminPathDelegate(graph.DirectionInbound)(tx, harness.LocalGroupSQL.ComputerA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserC.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))

		path, err = adAnalysis.CreateSQLAdminPathDelegate(graph.DirectionOutbound)(tx, harness.LocalGroupSQL.UserC)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserC.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))
	})
}

func TestCreateSQLAdminListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.CreateSQLAdminListDelegate(graph.DirectionInbound)(tx, harness.LocalGroupSQL.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, admins.Len())

		computers, err := adAnalysis.CreateSQLAdminListDelegate(graph.DirectionOutbound)(tx, harness.LocalGroupSQL.UserC, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 1, computers.Len())
	})
}

func TestCreateConstrainedDelegationPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateConstrainedDelegationPathDelegate(graph.DirectionInbound)(tx, harness.LocalGroupSQL.ComputerA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserD.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))

		path, err = adAnalysis.CreateConstrainedDelegationPathDelegate(graph.DirectionOutbound)(tx, harness.LocalGroupSQL.UserD)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserD.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))
	})
}

func TestCreateConstrainedDelegationListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionInbound)(tx, harness.LocalGroupSQL.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, admins.Len())

		computers, err := adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionOutbound)(tx, harness.LocalGroupSQL.UserD, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 1, computers.Len())
	})
}

func TestFetchOutboundADEntityControlPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OutboundControl.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		path, err := adAnalysis.FetchOutboundADEntityControlPaths(context.Background(), db, harness.OutboundControl.Controller)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Equal(t, 7, len(nodes))
		require.Contains(t, nodes, harness.OutboundControl.Controller.ID)
		require.Contains(t, nodes, harness.OutboundControl.GroupA.ID)
		require.Contains(t, nodes, harness.OutboundControl.UserC.ID)
		require.Contains(t, nodes, harness.OutboundControl.GroupB.ID)
		require.Contains(t, nodes, harness.OutboundControl.GroupC.ID)
		require.Contains(t, nodes, harness.OutboundControl.ComputerA.ID)
		require.Contains(t, nodes, harness.OutboundControl.ComputerC.ID)
	})
}

func TestFetchOutboundADEntityControl(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OutboundControl.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		control, err := adAnalysis.FetchOutboundADEntityControl(context.Background(), db, harness.OutboundControl.Controller, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 4, control.Len())
		ids := control.IDs()

		require.Contains(t, ids, harness.OutboundControl.GroupB.ID)
		require.Contains(t, ids, harness.OutboundControl.UserC.ID)
		require.Contains(t, ids, harness.OutboundControl.ComputerA.ID)
		require.Contains(t, ids, harness.OutboundControl.ComputerC.ID)

		control, err = adAnalysis.FetchOutboundADEntityControl(context.Background(), db, harness.OutboundControl.ControllerB, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 1, control.Len())
	})
}

func TestFetchInboundADEntityControllerPaths(t *testing.T) {
	t.Run("User", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.InboundControl.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			path, err := adAnalysis.FetchInboundADEntityControllerPaths(context.Background(), db, harness.InboundControl.ControlledUser)
			test.RequireNilErr(t, err)

			nodes := path.AllNodes().IDs()
			require.Equal(t, 5, len(nodes))
			require.Contains(t, nodes, harness.InboundControl.ControlledUser.ID)
			require.Contains(t, nodes, harness.InboundControl.GroupA.ID)
			require.Contains(t, nodes, harness.InboundControl.UserA.ID)
			require.Contains(t, nodes, harness.InboundControl.GroupB.ID)
			require.Contains(t, nodes, harness.InboundControl.UserD.ID)
		})
	})

	t.Run("Group", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.InboundControl.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			path, err := adAnalysis.FetchInboundADEntityControllerPaths(context.Background(), db, harness.InboundControl.ControlledGroup)
			test.RequireNilErr(t, err)

			nodes := path.AllNodes().IDs()
			require.Equal(t, 7, len(nodes))
			require.Contains(t, nodes, harness.InboundControl.ControlledGroup.ID)
			require.Contains(t, nodes, harness.InboundControl.UserE.ID)
			require.Contains(t, nodes, harness.InboundControl.UserF.ID)
			require.Contains(t, nodes, harness.InboundControl.GroupC.ID)
			require.Contains(t, nodes, harness.InboundControl.UserG.ID)
			require.Contains(t, nodes, harness.InboundControl.GroupD.ID)
			require.Contains(t, nodes, harness.InboundControl.UserH.ID)
		})
	})
}

func TestFetchInboundADEntityControllers(t *testing.T) {
	t.Run("User", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.InboundControl.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			control, err := adAnalysis.FetchInboundADEntityControllers(context.Background(), db, harness.InboundControl.ControlledUser, 0, 0)
			test.RequireNilErr(t, err)
			require.Equal(t, 4, control.Len())

			ids := control.IDs()
			require.Contains(t, ids, harness.InboundControl.UserD.ID)
			require.Contains(t, ids, harness.InboundControl.GroupB.ID)
			require.Contains(t, ids, harness.InboundControl.UserA.ID)
			require.Contains(t, ids, harness.InboundControl.GroupA.ID)

			control, err = adAnalysis.FetchInboundADEntityControllers(context.Background(), db, harness.InboundControl.ControlledUser, 0, 1)
			test.RequireNilErr(t, err)
			require.Equal(t, 1, control.Len())
		})
	})

	t.Run("Group", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.InboundControl.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			control, err := adAnalysis.FetchInboundADEntityControllers(context.Background(), db, harness.InboundControl.ControlledGroup, 0, 0)
			test.RequireNilErr(t, err)
			require.Equal(t, 6, control.Len())

			ids := control.IDs()
			require.Contains(t, ids, harness.InboundControl.GroupC.ID)
			require.Contains(t, ids, harness.InboundControl.GroupD.ID)
			require.Contains(t, ids, harness.InboundControl.UserE.ID)
			require.Contains(t, ids, harness.InboundControl.UserF.ID)
			require.Contains(t, ids, harness.InboundControl.UserG.ID)
			require.Contains(t, ids, harness.InboundControl.UserH.ID)

			control, err = adAnalysis.FetchInboundADEntityControllers(context.Background(), db, harness.InboundControl.ControlledGroup, 0, 1)
			test.RequireNilErr(t, err)
			require.Equal(t, 1, control.Len())
		})
	})
}

func TestCreateOUContainedPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OUHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.CreateOUContainedPathDelegate(ad.User)(tx, harness.OUHarness.OUA)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.OUHarness.OUA.ID)
		require.Contains(t, nodes, harness.OUHarness.UserA.ID)
		require.Contains(t, nodes, harness.OUHarness.OUC.ID)
		require.Contains(t, nodes, harness.OUHarness.UserB.ID)

		paths, err = adAnalysis.CreateOUContainedPathDelegate(ad.User)(tx, harness.OUHarness.OUB)
		test.RequireNilErr(t, err)
		nodes = paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.OUHarness.OUB.ID)
		require.Contains(t, nodes, harness.OUHarness.OUD.ID)
		require.Contains(t, nodes, harness.OUHarness.OUE.ID)
		require.Contains(t, nodes, harness.OUHarness.UserC.ID)
	})
}

func TestCreateOUContainedListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OUHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		contained, err := adAnalysis.CreateOUContainedListDelegate(ad.User)(tx, harness.OUHarness.OUA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, contained.Len())

		contained, err = adAnalysis.CreateOUContainedListDelegate(ad.User)(tx, harness.OUHarness.OUB, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 1, contained.Len())
	})
}

func TestFetchGroupMemberPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.MembershipHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.FetchGroupMemberPaths(tx, harness.MembershipHarness.GroupB)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.MembershipHarness.GroupB.ID)
		require.Contains(t, nodes, harness.MembershipHarness.UserA.ID)
		require.Contains(t, nodes, harness.MembershipHarness.ComputerA.ID)
	})
}

func TestFetchGroupMembers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.MembershipHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		members, err := adAnalysis.FetchGroupMembers(context.Background(), db, harness.MembershipHarness.GroupC, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 5, members.Len())
		require.Equal(t, 2, members.ContainingNodeKinds(ad.Computer).Len())
		require.Equal(t, 2, members.ContainingNodeKinds(ad.Group).Len())
		require.Equal(t, 1, members.ContainingNodeKinds(ad.User).Len())
	})
}

func TestFetchEntityGroupMembershipPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.MembershipHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchEntityGroupMembershipPaths(tx, harness.MembershipHarness.UserA)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.MembershipHarness.UserA.ID)
		require.Contains(t, nodes, harness.MembershipHarness.GroupB.ID)
		require.Contains(t, nodes, harness.MembershipHarness.GroupA.ID)
	})
}

func TestFetchEntityGroupMembership(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.MembershipHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		membership, err := adAnalysis.FetchEntityGroupMembership(tx, harness.MembershipHarness.UserA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 3, membership.Len())
	})
}

func TestCreateForeignEntityMembershipListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		members, err := adAnalysis.CreateForeignEntityMembershipListDelegate(ad.Group)(tx, harness.ForeignHarness.LocalDomain, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, members.Len())
		require.Equal(t, 1, members.ContainingNodeKinds(ad.Group).Len())

		members, err = adAnalysis.CreateForeignEntityMembershipListDelegate(ad.User)(tx, harness.ForeignHarness.LocalDomain, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 2, members.Len())
		require.Equal(t, 2, members.ContainingNodeKinds(ad.User).Len())
	})
}

func TestFetchCollectedDomains(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		domains, err := adAnalysis.FetchCollectedDomains(tx)
		test.RequireNilErr(t, err)
		for _, domain := range domains {
			collected, err := domain.Properties.Get(common.Collected.String()).Bool()
			test.RequireNilErr(t, err)
			require.True(t, collected)
		}
		require.Equal(t, harness.NumCollectedActiveDirectoryDomains, domains.Len())
		require.NotContains(t, domains.IDs(), harness.TrustDCSync.DomainC.ID)
	})
}

func TestCreateDomainTrustPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.CreateDomainTrustPathDelegate(graph.DirectionOutbound)(tx, harness.TrustDCSync.DomainA)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.TrustDCSync.DomainA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainB.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainC.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainD.ID)

		paths, err = adAnalysis.CreateDomainTrustPathDelegate(graph.DirectionInbound)(tx, harness.TrustDCSync.DomainA)

		test.RequireNilErr(t, err)
		nodes = paths.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.TrustDCSync.DomainA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainB.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainD.ID)
	})
}

func TestCreateDomainTrustListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		domains, err := adAnalysis.CreateDomainTrustListDelegate(graph.DirectionOutbound)(tx, harness.TrustDCSync.DomainA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 3, domains.Len())
		ids := domains.IDs()
		require.Contains(t, ids, harness.TrustDCSync.DomainB.ID)
		require.Contains(t, ids, harness.TrustDCSync.DomainC.ID)
		require.Contains(t, ids, harness.TrustDCSync.DomainD.ID)

		domains, err = adAnalysis.CreateDomainTrustListDelegate(graph.DirectionInbound)(tx, harness.TrustDCSync.DomainA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, domains.Len())
		ids = domains.IDs()
		require.Contains(t, ids, harness.TrustDCSync.DomainB.ID)
		require.Contains(t, ids, harness.TrustDCSync.DomainD.ID)
	})
}

func TestGetDCSyncers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	// XXX: Why does this need a WriteTransaction to run?
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		dcSyncers, err := analysis.GetDCSyncers(tx, harness.TrustDCSync.DomainA, true)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, len(dcSyncers))
		ids := make([]graph.ID, len(dcSyncers))
		for _, node := range dcSyncers {
			ids = append(ids, node.ID)
		}
		require.Contains(t, ids, harness.TrustDCSync.UserA.ID)
		require.Contains(t, ids, harness.TrustDCSync.UserB.ID)

		harness.TrustDCSync.UserA.Properties.Set(common.SystemTags.String(), ad.AdminTierZero)
		tx.UpdateNode(harness.TrustDCSync.UserA)

		dcSyncers, err = analysis.GetDCSyncers(tx, harness.TrustDCSync.DomainA, true)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, len(dcSyncers))
		ids = make([]graph.ID, len(dcSyncers))
		for _, node := range dcSyncers {
			ids = append(ids, node.ID)
		}
		require.NotContains(t, ids, harness.TrustDCSync.UserA.ID)
		require.Contains(t, ids, harness.TrustDCSync.UserB.ID)
	})
}

func TestFetchDCSyncers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		dcSyncers, err := adAnalysis.FetchDCSyncers(tx, harness.TrustDCSync.DomainA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, dcSyncers.Len())

		nodes := dcSyncers.IDs()
		require.Contains(t, nodes, harness.TrustDCSync.UserB.ID)
		require.Contains(t, nodes, harness.TrustDCSync.UserA.ID)
	})
}

func TestFetchDCSyncerPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchDCSyncerPaths(tx, harness.TrustDCSync.DomainA)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 5, len(nodes))
		require.Contains(t, nodes, harness.TrustDCSync.DomainA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.UserA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.GroupA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.GroupB.ID)
		require.Contains(t, nodes, harness.TrustDCSync.UserB.ID)
	})
}

func TestCreateForeignEntityMembershipPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.CreateForeignEntityMembershipPathDelegate(ad.Group)(tx, harness.ForeignHarness.LocalDomain)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 2, len(nodes))
		require.Contains(t, nodes, harness.ForeignHarness.ForeignGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.LocalGroup.ID)

		paths, err = adAnalysis.CreateForeignEntityMembershipPathDelegate(ad.User)(tx, harness.ForeignHarness.LocalDomain)

		test.RequireNilErr(t, err)
		nodes = paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.ForeignHarness.ForeignGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserA.ID)
		require.Contains(t, nodes, harness.ForeignHarness.LocalGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserB.ID)
	})
}

func TestFetchForeignAdmins(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.FetchForeignAdmins(tx, harness.ForeignHarness.LocalDomain, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, admins.Len())
		require.Equal(t, 2, admins.ContainingNodeKinds(ad.User).Len())
	})
}

func TestFetchForeignAdminPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchForeignAdminPaths(tx, harness.ForeignHarness.LocalDomain)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 5, len(nodes))
		require.Contains(t, nodes, harness.ForeignHarness.LocalComputer.ID)
		require.Contains(t, nodes, harness.ForeignHarness.LocalGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserA.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserB.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignGroup.ID)
	})
}

func TestFetchForeignGPOControllers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.FetchForeignGPOControllers(tx, harness.ForeignHarness.LocalDomain, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, admins.Len())
		require.Equal(t, 1, admins.ContainingNodeKinds(ad.User).Len())
		require.Equal(t, 1, admins.ContainingNodeKinds(ad.Group).Len())
	})
}

func TestFetchForeignGPOControllerPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchForeignGPOControllerPaths(tx, harness.ForeignHarness.LocalDomain)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserA.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.LocalGPO.ID)
	})
}

func TestFetchAllEnforcedGPOs(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		gpos, err := adAnalysis.FetchAllEnforcedGPOs(context.Background(), db, graph.NewNodeSet(harness.GPOEnforcement.OrganizationalUnitD))

		test.RequireNilErr(t, err)
		require.Equal(t, 2, gpos.Len())

		gpos, err = adAnalysis.FetchAllEnforcedGPOs(context.Background(), db, graph.NewNodeSet(harness.GPOEnforcement.OrganizationalUnitC))

		test.RequireNilErr(t, err)
		require.Equal(t, 1, gpos.Len())
	})
}

func TestFetchEntityLinkedGPOList(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		gpos, err := adAnalysis.FetchEntityLinkedGPOList(tx, harness.GPOEnforcement.Domain, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, gpos.Len())
	})
}

func TestFetchEntityLinkedGPOPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchEntityLinkedGPOPaths(tx, harness.GPOEnforcement.Domain)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
	})
}

func TestFetchLocalGroupCompleteness(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Completeness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		completeness, err := adAnalysis.FetchLocalGroupCompleteness(tx, harness.Completeness.DomainSid)

		test.RequireNilErr(t, err)
		require.Equal(t, .5, completeness)
	})
}

func TestFetchUserSessionCompleteness(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Completeness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		completeness, err := adAnalysis.FetchUserSessionCompleteness(tx, harness.Completeness.DomainSid)

		test.RequireNilErr(t, err)
		require.Equal(t, .5, completeness)
	})
}
