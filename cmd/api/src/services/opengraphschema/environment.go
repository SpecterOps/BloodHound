// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package opengraphschema

import (
	"context"
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (o *OpenGraphSchemaService) UpsertSchemaEnvironmentWithPrincipalKinds(ctx context.Context, environment model.SchemaEnvironment, principalKinds []model.SchemaEnvironmentPrincipalKind) error {
	// Validate environment
	if err := o.validateSchemaEnvironment(ctx, environment); err != nil {
		return fmt.Errorf("error validating schema environment: %w", err)
	}

	// Validate principal kinds
	for _, kind := range principalKinds {
		if err := o.validateSchemaEnvironmentPrincipalKind(ctx, kind.PrincipalKind); err != nil {
			return fmt.Errorf("error validating principal kind: %w", err)
		}
	}

	// Upsert the environment
	id, err := o.upsertSchemaEnvironment(ctx, environment)
	if err != nil {
		return fmt.Errorf("error upserting schema environment: %w", err)
	}

	// Upsert principal kinds for environment
	if err := o.upsertPrincipalKinds(ctx, id, principalKinds); err != nil {
		return fmt.Errorf("error upserting principal kinds: %w", err)
	}

	return nil
}

func (o *OpenGraphSchemaService) upsertSchemaEnvironment(ctx context.Context, graphSchema model.SchemaEnvironment) (int32, error) {
	if existing, err := o.openGraphSchemaRepository.GetSchemaEnvironmentById(ctx, graphSchema.ID); err != nil && !errors.Is(err, database.ErrNotFound) {
		return 0, fmt.Errorf("error retrieving schema environment id %d: %w", graphSchema.ID, err)
	} else if !errors.Is(err, database.ErrNotFound) {
		// Environment exists - delete it first
		if err := o.openGraphSchemaRepository.DeleteSchemaEnvironment(ctx, existing.ID); err != nil {
			return 0, fmt.Errorf("error deleting schema environment %d: %w", existing.ID, err)
		}
	}

	// Create Environment
	if created, err := o.openGraphSchemaRepository.CreateSchemaEnvironment(ctx, graphSchema.SchemaExtensionId, graphSchema.EnvironmentKindId, graphSchema.SourceKindId); err != nil {
		return 0, fmt.Errorf("error creating schema environment: %w", err)
	} else {
		return created.ID, nil
	}
}

/*
Validations: https://github.com/SpecterOps/BloodHound/blob/73b569a340ef5cd459b383e3e42e707b201193ee/rfc/bh-rfc-4.md#10-validation-rules-for-environments
 1. Ensure the specified environmentKind exists in kinds table
    ** QUESTION: Documentation states to use the kind table. Is there already a database method to query the kind table to do this validation or was I supposed to create it?
 2. Ensure the specified sourceKind exists in source_kinds table (create if it doesn't, reactivate if it does)
*/
func (o *OpenGraphSchemaService) validateSchemaEnvironment(ctx context.Context, graphSchema model.SchemaEnvironment) error {
	// Validate environment kind id exists in kinds table
	if _, err := o.openGraphSchemaRepository.GetKindById(ctx, graphSchema.EnvironmentKindId); err != nil {
		return fmt.Errorf("error retrieving environment kind: %w", err)
	}

	// Get all source kinds
	if sourceKinds, err := o.openGraphSchemaRepository.GetSourceKinds(ctx); err != nil {
		return fmt.Errorf("error retrieving source kinds: %w", err)
	} else {
		// Check if source kind exists
		found := false
		for _, kind := range sourceKinds {
			if graphSchema.SourceKindId == int32(kind.ID) {
				found = true
				break
			}
		}

		if !found {
			/*
				** QUESTION: Example Environment Schema: https://github.com/SpecterOps/BloodHound/blob/73b569a340ef5cd459b383e3e42e707b201193ee/rfc/bh-rfc-4.md#9-environments-and-principal-kinds
				The RFC example uses source kind names (strings) in the environment schema, but our model uses IDs (int32). To register a new source kind, we need the
				kind name/data, not just the ID. Cannot register with only an ID - kind id/name is required for registration.
				RegisterSourceKind(ctx context.Context) func(sourceKind graph.Kind) error

				For now, this validates that the source kind should exist.
			*/
			return fmt.Errorf("invalid source kind id %d", graphSchema.SourceKindId)
		}
	}

	return nil
}

func (o *OpenGraphSchemaService) upsertPrincipalKinds(ctx context.Context, environmentID int32, principalKinds []model.SchemaEnvironmentPrincipalKind) error {
	if existingKinds, err := o.openGraphSchemaRepository.GetSchemaEnvironmentPrincipalKindsByEnvironmentId(ctx, environmentID); err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("error retrieving existing principal kinds for environment %d: %w", environmentID, err)
	} else if !errors.Is(err, database.ErrNotFound) {
		// Delete all existing principal kinds
		for _, kind := range existingKinds {
			if err := o.openGraphSchemaRepository.DeleteSchemaEnvironmentPrincipalKind(ctx, kind.EnvironmentId, kind.PrincipalKind); err != nil {
				return fmt.Errorf("error deleting principal kind %d for environment %d: %w", kind.PrincipalKind, kind.EnvironmentId, err)
			}
		}
	}

	// Create the new principal kinds
	for _, kind := range principalKinds {
		if _, err := o.openGraphSchemaRepository.CreateSchemaEnvironmentPrincipalKind(ctx, environmentID, kind.PrincipalKind); err != nil {
			return fmt.Errorf("error creating principal kind %d for environment %d: %w", kind.PrincipalKind, environmentID, err)
		}
	}

	return nil
}

/*
Validations: https://github.com/SpecterOps/BloodHound/blob/73b569a340ef5cd459b383e3e42e707b201193ee/rfc/bh-rfc-4.md#10-validation-rules-for-environments
1. Ensure all principalKinds exist in kinds table.

** QUESTION: Documentation states to use the kind table. Is there already a database method to query the kind table to do this validation or was I supposed to create it?
*/
func (o *OpenGraphSchemaService) validateSchemaEnvironmentPrincipalKind(ctx context.Context, kindID int32) error {
	if _, err := o.openGraphSchemaRepository.GetKindById(ctx, kindID); err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("error retrieving kind by id: %w", err)
	} else if errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("invalid principal kind id %d", kindID)
	}

	return nil
}
