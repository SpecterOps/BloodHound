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
	"strings"

	"github.com/RoaringBitmap/roaring/roaring64"
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
	Nodes         graph.NodeKindSet
	AssignmentMap RoleAssignmentMap
}

func (s RoleAssignments) UsersWithoutRoles() graph.NodeSet {
	users := graph.NewNodeSet()

	for _, user := range s.Nodes.Get(azure.User) {
		if !s.AssignmentMap.UserHasRoles(user) {
			users.Add(user)
		}
	}

	return users
}

func (s RoleAssignments) NodesWithRole(roleTemplateIDs ...string) graph.NodeKindSet {
	members := graph.NewNodeKindSet()

	for userID, roleAssignments := range s.AssignmentMap {
		for _, roleTemplateID := range roleTemplateIDs {
			if _, hasRole := roleAssignments[roleTemplateID]; hasRole {
				members.Add(s.Nodes.GetNode(userID))
				break
			}
		}
	}

	return members
}

// NodesWithRolesExclusive will return nodes that *only* have a role/roles listed and exclude nodes that have other roles
func (s RoleAssignments) NodesWithRolesExclusive(roleTemplateIDs ...string) graph.NodeKindSet {
	var (
		members = graph.NewNodeKindSet()
		roleMap = make(map[string]struct{})
	)
	for _, roleTemplateID := range roleTemplateIDs {
		roleMap[roleTemplateID] = struct{}{}
	}

	for nodeID, roleAssignments := range s.AssignmentMap {
		var (
			hasIncludedRole = false
			hasExcludedRole = false
		)
		for assignment := range roleAssignments {
			if _, hasRole := roleMap[assignment]; hasRole {
				hasIncludedRole = true
			} else {
				hasExcludedRole = true
			}
		}
		if hasIncludedRole && !hasExcludedRole {
			members.Add(s.Nodes.GetNode(nodeID))
		}
	}

	return members
}

func (s RoleAssignments) NodeHasRole(id graph.ID, roleTemplateIDs ...string) bool {
	return s.AssignmentMap.HasRole(id, roleTemplateIDs...)
}

func TenantRoles(tx graph.Transaction, tenant *graph.Node, roleTemplateIDs ...string) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), tenant.ID),
			query.Kind(query.Relationship(), azure.Contains),
			query.Kind(query.End(), azure.Role),
			query.In(query.EndProperty(azure.RoleTemplateID.String()), roleTemplateIDs),
		)
	}))
}

func initTenantRoleAssignments(tx graph.Transaction, tenant *graph.Node) (RoleAssignments, error) {
	if tenantID, err := tenant.Properties.Get(azure.TenantID.String()).String(); err != nil {
		if graph.IsErrPropertyNotFound(err) {
			log.Errorf("Node %d is missing property %s", tenant.ID, azure.TenantID)
		}

		return RoleAssignments{}, err
	} else if roleMembers, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.NodeProperty(azure.TenantID.String()), tenantID),
			query.KindIn(query.Node(), azure.User, azure.Group, azure.ServicePrincipal),
		)
	})); err != nil && !graph.IsErrNotFound(err) {
		return RoleAssignments{}, err
	} else {
		return RoleAssignments{
			Nodes:         roleMembers.KindSet(),
			AssignmentMap: make(RoleAssignmentMap),
		}, nil
	}
}

func RoleMembers(tx graph.Transaction, tenant *graph.Node, roleTemplateIDs ...string) (graph.NodeSet, error) {
	if tenantRoles, err := TenantRoles(tx, tenant, roleTemplateIDs...); err != nil {
		return nil, err
	} else {
		return roleMembers(tx, tenantRoles)
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

func RoleMembersWithGrants(tx graph.Transaction, tenant *graph.Node, roleTemplateIDs ...string) (graph.NodeSet, error) {
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
		} else {
			return fetchedRoleAssignments.Nodes.EachNode(func(node *graph.Node) error {
				traversalPlan := ops.TraversalPlan{
					Root:      node,
					Direction: graph.DirectionOutbound,
					BranchQuery: func() graph.Criteria {
						return query.KindIn(query.Relationship(), azure.MemberOf, azure.HasRole)
					},
					DescentFilter: roleDescentFilter,
					PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
						return segment.Node.Kinds.ContainsOneOf(azure.Role)
					},
				}

				if roles, err := ops.AcyclicTraverseTerminals(tx, traversalPlan); err != nil {
					return err
				} else {
					roleTemplateIDs := make(map[string]struct{}, roles.Len())

					for _, roleNode := range roles {
						if rollTemplateID, err := roleNode.Properties.Get(azure.RoleTemplateID.String()).String(); err != nil {
							if !graph.IsErrPropertyNotFound(err) {
								return err
							}
						} else {
							roleTemplateIDs[rollTemplateID] = struct{}{}
						}
					}

					fetchedRoleAssignments.AssignmentMap[node.ID] = roleTemplateIDs
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

func FetchCollectedTenants(tx graph.Transaction) (graph.NodeSet, error) {
	return ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), azure.Tenant),
			query.Equals(query.NodeProperty(common.Collected.String()), true),
		)
	}))
}

