package opengraphschema

import (
	"context"
	"fmt"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
)

func (o *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, req v2.GraphSchemaExtension) error {
    return o.transactor.WithTransaction(ctx, func(repo OpenGraphSchemaRepository) error {
        txService := &OpenGraphSchemaService{
            openGraphSchemaRepository: repo,
            transactor:                o.transactor,
        }

		// Upsert environments with principal kinds
		// TODO: Temporary hardcoded extension ID
        if err := txService.UpsertSchemaEnvironmentWithPrincipalKinds(ctx, 1, req.Environments); err != nil {
			return fmt.Errorf("failed to upload environments with principal kinds: %w", err)
		}

		return nil
    })
}
