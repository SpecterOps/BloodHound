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
	"strings"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
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
		azure.CompanyAdministratorRole,
		azure.PrivilegedAuthenticationAdministratorRole,
		azure.PartnerTier2SupportRole,
		azure.HelpdeskAdministratorRole,
		azure.AuthenticationAdministratorRole,
		azure.UserAccountAdministratorRole,
		azure.PasswordAdministratorRole,
		azure.PartnerTier1SupportRole,
	}
}

func AddSecretRoleIDs() []string {
	return []string{
		azure.ApplicationAdministratorRole,
		azure.CloudApplicationAdministratorRole,
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

func PostProcessedRelationships() []graph.Kind {
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
		azure.SyncedToADUser,
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

func EndNodes(tx graph.Transaction, root *graph.Node, relationship graph.Kind, nodeKinds ...graph.Kind) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), root.ID),
			query.Kind(query.Relationship(), relationship),
			query.KindIn(query.End(), nodeKinds...),
		)
	}))
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
				} else if err := createAZMGAppRoleAssignmentReadWriteAllEdges(ctx, db, operation, tenantContainsServicePrincipalRelationships); err != nil {
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
				} else if err := createAZMGRoleManagementReadWriteDirectoryEdgesPart3(ctx, db, operation, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGRoleManagementReadWriteDirectoryEdgesPart4(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGRoleManagementReadWriteDirectoryEdgesPart5(ctx, db, operation, tenant, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := createAZMGServicePrincipalEndpointReadWriteAllEdges(ctx, db, operation, tenantContainsServicePrincipalRelationships); err != nil {
					return err
				} else if err := addSecret(operation, tenant); err != nil {
					return err
				}

				return nil
			}); err != nil {
				if err := operation.Done(); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error caught during azure AppRoleAssignments teardown: %v", err))
				}

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
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, targetRelationship := range append(tenantContainsServicePrincipalRelationships, tenantContainsAppRelationships...) {
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
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGAppRoleAssignmentReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.AppRoleAssignmentReadWriteAll); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGRoleManagementReadWriteDirectoryEdgesPart3(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func createAZMGServicePrincipalEndpointReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.ServicePrincipalEndpointReadWriteAll); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func addSecret(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node) error {
	return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if addSecretRoles, err := TenantRoles(tx, tenant, AddSecretRoleIDs()...); err != nil {
			return err
		} else if tenantAppsAndSPs, err := TenantApplicationsAndServicePrincipals(tx, tenant); err != nil {
			return err
		} else {
			for _, role := range addSecretRoles {
				for _, target := range tenantAppsAndSPs {
					slog.DebugContext(ctx, fmt.Sprintf("Adding AZAddSecret edge from role %s to %s %d", role.ID.String(), target.Kinds.Strings(), target.ID))
					nextJob := analysis.CreatePostRelationshipJob{
						FromID: role.ID,
						ToID:   target.ID,
						Kind:   azure.AddSecret,
					}

					if !channels.Submit(ctx, outC, nextJob) {
						return nil
					}
				}
			}
		}

		return nil
	})
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

						if err := operation.Operation.SubmitReader(func(ctx context.Context, _ graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
						}); err != nil {
							return err
						}
					}
				}
			}

			return nil
		}); err != nil {
			if err := operation.Done(); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error caught during azure ExecuteCommand teardown: %v", err))
			}

			return &operation.Stats, err
		}

		return &operation.Stats, operation.Done()
	}
}

func resetPassword(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], tenant *graph.Node, roleAssignments RoleAssignments) error {
	return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if pwResetRoles, err := TenantRoles(tx, tenant, ResetPasswordRoleIDs()...); err != nil {
			return err
		} else {
			for _, role := range pwResetRoles {
				if targets, err := resetPasswordEndNodeBitmapForRole(role, roleAssignments); err != nil {
					return fmt.Errorf("unable to continue processing azresetpassword for tenant node %d: %w", tenant.ID, err)
				} else {
					targets.Each(func(nextID uint64) bool {
						nextJob := analysis.CreatePostRelationshipJob{
							FromID: role.ID,
							ToID:   graph.ID(nextID),
							Kind:   azure.ResetPassword,
						}

						return channels.Submit(ctx, outC, nextJob)
					})
				}
			}
		}
		return nil
	})
}

