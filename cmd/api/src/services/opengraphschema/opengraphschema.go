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

// Mocks

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemaextension.go -package=mocks . OpenGraphSchemaExtensionRepository
//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemanodekindrepository.go -package=mocks . OpenGraphSchemaNodeKindRepository
//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemaedgekindrepository.go -package=mocks . OpenGraphSchemaEdgeKindRepository
//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemapropertyrepository.go -package=mocks . OpenGraphSchemaPropertyRepository

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphdbkindrepository.go -package=mocks . GraphDBKindRepository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
)

// OpenGraphSchemaExtensionRepository -
type OpenGraphSchemaExtensionRepository interface {
	CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string) (model.GraphSchemaExtension, error)
	GetGraphSchemaExtensionById(ctx context.Context, extensionId int32) (model.GraphSchemaExtension, error)
	GetGraphSchemaExtensions(ctx context.Context, extensionFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error)
	UpdateGraphSchemaExtension(ctx context.Context, extension model.GraphSchemaExtension) (model.GraphSchemaExtension, error)
	DeleteGraphSchemaExtension(ctx context.Context, extensionId int32) error
}

// OpenGraphSchemaNodeKindRepository -
type OpenGraphSchemaNodeKindRepository interface {
	CreateGraphSchemaNodeKind(ctx context.Context, name string, extensionId int32, displayName string, description string, isDisplayKind bool, icon, iconColor string) (model.GraphSchemaNodeKind, error)
	GetGraphSchemaNodeKindById(ctx context.Context, schemaNodeKindID int32) (model.GraphSchemaNodeKind, error)
	GetGraphSchemaNodeKinds(ctx context.Context, nodeKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaNodeKinds, int, error)
	UpdateGraphSchemaNodeKind(ctx context.Context, schemaNodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error)
	DeleteGraphSchemaNodeKind(ctx context.Context, schemaNodeKindId int32) error
}

// OpenGraphSchemaEdgeKindRepository -
type OpenGraphSchemaEdgeKindRepository interface {
	CreateGraphSchemaEdgeKind(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) (model.GraphSchemaEdgeKind, error)
	GetGraphSchemaEdgeKinds(ctx context.Context, edgeKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaEdgeKinds, int, error)
	GetGraphSchemaEdgeKindById(ctx context.Context, schemaEdgeKindId int32) (model.GraphSchemaEdgeKind, error)
	UpdateGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKind model.GraphSchemaEdgeKind) (model.GraphSchemaEdgeKind, error)
	DeleteGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKindId int32) error
}

// OpenGraphSchemaPropertyRepository -
type OpenGraphSchemaPropertyRepository interface {
	CreateGraphSchemaProperty(ctx context.Context, extensionId int32, name string, displayName string, dataType string, description string) (model.GraphSchemaProperty, error)
	GetGraphSchemaPropertyById(ctx context.Context, extensionPropertyId int32) (model.GraphSchemaProperty, error)
	GetGraphSchemaProperties(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaProperties, int, error)
	UpdateGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error)
	DeleteGraphSchemaProperty(ctx context.Context, propertyID int32) error
}

// GraphDBKindRepository -
type GraphDBKindRepository interface {
	// RefreshKinds refreshes the in memory kinds maps
	RefreshKinds(ctx context.Context) error
}

// OpenGraphSchemaService -
type OpenGraphSchemaService struct {
	openGraphSchemaExtensionRepository OpenGraphSchemaExtensionRepository
	openGraphSchemaNodeRepository      OpenGraphSchemaNodeKindRepository
	openGraphSchemaEdgeRepository      OpenGraphSchemaEdgeKindRepository
	openGraphSchemaPropertyRepository  OpenGraphSchemaPropertyRepository
	graphDBKindRepository              GraphDBKindRepository
}

func NewOpenGraphSchemaService(openGraphSchemaExtensionRepository OpenGraphSchemaExtensionRepository,
	openGraphSchemaNodeRepository OpenGraphSchemaNodeKindRepository, openGraphSchemaEdgeRepository OpenGraphSchemaEdgeKindRepository,
	openGraphSchemaPropertyRepository OpenGraphSchemaPropertyRepository, graphDBKindRepository GraphDBKindRepository) *OpenGraphSchemaService {
	return &OpenGraphSchemaService{
		openGraphSchemaExtensionRepository: openGraphSchemaExtensionRepository,
		openGraphSchemaNodeRepository:      openGraphSchemaNodeRepository,
		openGraphSchemaEdgeRepository:      openGraphSchemaEdgeRepository,
		openGraphSchemaPropertyRepository:  openGraphSchemaPropertyRepository,
		graphDBKindRepository:              graphDBKindRepository,
	}
}

