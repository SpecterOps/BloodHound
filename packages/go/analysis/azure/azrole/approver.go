package azrole

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
)

// CreateApproverEdge ...
func CreateApproverEdge(
	ctx context.Context,
	db graph.Database,
	tenantNode *graph.Node,
	operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob],
) error {
	tenantID, err := tenantNode.Properties.Get(azure.TenantID.String()).String()
	if err != nil {
		return err
	}
	tenantObjectID, err := tenantNode.Properties.Get(common.ObjectID.String()).String()
	if err != nil {
		return err
	}

	var fetchedAZRoles graph.NodeSet
	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		fetchedNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				// Step 1: Kind = AZRole and tenantId matches
				query.Kind(query.Node(), azure.Role),
				query.Equals(query.NodeProperty(azure.TenantID.String()), tenantObjectID),
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
		return err
	}

	for _, fetchedAZRole := range fetchedAZRoles {
		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			// Step 3a: Read the primaryApprovers list (group GUIDs)
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
			principalIDs := append(userApproversID, groupApproversID...)
			if len(principalIDs) == 0 {
				// Handle default admin roles...
				return handleDefaultAdminRoles(ctx, db, tx, outC, tenantID, fetchedAZRole)
			} else {
				// Handle principal IDs...
				return handlePrincipalApprovers(ctx, db, tx, outC, principalIDs, fetchedAZRole)
			}
		}); err != nil {
			return err
		}
	}

	return nil
}

func handleDefaultAdminRoles(ctx context.Context, db graph.Database, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, tenantID string, fetchedAZRole *graph.Node) error {
	var fetchedNodes graph.NodeSet
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		nodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.NodeProperty(azure.TenantID.String()), tenantID),
				query.Kind(query.Node(), azure.Role),
				query.Or(
					query.Kind(query.Node(), azure.GlobalAdmin),
					query.Kind(query.Node(), azure.PrivilegedRoleAdmin),
				),
			)
		}))
		if err != nil {
			return err
		}
		fetchedNodes = nodes
		return nil
	})
	if err != nil {
		return err
	}

	for _, fetchedNode := range fetchedNodes {
		// enqueue creation of AZRoleApprover edge: from fetchedNode → fetchedAZRole
		channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
			FromID: fetchedNode.ID,
			ToID:   fetchedAZRole.ID,
			Kind:   azure.AZRoleApprover,
		})
	}
	return nil
}

func handlePrincipalApprovers(ctx context.Context, db graph.Database, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, principalIDs []string, fetchedAZRole *graph.Node) error {
	for _, principalID := range principalIDs {
		var fetchedNode *graph.Node
		err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			node, err := tx.Nodes().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Node(), azure.Entity),
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
				slog.WarnContext(ctx, fmt.Sprintf("Entity node not found for principal ID: %s, skipping edge creation", principalID))
				continue
			} else {
				return err
			}
		}
		var nodeID graph.ID
		nodeID = fetchedNode.ID

		if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
			FromID: nodeID,
			ToID:   fetchedAZRole.ID,
			Kind:   azure.AZRoleApprover,
		}) {
			return nil
		}
	}
	return nil
}
