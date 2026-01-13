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
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
)

// UpsertSchemaEnvironmentWithPrincipalKinds creates or updates an environment with its principal kinds.
// If an environment with the same environment kind and source kind exists, it will be replaced.
func (s *BloodhoundDB) UpsertSchemaEnvironmentWithPrincipalKinds(ctx context.Context, schemaExtensionId int32, environmentKind string, sourceKind string, principalKinds []string) error {
	environment := model.SchemaEnvironment{
		SchemaExtensionId: schemaExtensionId,
	}

	envKind, err := s.validateAndTranslateEnvironmentKind(ctx, environmentKind)
	if err != nil {
		return err
	}

	sourceKindID, err := s.validateAndTranslateSourceKind(ctx, sourceKind)
	if err != nil {
		return err
	}

	translatedPrincipalKinds, err := s.validateAndTranslatePrincipalKinds(ctx, principalKinds)
	if err != nil {
		return err
	}

	environment.EnvironmentKindId = int32(envKind.ID)
	environment.SourceKindId = sourceKindID

	envID, err := s.upsertSchemaEnvironment(ctx, environment)
	if err != nil {
		return fmt.Errorf("error upserting schema environment: %w", err)
	}

	if err := s.upsertPrincipalKinds(ctx, envID, translatedPrincipalKinds); err != nil {
		return fmt.Errorf("error upserting principal kinds: %w", err)
	}

	return nil
}

// validateAndTranslateEnvironmentKind validates that the environment kind exists in the kinds table.
func (s *BloodhoundDB) validateAndTranslateEnvironmentKind(ctx context.Context, environmentKindName string) (model.Kind, error) {
	if envKind, err := s.GetKindByName(ctx, environmentKindName); err != nil && !errors.Is(err, ErrNotFound) {
		return model.Kind{}, fmt.Errorf("error retrieving environment kind '%s': %w", environmentKindName, err)
	} else if errors.Is(err, ErrNotFound) {
		return model.Kind{}, fmt.Errorf("environment kind '%s' not found", environmentKindName)
	} else {
		return envKind, nil
	}
}

// validateAndTranslateSourceKind validates that the source kind exists in the source_kinds table.
// If not found, it registers the source kind and returns its ID so it can be added to the Environment object.
func (s *BloodhoundDB) validateAndTranslateSourceKind(ctx context.Context, sourceKindName string) (int32, error) {
	if sourceKind, err := s.GetSourceKindByName(ctx, sourceKindName); err != nil && !errors.Is(err, ErrNotFound) {
		return 0, fmt.Errorf("error retrieving source kind '%s': %w", sourceKindName, err)
	} else if err == nil {
		return int32(sourceKind.ID), nil
	}

	// If source kind is not found, register it. If it exists and is inactive, it will automatically update as active.
	kindType := graph.StringKind(sourceKindName)
	if err := s.RegisterSourceKind(ctx)(kindType); err != nil {
		return 0, fmt.Errorf("error registering source kind '%s': %w", sourceKindName, err)
	}

	if sourceKind, err := s.GetSourceKindByName(ctx, sourceKindName); err != nil {
		return 0, fmt.Errorf("error retrieving newly registered source kind '%s': %w", sourceKindName, err)
	} else {
		return int32(sourceKind.ID), nil
	}
}

// validateAndTranslatePrincipalKinds ensures all principalKinds exist in the kinds table.
// It also translates them to IDs so they can be upserted into the database.
func (s *BloodhoundDB) validateAndTranslatePrincipalKinds(ctx context.Context, principalKindNames []string) ([]model.SchemaEnvironmentPrincipalKind, error) {
	principalKinds := make([]model.SchemaEnvironmentPrincipalKind, len(principalKindNames))

	for i, kindName := range principalKindNames {
		if kind, err := s.GetKindByName(ctx, kindName); err != nil && !errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("error retrieving principal kind by name '%s': %w", kindName, err)
		} else if errors.Is(err, ErrNotFound) {
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
// If an environment with the given kinds exists, it deletes it first before creating the new one.
func (s *BloodhoundDB) upsertSchemaEnvironment(ctx context.Context, graphSchema model.SchemaEnvironment) (int32, error) {
	if existing, err := s.GetSchemaEnvironmentByKinds(ctx, graphSchema.EnvironmentKindId, graphSchema.SourceKindId); err != nil && !errors.Is(err, ErrNotFound) {
		return 0, fmt.Errorf("error retrieving schema environment: %w", err)
	} else if !errors.Is(err, ErrNotFound) {
		// Environment exists - delete it first
		if err := s.DeleteSchemaEnvironment(ctx, existing.ID); err != nil {
			return 0, fmt.Errorf("error deleting schema environment %d: %w", existing.ID, err)
		}
	}

	// Create Environment
	if created, err := s.CreateSchemaEnvironment(ctx, graphSchema.SchemaExtensionId, graphSchema.EnvironmentKindId, graphSchema.SourceKindId); err != nil {
		return 0, fmt.Errorf("error creating schema environment: %w", err)
	} else {
		return created.ID, nil
	}
}

// upsertPrincipalKinds deletes all existing principal kinds for an environment and creates new ones.
func (s *BloodhoundDB) upsertPrincipalKinds(ctx context.Context, environmentID int32, principalKinds []model.SchemaEnvironmentPrincipalKind) error {
	if existingKinds, err := s.GetSchemaEnvironmentPrincipalKindsByEnvironmentId(ctx, environmentID); err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("error retrieving existing principal kinds for environment %d: %w", environmentID, err)
	} else if !errors.Is(err, ErrNotFound) {
		// Delete all existing principal kinds
		for _, kind := range existingKinds {
			if err := s.DeleteSchemaEnvironmentPrincipalKind(ctx, kind.EnvironmentId, kind.PrincipalKind); err != nil {
				return fmt.Errorf("error deleting principal kind %d for environment %d: %w", kind.PrincipalKind, kind.EnvironmentId, err)
			}
		}
	}

	// Create the new principal kinds
	for _, kind := range principalKinds {
		if _, err := s.CreateSchemaEnvironmentPrincipalKind(ctx, environmentID, kind.PrincipalKind); err != nil {
			return fmt.Errorf("error creating principal kind %d for environment %d: %w", kind.PrincipalKind, environmentID, err)
		}
	}

	return nil
}
