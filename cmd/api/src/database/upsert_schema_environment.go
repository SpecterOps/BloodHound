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
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// CreateEnvironmentWithPrincipalKinds translates FK names, creates the environment row,
// and reconciles its principal kinds.
func (s *BloodhoundDB) CreateEnvironmentWithPrincipalKinds(ctx context.Context, extensionId int32, input model.EnvironmentInput) (model.SchemaEnvironment, error) {
	if envKind, err := s.GetKindByName(ctx, input.EnvironmentKindName); err != nil {
		return model.SchemaEnvironment{}, fmt.Errorf("error retrieving environment kind '%s': %w", input.EnvironmentKindName, err)
	} else if sourceKind, err := s.UpsertKind(ctx, input.SourceKindName); err != nil {
		return model.SchemaEnvironment{}, fmt.Errorf("error registering source kind '%s': %w", input.SourceKindName, err)
	} else if principalKinds, err := s.buildPrincipalKindsFromNames(ctx, input.PrincipalKinds); err != nil {
		return model.SchemaEnvironment{}, err
	} else if created, err := s.CreateEnvironment(ctx, extensionId, envKind.ID, sourceKind.ID); err != nil {
		return model.SchemaEnvironment{}, fmt.Errorf("error creating environment: %w", err)
	} else if _, err := reconcile(ctx, principalKinds, nil, s.principalKindReconcileConfig(created.ID)); err != nil {
		return model.SchemaEnvironment{}, fmt.Errorf("error reconciling principal kinds: %w", err)
	} else {
		return created, nil
	}
}

// UpdateEnvironmentWithPrincipalKinds translates FK names, updates the environment row if
// source kind changed, and reconciles its principal kinds.
func (s *BloodhoundDB) UpdateEnvironmentWithPrincipalKinds(ctx context.Context, existing model.SchemaEnvironment, input model.EnvironmentInput) (model.SchemaEnvironment, error) {
	// Update the source kind only when it has changed.
	if sourceKind, err := s.UpsertKind(ctx, input.SourceKindName); err != nil {
		return model.SchemaEnvironment{}, fmt.Errorf("error registering source kind '%s': %w", input.SourceKindName, err)
	} else if existing.SourceKindId != sourceKind.ID {
		existing.SourceKindId = sourceKind.ID
		if existing, err = s.UpdateEnvironment(ctx, existing); err != nil {
			return model.SchemaEnvironment{}, fmt.Errorf("error updating environment: %w", err)
		}
	}

	// Always reconcile principal kinds against the (possibly updated) environment.
	if principalKindsInput, err := s.buildPrincipalKindsFromNames(ctx, input.PrincipalKinds); err != nil {
		return model.SchemaEnvironment{}, err
	} else if existingPrincipalKinds, err := s.GetPrincipalKindsByEnvironmentId(ctx, existing.ID); err != nil {
		return model.SchemaEnvironment{}, fmt.Errorf("error fetching existing principal kinds: %w", err)
	} else if _, err = reconcile(ctx, principalKindsInput, existingPrincipalKinds, s.principalKindReconcileConfig(existing.ID)); err != nil {
		return model.SchemaEnvironment{}, fmt.Errorf("error reconciling principal kinds: %w", err)
	} else {
		return existing, nil
	}
}

// buildPrincipalKindsFromNames looks up each kind by name and returns the corresponding
// principal kind structs. Returns an error if any kind does not exist.
func (s *BloodhoundDB) buildPrincipalKindsFromNames(ctx context.Context, names []string) ([]model.SchemaEnvironmentPrincipalKind, error) {
	principalKinds := make([]model.SchemaEnvironmentPrincipalKind, len(names))

	for i, name := range names {
		if kind, err := s.GetKindByName(ctx, name); err != nil {
			return nil, fmt.Errorf("error retrieving principal kind '%s': %w", name, err)
		} else {
			principalKinds[i] = model.SchemaEnvironmentPrincipalKind{PrincipalKind: kind.ID}
		}
	}

	return principalKinds, nil
}

// principalKindReconcileConfig returns the reconcileConfig for principal kinds.
// Inputs are pre-resolved SchemaEnvironmentPrincipalKind structs, environmentId is closed over by create.
func (s *BloodhoundDB) principalKindReconcileConfig(environmentId int32) reconcileConfig[model.SchemaEnvironmentPrincipalKind, model.SchemaEnvironmentPrincipalKind, int32] {
	return reconcileConfig[model.SchemaEnvironmentPrincipalKind, model.SchemaEnvironmentPrincipalKind, int32]{
		getInputKey:    func(input model.SchemaEnvironmentPrincipalKind) int32 { return input.PrincipalKind },
		getExistingKey: func(existing model.SchemaEnvironmentPrincipalKind) int32 { return existing.PrincipalKind },
		create: func(ctx context.Context, input model.SchemaEnvironmentPrincipalKind) (model.SchemaEnvironmentPrincipalKind, error) {
			return s.CreatePrincipalKind(ctx, environmentId, input.PrincipalKind)
		},
		// no-op
		update: func(_ context.Context, existing model.SchemaEnvironmentPrincipalKind, _ model.SchemaEnvironmentPrincipalKind) (model.SchemaEnvironmentPrincipalKind, error) {
			return existing, nil
		},
		delete: func(ctx context.Context, existing model.SchemaEnvironmentPrincipalKind) error {
			return s.DeletePrincipalKind(ctx, existing.EnvironmentId, existing.PrincipalKind)
		},
	}
}
