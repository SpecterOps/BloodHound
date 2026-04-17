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

// UpsertOpenGraphExtension reconciles the full state of a graph extension and all of its child
// entities (node kinds, relationship kinds, environments, and findings) against the provided input.
//
// Each entity set is diffed against the input: absent rows are deleted, matching rows are updated
// in place with their IDs preserved, and new rows are created. All mutations run inside a single
// transaction that rolls back on any error.
//
// Returns true if the extension already existed before this call, false if it was newly created.
// Returns ErrGraphExtensionBuiltIn if the named extension is a built-in and cannot be modified.
func (s *BloodhoundDB) UpsertOpenGraphExtension(ctx context.Context, graphExtensionInput model.GraphExtensionInput) (bool, error) {
	var (
		err          error
		schemaExists bool
		extension    model.GraphSchemaExtension

		tx                      = s.db.WithContext(ctx).Begin()
		bloodhoundDBTransaction = BloodhoundDB{db: tx, idResolver: s.idResolver}
	)

	defer func() {
		tx.Rollback()
	}()

	if err = tx.Error; err != nil {
		return schemaExists, err
	} else if extension, schemaExists, err = bloodhoundDBTransaction.findOrCreateExtension(ctx, graphExtensionInput.ExtensionInput); err != nil {
		return schemaExists, err
	} else if existingNodeKinds, err := bloodhoundDBTransaction.GetGraphSchemaNodeKindsByExtensionId(ctx, extension.ID); err != nil {
		return schemaExists, fmt.Errorf("failed to fetch existing node kinds: %w", err)
	} else if reconciledNodeKinds, err := reconcile(ctx, graphExtensionInput.NodeKindsInput, existingNodeKinds, bloodhoundDBTransaction.nodeKindReconcileConfig(extension.ID)); err != nil {
		return schemaExists, fmt.Errorf("failed to reconcile node kinds: %w", err)
	} else if err := bloodhoundDBTransaction.upsertCustomIcons(ctx, reconciledNodeKinds); err != nil {
		return schemaExists, fmt.Errorf("failed to upsert custom node icons: %w", err)
	} else if existingRelationshipKinds, err := bloodhoundDBTransaction.GetGraphSchemaRelationshipKindsByExtensionId(ctx, extension.ID); err != nil {
		return schemaExists, fmt.Errorf("failed to fetch existing relationship kinds: %w", err)
	} else if _, err := reconcile(ctx, graphExtensionInput.RelationshipKindsInput, existingRelationshipKinds, bloodhoundDBTransaction.relationshipKindReconcileConfig(extension.ID)); err != nil {
		return schemaExists, fmt.Errorf("failed to reconcile relationship kinds: %w", err)
	} else if existingEnvironments, err := bloodhoundDBTransaction.GetEnvironmentsByExtensionId(ctx, extension.ID); err != nil {
		return schemaExists, fmt.Errorf("failed to fetch existing environments: %w", err)
	} else if _, err := reconcile(ctx, graphExtensionInput.EnvironmentsInput, existingEnvironments, bloodhoundDBTransaction.environmentReconcileConfig(extension.ID)); err != nil {
		return schemaExists, fmt.Errorf("failed to reconcile environments: %w", err)
	} else if existingFindings, err := bloodhoundDBTransaction.GetSchemaFindingsByExtensionId(ctx, extension.ID); err != nil {
		return schemaExists, fmt.Errorf("failed to fetch existing findings: %w", err)
	} else if _, err := reconcile(ctx, graphExtensionInput.RelationshipFindingsInput, existingFindings, bloodhoundDBTransaction.findingReconcileConfig(extension.ID)); err != nil {
		return schemaExists, fmt.Errorf("failed to reconcile findings: %w", err)
	} else if err = tx.Commit().Error; err != nil {
		return schemaExists, err
	} else {
		return schemaExists, nil
	}
}

