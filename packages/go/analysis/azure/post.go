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
	"slices"
	"strings"

	"github.com/RoaringBitmap/roaring"
	"github.com/bloodhoundad/azurehound/v2/constants"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

func AddMemberAllGroupsTargetRoles() []string {
	return []string{
		azure.CompanyAdministratorRole,
		azure.PrivilegedRoleAdministratorRole,
	}
}

func AddMemberGroupNotRoleAssignableTargetRoles() []string {
	return []string{
		azure.GroupsAdministratorRole,
		azure.DirectoryWritersRole,
		azure.IdentityGovernanceAdministrator,
		azure.UserAccountAdministratorRole,
		azure.IntuneServiceAdministratorRole,
		azure.KnowledgeAdministratorRole,
		azure.KnowledgeManagerRole,
	}
}

func ResetPasswordRoleIDs() []string {
	return []string{
		constants.GlobalAdministratorRoleID,
		constants.PrivilegedAuthenticationAdministratorRoleID,
		constants.PartnerTier2SupportRoleID,
		constants.HelpdeskAdministratorRoleID,
		constants.AuthenticationAdministratorRoleID,
		constants.UserAdministratorRoleID,
		constants.PasswordAdministratorRoleID,
		constants.PartnerTier1SupportRoleID,
	}
}

func HelpdeskAdministratorPasswordResetTargetRoles() []string {
	return []string{
		azure.ReportsReaderRole,
		azure.MessageCenterReaderRole,
		azure.HelpdeskAdministratorRole,
		azure.GuestInviterRole,
		azure.DirectoryReadersRole,
		azure.PasswordAdministratorRole,
		azure.UsageSummaryReportsReaderRole,
	}
}

func AuthenticationAdministratorPasswordResetTargetRoles() []string {
	return []string{
		azure.AuthenticationAdministratorRole,
		azure.ReportsReaderRole,
		azure.MessageCenterReaderRole,
		azure.GuestInviterRole,
		azure.DirectoryReadersRole,
		azure.PasswordAdministratorRole,
		azure.UsageSummaryReportsReaderRole,
	}
}

func UserAdministratorPasswordResetTargetRoles() []string {
	return []string{
		azure.UserAccountAdministratorRole,
		azure.ReportsReaderRole,
		azure.MessageCenterReaderRole,
		azure.HelpdeskAdministratorRole,
		azure.GuestInviterRole,
		azure.DirectoryReadersRole,
		azure.PasswordAdministratorRole,
		azure.UsageSummaryReportsReaderRole,
		azure.GroupsAdministratorRole,
	}
}

func PasswordAdministratorPasswordResetTargetRoles() []string {
	return []string{
		azure.PasswordAdministratorRole,
		azure.GuestInviterRole,
		azure.DirectoryReadersRole,
	}
}

func AzurePostProcessedRelationships() []graph.Kind {
	return []graph.Kind{
		azure.AddSecret,
		azure.ExecuteCommand,
		azure.ResetPassword,
		azure.AddMembers,
		azure.GlobalAdmin,
		azure.PrivilegedRoleAdmin,
		azure.PrivilegedAuthAdmin,
		azure.AZMGAddMember,
		azure.AZMGAddOwner,
		azure.AZMGAddSecret,
		azure.AZMGGrantAppRoles,
		azure.AZMGGrantRole,
	}
}

func IsWindowsDevice(node *graph.Node) (bool, error) {
	if os, err := node.Properties.Get(common.OperatingSystem.String()).String(); err != nil {
		if graph.IsErrPropertyNotFound(err) {
			return false, nil
		}

		return false, err
	} else {
		return strings.Contains(strings.ToLower(os), "windows"), nil
	}
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

// TenantRoles returns the NodeSet of roles for a given tenant that match one of the given role template IDs. If no role template ID is provided, then all of the tenant role nodes are returned in the NodeSet.
func TenantRoles(tx graph.Transaction, tenant *graph.Node, roleTemplateIDs ...string) (graph.NodeSet, error) {
	defer log.LogAndMeasure(log.LevelInfo, "Tenant %d TenantRoles", tenant.ID)()

	if !IsTenantNode(tenant) {
		return nil, fmt.Errorf("cannot fetch tenant roles - node %d must be of kind %s", tenant.ID, azure.Tenant)
	}

	conditions := []graph.Criteria{
		query.Equals(query.StartID(), tenant.ID),
		query.Kind(query.Relationship(), azure.Contains),
		query.Kind(query.End(), azure.Role),
	}

	if len(roleTemplateIDs) > 0 {
		conditions = append(conditions, query.In(query.EndProperty(azure.RoleTemplateID.String()), roleTemplateIDs))
	}

	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(conditions...)
	}))
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

