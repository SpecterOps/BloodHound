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

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
)

// UpsertSchemaEnvironmentWithPrincipalKinds takes a slice of environments, validates and translates each environment.
// The translation is used to upsert the environments into the database.
// If an existing environment is found to already exist in the database, the existing environment will be removed and the new one will be uploaded.
func (o *OpenGraphSchemaService) UpsertSchemaEnvironmentWithPrincipalKinds(ctx context.Context, schemaExtensionId int32, environments []v2.Environment) error {
	for _, env := range environments {
		environment := model.SchemaEnvironment{
			SchemaExtensionId: schemaExtensionId,
		}

		if updatedEnv, principalKinds, err := o.validateAndTranslateEnvironment(ctx, environment, env); err != nil {
			return fmt.Errorf("error validating and translating environment: %w", err)
		} else if envID, err := o.upsertSchemaEnvironment(ctx, updatedEnv); err != nil {
			return fmt.Errorf("error upserting schema environment: %w", err)
		} else if err := o.upsertPrincipalKinds(ctx, envID, principalKinds); err != nil {
			return fmt.Errorf("error upserting principal kinds: %w", err)
		}
	}

	return nil
}

// validateAndTranslateEnvironment validates that the environment kind, source kind, and the principal kinds exist in the database.
// It is then translated from the API model to the Database model to prepare it for insert.
func (o *OpenGraphSchemaService) validateAndTranslateEnvironment(ctx context.Context, environment model.SchemaEnvironment, env v2.Environment) (model.SchemaEnvironment, []model.SchemaEnvironmentPrincipalKind, error) {
	if envKind, err := o.validateAndTranslateEnvironmentKind(ctx, env.EnvironmentKind); err != nil {
		return model.SchemaEnvironment{}, nil, err
	} else if sourceKindID, err := o.validateAndTranslateSourceKind(ctx, env.SourceKind); err != nil {
		return model.SchemaEnvironment{}, nil, err
	} else if principalKinds, err := o.validateAndTranslatePrincipalKinds(ctx, env.PrincipalKinds); err != nil {
		return model.SchemaEnvironment{}, nil, err
	} else {
		// Update environment with translated IDs
		environment.EnvironmentKindId = int32(envKind.ID)
		environment.SourceKindId = sourceKindID

		return environment, principalKinds, nil
	}
}

// validateAndTranslateEnvironmentKind validates that the environment kind exists in the kinds table.
func (o *OpenGraphSchemaService) validateAndTranslateEnvironmentKind(ctx context.Context, environmentKindName string) (model.Kind, error) {
	if envKind, err := o.openGraphSchemaRepository.GetKindByName(ctx, environmentKindName); err != nil && !errors.Is(err, database.ErrNotFound) {
		return model.Kind{}, fmt.Errorf("error retrieving environment kind '%s': %w", environmentKindName, err)
	} else if errors.Is(err, database.ErrNotFound){
		return model.Kind{}, fmt.Errorf("environment kind '%s' not found", environmentKindName)
	} else {
		return envKind, nil
	}
}

// validateAndTranslateSourceKind validates that the source kind exists in the source_kinds table.
// If not found, it registers the source kind and returns its ID so it can be added to the Environment object.
func (o *OpenGraphSchemaService) validateAndTranslateSourceKind(ctx context.Context, sourceKindName string) (int32, error) {
	if sourceKind, err := o.openGraphSchemaRepository.GetSourceKindByName(ctx, sourceKindName); err != nil && !errors.Is(err, database.ErrNotFound) {
		return 0, fmt.Errorf("error retrieving source kind '%s': %w", sourceKindName, err)
	} else if err == nil {
		return int32(sourceKind.ID), nil
	}

	// If source kind is not found, register it. If it exists and is inactive, it will automatically update as active.
	kindType := graph.StringKind(sourceKindName)
	if err := o.openGraphSchemaRepository.RegisterSourceKind(ctx)(kindType); err != nil {
		return 0, fmt.Errorf("error registering source kind '%s': %w", sourceKindName, err)
	}

	if sourceKind, err := o.openGraphSchemaRepository.GetSourceKindByName(ctx, sourceKindName); err != nil {
		return 0, fmt.Errorf("error retrieving newly registered source kind '%s': %w", sourceKindName, err)
	} else {
		return int32(sourceKind.ID), nil
	}
}

// validateAndTranslatePrincipalKinds ensures all principalKinds exist in the kinds table.
// It also translates them to IDs so they can be upserted into the database.
func (o *OpenGraphSchemaService) validateAndTranslatePrincipalKinds(ctx context.Context, principalKindNames []string) ([]model.SchemaEnvironmentPrincipalKind, error) {
	principalKinds := make([]model.SchemaEnvironmentPrincipalKind, len(principalKindNames))

	for i, kindName := range principalKindNames {
		if kind, err := o.openGraphSchemaRepository.GetKindByName(ctx, kindName); err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, fmt.Errorf("error retrieving principal kind by name '%s': %w", kindName, err)
		} else if errors.Is(err, database.ErrNotFound){
			return nil, fmt.Errorf("principal kind '%s' not found", kindName)
		} else {
			principalKinds[i] = model.SchemaEnvironmentPrincipalKind{
				PrincipalKind: int32(kind.ID),
			}
		}
	}

	return principalKinds, nil
}

// upsertSchemaEnvironment creates or updates a schema environment.
// If an environment with the given ID exists, it deletes it first before creating the new one.
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

// upsertPrincipalKinds deletes all existing principal kinds for an environment and creates new ones.
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