// UpsertGraphSchemaExtension - upserts the provided graph schema.
func (o *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, graphSchema model.GraphSchema) (bool, error) {
	var (
		err          error
		schemaExists bool

		extension model.GraphSchemaExtension

		existingGraphSchema = model.GraphSchema{
			GraphSchemaExtension:  extension,
			GraphSchemaNodeKinds:  make(model.GraphSchemaNodeKinds, 0),
			GraphSchemaProperties: make(model.GraphSchemaProperties, 0),
			GraphSchemaEdgeKinds:  make(model.GraphSchemaEdgeKinds, 0),
		}

		nodeKindActions MapDiffActions[model.GraphSchemaNodeKind]
		propertyActions MapDiffActions[model.GraphSchemaProperty]
		edgeKindActions MapDiffActions[model.GraphSchemaEdgeKind]
	)

	if err = validateGraphSchemaModel(graphSchema); err != nil {
		return schemaExists, fmt.Errorf("graph schema validation error: %w", err)
	} else if existingGraphSchema, err = o.getGraphSchemaByExtensionName(ctx, graphSchema.GraphSchemaExtension.Name); err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return schemaExists, err
		} else if errors.Is(err, database.ErrNotFound) {
			// extension does not exist so create extension
			if extension, err = o.openGraphSchemaExtensionRepository.CreateGraphSchemaExtension(ctx, graphSchema.GraphSchemaExtension.Name,
				graphSchema.GraphSchemaExtension.DisplayName, graphSchema.GraphSchemaExtension.Version); err != nil {
				return schemaExists, err
			}
		}
	} else {
		// extension exists, transfer model.Serial and update
		extension = graphSchema.GraphSchemaExtension
		schemaExists = true
		if extension.IsBuiltin {
			// TODO: Need rollback
			return schemaExists, fmt.Errorf("cannot modify a built-in graph schema extension")
		}
		extension.Serial = existingGraphSchema.GraphSchemaExtension.Serial
		if extension, err = o.openGraphSchemaExtensionRepository.UpdateGraphSchemaExtension(ctx, extension); err != nil {
			return schemaExists, err
		}
	}

	// perform map sync generating actions required for nodes, edges and properties
	nodeKindActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(), existingGraphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(), convertGraphSchemaNodeKinds)
	edgeKindActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(), existingGraphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(), convertGraphSchemaEdgeKinds)
	propertyActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaProperties.ToMapKeyedOnName(), existingGraphSchema.GraphSchemaProperties.ToMapKeyedOnName(), convertGraphSchemaProperties)

	if err = o.handleNodeKindDiffActions(ctx, extension.ID, nodeKindActions); err != nil {
		return schemaExists, err
	} else if err = o.handleEdgeKindDiffActions(ctx, extension.ID, edgeKindActions); err != nil {
		return schemaExists, err
	} else if err = o.handlePropertyDiffActions(ctx, extension.ID, propertyActions); err != nil {
		return schemaExists, err
	}

	// commit transaction

	// TODO: what to do, insert has already been committed however if the refresh fails then the new kinds wont be usable
	//  (this is what PMZ does with asset group tags)
	if err = o.graphDBKindRepository.RefreshKinds(ctx); err != nil {
		slog.WarnContext(ctx, "OpenGraphSchema: refreshing graph kind maps failed", attr.Error(err))
	}

	return schemaExists, nil
}