func EndNodes(tx graph.Transaction, root *graph.Node, relationship graph.Kind, nodeKinds ...graph.Kind) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), root.ID),
			query.Kind(query.Relationship(), relationship),
			query.KindIn(query.End(), nodeKinds...),
		)
	}))
}

func FetchTenants(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	var nodeSet graph.NodeSet
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if nodeSet, err = ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), azure.Tenant)
		})); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	} else {
		return nodeSet, nil
	}
}

func fetchAppOwnerRelationships(ctx context.Context, db graph.Database) ([]*graph.Relationship, error) {
	var appOwnerRels []*graph.Relationship
	return appOwnerRels, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if appOwnerRels, err = ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), azure.Entity),
				query.Kind(query.Relationship(), azure.Owns),
				query.Kind(query.End(), azure.App),
			)
		})); err != nil {
			return err
		} else {
			return nil
		}
	})
}

func fetchTenantContainsReadWriteAllGroupRelationships(tx graph.Transaction, tenant *graph.Node) ([]*graph.Relationship, error) {
	return ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), tenant.ID),
			query.Kind(query.Relationship(), azure.Contains),
			query.KindIn(query.End(), azure.Group),
			query.Equals(query.EndProperty(azure.IsAssignableToRole.String()), false),
		)
	}))
}

func fetchTenantContainsRelationships(tx graph.Transaction, tenant *graph.Node, nodeKinds ...graph.Kind) ([]*graph.Relationship, error) {
	return ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), tenant.ID),
			query.Kind(query.Relationship(), azure.Contains),
			query.KindIn(query.End(), nodeKinds...),
		)
	}))
}

func fetchReadWriteServicePrincipals(tx graph.Transaction, relationship graph.Kind, targetID graph.ID) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), azure.ServicePrincipal),
			query.Kind(query.Relationship(), relationship),
			query.Kind(query.End(), azure.ServicePrincipal),
			query.InIDs(query.EndID(), targetID),
		)
	}))
}

func aggregateSourceReadWriteServicePrincipals(tx graph.Transaction, tenantContainsServicePrincipalRelationships []*graph.Relationship, relationship graph.Kind) (graph.NodeSet, error) {
	sourceNodes := graph.NewNodeSet()
	for _, tenantContainsServicePrincipalRelationship := range tenantContainsServicePrincipalRelationships {
		if sourceServicePrincipals, err := fetchReadWriteServicePrincipals(tx, relationship, tenantContainsServicePrincipalRelationship.EndID); err != nil {
			return sourceNodes, err
		} else {
			sourceNodes.AddSet(sourceServicePrincipals)
		}
	}
	return sourceNodes, nil
}