func ListCollectedTenants(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	var tenants graph.NodeSet

	return tenants, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if collectedTenants, err := FetchCollectedTenants(tx); err != nil {
			return err
		} else {
			tenants = collectedTenants
		}

		return nil
	})
}

func FetchTenants(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	var (
		nodeSet graph.NodeSet
	)
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
	var (
		appOwnerRels []*graph.Relationship
	)
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
							ToID:   tenantContainsServicePrincipalRelationship.StartID, //the tenant
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
			//Hit done to close out the operation so it doesn't hang in the background
			operation.Done()
			return &operation.Stats, err
		} else {
			return &operation.Stats, operation.Done()
		}
	}
}

func fetchDeviceOwnerRelationships(ctx context.Context, db graph.Database) ([]*graph.Relationship, error) {
	var deviceOwnerRels []*graph.Relationship
	return deviceOwnerRels, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		deviceOwnerRels, err = ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Relationship(), azure.Owns),
				query.Kind(query.End(), azure.Device),
			)
		}))

		return err
	})
}

func ExecuteCommand(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if deviceOwnerRels, err := fetchDeviceOwnerRelationships(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else if tenants, err := FetchTenants(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "AZExecuteCommand Post Processing")
		for _, deviceOwner := range deviceOwnerRels {
			innerDeviceOwner := deviceOwner

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if end, err := ops.FetchNode(tx, innerDeviceOwner.EndID); err != nil {
					return err
				} else if isWindowsDevice, err := IsWindowsDevice(end); err != nil {
					return err
				} else if isWindowsDevice {
					nextJob := analysis.CreatePostRelationshipJob{
						FromID: innerDeviceOwner.StartID,
						ToID:   end.ID,
						Kind:   azure.ExecuteCommand,
					}

					if !channels.Submit(ctx, outC, nextJob) {
						return nil
					}
				}

				return nil
			})
		}

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
						operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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

func resetPassword(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], roleAssignments RoleAssignments) error {
	var (
		usersWithoutRoles = roleAssignments.UsersWithoutRoles()
	)

	if securityGroupOwners, err := getRoleEligibleSecurityGroupUsers(ctx, db, roleAssignments); err != nil {
		return err
	} else {
		for _, userID := range roleAssignments.Nodes.AllNodeIDs() {
			if err := resetPasswordCases(roleAssignments, operation, userID, usersWithoutRoles, securityGroupOwners); err != nil {
				log.Errorf("Unable to process AZResetPassword for node %d: %v", userID, err)
			}
		}

		return nil
	}
}

// getRoleEligibleSecurityGroupUsers finds Users who own or are members of a role eligible security group
func getRoleEligibleSecurityGroupUsers(ctx context.Context, db graph.Database, roleAssignments RoleAssignments) (*roaring64.Bitmap, error) {
	var (
		tenantGroups   = roleAssignments.Nodes.Get(azure.Group)
		securityGroups = make([]graph.ID, 0)
		groupUsers     = roaring64.New()
	)

	// find role eligible groups (security groups) in tenant
	for groupID, tenantGroup := range tenantGroups {
		if isRoleAssignable, err := tenantGroup.Properties.Get(azure.IsAssignableToRole.String()).Bool(); err != nil {
			if graph.IsErrPropertyNotFound(err) {
				log.Errorf("Node %d is missing property %s", tenantGroup.ID, azure.IsAssignableToRole)
			} else {
				return nil, err
			}
		} else if isRoleAssignable {
			securityGroups = append(securityGroups, groupID)
		}
	}

	return groupUsers, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		// find users that own or are a member of role eligible groups
		return tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.InIDs(query.StartID(), roleAssignments.Nodes.Get(azure.User).IDs()...),
				query.KindIn(query.Relationship(), azure.Owns, azure.MemberOf),
				query.Kind(query.End(), azure.Group),
				query.InIDs(query.EndID(), securityGroups...))
		}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for relationship := range cursor.Chan() {
				groupUsers.Add(relationship.StartID.Uint64())
			}
			return nil
		})
	})
}

