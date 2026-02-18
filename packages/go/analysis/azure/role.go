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

package azure

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/trace"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

func NewRoleEntityDetails(node *graph.Node) RoleDetails {
	return RoleDetails{
		Node: FromGraphNode(node),
	}
}

func RoleEntityDetails(ctx context.Context, db graph.Database, objectID string, hydrateCounts bool) (RoleDetails, error) {
	var details RoleDetails

	return details, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details = NewRoleEntityDetails(node)
			if hydrateCounts {
				if details, err = PopulateRoleEntityApprovers(tx, node, details); err != nil {
					return err
				}
				details, err = PopulateRoleEntityDetailsCounts(tx, node, details)
			}
			return err
		}
	})
}

func PopulateRoleEntityApprovers(tx graph.Transaction, node *graph.Node, details RoleDetails) (RoleDetails, error) {
	if approvers, err := FetchRoleApprovers(tx, node, 0, 0); err != nil {
		return details, err
	} else {
		details.Approvers = approvers.Len()
		return details, nil
	}
}

func PopulateRoleEntityDetailsCounts(tx graph.Transaction, node *graph.Node, details RoleDetails) (RoleDetails, error) {
	if activeAssignments, err := FetchEntityActiveAssignments(tx, node, 0, 0); err != nil {
		return details, err
	} else {
		details.ActiveAssignments = activeAssignments.Len()
	}

	if pimAssignments, err := FetchEntityPIMAssignments(tx, node, 0, 0); err != nil {
		return details, err
	} else {
		details.PIMAssignments = pimAssignments.Len()
	}

	return details, nil
}

type RoleAssignmentMap map[graph.ID]map[string]struct{}

func (s RoleAssignmentMap) UserHasRoles(user *graph.Node) bool {
	_, hasAssignments := s[user.ID]
	return hasAssignments
}

func (s RoleAssignmentMap) HasRole(id graph.ID, roleTemplateIDs ...string) bool {
	if roleAssignments, hasAssignments := s[id]; hasAssignments {
		for _, roleTemplateID := range roleTemplateIDs {
			if _, hasRole := roleAssignments[roleTemplateID]; hasRole {
				return true
			}
		}
	}

	return false
}

type RoleAssignments struct {
	TenantPrincipals              graph.NodeKindSet
	users                         cardinality.ImmutableDuplex[uint64]
	usersWithAnyRole              cardinality.ImmutableDuplex[uint64]
	usersWithoutRoles             cardinality.ImmutableDuplex[uint64]
	servicePrincipals             cardinality.ImmutableDuplex[uint64]
	RoleMap                       map[string]cardinality.Duplex[uint64]
	RoleAssignableGroupMembership cardinality.Duplex[uint64]
}

func (s RoleAssignments) GetNodeKindSet(bm cardinality.Duplex[uint64]) graph.NodeKindSet {
	result := graph.NewNodeKindSet()

	bm.Each(func(nextID uint64) bool {
		node := s.TenantPrincipals.GetNode(graph.ID(nextID))
		result.Add(node)

		return true
	})

	return result
}

func (s RoleAssignments) GetNodeSet(bm cardinality.Duplex[uint64]) graph.NodeSet {
	return s.GetNodeKindSet(bm).AllNodes()
}

func (s RoleAssignments) ServicePrincipals() cardinality.ImmutableDuplex[uint64] {
	return s.servicePrincipals
}

func (s RoleAssignments) Users() cardinality.ImmutableDuplex[uint64] {
	return s.users
}

func (s RoleAssignments) UsersWithAnyRole() cardinality.ImmutableDuplex[uint64] {
	return s.usersWithAnyRole
}

func (s RoleAssignments) UsersWithoutRoles() cardinality.ImmutableDuplex[uint64] {
	return s.usersWithoutRoles
}

func (s RoleAssignments) UsersWithRole(roleTemplateIDs ...string) cardinality.Duplex[uint64] {
	result := s.PrincipalsWithRole(roleTemplateIDs...)
	result.And(s.Users())
	return result
}

