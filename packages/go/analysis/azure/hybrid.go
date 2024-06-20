package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
)

func PostHybrid(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	var (
		err error
	)

	tenants, err := FetchTenants(ctx, db)
	if err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("hybrid post processing: %w", err)
	}

	operation := analysis.NewPostRelationshipOperation(ctx, db, "SyncedToEntraUser Post Processing")

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, tenant := range tenants {
			if tenantUsers, err := EndNodes(tx, tenant, azure.Contains, azure.User); err != nil {
				return err
			} else if len(tenantUsers) == 0 {
				return nil
			} else {
				for _, tenantUser := range tenantUsers {
					if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if onPremUserID, hasOnPremUser, err := HasOnPremUser(tenantUser); err != nil {
							return err
						} else if hasOnPremUser {
							if adUser, err := tx.Nodes().Filterf(func() graph.Criteria {
								return query.Where(query.Equals(query.Property(query.Node, common.ObjectID.String()), onPremUserID))
							}).First(); err != nil {
								return err
							} else {
								SyncedToEntraUserRelationship := analysis.CreatePostRelationshipJob{
									FromID: adUser.ID,
									ToID:   tenantUser.ID,
									Kind:   azure.AZMGGrantRole, // TODO: Create the SyncedToEntraUser relationship here
								}

								if !channels.Submit(ctx, outC, SyncedToEntraUserRelationship) {
									return nil
								}

								SyncedFromADUserRelationship := analysis.CreatePostRelationshipJob{
									FromID: tenantUser.ID,
									ToID:   adUser.ID,
									Kind:   ad.CanRDP, // TODO: Create the SyncedFromADUser relationship here
								}

								if !channels.Submit(ctx, outC, SyncedFromADUserRelationship) {
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

		return tx.Commit()
	})

	if opErr := operation.Done(); opErr != nil {
		return &operation.Stats, fmt.Errorf("marking operation as done: %w; transaction error (if any): %w")
	}

	return &operation.Stats, nil
}

func HasOnPremUser(node *graph.Node) (string, bool, error) {
	if onPremSyncEnabled, err := node.Properties.Get(azure.OnPremSyncEnabled.String()).String(); errors.Is(err, graph.ErrPropertyNotFound) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	} else if onPremID, err := node.Properties.Get(azure.OnPremID.String()).String(); errors.Is(err, graph.ErrPropertyNotFound) {
		return onPremID, false, nil
	} else if err != nil {
		return onPremID, false, err
	} else {
		return onPremID, (onPremSyncEnabled == "true" && len(onPremID) != 0), nil
	}
}
