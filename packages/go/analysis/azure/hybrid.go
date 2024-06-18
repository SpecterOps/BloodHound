package azure

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/azure"
)

func PostHybrid(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if syncedToEntraUser, err := SyncedToEntraUser(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed hybrid post processing: %w", err)
	} else {

	}

	return &operation.Stats, operation.Done()
}

func SyncedToEntraUser(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if azUsersWithOnPrem, err := fetchUsersWithOnPrem(tx, )
	})
	

	// if tenants, err := FetchTenants(ctx, db); err != nil {
	// 	return &analysis.AtomicPostProcessingStats{}, err
	// } else {
	// 	operation := analysis.NewPostRelationshipOperation(ctx, db, "SyncedToEntraUser Post Processing")

	// 	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			
	// 		for _, tenant := range tenants {
	// 			if tenantUsers, err := EndNodes(tx, tenant, azure.Contains, azure.User); err != nil {
	// 				return err
	// 			} else if tenantUsers.Len() == 0 {
	// 				return nil
	// 			} else {

	// 				for _, tenantUser := range tenantUsers {
	// 					innerTenantUser := tenantUser
	// 					operation.Operation.SubmitReader(func(ctx context.Context, _ graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
	// 						if hasOnPrem, err := HasOnPremUser(innerTenantUser); err != nil {
	// 							return err
	// 						} else {
	// 							if hasOnPrem {
	// 								if adUser, ad.FetchNodesByKind(ctx, db, ad.User)
	// 							}
	// 						}

	// 						return nil
	// 					})
	// 				}

	// 			}
	// 		}
	// 		return nil
	// 	}); err != nil {
	// 		operation.Done()
	// 		return &operation.Stats, err
	// 	}
	// 	return &operation.Stats, operation.Done()
	}

}

func HasOnPremUser(node *graph.Node) (bool, error) {
	if onPremSyncEnabled, err := node.Properties.Get(azure.OnPremSyncEnabled.String()).String(); err != nil {
		if graph.IsErrPropertyNotFound(err) {
			return false, nil
		}

		return false, err
	} else if onPremId, err := nodeProperties.Get(azure.OnPremID.String()).String(); err != nil {
		if graph.IsErrPropertyNotFound(err) {
			return false, nil
		}

		return false, err
	} else {
		return (onPremSyncEnabled == true && len(onPremID) != 0), nil
	}
}
