package opengraphschema

import (
	"context"
	"errors"
	"strconv"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type OpenGraphSchemaRepository interface {
	CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string) (model.GraphSchemaExtension, error)
	GetGraphSchemaExtensionById(ctx context.Context, extensionId int32) (model.GraphSchemaExtension, error)
	GetGraphSchemaExtensions(ctx context.Context, extensionFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error)
	UpdateGraphSchemaExtension(ctx context.Context, extension model.GraphSchemaExtension) (model.GraphSchemaExtension, error)
	DeleteGraphSchemaExtension(ctx context.Context, extensionId int32) error

	CreateGraphSchemaNodeKind(ctx context.Context, name string, extensionId int32, displayName string, description string, isDisplayKind bool, icon, iconColor string) (model.GraphSchemaNodeKind, error)
	GetGraphSchemaNodeKindById(ctx context.Context, schemaNodeKindID int32) (model.GraphSchemaNodeKind, error)
	GetGraphSchemaNodeKinds(ctx context.Context, nodeKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaNodeKinds, int, error)
	UpdateGraphSchemaNodeKind(ctx context.Context, schemaNodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error)
	DeleteGraphSchemaNodeKind(ctx context.Context, schemaNodeKindId int32) error

	CreateGraphSchemaProperty(ctx context.Context, extensionId int32, name string, displayName string, dataType string, description string) (model.GraphSchemaProperty, error)
	GetGraphSchemaPropertyById(ctx context.Context, extensionPropertyId int32) (model.GraphSchemaProperty, error)
	GetGraphSchemaProperties(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaProperties, int, error)
	UpdateGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error)
	DeleteGraphSchemaProperty(ctx context.Context, propertyID int32) error

	CreateGraphSchemaEdgeKind(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) (model.GraphSchemaEdgeKind, error)
	GetGraphSchemaEdgeKinds(ctx context.Context, edgeKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaEdgeKinds, int, error)
	GetGraphSchemaEdgeKindById(ctx context.Context, schemaEdgeKindId int32) (model.GraphSchemaEdgeKind, error)
	UpdateGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKind model.GraphSchemaEdgeKind) (model.GraphSchemaEdgeKind, error)
	DeleteGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKindId int32) error
}

type OpenGraphSchemaService struct {
	openGraphSchemaRepository OpenGraphSchemaRepository
}

func NewOpenGraphSchemaService(openGraphSchemaRepository OpenGraphSchemaRepository) *OpenGraphSchemaService {
	return &OpenGraphSchemaService{
		openGraphSchemaRepository: openGraphSchemaRepository,
	}
}

// UpsertGraphSchemaExtension -
func (o *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, graphSchema model.GraphSchema) error {
	var (
		err          error
		newExtension bool

		extensions = make(model.GraphSchemaExtensions, 0)
		extension  model.GraphSchemaExtension

		existingNodeKinds  = make(model.GraphSchemaNodeKinds, 0)
		existingProperties = make(model.GraphSchemaProperties, 0)
		existingEdgeKinds  = make(model.GraphSchemaEdgeKinds, 0)

		nodeKindActions MapSyncActions[model.GraphSchemaNodeKind]
		propertyActions MapSyncActions[model.GraphSchemaProperty]
		edgeKindActions MapSyncActions[model.GraphSchemaEdgeKind]
	)

	// Validate GraphSchema:
	// - ensure schema has a valid extension and that the node and edge slices are not empty
	// begin transaction

	// defer if err -> rollback

	// Check if extension exists
	if extensions, _, err = o.openGraphSchemaRepository.GetGraphSchemaExtensions(ctx, model.Filters{"name": []model.Filter{{
		Operator:    model.Equals,
		Value:       graphSchema.GraphSchemaExtension.Name,
		SetOperator: model.FilterAnd,
	}}}, model.Sort{}, 0, 1); err != nil && !errors.Is(err, database.ErrNotFound) {
		return err
	} else if len(extensions) == 0 {
		// extension does not exist
		newExtension = true
	} else {
		extension = extensions[0]
	}

	if newExtension {
		// extension does not exist so create extension
		extension, err = o.openGraphSchemaRepository.CreateGraphSchemaExtension(ctx, graphSchema.GraphSchemaExtension.Name,
			graphSchema.GraphSchemaExtension.DisplayName, graphSchema.GraphSchemaExtension.Version)
		if err != nil {
			return err
		}

	} else {
		existingNodeKinds, _, err = o.openGraphSchemaRepository.GetGraphSchemaNodeKinds(ctx, model.Filters{"schema_extension_id": []model.Filter{{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(extension.ID), 10),
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 0)
		if err != nil {
			return err
		}
		existingEdgeKinds, _, err = o.openGraphSchemaRepository.GetGraphSchemaEdgeKinds(ctx, model.Filters{"schema_extension_id": []model.Filter{{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(extension.ID), 10),
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 0)
		if err != nil {
			return err
		}
		existingProperties, _, err = o.openGraphSchemaRepository.GetGraphSchemaProperties(ctx, model.Filters{"schema_extension_id": []model.Filter{{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(extension.ID), 10),
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 0)
		if err != nil {
			return err
		}
	}

	// perform upsert merge
	// TODO: Do we need ValuesToUpsert? OnUpsert performs actions on pointers meaning the graphSchema structs are updated as well.
	nodeKindActions = DiffMapsToSyncActions(existingNodeKinds.ToMapKeyedOnName(), graphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(), convertGraphSchemaNodeKinds)
	edgeKindActions = DiffMapsToSyncActions(existingEdgeKinds.ToMapKeyedOnName(), graphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(), convertGraphSchemaEdgeKinds)
	propertyActions = DiffMapsToSyncActions(existingProperties.ToMapKeyedOnName(), graphSchema.GraphSchemaProperties.ToMapKeyedOnName(), convertGraphSchemaProperties)

	if nodeKindActions.ValuesToDelete != nil && len(nodeKindActions.ValuesToDelete) > 0 {
		for _, key := range nodeKindActions.ValuesToDelete {
			if err = o.openGraphSchemaRepository.DeleteGraphSchemaNodeKind(ctx, key.ID); err != nil {
				return err
			}
		}
	}
	if edgeKindActions.ValuesToDelete != nil && len(edgeKindActions.ValuesToDelete) > 0 {
		for _, key := range edgeKindActions.ValuesToDelete {
			if err = o.openGraphSchemaRepository.DeleteGraphSchemaEdgeKind(ctx, key.ID); err != nil {
				return err
			}
		}
	}
	if propertyActions.ValuesToDelete != nil && len(propertyActions.ValuesToDelete) > 0 {
		for _, key := range propertyActions.ValuesToDelete {
			if err = o.openGraphSchemaRepository.DeleteGraphSchemaProperty(ctx, key.ID); err != nil {
				return err
			}
		}
	}

	for _, newGraphSchemaNodeKind := range nodeKindActions.ValuesToUpsert {
		newGraphSchemaNodeKind.SchemaExtensionId = extension.ID // new node kinds need extension id for an existing extension
		if newGraphSchemaNodeKind.ID != 0 {
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
	for _, graphSchemaEdgeKind := range edgeKindActions.ValuesToUpsert {
		graphSchemaEdgeKind.SchemaExtensionId = extension.ID // new edge kinds need extension id for an existing extension
		if graphSchemaEdgeKind.ID != 0 {
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
	for _, graphSchemaProperty := range propertyActions.ValuesToUpsert {
		graphSchemaProperty.SchemaExtensionId = extension.ID // new properties need extension id for an existing extension
		if graphSchemaProperty.ID != 0 {
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

	// commit transaction

	return nil
}

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