// findOrCreateExtension looks up an extension by name. If one exists, its mutable metadata is
// updated and returned with existed=true. Otherwise a new row is created and returned with existed=false.
func (s *BloodhoundDB) findOrCreateExtension(ctx context.Context, extensionInput model.ExtensionInput) (model.GraphSchemaExtension, bool, error) {
	if existingExtensions, _, err := s.GetGraphSchemaExtensions(ctx,
		model.Filters{"name": []model.Filter{{
			Operator:    model.Equals,
			Value:       extensionInput.Name,
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 1); err != nil && !errors.Is(err, ErrNotFound) {
		return model.GraphSchemaExtension{}, false, err
	} else if len(existingExtensions) > 0 {
		if existingExtensions[0].IsBuiltin {
			return model.GraphSchemaExtension{}, true, model.ErrGraphExtensionBuiltIn
		}
		return s.updateExistingExtension(ctx, existingExtensions[0], extensionInput)
	} else {
		return s.createNewExtension(ctx, extensionInput)
	}
}

// updateExistingExtension updates the mutable metadata of an existing extension in place,
// preserving its ID and returning the updated row alongside a schemaExists flag of true.
func (s *BloodhoundDB) updateExistingExtension(ctx context.Context, existing model.GraphSchemaExtension, extensionInput model.ExtensionInput) (model.GraphSchemaExtension, bool, error) {
	existing.DisplayName = extensionInput.GetDisplayName()
	existing.Version = extensionInput.Version
	existing.Namespace = extensionInput.Namespace

	if updated, err := s.UpdateGraphSchemaExtension(ctx, existing); err != nil {
		return model.GraphSchemaExtension{}, false, fmt.Errorf("error updating existing extension: %w", err)
	} else {
		return updated, true, nil
	}
}

// createNewExtension creates a new extension row and returns it alongside a schemaExists flag of false.
func (s *BloodhoundDB) createNewExtension(ctx context.Context, extensionInput model.ExtensionInput) (model.GraphSchemaExtension, bool, error) {
	if created, err := s.CreateGraphSchemaExtension(ctx,
		extensionInput.Name, extensionInput.GetDisplayName(),
		extensionInput.Version, extensionInput.Namespace); err != nil {
		return model.GraphSchemaExtension{}, false, fmt.Errorf("error creating extension: %w", err)
	} else {
		return created, false, nil
	}
}

// nodeKindReconcileConfig returns the reconcileConfig for node kinds, keyed by name.
// extensionId is closed over by the create callback.
func (s *BloodhoundDB) nodeKindReconcileConfig(extensionId int32) reconcileConfig[model.NodeInput, model.GraphSchemaNodeKind, string] {
	return reconcileConfig[model.NodeInput, model.GraphSchemaNodeKind, string]{
		getInputKey:    func(input model.NodeInput) string { return input.Name },
		getExistingKey: func(existing model.GraphSchemaNodeKind) string { return existing.Name },
		create: func(ctx context.Context, input model.NodeInput) (model.GraphSchemaNodeKind, error) {
			return s.CreateGraphSchemaNodeKind(ctx, input.Name, extensionId,
				input.DisplayName, input.Description, input.IsDisplayKind, input.Icon, input.IconColor)
		},
		update: func(ctx context.Context, existing model.GraphSchemaNodeKind, input model.NodeInput) (model.GraphSchemaNodeKind, error) {
			existing.DisplayName = input.DisplayName
			existing.Description = input.Description
			existing.IsDisplayKind = input.IsDisplayKind
			existing.Icon = input.Icon
			existing.IconColor = input.IconColor
			return s.UpdateGraphSchemaNodeKind(ctx, existing)
		},
		delete: func(ctx context.Context, existing model.GraphSchemaNodeKind) error {
			return s.DeleteGraphSchemaNodeKind(ctx, existing.ID)
		},
	}
}

// relationshipKindReconcileConfig returns the reconcileConfig for relationship kinds, keyed by name.
// extensionId is closed over by the create callback.
func (s *BloodhoundDB) relationshipKindReconcileConfig(extensionId int32) reconcileConfig[model.RelationshipInput, model.GraphSchemaRelationshipKind, string] {
	return reconcileConfig[model.RelationshipInput, model.GraphSchemaRelationshipKind, string]{
		getInputKey:    func(input model.RelationshipInput) string { return input.Name },
		getExistingKey: func(existing model.GraphSchemaRelationshipKind) string { return existing.Name },
		create: func(ctx context.Context, input model.RelationshipInput) (model.GraphSchemaRelationshipKind, error) {
			return s.CreateGraphSchemaRelationshipKind(ctx, input.Name, extensionId,
				input.Description, input.IsTraversable)
		},
		update: func(ctx context.Context, existing model.GraphSchemaRelationshipKind, input model.RelationshipInput) (model.GraphSchemaRelationshipKind, error) {
			existing.Description = input.Description
			existing.IsTraversable = input.IsTraversable
			return s.UpdateGraphSchemaRelationshipKind(ctx, existing)
		},
		delete: func(ctx context.Context, existing model.GraphSchemaRelationshipKind) error {
			return s.DeleteGraphSchemaRelationshipKind(ctx, existing.ID)
		},
	}
}

// environmentReconcileConfig returns the reconcileConfig for environments.
// The create/update callbacks handle FK translation and principal kind reconciliation internally.
func (s *BloodhoundDB) environmentReconcileConfig(extensionId int32) reconcileConfig[model.EnvironmentInput, model.SchemaEnvironment, string] {
	return reconcileConfig[model.EnvironmentInput, model.SchemaEnvironment, string]{
		getInputKey:    func(input model.EnvironmentInput) string { return input.EnvironmentKindName },
		getExistingKey: func(existing model.SchemaEnvironment) string { return existing.EnvironmentKindName },
		create: func(ctx context.Context, input model.EnvironmentInput) (model.SchemaEnvironment, error) {
			return s.CreateEnvironmentWithPrincipalKinds(ctx, extensionId, input)
		},
		update: func(ctx context.Context, existing model.SchemaEnvironment, input model.EnvironmentInput) (model.SchemaEnvironment, error) {
			return s.UpdateEnvironmentWithPrincipalKinds(ctx, existing, input)
		},
		delete: func(ctx context.Context, existing model.SchemaEnvironment) error {
			return s.DeleteEnvironment(ctx, existing.ID)
		},
	}
}

// findingReconcileConfig returns the reconcileConfig for findings, keyed by name.
// The create/update callbacks handle FK translation and paired remediation create/update internally.
func (s *BloodhoundDB) findingReconcileConfig(extensionId int32) reconcileConfig[model.RelationshipFindingInput, model.SchemaFinding, string] {
	return reconcileConfig[model.RelationshipFindingInput, model.SchemaFinding, string]{
		getInputKey:    func(input model.RelationshipFindingInput) string { return input.Name },
		getExistingKey: func(existing model.SchemaFinding) string { return existing.Name },
		create: func(ctx context.Context, input model.RelationshipFindingInput) (model.SchemaFinding, error) {
			return s.CreateFindingWithRemediation(ctx, extensionId, input)
		},
		update: func(ctx context.Context, existing model.SchemaFinding, input model.RelationshipFindingInput) (model.SchemaFinding, error) {
			return s.UpdateFindingWithRemediation(ctx, existing, input)
		},
		delete: func(ctx context.Context, existing model.SchemaFinding) error {
			return s.DeleteSchemaFinding(ctx, existing.ID)
		},
	}
}

// upsertCustomIcons upserts custom icon definitions for the provided node kinds.
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

// parseIconDefinitionFromNodeKind builds a CustomNodeKind for use in create or update operations against the
// custom_node_kinds and schema_node_kinds tables. If an existingIcon is provided, its name and color are
// preserved for any fields not supplied by the node kind.
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
