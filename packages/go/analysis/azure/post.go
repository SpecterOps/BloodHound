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

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/delta"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/trace"
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

func AppRoleAssignments(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing App Role Assignments",
		attr.Namespace("analysis"),
		attr.Function("AppRoleAssignments"),
		attr.Scope("process"),
	)()

	if tenants, err := FetchTenants(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
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
					slog.ErrorContext(ctx, "Error caught during azure AppRoleAssignments teardown", attr.Error(err))
				}

				return &operation.Stats, err
			}
		}

		return &operation.Stats, operation.Done()
	}
}

func createAZMGApplicationReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if tenantContainsAppRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.App); err != nil {
			return err
		} else if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.ApplicationReadWriteAll); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, targetRelationship := range append(tenantContainsServicePrincipalRelationships, tenantContainsAppRelationships...) {
					for _, sourceNode := range sourceNodes {
						AZMGAddSecretRelationship := post.EnsureRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   targetRelationship.EndID,
							Kind:   azure.AZMGAddSecret,
						}

						if !channels.Submit(ctx, outC, AZMGAddSecretRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := post.EnsureRelationshipJob{
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

func createAZMGAppRoleAssignmentReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.AppRoleAssignmentReadWriteAll); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsServicePrincipalRelationship := range tenantContainsServicePrincipalRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGGrantAppRolesRelationship := post.EnsureRelationshipJob{
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

func createAZMGDirectoryReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.DirectoryReadWriteAll); err != nil {
			return err
		} else if tenantContainsGroupRelationships, err := fetchTenantContainsReadWriteAllGroupRelationships(tx, tenant); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsGroupRelationship := range tenantContainsGroupRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddMemberRelationship := post.EnsureRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddMember,
						}

						if !channels.Submit(ctx, outC, AZMGAddMemberRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := post.EnsureRelationshipJob{
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

func createAZMGGroupReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.GroupReadWriteAll); err != nil {
			return err
		} else if tenantContainsGroupRelationships, err := fetchTenantContainsReadWriteAllGroupRelationships(tx, tenant); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsGroupRelationship := range tenantContainsGroupRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddMemberRelationship := post.EnsureRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddMember,
						}

						if !channels.Submit(ctx, outC, AZMGAddMemberRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := post.EnsureRelationshipJob{
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

func createAZMGGroupMemberReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.GroupMemberReadWriteAll); err != nil {
			return err
		} else if tenantContainsGroupRelationships, err := fetchTenantContainsReadWriteAllGroupRelationships(tx, tenant); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsGroupRelationship := range tenantContainsGroupRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddMemberRelationship := post.EnsureRelationshipJob{
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

func createAZMGRoleManagementReadWriteDirectoryEdgesPart1(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else if tenantContainsRoleRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.Role); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsRoleRelationship := range tenantContainsRoleRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGGrantAppRolesRelationship := post.EnsureRelationshipJob{
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

func createAZMGRoleManagementReadWriteDirectoryEdgesPart2(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else if tenantContainsRoleRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.Role); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsRoleRelationship := range tenantContainsRoleRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGGrantRoleRelationship := post.EnsureRelationshipJob{
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

func createAZMGRoleManagementReadWriteDirectoryEdgesPart3(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsServicePrincipalRelationship := range tenantContainsServicePrincipalRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddSecretRelationship := post.EnsureRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsServicePrincipalRelationship.EndID,
							Kind:   azure.AZMGAddSecret,
						}

						if !channels.Submit(ctx, outC, AZMGAddSecretRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := post.EnsureRelationshipJob{
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

func createAZMGRoleManagementReadWriteDirectoryEdgesPart4(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else if tenantContainsAppRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.App); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsAppRelationship := range tenantContainsAppRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddSecretRelationship := post.EnsureRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsAppRelationship.EndID,
							Kind:   azure.AZMGAddSecret,
						}

						if !channels.Submit(ctx, outC, AZMGAddSecretRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := post.EnsureRelationshipJob{
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

func createAZMGRoleManagementReadWriteDirectoryEdgesPart5(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node, tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.RoleManagementReadWriteDirectory); err != nil {
			return err
		} else if tenantContainsGroupRelationships, err := fetchTenantContainsRelationships(tx, tenant, azure.Group); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsGroupRelationship := range tenantContainsGroupRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddMemberRelationship := post.EnsureRelationshipJob{
							FromID: sourceNode.ID,
							ToID:   tenantContainsGroupRelationship.EndID,
							Kind:   azure.AZMGAddMember,
						}

						if !channels.Submit(ctx, outC, AZMGAddMemberRelationship) {
							return nil
						}

						AZMGAddOwnerRelationship := post.EnsureRelationshipJob{
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

func createAZMGServicePrincipalEndpointReadWriteAllEdges(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenantContainsServicePrincipalRelationships []*graph.Relationship) error {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if sourceNodes, err := aggregateSourceReadWriteServicePrincipals(tx, tenantContainsServicePrincipalRelationships, azure.ServicePrincipalEndpointReadWriteAll); err != nil {
			return err
		} else {
			return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, tenantContainsServicePrincipalRelationship := range tenantContainsServicePrincipalRelationships {
					for _, sourceNode := range sourceNodes {
						AZMGAddOwnerRelationship := post.EnsureRelationshipJob{
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

func addSecret(operation analysis.StatTrackedOperation[post.EnsureRelationshipJob], tenant *graph.Node) error {
	return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		if addSecretRoles, err := TenantRoles(tx, tenant, AddSecretRoleIDs()...); err != nil {
			return err
		} else if tenantAppsAndSPs, err := TenantApplicationsAndServicePrincipals(tx, tenant); err != nil {
			return err
		} else {
			for _, role := range addSecretRoles {
				for _, target := range tenantAppsAndSPs {
					slog.DebugContext(
						ctx,
						"Adding AZAddSecret edge from role to target",
						slog.String("role_id", role.ID.String()),
						slog.String("target_kinds", strings.Join(target.Kinds.Strings(), ",")),
						slog.Any("target_id", target.ID),
					)
					nextJob := post.EnsureRelationshipJob{
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

func ExecuteCommand(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing ExecuteCommand",
		attr.Namespace("analysis"),
		attr.Function("ExecuteCommand"),
		attr.Scope("process"),
	)()

	if tenants, err := FetchTenants(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
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

						if err := operation.Operation.SubmitReader(func(ctx context.Context, _ graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
							if isWindowsDevice, err := IsWindowsDevice(innerTenantDevice); err != nil {
								return err
							} else if isWindowsDevice {
								for _, intuneAdmin := range intuneAdmins {
									nextJob := post.EnsureRelationshipJob{
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
				slog.ErrorContext(ctx, "Error caught during azure ExecuteCommand teardown", attr.Error(err))
			}

			return &operation.Stats, err
		}

		return &operation.Stats, operation.Done()
	}
}

func postAzureResetPassword(ctx context.Context, db graph.Database, sink *post.FilteredRelationshipSink, tenant *graph.Node, roleAssignments RoleAssignments) error {
	defer trace.Function(ctx, "postAzureResetPassword", attr.Operation("Reset Password Post Processing"))()

	return db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if pwResetRoles, err := TenantRoles(tx, tenant, ResetPasswordRoleIDs()...); err != nil {
			return err
		} else {
			for _, role := range pwResetRoles {
				if targets, err := resetPasswordEndNodeBitmapForRole(role, roleAssignments); err != nil {
					return fmt.Errorf("unable to continue processing azresetpassword for tenant node %d: %w", tenant.ID, err)
				} else {
					targets.Each(func(nextID uint64) bool {
						nextJob := post.EnsureRelationshipJob{
							FromID: role.ID,
							ToID:   graph.ID(nextID),
							Kind:   azure.ResetPassword,
						}

						return sink.Submit(ctx, nextJob)
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

func postAzureGlobalAdmins(ctx context.Context, sink *post.FilteredRelationshipSink, roleAssignments RoleAssignments, tenant *graph.Node) error {
	defer trace.Function(ctx, "postAzureGlobalAdmins", attr.Operation("Global Admins Post Processing"))()

	roleAssignments.PrincipalsWithRole(azure.CompanyAdministratorRole).Each(func(nextID uint64) bool {
		nextJob := post.EnsureRelationshipJob{
			FromID: graph.ID(nextID),
			ToID:   tenant.ID,
			Kind:   azure.GlobalAdmin,
		}

		return sink.Submit(ctx, nextJob)
	})

	return nil
}

func postAzurePrivilegedRoleAdmins(ctx context.Context, sink *post.FilteredRelationshipSink, roleAssignments RoleAssignments, tenant *graph.Node) {
	defer trace.Function(ctx, "postAzurePrivilegedRoleAdmins", attr.Operation("Privileged Role Admins Post Processing"))()

	roleAssignments.PrincipalsWithRole(azure.PrivilegedRoleAdministratorRole).Each(func(nextID uint64) bool {
		nextJob := post.EnsureRelationshipJob{
			FromID: graph.ID(nextID),
			ToID:   tenant.ID,
			Kind:   azure.PrivilegedRoleAdmin,
		}

		return sink.Submit(ctx, nextJob)
	})
}

func postAzurePrivilegedAuthAdmins(ctx context.Context, sink *post.FilteredRelationshipSink, roleAssignments RoleAssignments, tenant *graph.Node) {
	defer trace.Function(ctx, "postAzurePrivilegedAuthAdmins", attr.Operation("Privileged Auth Admins Post Processing"))()

	roleAssignments.PrincipalsWithRole(azure.PrivilegedAuthenticationAdministratorRole).Each(func(nextID uint64) bool {
		nextJob := post.EnsureRelationshipJob{
			FromID: graph.ID(nextID),
			ToID:   tenant.ID,
			Kind:   azure.PrivilegedAuthAdmin,
		}

		return sink.Submit(ctx, nextJob)
	})
}

func postAzureAddMembers(ctx context.Context, sink *post.FilteredRelationshipSink, roleAssignments RoleAssignments) error {
	defer trace.Function(ctx, "postAzureAddMembers", attr.Operation("Azure Add Members Post Processing"))()

	for tenantGroupID, tenantGroup := range roleAssignments.TenantPrincipals.Get(azure.Group) {
		roleAssignments.UsersWithRole(AddMemberAllGroupsTargetRoles()...).Each(func(nextID uint64) bool {
			nextJob := post.EnsureRelationshipJob{
				FromID: graph.ID(nextID),
				ToID:   tenantGroupID,
				Kind:   azure.AddMembers,
			}

			return sink.Submit(ctx, nextJob)
		})

		roleAssignments.ServicePrincipalsWithRole(AddMemberAllGroupsTargetRoles()...).Each(func(nextID uint64) bool {
			nextJob := post.EnsureRelationshipJob{
				FromID: graph.ID(nextID),
				ToID:   tenantGroupID,
				Kind:   azure.AddMembers,
			}

			return sink.Submit(ctx, nextJob)
		})

		if isRoleAssignable, err := tenantGroup.Properties.Get(azure.IsAssignableToRole.String()).Bool(); err != nil {
			if graph.IsErrPropertyNotFound(err) {
				slog.WarnContext(
					ctx,
					"Node is missing property",
					slog.Uint64("node_id", tenantGroup.ID.Uint64()),
					slog.String("property", azure.IsAssignableToRole.String()),
				)
			} else {
				return err
			}
		} else if !isRoleAssignable {
			roleAssignments.UsersWithRole(AddMemberGroupNotRoleAssignableTargetRoles()...).Each(func(nextID uint64) bool {
				nextJob := post.EnsureRelationshipJob{
					FromID: graph.ID(nextID),
					ToID:   tenantGroupID,
					Kind:   azure.AddMembers,
				}

				return sink.Submit(ctx, nextJob)
			})
		}

		if isRoleAssignable, err := tenantGroup.Properties.Get(azure.IsAssignableToRole.String()).Bool(); err != nil {
			if graph.IsErrPropertyNotFound(err) {
				slog.WarnContext(
					ctx,
					"Node is missing property",
					slog.Uint64("node_id", tenantGroup.ID.Uint64()),
					slog.String("property", azure.IsAssignableToRole.String()),
				)
			} else {
				return err
			}
		} else if !isRoleAssignable {
			roleAssignments.ServicePrincipalsWithRole(AddMemberGroupNotRoleAssignableTargetRoles()...).Each(func(nextID uint64) bool {
				nextJob := post.EnsureRelationshipJob{
					FromID: graph.ID(nextID),
					ToID:   tenantGroupID,
					Kind:   azure.AddMembers,
				}

				return sink.Submit(ctx, nextJob)
			})
		}
	}

	return nil
}

var userRoleAssignmentPostProcessedEdges = graph.Kinds{
	azure.ResetPassword,
	azure.GlobalAdmin,
	azure.PrivilegedRoleAdmin,
	azure.PrivilegedAuthAdmin,
	azure.AddMembers,
}

func UserRoleAssignments(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	ctx = trace.Context(ctx, slog.LevelInfo, "analysis", "Azure User Role Assignment Post-processing")
	defer trace.Function(ctx, "UserRoleAssignments")()

	if userRoleAssignmentTracker, err := delta.FetchTracker(ctx, db, userRoleAssignmentPostProcessedEdges); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else if tenantNodes, err := FetchTenants(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else {
		sink := post.NewFilteredRelationshipSink(ctx, "Azure User Role Assignments Post Processing", db, userRoleAssignmentTracker)
		defer sink.Done()

		for _, tenant := range tenantNodes {
			if roleAssignments, err := FetchTenantRoleAssignments(ctx, db, tenant); err != nil {
				return &post.AtomicPostProcessingStats{}, err
			} else {
				if err := postAzureResetPassword(ctx, db, sink, tenant, roleAssignments); err != nil {
					return &post.AtomicPostProcessingStats{}, err
				} else {
					postAzureGlobalAdmins(ctx, sink, roleAssignments, tenant)
					postAzurePrivilegedRoleAdmins(ctx, sink, roleAssignments, tenant)
					postAzurePrivilegedAuthAdmins(ctx, sink, roleAssignments, tenant)

					if err := postAzureAddMembers(ctx, sink, roleAssignments); err != nil {
						slog.Error("Azure AddMember Post-Processing Failure", attr.Error(err))
					}
				}
			}
		}

		return sink.Stats(), nil
	}
}

// CreateAZRoleApproverEdge creates AZRoleApprover edges from AZUser/AZGroup nodes to AZRole nodes.
//
// This function implements the AZRoleApprover edge creation logic according to the following requirements:
//
//  1. Identify each AZTenant labeled node in the database
//  2. For each AZTenant, find all AZRole nodes where:
//     - The AZRole's tenantid property matches the AZTenant's objectid property
//     - The AZRole's isApprovalRequired property is true
//     - The AZRole has primaryApprovers configured (user or group approvers)
//  3. For each qualifying AZRole node:
//     a. If primaryApprovers is empty/null:
//     - Create AZRoleApprover edges from "Global Administrator" and "Privileged Role Administrator"
//     roles in the same tenant to this AZRole
//     b. If primaryApprovers contains GUIDs:
//     - Create AZRoleApprover edges from each AZUser/AZGroup node matching those GUIDs to this AZRole
//
// Note: Groups for approvers can be nested as long as the root groups are not role eligible.
// Note: If no specific approvers are selected, privileged role administrators/global administrators
// become the default approvers (primaryApprovers array will be empty in this case).
//
// Returns post-processing statistics and any error encountered during processing.
func CreateAZRoleApproverEdge(
	ctx context.Context,
	db graph.Database,
) (
	*post.AtomicPostProcessingStats,
	error,
) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing Azure Role Approver Edges",
		attr.Namespace("analysis"),
		attr.Function("CreateAZRoleApproverEdge"),
		attr.Scope("process"),
	)()

	// Step 0: Identify each AZTenant labeled node in the database.
	operation := analysis.NewPostRelationshipOperation(ctx, db, "AZRoleApprover Post Processing")
	tenantNodes, err := FetchTenants(ctx, db)
	if err != nil {
		return &operation.Stats, err
	}

	// Process each tenant to create AZRoleApprover edges for roles requiring approval
	for _, tenantNode := range tenantNodes {
		if err := CreateApproverEdge(ctx, db, tenantNode, operation); err != nil {
			return &operation.Stats, err
		}
	}

	return &operation.Stats, operation.Done()
}

func FixManagementGroupNames(ctx context.Context, db graph.Database) error {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Fix Management Group Names",
		attr.Namespace("analysis"),
		attr.Function("FixManagementGroupNames"),
		attr.Scope("process"),
	)()

	if managementGroups, err := FetchManagementGroups(ctx, db); err != nil {
		return err
	} else if tenants, err := FetchTenants(ctx, db); err != nil {
		return err
	} else {
		tenantMap := make(map[string]string)
		for _, tenant := range tenants {
			if id, err := tenant.Properties.Get(common.ObjectID.String()).String(); err != nil {
				slog.ErrorContext(
					ctx,
					"Error getting tenant objectid",
					slog.Int64("tenant_id", tenant.ID.Int64()),
					attr.Error(err),
				)
				continue
			} else if tenantName, err := tenant.Properties.Get(common.Name.String()).String(); err != nil {
				slog.ErrorContext(
					ctx,
					"Error getting tenant name",
					slog.Int64("tenant_id", tenant.ID.Int64()),
					attr.Error(err),
				)
				continue
			} else {
				tenantMap[id] = tenantName
			}
		}

		return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
			for _, managementGroup := range managementGroups {
				if tenantId, err := managementGroup.Properties.Get(azure.TenantID.String()).String(); err != nil {
					slog.WarnContext(
						ctx,
						"Error getting tenantid for management group",
						slog.Int64("management_group_id", managementGroup.ID.Int64()),
						attr.Error(err),
					)
					continue
				} else if displayName, err := managementGroup.Properties.Get(common.DisplayName.String()).String(); err != nil {
					slog.WarnContext(
						ctx,
						"Error getting display name for management group",
						slog.Int64("management_group_id", managementGroup.ID.Int64()),
						attr.Error(err),
					)
					continue
				} else if tenantName, ok := tenantMap[tenantId]; !ok {
					slog.WarnContext(ctx, "Could not find a tenant that matches management group", slog.Int64("management_group_id", managementGroup.ID.Int64()))
					continue
				} else {
					managementGroup.Properties.Set(common.Name.String(), strings.ToUpper(fmt.Sprintf("%s@%s", displayName, tenantName)))
					if err := tx.UpdateNode(managementGroup); err != nil {
						return err
					}
				}
			}

			return nil
		})
	}
}
