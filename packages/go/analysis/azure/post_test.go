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

package azure_test

import (
	"context"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/bloodhoundad/azurehound/v2/constants"
	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/util/size"
	azschema "github.com/specterops/bloodhound/graphschema/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/specterops/bloodhound/analysis/azure"
)

var (
	user  = graph.NewNode(0, graph.NewProperties(), azschema.User)
	user2 = graph.NewNode(1, graph.NewProperties(), azschema.User)
	group = graph.NewNode(2, graph.NewProperties(), azschema.Group)
	app   = graph.NewNode(3, graph.NewProperties(), azschema.App)
)

// setupRoleAssignments is used to create a testable RoleAssignments struct. It is used in all RoleAssignments tests
// and may require adjusting tests if modified
func setupRoleAssignments() azure.RoleAssignments {
	roleMap := map[string]*roaring.Bitmap{
		constants.GlobalAdministratorRoleID:   roaring.New(),
		constants.ReportsReaderRoleID:         roaring.New(),
		constants.HelpdeskAdministratorRoleID: roaring.New(),
		constants.PartnerTier1SupportRoleID:   roaring.New(),
	}
	roleMap[constants.GlobalAdministratorRoleID].Add(uint32(user.ID))
	roleMap[constants.ReportsReaderRoleID].Add(uint32(group.ID))
	roleMap[constants.HelpdeskAdministratorRoleID].Add(uint32(group.ID))
	roleMap[constants.PartnerTier1SupportRoleID].Add(uint32(app.ID))

	return azure.RoleAssignments{
		// user2 has no roles! this is intentional
		Principals: graph.NewNodeSet(user, user2, group, app).KindSet(),
		RoleMap:    roleMap,
	}
}

func TestRoleAssignments_NodeHasRole(t *testing.T) {
	assignments := setupRoleAssignments()
	assert.True(t, assignments.NodeHasRole(user.ID, azschema.CompanyAdministratorRole))
	assert.False(t, assignments.NodeHasRole(user.ID, azschema.HelpdeskAdministratorRole))
	assert.True(t, assignments.NodeHasRole(group.ID, azschema.ReportsReaderRole))
	assert.True(t, assignments.NodeHasRole(group.ID, azschema.HelpdeskAdministratorRole))
	assert.False(t, assignments.NodeHasRole(group.ID, azschema.PartnerTier1SupportRole))
}

func TestRoleAssignments_UsersWithoutRoles(t *testing.T) {
	assignments := setupRoleAssignments()
	assert.False(t, assignments.UsersWithoutRoles().Contains(uint32(user.ID)))
	assert.True(t, assignments.UsersWithoutRoles().Contains(uint32(user2.ID)))
}

func TestRoleAssignments_NodesWithRole(t *testing.T) {
	assignments := setupRoleAssignments()
	assert.True(t, assignments.PrincipalsWithRole(constants.ReportsReaderRoleID, constants.GlobalAdministratorRoleID).Contains(uint32(user.ID)))
	assert.True(t, assignments.PrincipalsWithRole(constants.ReportsReaderRoleID, constants.GlobalAdministratorRoleID).Contains(uint32(group.ID)))
	assert.True(t, assignments.PrincipalsWithRole(constants.ReportsReaderRoleID, constants.HelpdeskAdministratorRoleID).Contains(uint32(group.ID)))
	assert.False(t, assignments.PrincipalsWithRole(constants.ReportsReaderRoleID).Contains(uint32(user.ID)))
}

func TestRoleAssignments_NodesWithRolesExclusive(t *testing.T) {
	assignments := setupRoleAssignments()
	assert.Equal(t, user, assignments.NodesWithRolesExclusive(azschema.ReportsReaderRole, azschema.CompanyAdministratorRole).Get(azschema.User).Get(user.ID))
	assert.Equal(t, graph.EmptyNodeSet().Get(0), assignments.NodesWithRolesExclusive(azschema.ReportsReaderRole, azschema.CompanyAdministratorRole).Get(azschema.Group).Get(group.ID))
	assert.Equal(t, group, assignments.NodesWithRolesExclusive(azschema.ReportsReaderRole, azschema.HelpdeskAdministratorRole).Get(azschema.Group).Get(group.ID))
	assert.Equal(t, graph.EmptyNodeSet().Get(0), assignments.NodesWithRolesExclusive(azschema.ReportsReaderRole).Get(azschema.User).Get(user.ID))
}