func resetPasswordEndNodeBitmapForRole(role *graph.Node, roleAssignments RoleAssignments) (cardinality.Duplex[uint64], error) {
	if roleTemplateIDProp := role.Properties.Get(azure.RoleTemplateID.String()); roleTemplateIDProp.IsNil() {
		return nil, fmt.Errorf("role node %d is missing property %s", role.ID, azure.RoleTemplateID)
	} else if roleTemplateID, err := roleTemplateIDProp.String(); err != nil {
		return nil, fmt.Errorf("role node %d property %s is not a string", role.ID, azure.RoleTemplateID)
	} else {
		result := cardinality.NewBitmap64()
		switch roleTemplateID {
		case azure.CompanyAdministratorRole, azure.PrivilegedAuthenticationAdministratorRole, azure.PartnerTier2SupportRole:
			result.Or(roleAssignments.Users())
		case azure.UserAccountAdministratorRole:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.Or(roleAssignments.UsersWithRolesExclusive(UserAdministratorPasswordResetTargetRoles()...))
			result.AndNot(roleAssignments.UsersWithRoleAssignableGroupMembership())
		case azure.HelpdeskAdministratorRole:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.Or(roleAssignments.UsersWithRolesExclusive(HelpdeskAdministratorPasswordResetTargetRoles()...))
			result.AndNot(roleAssignments.UsersWithRoleAssignableGroupMembership())
		case azure.AuthenticationAdministratorRole:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.Or(roleAssignments.UsersWithRolesExclusive(AuthenticationAdministratorPasswordResetTargetRoles()...))
			result.AndNot(roleAssignments.UsersWithRoleAssignableGroupMembership())
		case azure.PasswordAdministratorRole:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.Or(roleAssignments.UsersWithRolesExclusive(PasswordAdministratorPasswordResetTargetRoles()...))
			result.AndNot(roleAssignments.UsersWithRoleAssignableGroupMembership())
		case azure.PartnerTier1SupportRole:
			result.Or(roleAssignments.UsersWithoutRoles())
			result.AndNot(roleAssignments.UsersWithRoleAssignableGroupMembership())
		default:
			return nil, fmt.Errorf("role node %d has unsupported role template id '%s'", role.ID, roleTemplateID)
		}

		return result, nil
	}
}

func globalAdmins(roleAssignments RoleAssignments, tenant *graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		roleAssignments.PrincipalsWithRole(azure.CompanyAdministratorRole).Each(func(nextID uint64) bool {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: graph.ID(nextID),
				ToID:   tenant.ID,
				Kind:   azure.GlobalAdmin,
			}

			return channels.Submit(ctx, outC, nextJob)
		})

		return nil
	}); err != nil {
		slog.Error(fmt.Sprintf("Failed to submit azure global admins post processing job: %v", err))
	}
}

func privilegedRoleAdmins(roleAssignments RoleAssignments, tenant *graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		roleAssignments.PrincipalsWithRole(azure.PrivilegedRoleAdministratorRole).Each(func(nextID uint64) bool {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: graph.ID(nextID),
				ToID:   tenant.ID,
				Kind:   azure.PrivilegedRoleAdmin,
			}

			return channels.Submit(ctx, outC, nextJob)
		})

		return nil
	}); err != nil {
		slog.Error(fmt.Sprintf("Failed to submit privileged role admins post processing job: %v", err))
	}
}

func privilegedAuthAdmins(roleAssignments RoleAssignments, tenant *graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		roleAssignments.PrincipalsWithRole(azure.PrivilegedAuthenticationAdministratorRole).Each(func(nextID uint64) bool {
			nextJob := analysis.CreatePostRelationshipJob{
				FromID: graph.ID(nextID),
				ToID:   tenant.ID,
				Kind:   azure.PrivilegedAuthAdmin,
			}

			return channels.Submit(ctx, outC, nextJob)
		})

		return nil
	}); err != nil {
		slog.Error(fmt.Sprintf("Failed to submit azure privileged auth admins post processing job: %v", err))
	}
}