func AppRoleAssignments(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if tenants, err := FetchTenants(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "Azure App Role Assignments Post Processing")
		for _, tenant := range tenants {
			if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
				if tenantContainsServicePrincipalRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.ServicePrincipal); err != nil {
					return err
				} else if err := createAZMGApplicationReadWriteAllEdges(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGAppRoleAssignmentReadWriteAllEdges(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGDirectoryReadWriteAllEdges(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGGroupReadWriteAllEdges(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGGroupMemberReadWriteAllEdges(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGRoleManagementReadWriteDirectoryEdgesPart1(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGRoleManagementReadWriteDirectoryEdgesPart2(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGRoleManagementReadWriteDirectoryEdgesPart3(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGRoleManagementReadWriteDirectoryEdgesPart4(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGRoleManagementReadWriteDirectoryEdgesPart5(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGServicePrincipalEndpointReadWriteAllEdges(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else {
					return nil
				}
			}); err != nil {
				operation.Done()
				return &operation.Stats, err
			}
		}
		return &operation.Stats, operation.Done()
	}
}

func createAZMGApplicationReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if tenantContainsAppRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.App); err != nil {
			return err
		} else if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.ApplicationReadWriteAll); err != nil {
			return err
		} else {
			targetRelationships := append(tenantContainsServicePrincipalRelationships, tenantContainsAppRelationships...)

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, targetRelationship := range targetRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddSecretRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   targetRelationship.EndID,
							Kind:   azure.AZMGAddSecret,
						}

						if !channels.Submit(ctx, outC, AZMGAddSecretRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   targetRelationship.EndID,
							Kind:   azure.AZMGAddOwner,
						}

						if !channels.Submit(ctx, outC, AZMGAddOwnerRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGAppRoleAssignmentReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.AppRoleAssignmentReadWriteAll); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsServicePrincipalRelationship := range tenantContainsServicePrincipalRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGGrantAppRolesRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsServicePrincipalRelationship.StartID, // the tenant
							Kind:   azure.AZMGGrantAppRoles,
						}

						if !channels.Submit(ctx, outC, AZMGGrantAppRolesRelationship) {
							return nil
						}
					}
				}

				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGDirectoryReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.DirectoryReadWriteAll); err != nil {
			return err
		} else if tenantContainsGroupRelationships, err := fetchTenantContainsReadWriteAllGroupRelationships(tx, tenant); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsGroupRelationship := range tenantContainsGroupRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddMemberRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddMember,
						}

						if !channels.Submit(ctx, outC, AZMGAddMemberRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddOwner,
						}

						if !channels.Submit(ctx, outC, AZMGAddOwnerRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGGroupReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.GroupReadWriteAll); err != nil {
			return err
		} else if tenantContainsGroupRelationships, err := fetchTenantContainsReadWriteAllGroupRelationships(tx, tenant); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsGroupRelationship := range tenantContainsGroupRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddMemberRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddMember,
						}

						if !channels.Submit(ctx, outC, AZMGAddMemberRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddOwner,
						}

						if !channels.Submit(ctx, outC, AZMGAddOwnerRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGGroupMemberReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.GroupMemberReadWriteAll); err != nil {
			return err
		} else if tenantContainsGroupRelationships, err := fetchTenantContainsReadWriteAllGroupRelationships(tx, tenant); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsGroupRelationship := range tenantContainsGroupRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddMemberRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddMember,
						}

						if !channels.Submit(ctx, outC, AZMGAddMemberRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGRoleManagementReadWriteDirectoryEdgesPart1(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else if tenantContainsRoleRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.Role); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsRoleRelationship := range tenantContainsRoleRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGGrantAppRolesRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsRoleRelationship.StartID,
							Kind:   azure.AZMGGrantAppRoles,
						}

						if !channels.Submit(ctx, outC, AZMGGrantAppRolesRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGRoleManagementReadWriteDirectoryEdgesPart2(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else if tenantContainsRoleRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.Role); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsRoleRelationship := range tenantContainsRoleRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGGrantRoleRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsRoleRelationship.EndID,
							Kind:   azure.AZMGGrantRole,
						}

						if !channels.Submit(ctx, outC, AZMGGrantRoleRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGRoleManagementReadWriteDirectoryEdgesPart3(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsServicePrincipalRelationship := range tenantContainsServicePrincipalRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddSecretRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsServicePrincipalRelationship.EndID,
							Kind:   azure.AZMGAddSecret,
						}

						if !channels.Submit(ctx, outC, AZMGAddSecretRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsServicePrincipalRelationship.EndID,
							Kind:   azure.AZMGAddOwner,
						}

						if !channels.Submit(ctx, outC, AZMGAddOwnerRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGRoleManagementReadWriteDirectoryEdgesPart4(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else if tenantContainsAppRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.App); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsAppRelationship := range tenantContainsAppRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddSecretRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsAppRelationship.EndID,
							Kind:   azure.AZMGAddSecret,
						}

						if !channels.Submit(ctx, outC, AZMGAddSecretRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsAppRelationship.EndID,
							Kind:   azure.AZMGAddOwner,
						}

						if !channels.Submit(ctx, outC, AZMGAddOwnerRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGRoleManagementReadWriteDirectoryEdgesPart5(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else if tenantContainsGroupRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.Group); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsGroupRelationship := range tenantContainsGroupRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddMemberRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddMember,
						}

						if !channels.Submit(ctx, outC, AZMGAddMemberRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddOwner,
						}

						if !channels.Submit(ctx, outC, AZMGAddOwnerRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGServicePrincipalEndpointReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.ServicePrincipalEndpointReadWriteAll); err != nil {
			return err
		} else {
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, tenantContainsServicePrincipalRelationship := range tenantContainsServicePrincipalRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddOwnerRelationship := analysis.CreatePostRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsServicePrincipalRelationship.EndID,
							Kind:   azure.AZMGAddOwner,
						}

						if !channels.Submit(ctx, outC, AZMGAddOwnerRelationship) {
							return nil
						}
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func AddSecret(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if appOwnerRels, err := fetchAppOwnerRelationships(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else if tenants, err := FetchTenants(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "AZAddSecret Post Processing")

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for _, appOwner := range appOwnerRels {
				nextJob := analysis.CreatePostRelationshipJob{
					FromID: appOwner.StartID,
					ToID:   appOwner.EndID,
					Kind:   azure.AddSecret,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}

			return nil
		})

		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			for _, tenant := range tenants {
				if tenantContainsAppRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.App); err != nil {
					return err
				} else if len(tenantContainsAppRelationships) == 0 {
					return nil
				} else if roleMembers, err := RoleMembers(tx, tenant, azure.ApplicationAdministratorRole, azure.CloudApplicationAdministratorRole); err != nil {
					return err
				} else {
					for _, roleMember := range roleMembers {
						innerRoleMember := roleMember
						operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
							for _, tenantContainsAppsRelationship := range tenantContainsAppRelationships {
								nextJob := analysis.CreatePostRelationshipJob{
									FromID: innerRoleMember.ID,
									ToID:   tenantContainsAppsRelationship.EndID,
									Kind:   azure.AddSecret,
								}

								if !channels.Submit(ctx, outC, nextJob) {
									return nil
								}
							}

							return nil
						})
					}
				}
			}
			return nil
		}); err != nil {
			// Hit done to close out the operation so it doesn't hang in the background
			operation.Done()
			return &operation.Stats, err
		} else {
			return &operation.Stats, operation.Done()
		}
	}
}

