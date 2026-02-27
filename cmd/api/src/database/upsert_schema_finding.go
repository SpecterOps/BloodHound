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

// UpsertRelationshipFinding validates and upserts a relationship finding.
// If a finding with the same name exists, it will be deleted and re-created.
func (s *BloodhoundDB) UpsertRelationshipFinding(ctx context.Context, extensionId int32, relationshipKindName, environmentKind string, name, displayName string) (model.SchemaFinding, error) {
	relationshipKindId, err := s.validateAndTranslateRelationshipKind(ctx, relationshipKindName)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	environmentKindId, err := s.validateAndTranslateEnvironmentKind(ctx, environmentKind)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	// The unique constraint on environment_kind_id of the Schema Environment table ensures no
	// duplicates exist, enabling this logic.
	environment, err := s.GetEnvironmentByEnvironmentKindId(ctx, environmentKindId)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	// Note: All schema findings uploaded via extensions are currently set to relationship findings.
	finding, err := s.replaceFinding(ctx, model.SchemaFindingTypeRelationship, extensionId, relationshipKindId, environment.ID, name, displayName)
	if err != nil {
		return model.SchemaFinding{}, err
	}

	return finding, nil
}

// validateAndTranslateRelationshipKind validates that the relationship kind exists in the kinds table.
func (s *BloodhoundDB) validateAndTranslateRelationshipKind(ctx context.Context, relationshipKindName string) (int32, error) {
	if relationshipKind, err := s.GetKindByName(ctx, relationshipKindName); err != nil && !errors.Is(err, ErrNotFound) {
		return 0, fmt.Errorf("error retrieving relationship kind '%s': %w", relationshipKindName, err)
	} else if errors.Is(err, ErrNotFound) {
		return 0, fmt.Errorf("relationship kind '%s' not found", relationshipKindName)
	} else {
		return relationshipKind.ID, nil
	}
}

// replaceFinding creates or updates a schema finding.
// If a finding with the given name exists, it deletes it first before creating the new one.
func (s *BloodhoundDB) replaceFinding(ctx context.Context, findingType model.SchemaFindingType, extensionId, kindId, environmentId int32, name, displayName string) (model.SchemaFinding, error) {
	if existing, err := s.GetSchemaFindingByName(ctx, name); err != nil && !errors.Is(err, ErrNotFound) {
		return model.SchemaFinding{}, fmt.Errorf("error retrieving schema finding: %w", err)
	} else if err == nil {
		// Finding exists - delete it first
		if err := s.DeleteSchemaFinding(ctx, existing.ID); err != nil {
			return model.SchemaFinding{}, fmt.Errorf("error deleting schema finding %d: %w", existing.ID, err)
		}
	}

	finding, err := s.CreateSchemaFinding(ctx, findingType, extensionId, kindId, environmentId, name, displayName)
	if err != nil {
		return model.SchemaFinding{}, fmt.Errorf("error creating schema finding: %w", err)
	}

	return finding, nil

}
