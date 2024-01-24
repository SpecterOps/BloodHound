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
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

func FetchCollectedTenants(tx graph.Transaction) (graph.NodeSet, error) {
	return ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), azure.Tenant),
			query.Equals(query.NodeProperty(common.Collected.String()), true),
		)
	}))
}

func GetCollectedTenants(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
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

func FetchGraphDBTierZeroTaggedAssets(tx graph.Transaction, tenant *graph.Node) (graph.NodeSet, error) {
	defer log.LogAndMeasure(log.LevelInfo, "Tenant %d FetchGraphDBTierZeroTaggedAssets", tenant.ID)()

	if tenantObjectID, err := tenant.Properties.Get(common.ObjectID.String()).String(); err != nil {
		log.Errorf("Tenant node %d does not have a valid %s property: %v", tenant.ID, common.ObjectID, err)
		return nil, err
	} else {
		if nodeSet, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), azure.Entity),
				query.CaseInsensitiveStringContains(query.NodeProperty(azure.TenantID.String()), tenantObjectID), // I can't get this to work
				query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero),           // We should make a struct specific to Azure for AdminTierZero
			)
		})); err != nil {
			return nil, err
		} else {
			return nodeSet, nil
		}
	}
}

func FetchAzureAttackPathRoots(tx graph.Transaction, tenant *graph.Node) (graph.NodeSet, error) {
	defer log.LogAndMeasure(log.LevelDebug, "Tenant %d FetchAzureAttackPathRoots", tenant.ID)()

	attackPathRoots := graph.NewNodeKindSet()

	// Add the tenant as one of the critical path roots
	attackPathRoots.Add(tenant)

	// Pull in custom tier zero tagged assets
	if customTierZeroNodes, err := FetchGraphDBTierZeroTaggedAssets(tx, tenant); err != nil {
		return nil, err
	} else {
		attackPathRoots.AddSets(customTierZeroNodes)
	}

	// The CompanyAdministratorRole, PrivilegedRoleAdministratorRole, PrivilegedAuthenticationAdministratorRole, PartnerTier2SupportRole tenant roles are critical attack path roots
	if adminRoles, err := TenantRoles(tx, tenant, azure.CompanyAdministratorRole, azure.PrivilegedRoleAdministratorRole, azure.PrivilegedAuthenticationAdministratorRole, azure.PartnerTier2SupportRole); err != nil {
		return nil, err
	} else {
		attackPathRoots.AddSets(adminRoles)
	}

	// Find users that have CompanyAdministratorRole, PrivilegedRoleAdministratorRole, PrivilegedAuthenticationAdministratorRole, PartnerTier2SupportRole
	if adminRoleMembers, err := RoleMembersWithGrants(tx, tenant, azure.CompanyAdministratorRole, azure.PrivilegedRoleAdministratorRole, azure.PrivilegedAuthenticationAdministratorRole, azure.PartnerTier2SupportRole); err != nil {
		return nil, err
	} else {
		for _, adminRoleMember := range adminRoleMembers {
			// Add this role member as one of the critical path roots
			attackPathRoots.Add(adminRoleMember)
		}

		// Look for any apps that may run as a critical service principal
		if criticalServicePrincipals := adminRoleMembers.ContainingNodeKinds(azure.ServicePrincipal); criticalServicePrincipals.Len() > 0 {
			if criticalApps, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Start(), azure.App),
					query.Kind(query.Relationship(), azure.RunsAs),
					query.InIDs(query.EndID(), criticalServicePrincipals.IDs()...),
				)
			})); err != nil {
				return nil, err
			} else {
				for _, criticalApp := range criticalApps {
					// Add this app as one of the critical path roots
					attackPathRoots.Add(criticalApp)
				}
			}
		}
	}

	// Find any tenant virtual machines that are tied to an AD Admin Tier 0 security group
	if err := ops.ForEachEndNode(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), tenant.ID),
			query.Kind(query.Relationship(), azure.Contains),
			query.Kind(query.End(), azure.VM),
		)
	}), func(_ *graph.Relationship, tenantVM *graph.Node) error {
		if activeDirectoryTierZeroNodes, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      tenantVM,
			Direction: graph.DirectionOutbound,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.MemberOf)
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				terminalSystemTags, _ := segment.Node.Properties.GetOrDefault(common.SystemTags.String(), "").String()
				return strings.Contains(terminalSystemTags, ad.AdminTierZero)
			},
		}); err != nil {
			return err
		} else if activeDirectoryTierZeroNodes.Len() > 0 {
			// This VM is an AD computer with membership to an AD admin tier zero group. Track it as a critical
			// path root
			attackPathRoots.Add(tenantVM)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Any ResourceGroup that contains a critical attack path root is also a critical attack path root
	if err := ops.ForEachStartNode(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), azure.ResourceGroup),
			query.Kind(query.Relationship(), azure.Contains),
			query.InIDs(query.EndID(), attackPathRoots.AllNodeIDs()...),
		)
	}), func(_ *graph.Relationship, node *graph.Node) error {
		// This resource group contains a critical attack path root. Track it as a critical attack path root
		attackPathRoots.Add(node)
		return nil
	}); err != nil {
		return nil, err
	}

	// Any Subscription that contains a critical ResourceGroup is also a critical attack path root
	if err := ops.ForEachStartNode(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), azure.Subscription),
			query.Kind(query.Relationship(), azure.Contains),
			query.InIDs(query.EndID(), attackPathRoots.Get(azure.ResourceGroup).IDs()...),
		)
	}), func(_ *graph.Relationship, node *graph.Node) error {
		// This subscription contains a critical attack path root. Track it as a critical attack path root
		attackPathRoots.Add(node)
		return nil
	}); err != nil {
		return nil, err
	}

	// Any ManagementGroup that contains a critical Subscription is also a critical attack path root
	for _, criticalSubscription := range attackPathRoots.Get(azure.Subscription) {
		walkBitmap := roaring64.New()

		if criticalManagementGroups, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      criticalSubscription,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), azure.Contains)
			},
			DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				if nodeID := segment.Node.ID.Uint64(); !walkBitmap.Contains(nodeID) {
					walkBitmap.Add(nodeID)
					return true
				}

				return false
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(azure.ManagementGroup)
			},
		}); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSets(criticalManagementGroups)
		}
	}

	var (
		inboundNodes  = graph.NewNodeSet()
		tierZeroNodes = attackPathRoots.AllNodes()
	)

	// For each root look up collapsable inbound relationships to complete tier zero
	for _, attackPathRoot := range attackPathRoots.AllNodes() {
		if inboundCollapsablePaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      attackPathRoot,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), AzureNonDescentKinds()...)
			},
		}); err != nil {
			return nil, err
		} else {
			inboundNodes.AddSet(inboundCollapsablePaths.AllNodes())
		}
	}

	tierZeroNodes.AddSet(inboundNodes)
	return tierZeroNodes, nil
}

func FetchEntityRolePaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, fetchRolesTraversalPlan(node))
}

func FetchEntityRoles(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, fetchRolesTraversalPlan(node))
}

func FetchAbusableAppRoleAssignments(tx graph.Transaction, root *graph.Node, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        root,
		Direction:   direction,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterAbusableAppRoleAssignmentRelationships,
	})
}

func FetchAbusableAppRoleAssignmentsPaths(tx graph.Transaction, root *graph.Node, direction graph.Direction) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        root,
		Direction:   direction,
		BranchQuery: FilterAbusableAppRoleAssignmentRelationships,
	})
}

func FetchAppRoleAssignmentsTransitList(tx graph.Transaction, root *graph.Node, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        root,
		Direction:   direction,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterAppRoleAssignmentTransitRelationships,
	})
}

func FetchAppRoleAssignmentsTransitPaths(tx graph.Transaction, root *graph.Node, direction graph.Direction) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        root,
		Direction:   direction,
		BranchQuery: FilterAppRoleAssignmentTransitRelationships,
	})
}

func InboundControlDescentFilter(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
	// The first traversal must be a control relationship for expansion
	if segment.Depth() == 1 {
		return !segment.Edge.Kind.Is(azure.MemberOf)
	} else if segment.Depth() > 1 {
		return segment.Edge.Kind.Is(azure.MemberOf)
	}

	return true
}

func OutboundControlDescentFilter(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
	var (
		shouldDescend = true
		sawControlRel = false
	)

	// We want to ensure that AZMemberOf is expanded as well as controls relationships but with one exception: we do not
	// want to traverse more than one degree of control relationships. The question being answered for this entity query
	// is, "what does this entity have direct control of, including the entity's group memberships."
	segment.Path().WalkReverse(func(start, end *graph.Node, relationship *graph.Relationship) bool {
		if relationship.Kind.Is(azure.ControlRelationships()...) {
			if !sawControlRel {
				sawControlRel = true
			} else {
				// Reaching this condition means that this descent would result in a second control
				// relationship in this path, making this descendent ineligible for further traversal
				shouldDescend = false
				return false
			}
		}

		return true
	})

	return shouldDescend
}