func ExecuteCommand(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if tenants, err := FetchTenants(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "AZExecuteCommand Post Processing")
		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			for _, tenant := range tenants {
				if tenantDevices, err := EndNodes(tx, tenant, azure.Contains, azure.Device); err != nil {
					return err
				} else if tenantDevices.Len() == 0 {
					return nil
				} else if intuneAdmins, err := RoleMembers(tx, tenant, azure.IntuneServiceAdministratorRole); err != nil {
					return err
				} else {
					for _, tenantDevice := range tenantDevices {
						innerTenantDevice := tenantDevice
						operation.Operation.SubmitReader(func(ctx context.Context, _ graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
							if isWindowsDevice, err := IsWindowsDevice(innerTenantDevice); err != nil {
								return err
							} else if isWindowsDevice {
								for _, intuneAdmin := range intuneAdmins {
									nextJob := analysis.CreatePostRelationshipJob{
										FromID: intuneAdmin.ID,
										ToID:   innerTenantDevice.ID,
										Kind:   azure.ExecuteCommand,
									}

									if !channels.Submit(ctx, outC, nextJob) {
										return nil
									}
								}
							}

							return nil
						})
					}
				}
			}

			return nil
		}); err != nil {
			operation.Done()
			return &operation.Stats, err
		}

		return &operation.Stats, operation.Done()
	}
}

func resetPassword(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, roleAssignments RoleAssignments) error {
	return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if pwResetRoles, err := TenantRoles(tx, tenant, ResetPasswordRoleIDs()...); err != nil {
			return err
		} else {
			for _, role := range pwResetRoles {
				if targets, err := resetPasswordEndNodeBitmapForRole(role, roleAssignments); err != nil {
					return fmt.Errorf("unable to continue processing azresetpassword for tenant node %d: %w", tenant.ID, err)
				} else {
					iter := targets.Iterator()
					for iter.HasNext() {
						nextJob := analysis.CreatePostRelationshipJob{
							FromID: role.ID,
							ToID:   graph.ID(iter.Next()),
							Kind:   azure.ResetPassword,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}

					}
				}
			}
		}
		return nil
	})
}

func resetPasswordEndNodeBitmapForRole(role *graph.Node, roleAssignments RoleAssignments) (*roaring.Bitmap, error) {
	if roleTemplateIDProp := role.Properties.Get(azure.RoleTemplateID.String()); roleTemplateIDProp.IsNil() {
		return nil, fmt.Errorf("role node %d is missing property %s", role.ID, azure.RoleTemplateID)
	} else if roleTemplateID, err := roleTemplateIDProp.String(); err != nil {
		return nil, fmt.Errorf("role node %d property %s is not a string", role.ID, azure.RoleTemplateID)
	} else {
		result := roaring.New()
		switch roleTemplateID {
		case constants.GlobalAdministratorRoleID, constants.PrivilegedAuthenticationAdministratorRoleID, constants.PartnerTier2SupportRoleID:
			result.Or(roleAssignments.Users())
		case constants.UserAdministratorRoleID:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.Or(roleAssignments.UsersWithRolesExclusive(UserAdministratorPasswordResetTargetRoles()...))
		case constants.HelpdeskAdministratorRoleID:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.Or(roleAssignments.UsersWithRolesExclusive(HelpdeskAdministratorPasswordResetTargetRoles()...))
		case constants.AuthenticationAdministratorRoleID:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.Or(roleAssignments.UsersWithRolesExclusive(AuthenticationAdministratorPasswordResetTargetRoles()...))
		case constants.PasswordAdministratorRoleID:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.Or(roleAssignments.UsersWithRolesExclusive(PasswordAdministratorPasswordResetTargetRoles()...))
		case constants.PartnerTier1SupportRoleID:
			result.Or(roleAssignments.UsersWithoutRoles())
		default:
			return nil, fmt.Errorf("role node %d has unsupported role template id '%s'", role.ID, roleTemplateID)
		}
		return result, nil
	}
}

