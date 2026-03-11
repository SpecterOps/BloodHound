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

const CustomNodeIconType = "font-awesome"

// UpsertOpenGraphExtension - upserts the incoming graph extension by checking to see if the extension exists already,
// if so, deleting it and inserting the new extension.
//
// During development, it was decided to push the upsert logic down to the database layer due to difficulties of
// decoupling the database and service layers while still providing transactional guarantees. The following
// functions use models intended for the service layer and call the database public methods directly, rather
// than using an interface.
func (s *BloodhoundDB) UpsertOpenGraphExtension(ctx context.Context, graphExtensionInput model.GraphExtensionInput) (bool, error) {
	var (
		err              error
		schemaExists     bool
		createdExtension model.GraphSchemaExtension

		tx                      = s.db.WithContext(ctx).Begin()
		bloodhoundDBTransaction = BloodhoundDB{db: tx, idResolver: s.idResolver}
	)
	// Check for an immediate error after beginning the transaction
	if err = tx.Error; err != nil {
		return schemaExists, err
	}

	defer func() {
		tx.Rollback() // rollback is a no-op if the tx has already been committed
	}()

	if schemaExists, err = bloodhoundDBTransaction.cleanupExistingExtension(ctx, graphExtensionInput.ExtensionInput.Name); err != nil {
		return schemaExists, err
	} else if createdExtension, err = bloodhoundDBTransaction.CreateGraphSchemaExtension(ctx, graphExtensionInput.ExtensionInput.Name,
		graphExtensionInput.ExtensionInput.DisplayName, graphExtensionInput.ExtensionInput.Version, graphExtensionInput.ExtensionInput.Namespace); err != nil {
		return schemaExists, err
	} else if createdNodeKinds, err := bloodhoundDBTransaction.insertNodeKinds(ctx, createdExtension.ID,
		graphExtensionInput.NodeKindsInput); err != nil {
		return schemaExists, fmt.Errorf("failed to upsert node kinds: %w", err)
	} else if err = bloodhoundDBTransaction.upsertCustomIcons(ctx, createdNodeKinds); err != nil {
		return schemaExists, fmt.Errorf("failed to upsert custom node icons: %w", err)
	} else if err = bloodhoundDBTransaction.insertRelationshipKinds(ctx, createdExtension.ID,
		graphExtensionInput.RelationshipKindsInput); err != nil {
		return schemaExists, fmt.Errorf("failed to upsert edge kinds: %w", err)
	} else if err = bloodhoundDBTransaction.insertProperties(ctx,
		createdExtension.ID, graphExtensionInput.PropertiesInput); err != nil {
		return schemaExists, fmt.Errorf("failed to upsert properties: %w", err)
	} else if err = bloodhoundDBTransaction.upsertGraphEnvironments(ctx, createdExtension.ID,
		graphExtensionInput.EnvironmentsInput); err != nil {
		return schemaExists, err
	} else if err = bloodhoundDBTransaction.upsertFindingsAndRemediations(ctx, createdExtension.ID,
		graphExtensionInput.RelationshipFindingsInput); err != nil {
		return schemaExists, err
	} else if err = tx.Commit().Error; err != nil {
		return schemaExists, err
	} else {
		return schemaExists, nil
	}
}