// getGraphSchemaByExtensionName - returns a graph schema extension with nodes, edges and properties. Will return database.ErrNotFound
// if the extension does not exist.
func (o *OpenGraphSchemaService) getGraphSchemaByExtensionName(ctx context.Context, extensionName string) (model.GraphSchema, error) {
	var graphSchema = model.GraphSchema{
		GraphSchemaProperties: make(model.GraphSchemaProperties, 0),
		GraphSchemaEdgeKinds:  make(model.GraphSchemaEdgeKinds, 0),
		GraphSchemaNodeKinds:  make(model.GraphSchemaNodeKinds, 0),
	}

	if extensions, totalRecords, err := o.openGraphSchemaExtensionRepository.GetGraphSchemaExtensions(ctx,
		model.Filters{"name": []model.Filter{{ // check to see if extension exists
			Operator:    model.Equals,
			Value:       extensionName,
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 1); err != nil && !errors.Is(err, database.ErrNotFound) {
		return model.GraphSchema{}, err
	} else if totalRecords == 0 || errors.Is(err, database.ErrNotFound) {
		return model.GraphSchema{}, database.ErrNotFound
	} else {
		graphSchema.GraphSchemaExtension = extensions[0]
		if graphSchema.GraphSchemaNodeKinds, _, err = o.openGraphSchemaNodeRepository.GetGraphSchemaNodeKinds(ctx,
			model.Filters{"schema_extension_id": []model.Filter{{
				Operator:    model.Equals,
				Value:       strconv.FormatInt(int64(graphSchema.GraphSchemaExtension.ID), 10),
				SetOperator: model.FilterAnd,
			}}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, database.ErrNotFound) {
			return model.GraphSchema{}, err
		} else if graphSchema.GraphSchemaEdgeKinds, _, err = o.openGraphSchemaEdgeRepository.GetGraphSchemaEdgeKinds(ctx,
			model.Filters{"schema_extension_id": []model.Filter{{
				Operator:    model.Equals,
				Value:       strconv.FormatInt(int64(graphSchema.GraphSchemaExtension.ID), 10),
				SetOperator: model.FilterAnd,
			}}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, database.ErrNotFound) {
			return model.GraphSchema{}, err
		} else if graphSchema.GraphSchemaProperties, _, err = o.openGraphSchemaPropertyRepository.GetGraphSchemaProperties(ctx,
			model.Filters{"schema_extension_id": []model.Filter{{
				Operator:    model.Equals,
				Value:       strconv.FormatInt(int64(graphSchema.GraphSchemaExtension.ID), 10),
				SetOperator: model.FilterAnd,
			}}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, database.ErrNotFound) {
			return model.GraphSchema{}, err
		}
		return graphSchema, nil
	}
}

func validateGraphSchemaModel(graphSchema model.GraphSchema) error {
	if graphSchema.GraphSchemaExtension.Name == "" {
		return errors.New("graph schema extension name is required")
	} else if len(graphSchema.GraphSchemaNodeKinds) == 0 {
		return errors.New("graph schema node kinds is required")
	}
	return nil
}

// TODO: Consolidate the functions below

func (o *OpenGraphSchemaService) handlePropertyDiffActions(ctx context.Context, extensionId int32, actions MapDiffActions[model.GraphSchemaProperty]) error {
	var err error
	if len(actions.ItemsToDelete) > 0 {
		for _, deletedGraphSchemaProperty := range actions.ItemsToDelete {
			if err = o.openGraphSchemaPropertyRepository.DeleteGraphSchemaProperty(ctx, deletedGraphSchemaProperty.ID); err != nil {
				return err
			}
		}
	}
	if len(actions.ItemsToUpdate) > 0 {
		for _, updatedGraphSchemaProperty := range actions.ItemsToUpdate {
			updatedGraphSchemaProperty.SchemaExtensionId = extensionId // new properties need extension id for an existing extension
			if _, err = o.openGraphSchemaPropertyRepository.UpdateGraphSchemaProperty(ctx, updatedGraphSchemaProperty); err != nil {
				return err
			}
		}
	}
	if len(actions.ItemsToInsert) > 0 {
		for _, newGraphSchemaProperty := range actions.ItemsToInsert {
			newGraphSchemaProperty.SchemaExtensionId = extensionId // new properties need extension id for an existing extension
			if _, err = o.openGraphSchemaPropertyRepository.CreateGraphSchemaProperty(ctx, newGraphSchemaProperty.SchemaExtensionId,
				newGraphSchemaProperty.Name, newGraphSchemaProperty.DisplayName, newGraphSchemaProperty.DataType,
				newGraphSchemaProperty.Description); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *OpenGraphSchemaService) handleNodeKindDiffActions(ctx context.Context, extensionId int32, actions MapDiffActions[model.GraphSchemaNodeKind]) error {
	var err error
	if len(actions.ItemsToDelete) > 0 {
		for _, deletedGraphSchemaNodeKind := range actions.ItemsToDelete {
			if err = o.openGraphSchemaNodeRepository.DeleteGraphSchemaNodeKind(ctx, deletedGraphSchemaNodeKind.ID); err != nil {
				return err
			}
		}
	}
	if len(actions.ItemsToUpdate) > 0 {
		for _, updatedGraphSchemaNodeKind := range actions.ItemsToUpdate {
			updatedGraphSchemaNodeKind.SchemaExtensionId = extensionId
			if _, err = o.openGraphSchemaNodeRepository.UpdateGraphSchemaNodeKind(ctx, updatedGraphSchemaNodeKind); err != nil {
				return err
			}

		}
	}
	if len(actions.ItemsToInsert) > 0 {
		for _, newGraphSchemaNodeKind := range actions.ItemsToInsert {
			newGraphSchemaNodeKind.SchemaExtensionId = extensionId // new node kinds need extension id
			if _, err = o.openGraphSchemaNodeRepository.CreateGraphSchemaNodeKind(ctx, newGraphSchemaNodeKind.Name,
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
	if len(actions.ItemsToDelete) > 0 {
		for _, deletedGraphSchemaKind := range actions.ItemsToDelete {
			if err = o.openGraphSchemaEdgeRepository.DeleteGraphSchemaEdgeKind(ctx, deletedGraphSchemaKind.ID); err != nil {
				return err
			}
		}
	}

	if len(actions.ItemsToUpdate) > 0 {
		for _, updatedGraphSchemaKind := range actions.ItemsToUpdate {
			updatedGraphSchemaKind.SchemaExtensionId = extensionId
			if _, err = o.openGraphSchemaEdgeRepository.UpdateGraphSchemaEdgeKind(ctx, updatedGraphSchemaKind); err != nil {
				return err
			}
		}
	}
	if len(actions.ItemsToInsert) > 0 {
		for _, newGraphSchemaEdgeKind := range actions.ItemsToInsert {
			newGraphSchemaEdgeKind.SchemaExtensionId = extensionId
			if _, err = o.openGraphSchemaEdgeRepository.CreateGraphSchemaEdgeKind(ctx, newGraphSchemaEdgeKind.Name,
				newGraphSchemaEdgeKind.SchemaExtensionId, newGraphSchemaEdgeKind.Description, newGraphSchemaEdgeKind.IsTraversable); err != nil {
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