func (s RoleAssignments) ServicePrincipalsWithRole(roleTemplateIDs ...string) cardinality.Duplex[uint64] {
	result := s.PrincipalsWithRole(roleTemplateIDs...)
	result.And(s.ServicePrincipals())
	return result
}

func (s RoleAssignments) UsersWithRolesExclusive(roleTemplateIDs ...string) cardinality.Duplex[uint64] {
	result := s.PrincipalsWithRolesExclusive(roleTemplateIDs...)
	result.And(s.Users())
	return result
}

func (s RoleAssignments) UsersWithRoleAssignableGroupMembership() cardinality.Duplex[uint64] {
	// this field is a bitmap of all user IDs who are members of role assignable groups.
	return s.RoleAssignableGroupMembership
}

// PrincipalsWithRole returns a roaring bitmap of principals that have been assigned one or more of the matching roles from list of role template IDs
func (s RoleAssignments) PrincipalsWithRole(roleTemplateIDs ...string) cardinality.Duplex[uint64] {
	result := cardinality.NewBitmap64()
	for _, roleTemplateID := range roleTemplateIDs {
		if bitmap, ok := s.RoleMap[roleTemplateID]; ok {
			result.Or(bitmap)
		}
	}
	return result
}

// PrincipalsWithRole returns a roaring bitmap of principals that have been assigned one or more of the matching roles from list of role template IDs but excluding principals with non-matching roles
func (s RoleAssignments) PrincipalsWithRolesExclusive(roleTemplateIDs ...string) cardinality.Duplex[uint64] {
	var (
		result             = cardinality.NewBitmap64()
		excludedPrincipals = cardinality.NewBitmap64()
	)
	for roleID, bitmap := range s.RoleMap {
		if slices.Contains(roleTemplateIDs, roleID) {
			result.Or(bitmap)
		} else {
			excludedPrincipals.Or(bitmap)
		}
	}
	result.AndNot(excludedPrincipals)
	return result
}

// NodesWithRolesExclusive will return nodes that *only* have a role/roles listed and exclude nodes that have other roles
func (s RoleAssignments) NodesWithRolesExclusive(roleTemplateIDs ...string) graph.NodeKindSet {
	bm := s.PrincipalsWithRolesExclusive(roleTemplateIDs...)
	return s.GetNodeKindSet(bm)
}

func (s RoleAssignments) NodeHasRole(id graph.ID, roleTemplateIDs ...string) bool {
	for _, roleID := range roleTemplateIDs {
		if bm, ok := s.RoleMap[roleID]; ok {
			if bm.Contains(id.Uint64()) {
				return true
			}
		}
	}
	return false
}

func NewTenantRoleAssignments(tenant *graph.Node, tenantPrincipals graph.NodeKindSet, roleAssignableGroupMembership cardinality.Duplex[uint64], roleMap map[string]cardinality.Duplex[uint64]) RoleAssignments {
	var (
		users             = tenantPrincipals.Get(azure.User).IDBitmap()
		usersWithAnyRole  = cardinality.NewBitmap64()
		usersWithoutRoles = cardinality.NewBitmap64()
		servicePrincipals = tenantPrincipals.Get(azure.ServicePrincipal).IDBitmap()
	)

	// Calculate users with any role first
	for _, bitmap := range roleMap {
		usersWithAnyRole.Or(bitmap)
	}

	usersWithAnyRole.And(users)

	// Calculate users without roles next
	usersWithoutRoles.Or(users)
	usersWithoutRoles.AndNot(usersWithAnyRole)

	slog.Info("Tenant Role Assignment Details",
		slog.Uint64("num_users", users.Cardinality()),
		slog.Uint64("num_service_principals", servicePrincipals.Cardinality()),
	)

	return RoleAssignments{
		TenantPrincipals:              tenantPrincipals,
		users:                         users,
		usersWithAnyRole:              usersWithAnyRole,
		usersWithoutRoles:             usersWithoutRoles,
		servicePrincipals:             servicePrincipals,
		RoleMap:                       roleMap,
		RoleAssignableGroupMembership: roleAssignableGroupMembership,
	}
}