// cleanupExistingExtension - checks to see if an extension exists for the given name, if so it will delete it.
// Returns whether the extension existed or not and the first error if encountered.
func (s *BloodhoundDB) cleanupExistingExtension(ctx context.Context, extensionName string) (bool, error) {
	var (
		err                     error
		existingGraphExtensions model.GraphSchemaExtensions
	)

	if existingGraphExtensions, _, err = s.GetGraphSchemaExtensions(ctx,
		model.Filters{"name": []model.Filter{{ // check to see if extension exists
			Operator:    model.Equals,
			Value:       extensionName,
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 1); err != nil && !errors.Is(err, ErrNotFound) {
		return false, err
	} else if len(existingGraphExtensions) > 0 {
		existingGraphExtension := existingGraphExtensions[0]
		if err = s.DeleteGraphSchemaExtension(ctx, existingGraphExtension.ID); err != nil {
			return false, err
		}
	}
	return len(existingGraphExtensions) > 0, nil
}

// insertProperties - inserts a slice of new properties for the provided extension.
func (s *BloodhoundDB) insertProperties(ctx context.Context, extensionId int32, newGraphSchemaProperties model.PropertiesInput) error {
	var (
		err error
	)

	for _, property := range newGraphSchemaProperties {
		if _, err = s.CreateGraphSchemaProperty(ctx, extensionId, property.Name,
			property.DisplayName, property.DataType, property.Description); err != nil {
			return err
		}
	}

	return nil
}

// insertRelationshipKinds - inserts a slice of new relationship kinds for the provided extension.
func (s *BloodhoundDB) insertRelationshipKinds(ctx context.Context, extensionId int32, newRelationshipKinds model.RelationshipsInput) error {
	var err error

	for _, relationshipKind := range newRelationshipKinds {
		if _, err = s.CreateGraphSchemaRelationshipKind(ctx, relationshipKind.Name, extensionId,
			relationshipKind.Description, relationshipKind.IsTraversable); err != nil {
			return err
		}
	}

	return nil
}

// insertNodeKinds - inserts a slice of new node kinds for the provided extension.
func (s *BloodhoundDB) insertNodeKinds(ctx context.Context, extensionId int32, newGraphSchemaNodeKinds model.NodesInput) (model.GraphSchemaNodeKinds, error) {
	createdNodeKinds := make(model.GraphSchemaNodeKinds, 0, len(newGraphSchemaNodeKinds))

	for _, nodeKind := range newGraphSchemaNodeKinds {
		if createdNodeKind, err := s.CreateGraphSchemaNodeKind(ctx, nodeKind.Name, extensionId,
			nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor); err != nil {
			return createdNodeKinds, err
		} else {
			createdNodeKinds = append(createdNodeKinds, createdNodeKind)
		}
	}
	return createdNodeKinds, nil
}

// upsertCustomIcons - upserts any new custom icon definitions for the provided node kinds.
func (s *BloodhoundDB) upsertCustomIcons(ctx context.Context, nodeKinds model.GraphSchemaNodeKinds) error {
	var customNodeKindsToCreate model.CustomNodeKinds
	var customNodeKindsToUpdate model.CustomNodeKinds
	if existingIconsMap, err := getExistingIconsMap(ctx, s); err != nil {
		return err
	} else {
		for _, nodeKind := range nodeKinds {
			if nodeKind.IsDisplayKind {
				if existingIcon, ok := existingIconsMap[nodeKind.Name]; ok {
					customNodeKindDefinition := parseIconDefinitionFromNodeKind(nodeKind, &existingIcon)
					customNodeKindsToUpdate = append(customNodeKindsToUpdate, customNodeKindDefinition)
				} else {
					customNodeKindDefinition := parseIconDefinitionFromNodeKind(nodeKind, nil)
					customNodeKindsToCreate = append(customNodeKindsToCreate, customNodeKindDefinition)
				}
			}

		}
		if len(customNodeKindsToCreate) > 0 {
			if _, err := s.CreateCustomNodeKinds(ctx, customNodeKindsToCreate); err != nil {
				return err
			}
		}

		if len(customNodeKindsToUpdate) > 0 {
			for _, kindDefinition := range customNodeKindsToUpdate {
				if _, err := s.UpdateCustomNodeKind(ctx, kindDefinition); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// upsertGraphEnvironments - inserts a slice of new environments for the provided extension.
func (s *BloodhoundDB) upsertGraphEnvironments(ctx context.Context, extensionID int32, environments model.EnvironmentsInput) error {
	for _, env := range environments {
		if err := s.UpsertSchemaEnvironmentWithPrincipalKinds(ctx, extensionID, env.EnvironmentKindName, env.SourceKindName, env.PrincipalKinds); err != nil {
			return fmt.Errorf("failed to upsert environment with principal kinds: %w", err)
		}
	}
	return nil
}

// upsertFindingsAndRemediations - inserts a slice of new findings/remediations for the provided extension.
func (s *BloodhoundDB) upsertFindingsAndRemediations(ctx context.Context, extensionId int32, findings model.RelationshipFindingsInput) error {
	for _, finding := range findings {
		if schemaFinding, err := s.UpsertRelationshipFinding(ctx, extensionId, finding.RelationshipKindName,
			finding.EnvironmentKindName, finding.Name, finding.DisplayName); err != nil {
			return fmt.Errorf("failed to upsert finding: %w", err)
		} else {
			if err := s.UpsertRemediation(ctx, schemaFinding.ID, finding.RemediationInput.ShortDescription,
				finding.RemediationInput.LongDescription, finding.RemediationInput.ShortRemediation, finding.RemediationInput.LongRemediation); err != nil {
				return fmt.Errorf("failed to upsert remediation: %w", err)
			}
		}
	}
	return nil
}

// getExistingIconsMap creates a map of existing icons for quick lookups.
func getExistingIconsMap(ctx context.Context, db *BloodhoundDB) (map[string]model.CustomNodeKind, error) {
	existingIconMap := make(map[string]model.CustomNodeKind)
	if existingIcons, err := db.GetCustomNodeKinds(ctx, nil); err != nil {
		return existingIconMap, fmt.Errorf("failed to get custom node kinds from database: %w", err)
	} else {
		for _, icon := range existingIcons {
			existingIconMap[icon.KindName] = icon
		}
	}
	return existingIconMap, nil
}

// parseIconDefinitionFromNodeKinds creates a CustomNodeKind that will be used in either Creating or Updating a custom node kind in the custom_node_kind & schema_node_kinds tables. If an icon exists, the name and color will be preserved if not provided.
func parseIconDefinitionFromNodeKind(nodeKind model.GraphSchemaNodeKind, existingIcon *model.CustomNodeKind) model.CustomNodeKind {
	var customNodeKind model.CustomNodeKind
	var customNodeIcon = model.CustomNodeIcon{Type: CustomNodeIconType}

	if nodeKind.Icon != "" {
		customNodeIcon.Name = nodeKind.Icon
	} else if existingIcon != nil {
		// preserve existing icon name if not provided
		customNodeIcon.Name = existingIcon.Config.Icon.Name
	}

	if nodeKind.IconColor != "" {
		customNodeIcon.Color = nodeKind.IconColor
	} else if existingIcon != nil {
		// preserve existing icon color if not provided
		customNodeIcon.Color = existingIcon.Config.Icon.Color
	}

	customNodeKind.KindName = nodeKind.Name
	customNodeKind.Config = model.CustomNodeKindConfig{Icon: customNodeIcon}
	customNodeKind.SchemaNodeKindId = &nodeKind.ID
	return customNodeKind

}
