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
)

// UpsertFinding validates and upserts a finding.
// If a finding with the same name exists, it will be deleted and re-created.
func (s *BloodhoundDB) UpsertFinding(ctx context.Context, extensionId int32, sourceKindName, kindName, environmentKind string, name, displayName string) (model.SchemaFinding, error) {
	kindId, err := s.validateAndTranslateKind(ctx, kindName)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	environmentKindId, err := s.validateAndTranslateEnvironmentKind(ctx, environmentKind)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	sourceKindId, err := s.validateAndTranslateSourceKind(ctx, sourceKindName)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	// The unique constraint on (environment_kind_id, source_kind_id) of the Schema Environment table ensures no
	// duplicate pairs exist, enabling this logic.
	environment, err := s.GetEnvironmentByKinds(ctx, environmentKindId, sourceKindId)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	finding, err := s.replaceFinding(ctx, extensionId, kindId, environment.ID, name, displayName)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	return finding, nil
}

// validateAndTranslateKind validates that the kind exists in the kinds table.
func (s *BloodhoundDB) validateAndTranslateKind(ctx context.Context, kindName string) (int32, error) {
	if relationshipKind, err := s.GetKindByName(ctx, kindName); err != nil && !errors.Is(err, ErrNotFound) {
		return 0, fmt.Errorf("error retrieving kind '%s': %w", kindName, err)
	} else if errors.Is(err, ErrNotFound) {
		return 0, fmt.Errorf("kind '%s' not found", kindName)
	} else {
		return relationshipKind.ID, nil
	}
}

// replaceFinding creates or updates a schema relationship finding.
// If a finding with the given name exists, it deletes it first before creating the new one.
func (s *BloodhoundDB) replaceFinding(ctx context.Context, extensionId, kindId, environmentId int32, name, displayName string) (model.SchemaFinding, error) {
	if existing, err := s.GetSchemaFindingByName(ctx, name); err != nil && !errors.Is(err, ErrNotFound) {
		return model.SchemaFinding{}, fmt.Errorf("error retrieving schema relationship finding: %w", err)
	} else if err == nil {
		// Finding exists - delete it first
		if err := s.DeleteSchemaFinding(ctx, existing.ID); err != nil {
			return model.SchemaFinding{}, fmt.Errorf("error deleting schema relationship finding %d: %w", existing.ID, err)
		}
	}

	// Note: All schema findings uploaded via extensions are currently set to relationship findings.
	finding, err := s.CreateSchemaFinding(ctx, model.SchemaFindingTypeRelationship, extensionId, kindId, environmentId, name, displayName)
	if err != nil {
		return model.SchemaFinding{}, fmt.Errorf("error creating schema relationship finding: %w", err)
	}

	return finding, nil

}
