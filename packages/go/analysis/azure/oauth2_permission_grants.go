package azure

import (
    "context"
    "github.com/specterops/dawgs/graph"
)

func ListEntityOAuth2PermissionGrantPaths(ctx context.Context, db graph.Database, objectID string) (graph.PathSet, error) {
    var paths graph.PathSet

    return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
        node, err := FetchEntityByObjectID(tx, objectID)
        if err != nil {
            return err
        }

        // Für Grants modellierst du Kanten OUTBOUND (User/Tenant -> ServicePrincipal)
        paths, err = FetchOutboundEntityOAuth2PermissionGrantPaths(tx, node)
        return err
    })
}