func addMembers(roleAssignments RoleAssignments, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) {
	for tenantGroupID, tenantGroup := range roleAssignments.Principals.Get(azure.Group) {
		var (
			innerGroupID = tenantGroupID
			innerGroup   = tenantGroup
		)

		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			roleAssignments.UsersWithRole(AddMemberAllGroupsTargetRoles()...).Each(func(nextID uint64) bool {
				nextJob := analysis.CreatePostRelationshipJob{
					FromID: graph.ID(nextID),
					ToID:   innerGroupID,
					Kind:   azure.AddMembers,
				}

				return channels.Submit(ctx, outC, nextJob)
			})

			return nil
		}); err != nil {
			slog.Error(fmt.Sprintf("Failed to submit azure add members AddMemberAllGroupsTargetRoles post processing job: %v", err))
		}

		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if isRoleAssignable, err := innerGroup.Properties.Get(azure.IsAssignableToRole.String()).Bool(); err != nil {
				if graph.IsErrPropertyNotFound(err) {
					slog.WarnContext(ctx, fmt.Sprintf("Node %d is missing property %s", innerGroup.ID, azure.IsAssignableToRole))
				} else {
					return err
				}
			} else if !isRoleAssignable {
				roleAssignments.UsersWithRole(AddMemberGroupNotRoleAssignableTargetRoles()...).Each(func(nextID uint64) bool {
					nextJob := analysis.CreatePostRelationshipJob{
						FromID: graph.ID(nextID),
						ToID:   innerGroupID,
						Kind:   azure.AddMembers,
					}

					return channels.Submit(ctx, outC, nextJob)
				})
			}

			return nil
		}); err != nil {
			slog.Error(fmt.Sprintf("Failed to submit azure add members AddMemberGroupNotRoleAssignableTargetRoles post processing job: %v", err))
		}
	}
}