func globalAdmins(roleAssignments RoleAssignments, tenant *graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		iter := roleAssignments.PrincipalsWithRole(constants.GlobalAdministratorRoleID).Iterator()
		for iter.HasNext() {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: graph.ID(iter.Next()),
				ToID:   tenant.ID,
				Kind:   azure.GlobalAdmin,
			}

			if !channels.Submit(ctx, outC, nextJob) {
				return nil
			}

		}
		return nil
	})
}

func privilegedRoleAdmins(roleAssignments RoleAssignments, tenant *graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		iter := roleAssignments.PrincipalsWithRole(constants.PrivilegedRoleAdministratorRoleID).Iterator()
		for iter.HasNext() {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: graph.ID(iter.Next()),
				ToID:   tenant.ID,
				Kind:   azure.PrivilegedRoleAdmin,
			}

			if !channels.Submit(ctx, outC, nextJob) {
				return nil
			}

		}
		return nil
	})
}

func privilegedAuthAdmins(roleAssignments RoleAssignments, tenant *graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		iter := roleAssignments.PrincipalsWithRole(constants.PrivilegedAuthenticationAdministratorRoleID).Iterator()
		for iter.HasNext() {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: graph.ID(iter.Next()),
				ToID:   tenant.ID,
				Kind:   azure.PrivilegedAuthAdmin,
			}

			if !channels.Submit(ctx, outC, nextJob) {
				return nil
			}

		}

		return nil
	})
}

func addMembers(roleAssignments RoleAssignments, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	tenantGroups := roleAssignments.Principals.Get(azure.Group)

	for tenantGroupID, tenantGroup := range tenantGroups {
		innerGroupID := tenantGroupID
		innerGroup := tenantGroup
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			iter := roleAssignments.UsersWithRole(AddMemberAllGroupsTargetRoles()...).Iterator()
			for iter.HasNext() {
				nextJob := analysis.CreatePostRelationshipJob{
					FromID: graph.ID(iter.Next()),
					ToID:   innerGroupID,
					Kind:   azure.AddMembers,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}

			}
			return nil
		})

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if isRoleAssignable, err := innerGroup.Properties.Get(azure.IsAssignableToRole.String()).Bool(); err != nil {
				if graph.IsErrPropertyNotFound(err) {
					log.Errorf("Node %d is missing property %s", innerGroup.ID, azure.IsAssignableToRole)
				} else {
					return err
				}
			} else if !isRoleAssignable {
				iter := roleAssignments.UsersWithRole(AddMemberGroupNotRoleAssignableTargetRoles()...).Iterator()
				for iter.HasNext() {
					nextJob := analysis.CreatePostRelationshipJob{
						FromID: graph.ID(iter.Next()),
						ToID:   innerGroupID,
						Kind:   azure.AddMembers,
					}

					if !channels.Submit(ctx, outC, nextJob) {
						return nil
					}

				}
			}

			return nil
		})
	}
}

func UserRoleAssignments(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if tenantNodes, err := FetchTenants(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "Azure User Role Assignments Post Processing")
		for _, tenant := range tenantNodes {
			if roleAssignments, err := TenantRoleAssignments(ctx, db, tenant); err != nil {
				operation.Done()
				return &analysis.AtomicPostProcessingStats{}, err
			} else {
				if err := resetPassword(ctx, db, operation, tenant, roleAssignments); err != nil {
					operation.Done()
					return &analysis.AtomicPostProcessingStats{}, err
				} else {
					globalAdmins(roleAssignments, tenant, operation)
					privilegedRoleAdmins(roleAssignments, tenant, operation)
					privilegedAuthAdmins(roleAssignments, tenant, operation)
					addMembers(roleAssignments, operation)
				}
			}
		}

		return &operation.Stats, operation.Done()
	}
}
