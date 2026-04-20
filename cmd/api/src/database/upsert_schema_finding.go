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

// resolveFindingFKs translates the FK names in a RelationshipFindingInput to their corresponding IDs.
func (s *BloodhoundDB) resolveFindingFKs(ctx context.Context, input model.RelationshipFindingInput) (int32, int32, error) {
	if relKind, err := s.GetKindByName(ctx, input.RelationshipKindName); err != nil {
		return 0, 0, fmt.Errorf("error retrieving relationship kind '%s': %w", input.RelationshipKindName, err)
	} else if environment, err := s.GetEnvironmentKindName(ctx, input.EnvironmentKindName); err != nil {
		return 0, 0, fmt.Errorf("error retrieving environment kind '%s': %w", input.EnvironmentKindName, err)
	} else {
		return relKind.ID, environment.ID, nil
	}
}

// applyFindingInput applies resolved FK IDs and mutable fields from input onto an existing finding,
// returning the updated struct.
func applyFindingInput(existing model.SchemaFinding, relKindId, environmentId int32, input model.RelationshipFindingInput) model.SchemaFinding {
	existing.Type = model.SchemaFindingTypeRelationship
	existing.DisplayName = input.DisplayName
	existing.KindId = relKindId
	existing.EnvironmentId = environmentId
	return existing
}

// CreateFindingWithRemediation translates FK names, creates the finding, and creates its 1:1 remediation.
func (s *BloodhoundDB) CreateFindingWithRemediation(ctx context.Context, extensionId int32, input model.RelationshipFindingInput) (model.SchemaFinding, error) {
	if relKindId, environmentId, err := s.resolveFindingFKs(ctx, input); err != nil {
		return model.SchemaFinding{}, err
	} else if finding, err := s.CreateSchemaFinding(ctx, model.SchemaFindingTypeRelationship,
		extensionId, relKindId, environmentId, input.Name, input.DisplayName); err != nil {
		return model.SchemaFinding{}, fmt.Errorf("error creating finding: %w", err)
	} else if _, err := s.CreateRemediation(ctx, finding.ID,
		input.RemediationInput.ShortDescription, input.RemediationInput.LongDescription,
		input.RemediationInput.ShortRemediation, input.RemediationInput.LongRemediation); err != nil {
		return model.SchemaFinding{}, fmt.Errorf("error creating remediation: %w", err)
	} else {
		return finding, nil
	}
}

// UpdateFindingWithRemediation translates FK names, updates the finding, and updates its 1:1 remediation.
func (s *BloodhoundDB) UpdateFindingWithRemediation(ctx context.Context, existing model.SchemaFinding, input model.RelationshipFindingInput) (model.SchemaFinding, error) {
	if relKindId, environmentId, err := s.resolveFindingFKs(ctx, input); err != nil {
		return model.SchemaFinding{}, err
	} else if updated, err := s.UpdateSchemaFinding(ctx, applyFindingInput(existing, relKindId, environmentId, input)); err != nil {
		return model.SchemaFinding{}, fmt.Errorf("error updating finding: %w", err)
	} else if _, err := s.UpdateRemediation(ctx, updated.ID,
		input.RemediationInput.ShortDescription, input.RemediationInput.LongDescription,
		input.RemediationInput.ShortRemediation, input.RemediationInput.LongRemediation); err != nil {
		return model.SchemaFinding{}, fmt.Errorf("error updating remediation: %w", err)
	} else {
		return updated, nil
	}
}