func UserRoleAssignments(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if tenantNodes, err := FetchTenants(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "Azure User Role Assignments Post Processing")

		for _, tenant := range tenantNodes {
			if roleAssignments, err := TenantRoleAssignments(ctx, db, tenant); err != nil {
				if err := operation.Done(); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error caught during azure UserRoleAssignments.TenantRoleAssignments teardown: %v", err))
				}

				return &analysis.AtomicPostProcessingStats{}, err
			} else {
				if err := resetPassword(operation, tenant, roleAssignments); err != nil {
					if err := operation.Done(); err != nil {
						slog.ErrorContext(ctx, fmt.Sprintf("Error caught during azure UserRoleAssignments.resetPassword teardown: %v", err))
					}

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

// CreateAZRoleApproverEdge will create AZRoleApprover edges from AZUser/AZGroup nodes to AZRole nodes
// according to the following task steps:
// Step 0: For each AZTenant in the graph...
func CreateAZRoleApproverEdge(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	// Step 0: Identify each AZTenant labeled node in the database.
	processingStats := &analysis.AtomicPostProcessingStats{}
	tenantNodes, err := FetchTenants(ctx, db)
	if err != nil {
		return processingStats, err
	}

	for _, tenantNode := range tenantNodes {
		// Read the tenantId property from the AZTenant node
		tenantID := tenantNode.Properties.Get(azure.TenantID.String())
		if tenantID.IsNil() {
			return processingStats, fmt.Errorf("read tenant ID property value is nil")
		}

		// Step 1 & 2: Within this tenant, find all AZRole nodes
		//             whose tenantId matches AND where
		//             end-user approval is required AND
		//             at least one of the primaryApprovers props is non-null.
		var fetchedAZRoles graph.NodeSet
		err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			fetchedNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
				return query.And(
					// Step 1: Kind = AZRole and tenantId matches
					query.Kind(query.Node(), azure.Role),
					query.Equals(query.NodeProperty(azure.TenantID.String()), tenantID),
					// Step 2: isApprovalRequired == true
					query.Equals(
						query.NodeProperty(azure.EndUserAssignmentRequiresApproval.String()),
						true,
					),
					// Step 2: primaryApprovers (user or group) is not null
					query.Or(
						query.IsNotNull(
							query.NodeProperty(azure.EndUserAssignmentUserApprovers.String()),
						),
						query.IsNotNull(
							query.NodeProperty(azure.EndUserAssignmentGroupApprovers.String()),
						),
					),
				)
			}))
			if err != nil {
				return err
			}
			fetchedAZRoles = fetchedNodes
			return nil
		})
		if err != nil {
			return processingStats, err
		}

		// Step 3: For each AZRole that requires approval...
		for _, fetchedAZRole := range fetchedAZRoles {
			// Step 3a: Read the primaryApprovers list (group GUIDs)
			userApproversID, err := fetchedAZRole.Properties.Get(
				azure.EndUserAssignmentUserApprovers.String(),
			).StringSlice()
			if err != nil {
				return processingStats, err
			}
			groupApproversID, err := fetchedAZRole.Properties.Get(
				azure.EndUserAssignmentGroupApprovers.String(),
			).StringSlice()
			if err != nil {
				return processingStats, err
			}
			principalIDs := append(userApproversID, groupApproversID...)

			if len(principalIDs) == 0 {
				// Step 3b: primaryApprovers is null/empty
				// Step 3b.i: Use tenantId from this AZRole (already have tenantID)
				// Step 3b.ii: Find GlobalAdmin AND PrivilegedRoleAdmin roles in this tenant
				//             and create an AZRoleApprover edge from each to this role.
				err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
					fetchedNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
						return query.And(
							query.Equals(query.NodeProperty(azure.TenantID.String()), tenantID),
							query.Kind(query.Node(), azure.Role),
							query.Kind(query.Node(), azure.GlobalAdmin),
							query.Kind(query.Node(), azure.PrivilegedRoleAdmin),
						)
					}))
					if err != nil {
						return err
					}

					for _, fetchedNode := range fetchedNodes {
						// enqueue creation of AZRoleApprover edge: from fetchedNode → fetchedAZRole
						var outC chan<- analysis.CreatePostRelationshipJob
						channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
							FromID: fetchedNode.ID,
							ToID:   fetchedAZRole.ID,
							Kind:   azure.AZRoleApprover,
						})
					}

					return nil
				})
				if err != nil {
					return processingStats, err
				}
			} else {
				// Step 3c: primaryApprovers is NOT null
				// Step 3c.i & 3c.ii: For each GUID, find-or-create an AZBase node and
				//                    create an AZRoleApprover edge to the role.
				for _, principalID := range principalIDs {
					err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
						// look up existing entity node by objectId
						fetchedNode, err := tx.Nodes().Filterf(func() graph.Criteria {
							return query.And(
								query.Kind(query.Node(), azure.Entity),
								query.Equals(query.NodeProperty(common.ObjectID.String()), principalID),
							)
						}).First()
						if err != nil {
							if graph.IsErrNotFound(err) {
								// not found → create new AZBase/Entity node
								createdNode, err := tx.CreateNode(
									graph.AsProperties(graph.PropertyMap{
										common.ObjectID: principalID,
									}),
									azure.Entity,
								)
								if err != nil {
									return err
								}
								fetchedNode = createdNode
							} else {
								return err
							}
						}

						// enqueue creation of AZRoleApprover edge: from fetchedNode → fetchedAZRole
						var outC chan<- analysis.CreatePostRelationshipJob
						channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
							FromID: fetchedNode.ID,
							ToID:   fetchedAZRole.ID,
							Kind:   azure.AZRoleApprover,
						})

						return nil
					})
					if err != nil {
						return processingStats, err
					}
				}
			}
		}
	}

	return processingStats, nil
}
