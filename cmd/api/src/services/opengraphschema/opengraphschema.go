// Copyright 2025 Specter Ops, Inc.
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
package opengraphschema

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/opengraphschema.go -package=mocks . OpenGraphSchemaRepository

import (
	"context"
	"errors"
	"fmt"

	// "log/slog"
	"strconv"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// OpenGraphSchemaRepository -
type OpenGraphSchemaRepository interface {
	// Extension
	CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string) (model.GraphSchemaExtension, error)
	GetGraphSchemaExtensionById(ctx context.Context, extensionId int32) (model.GraphSchemaExtension, error)
	GetGraphSchemaExtensions(ctx context.Context, extensionFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error)
	UpdateGraphSchemaExtension(ctx context.Context, extension model.GraphSchemaExtension) (model.GraphSchemaExtension, error)
	DeleteGraphSchemaExtension(ctx context.Context, extensionId int32) error

	// Schema Node Kind
	CreateGraphSchemaNodeKind(ctx context.Context, name string, extensionId int32, displayName string, description string, isDisplayKind bool, icon, iconColor string) (model.GraphSchemaNodeKind, error)
	GetGraphSchemaNodeKindById(ctx context.Context, schemaNodeKindID int32) (model.GraphSchemaNodeKind, error)
	GetGraphSchemaNodeKinds(ctx context.Context, nodeKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaNodeKinds, int, error)
	UpdateGraphSchemaNodeKind(ctx context.Context, schemaNodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error)
	DeleteGraphSchemaNodeKind(ctx context.Context, schemaNodeKindId int32) error

	// Property
	CreateGraphSchemaProperty(ctx context.Context, extensionId int32, name string, displayName string, dataType string, description string) (model.GraphSchemaProperty, error)
	GetGraphSchemaPropertyById(ctx context.Context, extensionPropertyId int32) (model.GraphSchemaProperty, error)
	GetGraphSchemaProperties(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaProperties, int, error)
	UpdateGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error)
	DeleteGraphSchemaProperty(ctx context.Context, propertyID int32) error

	// Edge Kind
	CreateGraphSchemaEdgeKind(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) (model.GraphSchemaEdgeKind, error)
	GetGraphSchemaEdgeKinds(ctx context.Context, edgeKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaEdgeKinds, int, error)
	GetGraphSchemaEdgeKindById(ctx context.Context, schemaEdgeKindId int32) (model.GraphSchemaEdgeKind, error)
	UpdateGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKind model.GraphSchemaEdgeKind) (model.GraphSchemaEdgeKind, error)
	DeleteGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKindId int32) error

	// Environment
	CreateSchemaEnvironment(ctx context.Context, schemaExtensionId int32, environmentKindId int32, sourceKindId int32) (model.SchemaEnvironment, error)
	GetSchemaEnvironmentById(ctx context.Context, schemaEnvironmentId int32) (model.SchemaEnvironment, error)
	DeleteSchemaEnvironment(ctx context.Context, schemaEnvironmentId int32) error

	// Relationship Finding
	// CreateSchemaRelationshipFinding(ctx context.Context, schemaExtensionId int32, relationshipKindId int32, environmentId int32, name string, displayName string) (model.SchemaRelationshipFinding, error)
	// GetSchemaRelationshipFindingById(ctx context.Context, schemaRelationshipFindingId int32) (model.SchemaRelationshipFinding, error)
	// DeleteSchemaRelationshipFinding(ctx context.Context, schemaRelationshipFindingId int32) error

	// Source Kinds
	GetSourceKinds(ctx context.Context) ([]database.SourceKind, error)
}

type OpenGraphSchemaService struct {
	openGraphSchemaRepository OpenGraphSchemaRepository
}

func NewOpenGraphSchemaService(openGraphSchemaRepository OpenGraphSchemaRepository) *OpenGraphSchemaService {
	return &OpenGraphSchemaService{
		openGraphSchemaRepository: openGraphSchemaRepository,
	}
}

