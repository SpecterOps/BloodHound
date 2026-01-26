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

// GetGraphFindingsBySchemaExtensionId - retrieves a model.GraphFindings using the provided extension id. If no graph
// findings exist, ErrNotFound is returned.
func (s *BloodhoundDB) GetGraphFindingsBySchemaExtensionId(ctx context.Context, extensionId int32) (model.GraphFindings, error) {
	var (
		err            error
		schemaFindings []model.SchemaRelationshipFinding
		findings       = make(model.GraphFindings, 0)
	)

	if schemaFindings, err = s.GetSchemaRelationshipFindingsBySchemaExtensionId(ctx, extensionId); err != nil {
		return nil, err
	}
	for _, finding := range schemaFindings {
		var (
			envKind, edgeKind model.Kind
			sourceKind        SourceKind
			remediation       model.Remediation
			environment       model.SchemaEnvironment
		)
		if edgeKind, err = s.GetKindById(ctx, finding.RelationshipKindId); err != nil {
			return nil, fmt.Errorf("unable to retrieve finding edge kind by id: %w", err)
		} else if environment, err = s.GetEnvironmentById(ctx, finding.EnvironmentId); err != nil {
			return nil, fmt.Errorf("unable to retrieve finding environment by id: %w", err)
		} else if envKind, err = s.GetKindById(ctx, environment.EnvironmentKindId); err != nil {
			return nil, fmt.Errorf("unable to retrieve finding environment kind by id: %w", err)
		} else if sourceKind, err = s.GetSourceKindById(ctx, int(environment.SourceKindId)); err != nil {
			return nil, fmt.Errorf("unable to retrieve finding environment source by id: %w", err)
		} else if remediation, err = s.GetRemediationByFindingId(ctx, finding.ID); err != nil && !errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("unable to retrieve remediation by finding id: %w", err)
		}

		findings = append(findings, model.GraphFinding{
			ID:                finding.ID,
			Name:              finding.Name,
			SchemaExtensionId: finding.SchemaExtensionId,
			SourceKind:        sourceKind.Name.String(),
			DisplayName:       finding.DisplayName,
			RelationshipKind:  edgeKind.Name,
			EnvironmentKind:   envKind.Name,
			Remediation:       remediation,
		})
	}

	return findings, nil
}

func (s *BloodhoundDB) upsertFindingsAndRemediations(ctx context.Context, extensionId int32, findings []model.GraphFinding) error {
	for _, finding := range findings {
		if schemaFinding, err := s.UpsertFinding(ctx, extensionId, finding.SourceKind,
			finding.RelationshipKind, finding.EnvironmentKind, finding.Name, finding.DisplayName); err != nil {
			return fmt.Errorf("failed to upsert finding: %w", err)
		} else {
			if err := s.UpsertRemediation(ctx, schemaFinding.ID, finding.Remediation.ShortDescription,
				finding.Remediation.LongDescription, finding.Remediation.ShortRemediation, finding.Remediation.LongRemediation); err != nil {
				return fmt.Errorf("failed to upsert remediation: %w", err)
			}
		}
	}
	return nil
}

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

	// The unique constraint on (environment_kind_id, source_kind_id) of the Schema Environment table ensures no
	// duplicate pairs exist, enabling this logic.
	environment, err := s.GetEnvironmentByKinds(ctx, environmentKindId, sourceKindId)
	if err != nil {
		return model.SchemaRelationshipFinding{}, err
	}

	finding, err := s.replaceFinding(ctx, extensionId, relationshipKindId, environment.ID, name, displayName)
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

// replaceFinding creates or updates a schema relationship finding.
// If a finding with the given name exists, it deletes it first before creating the new one.
func (s *BloodhoundDB) replaceFinding(ctx context.Context, extensionId, relationshipKindId, environmentId int32, name, displayName string) (model.SchemaRelationshipFinding, error) {
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