func OutboundControlPathFilter(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
	return !segment.Edge.Kind.Is(azure.MemberOf)
}

func FetchOutboundEntityObjectControlPaths(tx graph.Transaction, root *graph.Node, direction graph.Direction) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:          root,
		Direction:     direction,
		BranchQuery:   FilterControlsRelationships,
		DescentFilter: OutboundControlDescentFilter,
		PathFilter:    OutboundControlPathFilter,
	})
}

func FetchInboundEntityObjectControlPaths(tx graph.Transaction, root *graph.Node, direction graph.Direction) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:          root,
		Direction:     direction,
		BranchQuery:   FilterControlsRelationships,
		DescentFilter: InboundControlDescentFilter,
	})
}

func FetchOutboundEntityObjectControl(tx graph.Transaction, root *graph.Node, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:          root,
		Direction:     direction,
		Skip:          skip,
		Limit:         limit,
		BranchQuery:   FilterControlsRelationships,
		DescentFilter: OutboundControlDescentFilter,
		PathFilter:    OutboundControlPathFilter,
	})
}

func FetchInboundEntityObjectControllers(tx graph.Transaction, root *graph.Node, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:          root,
		Direction:     direction,
		Skip:          skip,
		Limit:         limit,
		BranchQuery:   FilterControlsRelationships,
		DescentFilter: InboundControlDescentFilter,
	}, func(node *graph.Node) bool {
		return root.ID != node.ID
	})
}

func FetchEntityActiveAssignmentPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterEntityActiveAssignments,
	})
}

func FetchEntityPIMAssignmentPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterEntityPIMAssignments,
	})
}

func FetchEntityActiveAssignments(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterEntityActiveAssignments,
	})
}

func FetchEntityPIMAssignments(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterEntityPIMAssignments,
	})
}

func FetchEntityGroupMembershipPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionOutbound,
		BranchQuery: FilterGroupMembership,
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Depth() <= 1
		},
	})
}

func FetchEntityGroupMembership(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionOutbound,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterGroupMembership,
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Depth() <= 1
		},
	})
}

func FetchGroupMemberPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterGroupMembers,
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Depth() <= 1
		},
	})
}

func FetchGroupMembers(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		Skip:      skip,
		Limit:     limit,
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Depth() <= 1
		},
		BranchQuery: FilterGroupMembers,
	})
}

func FetchInboundEntityExecutionPrivilegePaths(tx graph.Transaction, node *graph.Node, direction graph.Direction) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:          node,
		Direction:     direction,
		BranchQuery:   FilterExecutionPrivileges,
		DescentFilter: InboundControlDescentFilter,
	})
}

func FetchInboundEntityExecutionPrivileges(tx graph.Transaction, node *graph.Node, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:          node,
		Direction:     direction,
		Skip:          skip,
		Limit:         limit,
		BranchQuery:   FilterExecutionPrivileges,
		DescentFilter: InboundControlDescentFilter,
	})
}

func FetchOutboundEntityExecutionPrivilegePaths(tx graph.Transaction, node *graph.Node, direction graph.Direction) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   direction,
		BranchQuery: FilterExecutionPrivileges,
		PathFilter:  OutboundControlPathFilter,
	})
}

func FetchOutboundEntityExecutionPrivileges(tx graph.Transaction, node *graph.Node, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   direction,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterExecutionPrivileges,
		PathFilter:  OutboundControlPathFilter,
	})
}

func FetchApplicationServicePrincipals(tx graph.Transaction, app *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), app.ID),
			query.Kind(query.Relationship(), azure.RunsAs),
			query.Kind(query.End(), azure.ServicePrincipal),
		)
	}))
}

func FetchServicePrincipalApplications(tx graph.Transaction, servicePrincipal *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), azure.App),
			query.Kind(query.Relationship(), azure.RunsAs),
			query.Equals(query.EndID(), servicePrincipal.ID),
		)
	}))
}

func FetchKeyVaultReaderPaths(tx graph.Transaction, keyVault *graph.Node) (graph.PathSet, error) {
	keyVaultReaders := graph.NewPathSet()

	if readers, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterKeyReaders,
	}); err != nil {
		return nil, err
	} else {
		keyVaultReaders.AddPathSet(readers)
	}

	if readers, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterCertificateReaders,
	}); err != nil {
		return nil, err
	} else {
		keyVaultReaders.AddPathSet(readers)
	}

	if readers, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterSecretReaders,
	}); err != nil {
		return nil, err
	} else {
		keyVaultReaders.AddPathSet(readers)
	}

	return keyVaultReaders, nil
}

