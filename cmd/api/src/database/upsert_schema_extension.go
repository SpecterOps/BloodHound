package database

import (
	"context"
	"fmt"
)

type EnvironmentInput struct {
	EnvironmentKindName string
	SourceKindName      string
	PrincipalKinds      []string
}

func (s *BloodhoundDB) UpsertGraphSchemaExtension(ctx context.Context, extensionID int32, environments []EnvironmentInput) error {
	return s.Transaction(ctx, func(tx *BloodhoundDB) error {
		for _, env := range environments {
			if err := tx.UpsertSchemaEnvironmentWithPrincipalKinds(ctx, extensionID, env.EnvironmentKindName, env.SourceKindName, env.PrincipalKinds); err != nil {
				return fmt.Errorf("failed to upload environment with principal kinds: %w", err)
			}
		}

		return nil
	})
}
