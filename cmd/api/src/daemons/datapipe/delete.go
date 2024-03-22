package datapipe

import (
	"context"
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
)

func DeleteCollectedGraphData(ctx context.Context, graphDB graph.Database) error {
	var nodeIDs []graph.ID

	if err := graphDB.ReadTransaction(ctx,
		func(tx graph.Transaction) error {
			fetchedNodeIDs, err := ops.FetchNodeIDs(tx.Nodes())

			nodeIDs = append(nodeIDs, fetchedNodeIDs...)
			return err
		},
	); err != nil {
		return fmt.Errorf("error fetching all nodes: %w", err)
	} else if err := graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, nodeId := range nodeIDs {
			// deleting a node also deletes all of its edges due to a sql trigger
			if err := batch.DeleteNode(nodeId); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error deleting all nodes: %w", err)
	} else {
		// if successful, handle audit log and kick off analysis
		return nil
	}
}
