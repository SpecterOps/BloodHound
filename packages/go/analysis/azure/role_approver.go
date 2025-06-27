package azure

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

// CreateApproverEdge processes a single AZTenant node to create AZRoleApprover edges for all qualifying AZRole nodes.
//
// This function implements the core logic for AZRoleApprover edge creation:
//
// 1. Extracts the tenant's objectid to match against AZRole tenantid properties
// 2. Finds all AZRole nodes in this tenant where:
//   - tenantId matches the tenant's objectid
//   - isApprovalRequired == true
//   - At least one of EndUserAssignmentUserApprovers or EndUserAssignmentGroupApprovers is not null
//
// 3. For each qualifying AZRole, determines the appropriate approvers:
//   - If no specific approvers configured: uses default admin roles (Global Admin, Privileged Role Admin)
//   - If specific approvers configured: uses the specified user/group GUIDs
//
// 4. Creates AZRoleApprover edges from approver nodes to the AZRole node
//
// Parameters:
//   - ctx: Context for the operation
//   - db: Graph database instance
//   - tenantNode: The AZTenant node to process
//   - operation: Post-processing operation tracker for creating edges
//
// Returns error if any step fails during processing.
func CreateApproverEdge(
	ctx context.Context,
	db graph.Database,
	tenantNode *graph.Node,
	operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob],
) error {
	// Extract the tenant's objectid to match against AZRole tenantid properties
	tenantObjectID, err := tenantNode.Properties.Get(common.ObjectID.String()).String()
	if err != nil {
		return err
	}

	// Step 1 & 2: Find all AZRole nodes in this tenant that require approval and have approvers configured
	var fetchedAZRoles graph.NodeSet
	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		fetchedNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				// Step 1: Kind = AZRole and tenantId matches the current tenant's objectid
				query.Kind(query.Node(), azure.Role),
				query.Equals(query.NodeProperty(azure.TenantID.String()), tenantObjectID),
				// Step 2: isApprovalRequired == true (role requires approval for assignment)
				query.Equals(
					query.NodeProperty(azure.EndUserAssignmentRequiresApproval.String()),
					true,
				),
				// Step 2: primaryApprovers (user or group) is not null - at least one approver type configured
				query.Or(
					query.IsNotNull(query.NodeProperty(azure.EndUserAssignmentUserApprovers.String())),
					query.IsNotNull(query.NodeProperty(azure.EndUserAssignmentGroupApprovers.String())),
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
		return err
	}

	// Step 3: Process each qualifying AZRole to create appropriate AZRoleApprover edges
	for _, fetchedAZRole := range fetchedAZRoles {
		if err := operation.Operation.SubmitReader(func(
			ctx context.Context,
			tx graph.Transaction,
			outC chan<- analysis.CreatePostRelationshipJob,
		) error {
			// Step 3a: Read the primaryApprovers lists (user and group GUIDs)
			userApproversID, err := fetchedAZRole.Properties.Get(
				azure.EndUserAssignmentUserApprovers.String(),
			).StringSlice()
			if err != nil {
				return err
			}
			groupApproversID, err := fetchedAZRole.Properties.Get(
				azure.EndUserAssignmentGroupApprovers.String(),
			).StringSlice()
			if err != nil {
				return err
			}

			// Combine user and group approver GUIDs into a single list
			principalIDs := append(userApproversID, groupApproversID...)

			// Step 3b: Determine approver strategy based on whether specific approvers are configured
			if len(principalIDs) == 0 {
				// Step 3b.i-iii: No specific approvers - use default admin roles
				return handleDefaultAdminRoles(ctx, db, outC, tenantNode, fetchedAZRole)
			} else {
				// Step 3c.i-ii: Specific approvers configured - use the specified GUIDs
				return handlePrincipalApprovers(ctx, db, outC, principalIDs, fetchedAZRole)
			}
		}); err != nil {
			return err
		}
	}

	return nil
}

// handleDefaultAdminRoles implements Step 3b logic: when no specific approvers are configured,
// use Global Administrator and Privileged Role Administrator roles as default approvers.
//
// This function:
// 1. Finds "Global Administrator" and "Privileged Role Administrator" AZRole nodes in the same tenant
// 2. Creates AZRoleApprover edges from each of these admin roles to the target AZRole
//
// This handles the case where primaryApprovers is null/empty, meaning the organization
// relies on default administrative roles for approval rather than specific users/groups.
func handleDefaultAdminRoles(
	ctx context.Context,
	db graph.Database,
	outC chan<- analysis.CreatePostRelationshipJob,
	tenantNode, fetchedAZRole *graph.Node,
) error {
	// Step 3b.ii: Find Global Administrator and Privileged Role Administrator roles in this tenant
	var fetchedNodes graph.NodeSet
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		roleAssignments, err := TenantRoles(
			tx,
			tenantNode,
			azure.PrivilegedRoleAdministratorRole, // Privileged Role Administrator
			azure.CompanyAdministratorRole,        // Global Administrator
		)
		if err != nil {
			return err
		}

		fetchedNodes = roleAssignments
		return nil
	})
	if err != nil {
		return err
	}

	// Step 3b.iii: Create AZRoleApprover edges from each default admin role to the target AZRole
	for _, fetchedNode := range fetchedNodes {
		// Enqueue creation of AZRoleApprover edge: from admin role → target AZRole
		channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
			FromID: fetchedNode.ID,
			ToID:   fetchedAZRole.ID,
			Kind:   azure.AZRoleApprover,
		})
	}
	return nil
}

// handlePrincipalApprovers implements Step 3c logic: when specific approvers are configured,
// create AZRoleApprover edges from the specified user/group nodes to the target AZRole.
//
// This function:
// 1. Iterates through each GUID in the primaryApprovers list
// 2. Finds the corresponding AZUser or AZGroup node with matching objectid
// 3. Creates an AZRoleApprover edge from that node to the target AZRole
//
// This handles the case where specific users or groups have been designated as approvers
// for role assignments, rather than relying on default administrative roles.
//
// Note: Groups for approvers can be nested as long as the root groups are not role eligible.
func handlePrincipalApprovers(
	ctx context.Context,
	db graph.Database,
	outC chan<- analysis.CreatePostRelationshipJob,
	principalIDs []string,
	fetchedAZRole *graph.Node,
) error {
	// Step 3c.ii: Process each GUID in the primaryApprovers list
	for _, principalID := range principalIDs {
		var fetchedNode *graph.Node

		// Step 3c.ii.1: Find the AZUser or AZGroup node with matching objectid
		err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			node, err := tx.Nodes().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Node(), azure.Entity), // Matches AZUser, AZGroup, etc.
					query.Equals(query.NodeProperty(common.ObjectID.String()), principalID),
				)
			}).First()
			if err != nil {
				return err
			}
			fetchedNode = node
			return nil
		})
		if err != nil {
			if graph.IsErrNotFound(err) {
				// Log warning if approver node not found (may have been deleted or not yet ingested)
				slog.WarnContext(
					ctx,
					fmt.Sprintf(
						"Entity node not found for principal ID: %s, skipping edge creation",
						principalID,
					),
				)
				continue
			} else {
				return err
			}
		}

		// Step 3c.ii.2: Create AZRoleApprover edge from approver node to target AZRole
		if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
			FromID: fetchedNode.ID,
			ToID:   fetchedAZRole.ID,
			Kind:   azure.AZRoleApprover,
		}) {
			return nil
		}
	}
	return nil
}
