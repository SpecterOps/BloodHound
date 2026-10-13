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
	"time"

	"github.com/gofrs/uuid"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
)

const CustomNodeIconType = "font-awesome"

// UpsertOpenGraphExtension - upserts the incoming graph extension by checking to see if the extension exists already,
// if so, deleting it and inserting the new extension.
//
// Each entity set is diffed against the input: absent rows are deleted, matching rows are updated
// in place with their IDs preserved, and new rows are created. All mutations run inside a single
// transaction that rolls back on any error.
//
// Returns the persisted extension and reconciliation results after a successful commit.
// Returns ErrGraphExtensionBuiltIn if the named extension is a built-in and cannot be modified.
func (s *BloodhoundDB) UpsertOpenGraphExtension(ctx context.Context, graphExtensionInput model.GraphExtensionInput) (model.GraphExtensionUpsertResult, error) {
	var (
		err                         error
		schemaExists                bool
		extension                   model.GraphSchemaExtension
		reconciledNodeKinds         kindReconcileResult[model.GraphSchemaNodeKind]
		reconciledRelationshipKinds kindReconcileResult[model.GraphSchemaRelationshipKind]
		reconciledEnvironments      model.ReconcileResult[model.SchemaEnvironment]
		reconciledFindings          model.ReconcileResult[model.SchemaFinding]
		reconciledSavedQueries      model.ReconcileResult[model.SavedQuery]

		tx                      = s.db.WithContext(ctx).Begin()
		bloodhoundDBTransaction = BloodhoundDB{db: tx, idResolver: s.idResolver}
	)

	defer func() {
		tx.Rollback()
	}()

	if err = tx.Error; err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed creating pgsql transaction: %w", err)
	} else if extension, schemaExists, err = bloodhoundDBTransaction.findOrCreateExtension(ctx, graphExtensionInput.ExtensionInput); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to find or create opengraph extension: %w", err)
	} else if existingNodeKinds, err := bloodhoundDBTransaction.GetGraphSchemaNodeKindsByExtensionId(ctx, extension.ID); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to fetch existing node kinds: %w", err)
	} else if reconciledNodeKinds, err = bloodhoundDBTransaction.reconcileNodeKinds(ctx, extension.ID, graphExtensionInput.NodeKindsInput, existingNodeKinds); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to reconcile node kinds: %w", err)
	} else if err := bloodhoundDBTransaction.upsertCustomIcons(ctx, append(append(model.GraphSchemaNodeKinds{}, reconciledNodeKinds.Kinds.Created...), reconciledNodeKinds.Kinds.Updated...)); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to upsert custom node icons: %w", err)
	} else if existingRelationshipKinds, err := bloodhoundDBTransaction.GetGraphSchemaRelationshipKindsByExtensionId(ctx, extension.ID); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to fetch existing relationship kinds: %w", err)
	} else if reconciledRelationshipKinds, err = bloodhoundDBTransaction.reconcileRelationshipKinds(ctx, extension.ID, graphExtensionInput.RelationshipKindsInput, existingRelationshipKinds); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to reconcile relationship kinds: %w", err)
	} else if existingEnvironments, err := bloodhoundDBTransaction.GetEnvironmentsByExtensionId(ctx, extension.ID); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to fetch existing environments: %w", err)
	} else if reconciledEnvironments, err = reconcile(ctx, graphExtensionInput.EnvironmentsInput, existingEnvironments, bloodhoundDBTransaction.environmentReconcileConfig(extension.ID)); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to reconcile environments: %w", err)
	} else if existingFindings, err := bloodhoundDBTransaction.GetSchemaFindingsByExtensionId(ctx, extension.ID); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to fetch existing findings: %w", err)
	} else if reconciledFindings, err = reconcile(ctx, graphExtensionInput.RelationshipFindingsInput, existingFindings, bloodhoundDBTransaction.findingReconcileConfig(extension.ID)); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to reconcile findings: %w", err)
	} else if existingSavedQueries, err := bloodhoundDBTransaction.GetSavedQueriesByExtensionID(ctx, extension.ID); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to fetch existing saved queries: %w", err)
	} else if reconciledSavedQueries, err = reconcile(ctx, graphExtensionInput.SavedQueriesInput, existingSavedQueries, bloodhoundDBTransaction.savedQueryReconcileConfig(extension.ID)); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to reconcile saved queries: %w", err)
	} else if existingSelectors, err := bloodhoundDBTransaction.GetAssetGroupTagSelectorsByExtensionId(ctx, extension.ID); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to fetch existing asset group tag selectors: %w", err)
	} else if reconciledPZRules, err := reconcile(ctx, graphExtensionInput.PZRulesInput, existingSelectors, bloodhoundDBTransaction.pzRulesReconcileConfig(extension.ID)); err != nil {
		return model.GraphExtensionUpsertResult{}, fmt.Errorf("failed to reconcile PZ rules: %w", err)
	} else if err = tx.Commit().Error; err != nil {
		return model.GraphExtensionUpsertResult{}, err
	} else {
		return model.GraphExtensionUpsertResult{
			ExtensionExisted:           schemaExists,
			Extension:                  extension,
			NodeKindsResult:            reconciledNodeKinds.Kinds,
			RelationshipKindsResult:    reconciledRelationshipKinds.Kinds,
			KindInfosResult:            mergeReconcileResults(reconciledNodeKinds.KindInfo, reconciledRelationshipKinds.KindInfo),
			EnvironmentsResult:         reconciledEnvironments,
			RelationshipFindingsResult: reconciledFindings,
			SavedQueriesResult:         reconciledSavedQueries,
			PZRulesResult:              reconciledPZRules,
		}, nil
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

func (s *BloodhoundDB) pzRulesReconcileConfig(extensionID int32) reconcileConfig[model.PZRuleInput, model.AssetGroupTagSelector, string] {
	return reconcileConfig[model.PZRuleInput, model.AssetGroupTagSelector, string]{
		getInputKey: func(input model.PZRuleInput) string { return input.ExtensionRuleId },
		getExistingKey: func(existing model.AssetGroupTagSelector) string {
			if existing.RuleKey.Valid {
				return existing.RuleKey.String
			} else {
				return ""
			}
		},
		create: func(ctx context.Context, input model.PZRuleInput) (model.AssetGroupTagSelector, error) {
			if selector, err := s.agtSelectorFromPzRule(ctx, extensionID, input); err != nil {
				return model.AssetGroupTagSelector{}, err
			} else if created, err := s.CreateAssetGroupTagSelector(ctx, model.User{PrincipalName: model.AssetGroupActorOpenGraphExtensionManagement}, selector); err != nil {
				return model.AssetGroupTagSelector{}, fmt.Errorf("failed to create extension privilege zone rule %q: %w", input.ExtensionRuleId, err)
			} else {
				return created, nil
			}
		},
		update: func(ctx context.Context, existing model.AssetGroupTagSelector, input model.PZRuleInput) (model.AssetGroupTagSelector, error) {
			if selector, err := s.agtSelectorFromPzRule(ctx, extensionID, input); err != nil {
				return model.AssetGroupTagSelector{}, err
			} else {
				if !input.Enabled && existing.DisabledAt.Valid {
					selector.DisabledAt = existing.DisabledAt
					selector.DisabledBy = existing.DisabledBy
				}
				return s.UpdateOpenGraphAssetGroupTagSelector(ctx, extensionID, selector)
			}
		},
		delete: func(ctx context.Context, existing model.AssetGroupTagSelector) error {
			return s.DeleteAssetGroupTagSelector(ctx, model.User{PrincipalName: model.AssetGroupActorOpenGraphExtensionManagement}, existing)
		},
	}
}

func (s *BloodhoundDB) agtSelectorFromPzRule(ctx context.Context, extensionID int32, input model.PZRuleInput) (model.AssetGroupTagSelector, error) {
	var (
		assetGroupTags model.AssetGroupTags
		selectorSeeds  = make([]model.SelectorSeed, 0, len(input.Seeds))
		selector       model.AssetGroupTagSelector
		err            error
	)

	if assetGroupTags, err = s.GetAssetGroupTags(ctx, model.SQLFilter{
		SQLString: "type = ? AND position = ?",
		Params:    []any{model.AssetGroupTagTypeTier, model.AssetGroupTierZeroPosition},
	}); err != nil {
		return model.AssetGroupTagSelector{}, fmt.Errorf("failed to fetch tier zero asset group tag: %w", err)
	} else if len(assetGroupTags) == 0 {
		return model.AssetGroupTagSelector{}, errors.New("tier zero asset group tag not found")
	}

	selector = model.AssetGroupTagSelector{
		AssetGroupTagId: assetGroupTags[0].ID,
		IsDefault:       true,
		Name:            input.Name,
		Description:     input.Description,
		AutoCertify:     model.SelectorAutoCertifyMethodDisabled,
		AllowDisable:    input.AllowDisable,
		RuleKey:         null.StringFrom(input.ExtensionRuleId),
		ExtensionId:     null.Int32From(extensionID),
	}

	for _, seed := range input.Seeds {
		selectorSeeds = append(selectorSeeds, model.SelectorSeed{Type: seed.Type, Value: seed.Value})
	}
	selector.Seeds = selectorSeeds

	if !input.Enabled {
		selector.DisabledAt = null.TimeFrom(time.Now())
		selector.DisabledBy = null.StringFrom(model.AssetGroupActorOpenGraphExtensionManagement)
	}

	return selector, nil
}

func (s *BloodhoundDB) createExtensionSavedQuery(ctx context.Context, extensionID int32, input model.SavedQueryInput) (model.SavedQuery, error) {
	queryKey := input.QueryKey
	if created, err := s.CreateSavedQuery(ctx, uuid.Nil, input.Name, input.Query, input.Description, &extensionID, &queryKey, input.Category); err != nil {
		return model.SavedQuery{}, fmt.Errorf("failed to create extension saved query %q: %w", input.QueryKey, err)
	} else if _, err := s.CreateSavedQueryPermissionToPublic(ctx, created.ID); err != nil {
		return model.SavedQuery{}, fmt.Errorf("failed to make extension saved query %q public: %w", input.QueryKey, err)
	} else {
		return created, nil
	}
}

func (s *BloodhoundDB) savedQueryReconcileConfig(extensionID int32) reconcileConfig[model.SavedQueryInput, model.SavedQuery, string] {
	return reconcileConfig[model.SavedQueryInput, model.SavedQuery, string]{
		getInputKey: func(input model.SavedQueryInput) string { return input.QueryKey },
		getExistingKey: func(existing model.SavedQuery) string {
			if existing.QueryKey == nil {
				return ""
			}

			return *existing.QueryKey
		},
		create: func(ctx context.Context, input model.SavedQueryInput) (model.SavedQuery, error) {
			return s.createExtensionSavedQuery(ctx, extensionID, input)
		},
		update: func(ctx context.Context, existing model.SavedQuery, input model.SavedQueryInput) (model.SavedQuery, error) {
			existing.Category = input.Category
			existing.Name = input.Name
			existing.Query = input.Query
			existing.Description = input.Description
			return s.UpdateSavedQuery(ctx, existing)
		},
		delete: func(ctx context.Context, existing model.SavedQuery) error {
			return s.DeleteSavedQuery(ctx, existing.ID)
		},
	}
}

// kindReconcileResult contains the persisted kind and nested kind-info outcomes from a reconciliation.
type kindReconcileResult[T any] struct {
	Kinds    model.ReconcileResult[T]
	KindInfo model.ReconcileResult[model.GraphSchemaKindInfo]
}

// reconcileNodeKinds reconciles an extension's node kinds and their nested kind-info records.
func (s *BloodhoundDB) reconcileNodeKinds(ctx context.Context, extensionId int32, inputs model.NodesInput, existingKinds model.GraphSchemaNodeKinds) (kindReconcileResult[model.GraphSchemaNodeKind], error) {
	var (
		result            kindReconcileResult[model.GraphSchemaNodeKind]
		err               error
		existingKindInfos = map[int32][]model.GraphSchemaKindInfo{}
	)

	for _, existingKind := range existingKinds {
		if kindInfos, err := s.GetKindInfos(ctx, existingKind.KindId); err != nil {
			return result, fmt.Errorf("failed to fetch existing node kind info: %w", err)
		} else {
			existingKindInfos[existingKind.KindId] = kindInfos
		}
	}

	config := reconcileConfig[model.NodeInput, model.GraphSchemaNodeKind, string]{
		getInputKey:    func(input model.NodeInput) string { return input.Name },
		getExistingKey: func(existing model.GraphSchemaNodeKind) string { return existing.Name },
		create: func(ctx context.Context, input model.NodeInput) (model.GraphSchemaNodeKind, error) {
			// Create the node kind first
			if createdKind, err := s.CreateGraphSchemaNodeKind(ctx, input.Name, extensionId,
				input.DisplayName, input.Description, input.IsDisplayKind, input.Icon, input.IconColor); err != nil {
				return model.GraphSchemaNodeKind{}, err
			} else if kindInfoResult, err := reconcile(ctx, input.Info, []model.GraphSchemaKindInfo{}, s.kindInfoReconcileConfig(createdKind.KindId, &createdKind.ID, nil)); err != nil {
				return model.GraphSchemaNodeKind{}, fmt.Errorf("failed to create node kind info: %w", err)
			} else {
				result.KindInfo = mergeReconcileResults(result.KindInfo, kindInfoResult)
				return createdKind, nil
			}
		},
		update: func(ctx context.Context, existing model.GraphSchemaNodeKind, input model.NodeInput) (model.GraphSchemaNodeKind, error) {
			existing.DisplayName = input.DisplayName
			existing.Description = input.Description
			existing.IsDisplayKind = input.IsDisplayKind
			existing.Icon = input.Icon
			existing.IconColor = input.IconColor

			// Update the node kind first
			if updatedKind, err := s.UpdateGraphSchemaNodeKind(ctx, existing); err != nil {
				return model.GraphSchemaNodeKind{}, err
			} else {
				// Now reconcile info entries for this node kind
				if kindInfoResult, err := reconcile(ctx, input.Info, existingKindInfos[existing.KindId], s.kindInfoReconcileConfig(updatedKind.KindId, &updatedKind.ID, nil)); err != nil {
					return model.GraphSchemaNodeKind{}, fmt.Errorf("failed to reconcile node kind info: %w", err)
				} else {
					result.KindInfo = mergeReconcileResults(result.KindInfo, kindInfoResult)
				}

				return updatedKind, nil
			}
		},
		delete: func(ctx context.Context, existing model.GraphSchemaNodeKind) error {
			if !existing.IsDisplayKind {
				// Before deleting the schema_node_kind, create a stub in custom_node_kinds for non-display node kinds. Non-display kinds are
				// not tracked in custom_node_kinds while the schema is active. It is possible that a schemaless node kind of the same name
				// existed before this extension upserted it to a non-display node kind, which removes it from the custom node table.
				// Because of this edge case, inserting a stub here for the non-display node kind ensures that node kind will
				// remain tracked even after the schema_node_kind is gone, in the event any nodes of those kinds remain in the graph.
				// Ignore ErrDuplicateCustomNodeKindName. This indicates the kind is already tracked, nothing to do.
				if _, err := s.CreateCustomNodeKinds(ctx, model.CustomNodeKinds{{KindName: existing.Name, Config: CustomNodeKindStubConfig}}); err != nil && !errors.Is(err, ErrDuplicateCustomNodeKindName) {
					return fmt.Errorf("failed to ensure stub for non-display schema node kind %q: %w", existing.Name, err)
				}
			}

			// Deleting from schema_node_kinds automatically nulls the schema_node_kind_id FK
			// in custom_node_kinds via ON DELETE SET NULL.
			if err := s.DeleteGraphSchemaNodeKind(ctx, existing.ID); err != nil {
				return err
			}

			result.KindInfo = mergeReconcileResults(result.KindInfo, model.ReconcileResult[model.GraphSchemaKindInfo]{Deleted: existingKindInfos[existing.KindId]})
			return nil
		},
	}

	result.Kinds, err = reconcile(ctx, inputs, existingKinds, config)
	return result, err
}

// reconcileRelationshipKinds reconciles an extension's relationship kinds and their nested kind-info records.
func (s *BloodhoundDB) reconcileRelationshipKinds(ctx context.Context, extensionID int32, inputs model.RelationshipsInput, existingKinds model.GraphSchemaRelationshipKinds) (kindReconcileResult[model.GraphSchemaRelationshipKind], error) {
	var (
		result            kindReconcileResult[model.GraphSchemaRelationshipKind]
		err               error
		existingKindInfos = map[int32][]model.GraphSchemaKindInfo{}
	)

	for _, existingKind := range existingKinds {
		if kindInfos, err := s.GetKindInfos(ctx, existingKind.KindId); err != nil {
			return result, fmt.Errorf("failed to fetch existing relationship kind info: %w", err)
		} else {
			existingKindInfos[existingKind.KindId] = kindInfos
		}
	}

	config := reconcileConfig[model.RelationshipInput, model.GraphSchemaRelationshipKind, string]{
		getInputKey:    func(input model.RelationshipInput) string { return input.Name },
		getExistingKey: func(existing model.GraphSchemaRelationshipKind) string { return existing.Name },
		create: func(ctx context.Context, input model.RelationshipInput) (model.GraphSchemaRelationshipKind, error) {
			// Create the relationship kind first
			if createdKind, err := s.CreateGraphSchemaRelationshipKind(ctx, input.Name, extensionID,
				input.Description, input.IsTraversable); err != nil {
				return model.GraphSchemaRelationshipKind{}, err
			} else if kindInfoResult, err := reconcile(ctx, input.Info, []model.GraphSchemaKindInfo{}, s.kindInfoReconcileConfig(createdKind.KindId, nil, &createdKind.ID)); err != nil {
				return model.GraphSchemaRelationshipKind{}, fmt.Errorf("failed to create relationship kind info: %w", err)
			} else {
				result.KindInfo = mergeReconcileResults(result.KindInfo, kindInfoResult)
				return createdKind, nil
			}
		},
		update: func(ctx context.Context, existing model.GraphSchemaRelationshipKind, input model.RelationshipInput) (model.GraphSchemaRelationshipKind, error) {
			existing.Description = input.Description
			existing.IsTraversable = input.IsTraversable

			// Update the relationship kind first
			if updatedKind, err := s.UpdateGraphSchemaRelationshipKind(ctx, existing); err != nil {
				return model.GraphSchemaRelationshipKind{}, err
			} else {
				// Now reconcile info entries for this relationship kind
				if kindInfoResult, err := reconcile(ctx, input.Info, existingKindInfos[existing.KindId], s.kindInfoReconcileConfig(updatedKind.KindId, nil, &updatedKind.ID)); err != nil {
					return model.GraphSchemaRelationshipKind{}, fmt.Errorf("failed to reconcile relationship kind info: %w", err)
				} else {
					result.KindInfo = mergeReconcileResults(result.KindInfo, kindInfoResult)
				}

				return updatedKind, nil
			}
		},
		delete: func(ctx context.Context, existing model.GraphSchemaRelationshipKind) error {
			if err := s.DeleteGraphSchemaRelationshipKind(ctx, existing.ID); err != nil {
				return err
			}
			result.KindInfo = mergeReconcileResults(result.KindInfo, model.ReconcileResult[model.GraphSchemaKindInfo]{Deleted: existingKindInfos[existing.KindId]})
			return nil
		},
	}

	result.Kinds, err = reconcile(ctx, inputs, existingKinds, config)
	return result, err
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

// kindInfoReconcileConfig returns the reconcileConfig for kind info entries, keyed by InfoKey.
// This is used for both node kinds and relationship kinds.
func (s *BloodhoundDB) kindInfoReconcileConfig(kindID int32, nodeKindID, relationshipKindID *int32) reconcileConfig[model.KindInfoInput, model.GraphSchemaKindInfo, string] {
	return reconcileConfig[model.KindInfoInput, model.GraphSchemaKindInfo, string]{
		getInputKey:    func(input model.KindInfoInput) string { return input.InfoKey },
		getExistingKey: func(existing model.GraphSchemaKindInfo) string { return existing.InfoKey },
		create: func(ctx context.Context, input model.KindInfoInput) (model.GraphSchemaKindInfo, error) {
			return s.CreateKindInfo(ctx, kindID, nodeKindID, relationshipKindID, input)
		},
		update: func(ctx context.Context, existing model.GraphSchemaKindInfo, input model.KindInfoInput) (model.GraphSchemaKindInfo, error) {
			existing.Title = input.Title
			existing.Position = input.Position
			existing.Content = input.Content
			return s.UpdateKindInfo(ctx, existing)
		},
		delete: func(ctx context.Context, existing model.GraphSchemaKindInfo) error {
			return s.DeleteKindInfo(ctx, existing.ID)
		},
	}
}

// upsertCustomIcons upserts custom icon definitions for the provided node kinds.
func (s *BloodhoundDB) upsertCustomIcons(ctx context.Context, nodeKinds model.GraphSchemaNodeKinds) error {
	var (
		customNodeKindsToCreate     model.CustomNodeKinds
		customNodeKindsToUpdate     model.CustomNodeKinds
		customNodeKindNamesToDelete []string
	)

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
			} else if _, ok := existingIconsMap[nodeKind.Name]; ok {
				// The kind is no longer a display kind, so drop it from custom_node_kinds to keep
				// custom_node_kinds the source of truth for display node kinds
				customNodeKindNamesToDelete = append(customNodeKindNamesToDelete, nodeKind.Name)
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

		if len(customNodeKindNamesToDelete) > 0 {
			for _, kindName := range customNodeKindNamesToDelete {
				if err := s.DeleteCustomNodeKind(ctx, kindName); err != nil {
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
	if existingIcons, err := db.GetCustomNodeKinds(ctx); err != nil {
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
	var customNodeKind = model.CustomNodeKind{
		KindName:         nodeKind.Name,
		SchemaNodeKindId: &nodeKind.ID,
		Config: model.CustomNodeKindConfig{
			Icon: graphschema.DisplayNodeIcon{Type: graphschema.DisplayNodeTypeFontAwesome},
		},
	}

	// fallback to existing icon if provided
	if existingIcon != nil {
		customNodeKind.Config.Icon = existingIcon.Config.Icon
	}

	if nodeKind.Icon != "" {
		customNodeKind.Config.Icon.Name = nodeKind.Icon
	}

	if nodeKind.IconColor != "" {
		customNodeKind.Config.Icon.Color = nodeKind.IconColor
	}

	return customNodeKind

}
