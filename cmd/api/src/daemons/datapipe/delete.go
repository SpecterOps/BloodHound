package datapipe

import (
	"context"
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/util/channels"
)

func DeleteCollectedGraphData(ctx context.Context, graphDB graph.Database) error {
	operation := ops.StartNewOperation[graph.ID](ops.OperationContext{
		Parent:     ctx,
		DB:         graphDB,
		NumReaders: 1,
		NumWriters: 1,
	})

	operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan graph.ID) error {
		for {
			if nextID, hasNextID := channels.Receive(ctx, inC); hasNextID {
				if err := batch.DeleteRelationship(nextID); err != nil {
					return err
				}
			} else {
				break
			}
		}

		return nil
	})

	operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- graph.ID) error {
		return tx.Relationships().FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
			channels.PipeAll(ctx, cursor.Chan(), outC)
			return cursor.Error()
		})
	})

	if err := operation.Done(); err != nil {
		return fmt.Errorf("error deleting graph relationships: %w", err)
	}

	operation = ops.StartNewOperation[graph.ID](ops.OperationContext{
		Parent:     ctx,
		DB:         graphDB,
		NumReaders: 1,
		NumWriters: 1,
	})

	operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan graph.ID) error {
		for {
			if nextID, hasNextID := channels.Receive(ctx, inC); hasNextID {
				if err := batch.DeleteNode(nextID); err != nil {
					return err
				}
			} else {
				break
			}
		}

		return nil
	})

	operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- graph.ID) error {
		return tx.Nodes().FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
			channels.PipeAll(ctx, cursor.Chan(), outC)
			return cursor.Error()
		})
	})

	if err := operation.Done(); err != nil {
		return fmt.Errorf("error deleting graph nodes: %w", err)
	}

	return nil
}