func FetchTenantRoleAssignments(ctx context.Context, db graph.Database, tenant *graph.Node) (RoleAssignments, error) {
	defer trace.Function(ctx, "FetchTenantRoleAssignments")()

	var roleAssignments RoleAssignments

	if !IsTenantNode(tenant) {
		return RoleAssignments{}, fmt.Errorf("cannot initialize tenant role assignments - node %d must be of kind %s", tenant.ID, azure.Tenant)
	}

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if tenantPrincipalsNodeSet, err := TenantPrincipals(tx, tenant); err != nil && !graph.IsErrNotFound(err) {
			return err
		} else if roles, err := TenantRoles(tx, tenant); err != nil {
			return err
		} else {
			var (
				tenantPrincipalsNodeKindSet   = tenantPrincipalsNodeSet.KindSet()
				roleAssignableGroupMembership = cardinality.NewBitmap64()
				roleMap                       = map[string]cardinality.Duplex[uint64]{}
			)

			// for each of the role assignable groups returned, fetch the users who are members
			for _, group := range tenantPrincipalsNodeKindSet.Get(azure.Group) {
				if members, err := FetchRoleAssignableGroupMembersUsers(tx, group, 0, 0); err != nil {
					return err
				} else {
					// set all users who have role assignable group membership
					roleAssignableGroupMembership.Or(members.IDBitmap())
				}
			}

			for _, node := range roles {
				if roleTemplateID, err := node.Properties.Get(azure.RoleTemplateID.String()).String(); err != nil {
					if !graph.IsErrPropertyNotFound(err) {
						return err
					}
				} else if members, err := RoleMembers(tx, tenant, roleTemplateID); err != nil {
					if !graph.IsErrNotFound(err) {
						return err
					}
				} else {
					roleMap[roleTemplateID] = members.IDBitmap()
				}
			}

			roleAssignments = NewTenantRoleAssignments(tenant, tenantPrincipalsNodeKindSet, roleAssignableGroupMembership, roleMap)
			return nil
		}
	}); err != nil {
		return RoleAssignments{}, err
	}

	return roleAssignments, nil
}

// RoleMembers returns the NodeSet of members for a given set of roles
func RoleMembers(tx graph.Transaction, tenant *graph.Node, roleTemplateIDs ...string) (graph.NodeSet, error) {
	if tenantRoles, err := TenantRoles(tx, tenant, roleTemplateIDs...); err != nil {
		return nil, err
	} else if members, err := roleMembers(tx, tenantRoles); err != nil {
		return nil, err
	} else {
		for _, role := range tenantRoles {
			members.Remove(role.ID)
		}
		return members, nil
	}
}

func roleMembers(tx graph.Transaction, tenantRoles graph.NodeSet, additionalRelationships ...graph.Kind) (graph.NodeSet, error) {
	members := graph.NewNodeSet()

	for _, tenantRole := range tenantRoles {
		if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      tenantRole,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.And(
					query.KindIn(query.Relationship(), append(additionalRelationships, azure.MemberOf, azure.HasRole)...),
				)
			},
			DescentFilter: roleDescentFilter,
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(azure.User, azure.Group, azure.ServicePrincipal)
			},
		}); err != nil {
			return nil, err
		} else {
			// TODO: This could be more optimal by iterating in place instead of aggregating all results
			members.AddSet(paths.AllNodes())
		}
	}

	return members, nil
}

// RoleMembersWithGrants returns the NodeSet of members for a given set of roles, including those members who may be able to grant themselves one of the given roles
// NOTE: The current implementation also includes the role nodes in the returned set. It may be worth considering removing those nodes from the set if doing so doesn't break tier zero/high value assignment
func RoleMembersWithGrants(tx graph.Transaction, tenant *graph.Node, roleTemplateIDs ...string) (graph.NodeSet, error) {
	defer measure.LogAndMeasureWithThreshold(slog.LevelInfo, "RoleMembersWithGrants", slog.Int64("tenant_id", tenant.ID.Int64()))()

	if tenantRoles, err := TenantRoles(tx, tenant, roleTemplateIDs...); err != nil {
		return nil, err
	} else {
		return roleMembers(tx, tenantRoles, azure.GrantSelf)
	}
}