func resetPasswordCases(roleAssignments RoleAssignments, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], userID graph.ID, usersWithoutRoles graph.NodeSet, securityGroupUsers *roaring64.Bitmap) error {
	if roleAssignments.NodeHasRole(userID, azure.CompanyAdministratorRole, azure.PrivilegedAuthenticationAdministratorRole, azure.PartnerTier2SupportRole) {
		// GA, PAA, and PT2S roles can reset all user passwords in the tenant
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID := range roleAssignments.Nodes.Get(azure.User) {
				if userID == targetID {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}

			return nil
		})
	} else if roleAssignments.NodeHasRole(userID, azure.HelpdeskAdministratorRole) {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID, targetNode := range roleAssignments.NodesWithRolesExclusive(HelpdeskAdministratorPasswordResetTargetRoles()...).Get(azure.User) {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetNode.ID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}

			return nil
		})

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID := range usersWithoutRoles {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}
			return nil
		})

	} else if roleAssignments.NodeHasRole(userID, azure.AuthenticationAdministratorRole) {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID, targetNode := range roleAssignments.NodesWithRolesExclusive(AuthenticationAdministratorPasswordResetTargetRoles()...).Get(azure.User) {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetNode.ID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}

			return nil
		})

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID := range usersWithoutRoles {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}
			return nil
		})

	} else if roleAssignments.NodeHasRole(userID, azure.UserAccountAdministratorRole) {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID, targetNode := range roleAssignments.NodesWithRolesExclusive(UserAdministratorPasswordResetTargetRoles()...).Get(azure.User) {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetNode.ID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}
			return nil
		})

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID := range usersWithoutRoles {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}

			return nil
		})
	} else if roleAssignments.NodeHasRole(userID, azure.PasswordAdministratorRole) {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID := range roleAssignments.NodesWithRolesExclusive(PasswordAdministratorPasswordResetTargetRoles()...).Get(azure.User) {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}

			return nil
		})

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID := range usersWithoutRoles {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}
			return nil
		})
	} else if roleAssignments.NodeHasRole(userID, azure.PartnerTier1SupportRole) {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for targetID, targetNode := range usersWithoutRoles {
				if userID == targetID || securityGroupUsers.Contains(targetID.Uint64()) {
					continue
				}

				nextJob := analysis.CreatePostRelationshipJob{
					FromID: userID,
					ToID:   targetNode.ID,
					Kind:   azure.ResetPassword,
				}

				if !channels.Submit(ctx, outC, nextJob) {
					return nil
				}
			}

			return nil
		})

	}

	return nil
}

func globalAdmins(roleAssignments RoleAssignments, tenant *graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, roleMember := range roleAssignments.NodesWithRole(azure.CompanyAdministratorRole).GetCombined(azure.User, azure.ServicePrincipal, azure.Group) {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: roleMember.ID,
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
		for _, roleMember := range roleAssignments.NodesWithRole(azure.PrivilegedRoleAdministratorRole).GetCombined(azure.User, azure.ServicePrincipal, azure.Group) {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: roleMember.ID,
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
		for _, roleMember := range roleAssignments.NodesWithRole(azure.PrivilegedAuthenticationAdministratorRole).GetCombined(azure.User, azure.ServicePrincipal, azure.Group) {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: roleMember.ID,
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
	var (
		tenantGroups = roleAssignments.Nodes.Get(azure.Group)
	)

	for tenantGroupID, tenantGroup := range tenantGroups {
		innerGroupID := tenantGroupID
		innerGroup := tenantGroup
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for _, tenantUser := range roleAssignments.NodesWithRole(AddMemberAllGroupsTargetRoles()...).Get(azure.User) {
				nextJob := analysis.CreatePostRelationshipJob{
					FromID: tenantUser.ID,
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
				for _, tenantUser := range roleAssignments.NodesWithRole(AddMemberGroupNotRoleAssignableTargetRoles()...).Get(azure.User) {
					nextJob := analysis.CreatePostRelationshipJob{
						FromID: tenantUser.ID,
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
				if err := resetPassword(ctx, db, operation, roleAssignments); err != nil {
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