// UpsertGraphSchemaExtension -  the incoming graph schema using GenerateMapSynchronizationDiffActions, this function generates the upsert
func (o *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, graphSchema model.GraphSchema) error {
	var (
		err error

		extensions = make(model.GraphSchemaExtensions, 0)
		extension  model.GraphSchemaExtension

		existingNodeKinds  = make(model.GraphSchemaNodeKinds, 0)
		existingProperties = make(model.GraphSchemaProperties, 0)
		existingEdgeKinds  = make(model.GraphSchemaEdgeKinds, 0)
		existingEnvironments  = make(model.GraphSchemaEnvironments, 0)

		nodeKindActions MapDiffActions[model.GraphSchemaNodeKind]
		propertyActions MapDiffActions[model.GraphSchemaProperty]
		edgeKindActions MapDiffActions[model.GraphSchemaEdgeKind]
		environmentActions MapDiffActions[model.SchemaEnvironment]
	)

	// defer func() {
	// 	if err != nil {
	// 		slog.ErrorContext(ctx, "failed to upsert graph schema extension: %v", err)
	// 	} else {
	// 		slog.DebugContext(ctx, "upsert graph schema extension successfully")
	// 	}
	// }()

	if err = validateGraphSchemaModel(graphSchema); err != nil {
		return fmt.Errorf("graph schema validation error: %w", err)
	} else if extensions, _, err = o.openGraphSchemaRepository.GetGraphSchemaExtensions(ctx, model.Filters{"name": []model.Filter{{ // check to see if extension exists
		Operator:    model.Equals,
		Value:       graphSchema.GraphSchemaExtension.Name,
		SetOperator: model.FilterAnd,
	}}}, model.Sort{}, 0, 1); err != nil && !errors.Is(err, database.ErrNotFound) {
		return err
	} else if errors.Is(err, database.ErrNotFound) {
		// extension does not exist so create extension
		if extension, err = o.openGraphSchemaRepository.CreateGraphSchemaExtension(ctx, graphSchema.GraphSchemaExtension.Name,
			graphSchema.GraphSchemaExtension.DisplayName, graphSchema.GraphSchemaExtension.Version); err != nil {
			return err
		}
	} else {
		// extension exists
		extension = extensions[0]
		if extension, err = o.openGraphSchemaRepository.UpdateGraphSchemaExtension(ctx, extension); err != nil {
			return err
		} else if existingNodeKinds, _, err = o.openGraphSchemaRepository.GetGraphSchemaNodeKinds(ctx, model.Filters{"schema_extension_id": []model.Filter{{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(extension.ID), 10),
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 0); err != nil {
			return err
		} else if existingEdgeKinds, _, err = o.openGraphSchemaRepository.GetGraphSchemaEdgeKinds(ctx, model.Filters{"schema_extension_id": []model.Filter{{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(extension.ID), 10),
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 0); err != nil {
			return err
		} else if existingProperties, _, err = o.openGraphSchemaRepository.GetGraphSchemaProperties(ctx, model.Filters{"schema_extension_id": []model.Filter{{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(extension.ID), 10),
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 0); err != nil {
			return err
		}
	}

	// perform map sync generating actions required for nodes, edges and properties
	// TODO: Do we need ItemsToUpsert? OnUpsert performs on pointers to src and dst structs.
	nodeKindActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(), existingNodeKinds.ToMapKeyedOnName(), convertGraphSchemaNodeKinds)
	edgeKindActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(), existingEdgeKinds.ToMapKeyedOnName(), convertGraphSchemaEdgeKinds)
	propertyActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaProperties.ToMapKeyedOnName(), existingProperties.ToMapKeyedOnName(), convertGraphSchemaProperties)
	environmentActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaEnvironment.ToMapKeyedOnEnvironmentAndSource(), existingEnvironments.ToMapKeyedOnEnvironmentAndSource(), convertGraphSchemaEnvironment)

	if err = o.handleNodeKindDiffActions(ctx, extension.ID, nodeKindActions); err != nil {
		return err
	} else if err = o.handleEdgeKindDiffActions(ctx, extension.ID, edgeKindActions); err != nil {
		return err
	} else if err = o.handlePropertyDiffActions(ctx, extension.ID, propertyActions); err != nil {
		return err
	} else if err = o.handleEnvironmentDiffActions(ctx, extension.ID, environmentActions); err != nil {
		return err
	}

	// commit transaction

	return nil
}

func validateGraphSchemaModel(graphSchema model.GraphSchema) error {
	if graphSchema.GraphSchemaExtension.Name == "" {
		return errors.New("graph schema extension name is required")
	} else if graphSchema.GraphSchemaNodeKinds == nil || len(graphSchema.GraphSchemaNodeKinds) == 0 {
		return errors.New("graph schema node kinds is required")
	}
	return nil
}

// TODO: Consolidate the functions below

func (o *OpenGraphSchemaService) handlePropertyDiffActions(ctx context.Context, extensionId int32, actions MapDiffActions[model.GraphSchemaProperty]) error {
	var err error
	if actions.ItemsToDelete != nil && len(actions.ItemsToDelete) > 0 {
		for _, key := range actions.ItemsToDelete {
			if err = o.openGraphSchemaRepository.DeleteGraphSchemaProperty(ctx, key.ID); err != nil {
				return err
			}
		}
	}

	for _, graphSchemaProperty := range actions.ItemsToDelete {
		graphSchemaProperty.SchemaExtensionId = extensionId // new properties need extension id for an existing extension
		if graphSchemaProperty.ID != 0 {                    // An existing property would have had its id transferred to it during onUpsert func within the map merge.
			// property already exists
			if _, err = o.openGraphSchemaRepository.UpdateGraphSchemaProperty(ctx, graphSchemaProperty); err != nil {
				return err
			}
		} else {
			// create new property
			if _, err = o.openGraphSchemaRepository.CreateGraphSchemaProperty(ctx, graphSchemaProperty.SchemaExtensionId,
				graphSchemaProperty.Name, graphSchemaProperty.DisplayName, graphSchemaProperty.DataType,
				graphSchemaProperty.Description); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *OpenGraphSchemaService) handleNodeKindDiffActions(ctx context.Context, extensionId int32, actions MapDiffActions[model.GraphSchemaNodeKind]) error {
	var err error
	if actions.ItemsToDelete != nil && len(actions.ItemsToDelete) > 0 {
		for _, key := range actions.ItemsToDelete {
			if err = o.openGraphSchemaRepository.DeleteGraphSchemaNodeKind(ctx, key.ID); err != nil {
				return err
			}
		}
	}
	for _, newGraphSchemaNodeKind := range actions.ItemsToUpsert {
		newGraphSchemaNodeKind.SchemaExtensionId = extensionId // new node kinds need extension id for an existing extension
		if newGraphSchemaNodeKind.ID != 0 {                    // An existing node kind would have had its id transferred to it during onUpsert func within the map merge.
			if _, err = o.openGraphSchemaRepository.UpdateGraphSchemaNodeKind(ctx, newGraphSchemaNodeKind); err != nil {
				return err
			}
		} else {
			if _, err = o.openGraphSchemaRepository.CreateGraphSchemaNodeKind(ctx, newGraphSchemaNodeKind.Name,
				newGraphSchemaNodeKind.SchemaExtensionId, newGraphSchemaNodeKind.DisplayName, newGraphSchemaNodeKind.Description,
				newGraphSchemaNodeKind.IsDisplayKind, newGraphSchemaNodeKind.Icon, newGraphSchemaNodeKind.IconColor); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *OpenGraphSchemaService) handleEdgeKindDiffActions(ctx context.Context, extensionId int32, actions MapDiffActions[model.GraphSchemaEdgeKind]) error {
	var err error
	// Delete Edge Kinds that are not in incoming schema
	if actions.ItemsToDelete != nil && len(actions.ItemsToDelete) > 0 {
		for _, key := range actions.ItemsToDelete {
			if err = o.openGraphSchemaRepository.DeleteGraphSchemaEdgeKind(ctx, key.ID); err != nil {
				return err
			}
		}
	}

	// Insert or Update incoming edge kinds
	for _, graphSchemaEdgeKind := range actions.ItemsToUpsert {
		graphSchemaEdgeKind.SchemaExtensionId = extensionId // new edge kinds need extension id for an existing extension
		if graphSchemaEdgeKind.ID != 0 {                    // An existing property would have had its id transferred to it during onUpsert func within the map merge.
			if _, err = o.openGraphSchemaRepository.UpdateGraphSchemaEdgeKind(ctx, graphSchemaEdgeKind); err != nil {
				return err
			}
		} else {
			if _, err = o.openGraphSchemaRepository.CreateGraphSchemaEdgeKind(ctx, graphSchemaEdgeKind.Name,
				graphSchemaEdgeKind.SchemaExtensionId, graphSchemaEdgeKind.Description, graphSchemaEdgeKind.IsTraversable); err != nil {
				return err
			}
		}
	}
	return nil
}

// TODO: Consolidate the functions below

// convertGraphSchemaNodeKinds - reassigns model.Serial and SchemaExtensionId data from dst to src if dst is not nil.
func convertGraphSchemaNodeKinds(src, dst *model.GraphSchemaNodeKind) {
	if dst == nil {
		return
	}

	src.Serial = dst.Serial
	src.SchemaExtensionId = dst.SchemaExtensionId
}

func convertGraphSchemaEdgeKinds(src, dst *model.GraphSchemaEdgeKind) {
	if dst == nil {
		return
	}
	src.Serial = dst.Serial
	src.SchemaExtensionId = dst.SchemaExtensionId
}

func convertGraphSchemaProperties(src, dst *model.GraphSchemaProperty) {
	if dst == nil {
		return
	}
	src.Serial = dst.Serial
	src.SchemaExtensionId = dst.SchemaExtensionId
}
