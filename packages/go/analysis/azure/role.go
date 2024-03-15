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
	"github.com/RoaringBitmap/roaring"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/log"
	"slices"

	"github.com/specterops/bloodhound/dawgs/graph"
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
				details, err = PopulateRoleEntityDetailsCounts(tx, node, details)
			}
			return err
		}
	})
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
	Principals graph.NodeKindSet
	RoleMap    map[string]*roaring.Bitmap
}

func (s RoleAssignments) GetNodeKindSet(bm *roaring.Bitmap) graph.NodeKindSet {
	var (
		result = graph.NewNodeKindSet()
		iter   = bm.Iterator()
	)
	for iter.HasNext() {
		node := s.Principals.GetNode(graph.ID(iter.Next()))
		result.Add(node)
	}
	return result
}

func (s RoleAssignments) GetNodeSet(bm *roaring.Bitmap) graph.NodeSet {
	return s.GetNodeKindSet(bm).AllNodes()
}

func (s RoleAssignments) Users() *roaring.Bitmap {
	return s.Principals.Get(azure.User).IDBitmap()
}

func (s RoleAssignments) UsersWithAnyRole() *roaring.Bitmap {
	users := s.Users()

	principalsWithRoles := roaring.New()
	for _, bitmap := range s.RoleMap {
		principalsWithRoles.Or(bitmap)
	}
	principalsWithRoles.And(users)
	return principalsWithRoles
}

func (s RoleAssignments) UsersWithoutRoles() *roaring.Bitmap {
	result := s.Users()
	result.AndNot(s.UsersWithAnyRole())
	return result
}

func (s RoleAssignments) UsersWithRole(roleTemplateIDs ...string) *roaring.Bitmap {
	result := s.PrincipalsWithRole(roleTemplateIDs...)
	result.And(s.Users())
	return result
}

func (s RoleAssignments) UsersWithRolesExclusive(roleTemplateIDs ...string) *roaring.Bitmap {
	result := s.PrincipalsWithRolesExclusive(roleTemplateIDs...)
	result.And(s.Users())
	return result
}

// PrincipalsWithRole returns a roaring bitmap of principals that have been assigned one or more of the matching roles from list of role template IDs
func (s RoleAssignments) PrincipalsWithRole(roleTemplateIDs ...string) *roaring.Bitmap {
	result := roaring.New()
	for _, roleTemplateID := range roleTemplateIDs {
		if bitmap, ok := s.RoleMap[roleTemplateID]; ok {
			result.Or(bitmap)
		}
	}
	return result
}

// PrincipalsWithRole returns a roaring bitmap of principals that have been assigned one or more of the matching roles from list of role template IDs but excluding principals with non-matching roles
func (s RoleAssignments) PrincipalsWithRolesExclusive(roleTemplateIDs ...string) *roaring.Bitmap {
	var (
		result             = roaring.New()
		excludedPrincipals = roaring.New()
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
			if bm.Contains(uint32(id)) {
				return true
			}
		}
	}
	return false
}

func initTenantRoleAssignments(tx graph.Transaction, tenant *graph.Node) (RoleAssignments, error) {
	if !IsTenantNode(tenant) {
		return RoleAssignments{}, fmt.Errorf("cannot initialize tenant role assignments - node %d must be of kind %s", tenant.ID, azure.Tenant)
	} else if roleMembers, err := TenantPrincipals(tx, tenant); err != nil && !graph.IsErrNotFound(err) {
		return RoleAssignments{}, err
	} else {
		return RoleAssignments{
			Principals: roleMembers.KindSet(),
			RoleMap:    make(map[string]*roaring.Bitmap),
		}, nil
	}
}

func TenantRoleAssignments(ctx context.Context, db graph.Database, tenant *graph.Node) (RoleAssignments, error) {
	var roleAssignments RoleAssignments
	return roleAssignments, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedRoleAssignments, err := initTenantRoleAssignments(tx, tenant); err != nil {
			return err
		} else if roles, err := TenantRoles(tx, tenant); err != nil {
			return err
		} else {
			return roles.KindSet().EachNode(func(node *graph.Node) error {
				if roleTemplateID, err := node.Properties.Get(azure.RoleTemplateID.String()).String(); err != nil {
					if !graph.IsErrPropertyNotFound(err) {
						return err
					}
				} else if members, err := RoleMembers(tx, tenant, roleTemplateID); err != nil {
					if !graph.IsErrNotFound(err) {
						return err
					}
				} else {
					fetchedRoleAssignments.RoleMap[roleTemplateID] = members.IDBitmap()
				}
				roleAssignments = fetchedRoleAssignments
				return nil
			})
		}
	})
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
	defer log.LogAndMeasure(log.LevelInfo, "Tenant %d RoleMembersWithGrants", tenant.ID)()

	if tenantRoles, err := TenantRoles(tx, tenant, roleTemplateIDs...); err != nil {
		return nil, err
	} else {
		return roleMembers(tx, tenantRoles, azure.GrantSelf)
	}
}
