package post

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

func MigrationForDCAPostProcessedEdges(ctx context.Context, db graph.Database, migratedRelationships graph.Kinds) error {
	for _, kind := range migratedRelationships {
		var relationshipIDs []graph.ID

		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			fetchedRelationshipIDs, err := ops.FetchRelationshipIDs(tx.Relationships().Filterf(func() graph.Criteria {
				// Only remove existing post-processed edges if they contain a `lastseen` property
				return query.And(
					query.Kind(query.Relationship(), kind),
					query.Exists(query.RelationshipProperty(common.LastSeen.String())),
				)
			}))

			relationshipIDs = fetchedRelationshipIDs
			return err
		}); err != nil {
			return err
		}

		// Only run deletion which outputs measurements to the log if there are relationships to delete
		if len(relationshipIDs) > 0 {
			measuref := measure.ContextMeasure(
				ctx,
				slog.LevelInfo,
				fmt.Sprintf("Deleted %d %s relationships for DCA Post-Processing migration", len(relationshipIDs), kind.String()),
				attr.Namespace("analysis"),
				attr.Function("MigrationForDACPostProcessedEdges"),
				attr.Scope("process"),
			)

			if err := db.BatchOperation(ctx, func(batch graph.Batch) error {
				for _, relationshipID := range relationshipIDs {
					if err := batch.DeleteRelationship(relationshipID); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return err
			}

			// Output the measurement if there is no deletion error
			measuref()
		}
	}

	return nil
}