func TestTenantRoles(t *testing.T) {
	var (
		ctrl       = gomock.NewController(t)
		mockTx     = graph_mocks.NewMockTransaction(ctrl)
		stubTenant = &graph.Node{
			ID:         1,
			Kinds:      graph.Kinds{azschema.Entity, azschema.Tenant},
			Properties: &graph.Properties{},
		}
		stubIntuneAdminRole = &graph.Node{
			ID:    2,
			Kinds: graph.Kinds{azschema.Entity, azschema.Role},
			Properties: &graph.Properties{
				Map: map[string]any{
					"templateid": constants.IntuneAdministratorRoleID,
				},
			},
		}
	)
	defer ctrl.Finish()

	mockRelQuery1 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockTx.EXPECT().Relationships().Return(mockRelQuery1).Times(1)

	mockRelQuery2 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockRelQuery1.EXPECT().Filterf(gomock.AssignableToTypeOf(func() graph.Criteria { return nil })).Return(mockRelQuery2)
	mockRelQuery2.EXPECT().
		FetchDirection(gomock.Any(), gomock.AssignableToTypeOf(func(graph.Cursor[graph.DirectionalResult]) error { return nil })).
		DoAndReturn(func(_ any, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
			mockCursor := graph_mocks.NewMockCursor[graph.DirectionalResult](ctrl)
			c := make(chan graph.DirectionalResult, 1)
			go func() {
				defer close(c)
				c <- graph.DirectionalResult{Node: stubIntuneAdminRole}
			}()
			mockCursor.EXPECT().Chan().Return(c)
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})

	roles, err := azure.TenantRoles(mockTx, stubTenant, constants.IntuneAdministratorRoleID)
	require.Nil(t, err)
	assert.Equal(t, 1, roles.Len())
	assert.Contains(t, roles.Slice(), stubIntuneAdminRole)
}

func TestRoleMembers(t *testing.T) {
	var (
		ctrl       = gomock.NewController(t)
		mockTx     = graph_mocks.NewMockTransaction(ctrl)
		stubTenant = &graph.Node{
			ID:         1,
			Kinds:      graph.Kinds{azschema.Entity, azschema.Tenant},
			Properties: &graph.Properties{},
		}
		stubIntuneAdminRole = &graph.Node{
			ID:    2,
			Kinds: graph.Kinds{azschema.Entity, azschema.Role},
			Properties: &graph.Properties{
				Map: map[string]any{
					"templateid": constants.IntuneAdministratorRoleID,
				},
			},
		}
		stubIntuneAdmin1 = &graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azschema.Entity, azschema.User},
		}
		stubIntuneAdmin2 = &graph.Node{
			ID:    4,
			Kinds: graph.Kinds{azschema.Entity, azschema.User},
		}
		stubIntuneAdmin3 = &graph.Node{
			ID:    5,
			Kinds: graph.Kinds{azschema.Entity, azschema.User},
		}
	)
	defer ctrl.Finish()

	mockTx.EXPECT().TraversalMemoryLimit().Return(1 * size.Gibibyte).AnyTimes()

	mockRelQuery1 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockRelQuery2 := graph_mocks.NewMockRelationshipQuery(ctrl)
	gomock.InOrder(
		mockTx.EXPECT().Relationships().Return(mockRelQuery1),
		mockTx.EXPECT().Relationships().Return(mockRelQuery2).AnyTimes(),
	)

	mockFilterf1 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockRelQuery1.EXPECT().Filterf(gomock.AssignableToTypeOf(func() graph.Criteria { return nil })).Return(mockFilterf1)
	mockFilterf1.EXPECT().
		FetchDirection(gomock.Any(), gomock.AssignableToTypeOf(func(graph.Cursor[graph.DirectionalResult]) error { return nil })).
		DoAndReturn(func(_ any, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
			mockCursor := graph_mocks.NewMockCursor[graph.DirectionalResult](ctrl)
			c := make(chan graph.DirectionalResult, 1)
			go func() {
				defer close(c)
				c <- graph.DirectionalResult{Node: stubIntuneAdminRole}
			}()
			mockCursor.EXPECT().Chan().Return(c)
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})

	mockFilterf2 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockRelQuery2.EXPECT().Filterf(gomock.AssignableToTypeOf(func() graph.Criteria { return nil })).Return(mockFilterf2).AnyTimes()
	mockFilterf2.EXPECT().
		FetchDirection(gomock.Any(), gomock.AssignableToTypeOf(func(graph.Cursor[graph.DirectionalResult]) error { return nil })).
		DoAndReturn(func(_ any, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
			mockCursor := graph_mocks.NewMockCursor[graph.DirectionalResult](ctrl)
			c := make(chan graph.DirectionalResult, 3)
			go func() {
				defer close(c)
				c <- graph.DirectionalResult{
					Node: stubIntuneAdmin1,
					Relationship: &graph.Relationship{
						ID:      101,
						Kind:    azschema.HasRole,
						StartID: stubIntuneAdmin1.ID,
						EndID:   stubIntuneAdminRole.ID,
					},
				}
				c <- graph.DirectionalResult{
					Node: stubIntuneAdmin2,
					Relationship: &graph.Relationship{
						ID:      102,
						Kind:    azschema.HasRole,
						StartID: stubIntuneAdmin2.ID,
						EndID:   stubIntuneAdminRole.ID,
					},
				}
				c <- graph.DirectionalResult{
					Node: stubIntuneAdmin3,
					Relationship: &graph.Relationship{
						ID:      103,
						Kind:    azschema.HasRole,
						StartID: stubIntuneAdmin3.ID,
						EndID:   stubIntuneAdminRole.ID,
					},
				}
			}()
			mockCursor.EXPECT().Chan().Return(c)
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		}).AnyTimes()

	members, err := azure.RoleMembers(mockTx, stubTenant, constants.IntuneAdministratorRoleID)
	require.Nil(t, err)
	assert.Equal(t, 3, members.Len())
	assert.NotContains(t, members.Slice(), stubIntuneAdminRole)
	assert.Contains(t, members.Slice(), stubIntuneAdmin1)
	assert.Contains(t, members.Slice(), stubIntuneAdmin2)
	assert.Contains(t, members.Slice(), stubIntuneAdmin3)
}