func FetchKeyVaultReaders(tx graph.Transaction, keyVault *graph.Node, skip, limit int) (graph.NodeSet, error) {
	keyVaultReaders := graph.NewNodeSet()

	if readers, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterKeyReaders,
	}); err != nil {
		return nil, err
	} else {
		keyVaultReaders.AddSet(readers)
	}

	if readers, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterCertificateReaders,
	}); err != nil {
		return nil, err
	} else {
		keyVaultReaders.AddSet(readers)
	}

	if readers, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterSecretReaders,
	}); err != nil {
		return nil, err
	} else {
		keyVaultReaders.AddSet(readers)
	}

	return keyVaultReaders, nil
}

func FetchKeyVaultReaderCounts(tx graph.Transaction, keyVault *graph.Node) (KeyVaultReaderCounts, error) {
	var keyVaultReaders KeyVaultReaderCounts
	AllReaders := graph.NewNodeSet()

	if readers, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterKeyReaders,
	}); err != nil {
		return KeyVaultReaderCounts{}, err
	} else {
		keyVaultReaders.KeyReaders = readers.Len()
		AllReaders.AddSet(readers)
	}

	if readers, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterCertificateReaders,
	}); err != nil {
		return KeyVaultReaderCounts{}, err
	} else {
		keyVaultReaders.CertificateReaders = readers.Len()
		AllReaders.AddSet(readers)
	}

	if readers, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        keyVault,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterSecretReaders,
	}); err != nil {
		return KeyVaultReaderCounts{}, err
	} else {
		keyVaultReaders.SecretReaders = readers.Len()
		AllReaders.AddSet(readers)
	}

	keyVaultReaders.AllReaders = AllReaders.Len()

	return keyVaultReaders, nil
}

func EntityDescendentsTraversal(root *graph.Node, _ ...graph.Kind) ops.TraversalPlan {
	return ops.TraversalPlan{
		Root:        root,
		Direction:   graph.DirectionOutbound,
		BranchQuery: FilterContains,
	}
}

func FetchEntityDescendentPaths(tx graph.Transaction, root *graph.Node, descendentKinds ...graph.Kind) (graph.PathSet, error) {
	return ops.TraverseIntermediaryPaths(tx, EntityDescendentsTraversal(root, descendentKinds...), func(node *graph.Node) bool {
		return node.Kinds.ContainsOneOf(descendentKinds...)
	})
}

func FetchEntityDescendents(tx graph.Transaction, root *graph.Node, skip, limit int, descendentKinds ...graph.Kind) (graph.NodeSet, error) {
	if paths, err := FetchEntityDescendentPaths(tx, root, descendentKinds...); err != nil {
		return nil, err
	} else {
		nodes := paths.AllNodes()
		nodes.Remove(root.ID)
		return nodes.ContainingNodeKinds(descendentKinds...), nil
	}
}

func FetchEntityDescendentCounts(tx graph.Transaction, root *graph.Node, skip, limit int, descendentKinds ...graph.Kind) (Descendents, error) {
	if paths, err := FetchEntityDescendentPaths(tx, root, descendentKinds...); err != nil {
		return Descendents{}, err
	} else {
		details := Descendents{
			DescendentCounts: map[string]int{},
		}
		kindSet := paths.AllNodes().KindSet()
		kindSet.RemoveNode(root.ID)
		for _, kind := range descendentKinds {
			details.DescendentCounts[kind.String()] = int(kindSet.Count(kind))
		}

		return details, nil

	}
}

func fetchRolesTraversalPlan(root *graph.Node) ops.TraversalPlan {
	return ops.TraversalPlan{
		Root:      root,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.KindIn(query.Relationship(), azure.MemberOf, azure.HasRole),
			)
		},
		DescentFilter: roleDescentFilter,
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.Kinds.ContainsOneOf(azure.Role)
		},
	}
}

// FetchEntityByObjectID pulls a node by its ObjectID. It requires a kind to perform index lookups against.
func FetchEntityByObjectID(tx graph.Transaction, objectID string) (*graph.Node, error) {
	return tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), azure.Entity),
			query.Equals(query.NodeProperty(common.ObjectID.String()), objectID),
		)
	}).First()
}
