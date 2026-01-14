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
func (s *BloodhoundDB) UpsertFinding(ctx context.Context, extensionId int32, sourceKindName, relationshipKindName, environmentKind string, name, displayName string) (model.SchemaRelationshipFinding, error) {
	relationshipKindId, err := s.validateAndTranslateRelationshipKind(ctx, relationshipKindName)
	if err != nil {
		return model.SchemaRelationshipFinding{}, err
	}

	environmentKindId, err := s.validateAndTranslateEnvironmentKind(ctx, environmentKind)
	if err != nil {
		return model.SchemaRelationshipFinding{}, err
	}

	sourceKindId, err := s.validateAndTranslateSourceKind(ctx, sourceKindName)
	if err != nil {
		return model.SchemaRelationshipFinding{}, err
	}

	environment, err := s.GetSchemaEnvironmentByKinds(ctx, environmentKindId, sourceKindId)
	if err != nil {
		return model.SchemaRelationshipFinding{}, err
	}

	finding, err := s.upsertFinding(ctx, extensionId, relationshipKindId, environment.ID, name, displayName)
	if err != nil {
		return model.SchemaRelationshipFinding{}, err
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

// upsertFinding creates or updates a schema relationship finding.
// If an environment with the given kinds exists, it deletes it first before creating the new one.
func (s *BloodhoundDB) upsertFinding(ctx context.Context, extensionId, relationshipKindId, environmentId int32, name, displayName string) (model.SchemaRelationshipFinding, error) {
	if existing, err := s.GetSchemaRelationshipFindingByName(ctx, name); err != nil && !errors.Is(err, ErrNotFound) {
		return model.SchemaRelationshipFinding{}, fmt.Errorf("error retrieving schema relationship finding: %w", err)
	} else if err == nil {
		// Finding exists - delete it first
		if err := s.DeleteSchemaRelationshipFinding(ctx, existing.ID); err != nil {
			return model.SchemaRelationshipFinding{}, fmt.Errorf("error deleting schema relationship finding %d: %w", existing.ID, err)
		}
	}

	finding, err := s.CreateSchemaRelationshipFinding(ctx, extensionId, relationshipKindId, environmentId, name, displayName)
	if err != nil {
		return model.SchemaRelationshipFinding{}, fmt.Errorf("error creating schema relationship finding: %w", err)
	}

	return finding, nil

}