func TestFetchTenants(t *testing.T) {
	var (
		ctrl       = gomock.NewController(t)
		mockDB     = graph_mocks.NewMockDatabase(ctrl)
		stubTenant = &graph.Node{
			ID:         1,
			Kinds:      graph.Kinds{azschema.Entity, azschema.Tenant},
			Properties: &graph.Properties{},
		}
	)

	mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.AssignableToTypeOf(func(graph.Transaction) error { return nil }), gomock.Any()).DoAndReturn(func(_ any, delegate graph.TransactionDelegate, _ ...any) error {
		var (
			mockTx        = graph_mocks.NewMockTransaction(ctrl)
			mockNodeQuery = graph_mocks.NewMockNodeQuery(ctrl)
			mockFilterf   = graph_mocks.NewMockNodeQuery(ctrl)
		)
		mockTx.EXPECT().Nodes().Return(mockNodeQuery)
		mockNodeQuery.EXPECT().Filterf(gomock.Any()).Return(mockFilterf)
		mockFilterf.EXPECT().Fetch(gomock.AssignableToTypeOf(func(graph.Cursor[*graph.Node]) error { return nil })).DoAndReturn(func(delegate func(graph.Cursor[*graph.Node]) error) error {
			mockCursor := graph_mocks.NewMockCursor[*graph.Node](ctrl)
			c := make(chan *graph.Node, 1)
			go func() {
				defer close(c)
				c <- stubTenant
			}()

			mockCursor.EXPECT().Chan().Return(c)
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})

		return delegate(mockTx)
	})

	tenants, err := azure.FetchTenants(context.Background(), mockDB)
	require.Nil(t, err)
	assert.Equal(t, 1, tenants.Len())
	assert.Contains(t, tenants.Slice(), stubTenant)
}

func TestEndNodes(t *testing.T) {
	var (
		ctrl         = gomock.NewController(t)
		mockTx       = graph_mocks.NewMockTransaction(ctrl)
		mockRelQuery = graph_mocks.NewMockRelationshipQuery(ctrl)
		mockFilterf  = graph_mocks.NewMockRelationshipQuery(ctrl)
		stubTenant   = &graph.Node{
			ID:         1,
			Kinds:      graph.Kinds{azschema.Entity, azschema.Tenant},
			Properties: &graph.Properties{},
		}
		stubDevice1 = &graph.Node{
			ID:    1,
			Kinds: graph.Kinds{azschema.Entity, azschema.Device},
		}
		stubDevice2 = &graph.Node{
			ID:    2,
			Kinds: graph.Kinds{azschema.Entity, azschema.Device},
		}
		stubDevice3 = &graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azschema.Entity, azschema.Device},
		}
	)
	mockTx.EXPECT().Relationships().Return(mockRelQuery)
	mockRelQuery.EXPECT().Filterf(gomock.Any()).Return(mockFilterf)
	mockFilterf.EXPECT().
		FetchDirection(gomock.Any(), gomock.AssignableToTypeOf(func(graph.Cursor[graph.DirectionalResult]) error { return nil })).
		DoAndReturn(func(_ any, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
			mockCursor := graph_mocks.NewMockCursor[graph.DirectionalResult](ctrl)
			c := make(chan graph.DirectionalResult, 3)
			go func() {
				defer close(c)
				c <- graph.DirectionalResult{Node: stubDevice1}
				c <- graph.DirectionalResult{Node: stubDevice2}
				c <- graph.DirectionalResult{Node: stubDevice3}
			}()
			mockCursor.EXPECT().Chan().Return(c)
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})

	nodes, err := azure.EndNodes(mockTx, stubTenant, azschema.Contains, azschema.Device)
	require.Nil(t, err)
	assert.Equal(t, 3, nodes.Len())
	assert.NotContains(t, nodes.Slice(), stubTenant)
	assert.Contains(t, nodes.Slice(), stubDevice1)
	assert.Contains(t, nodes.Slice(), stubDevice2)
	assert.Contains(t, nodes.Slice(), stubDevice3)
}
