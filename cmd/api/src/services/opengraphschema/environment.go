package opengraphschema

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (o *OpenGraphSchemaService) UpsertSchemaEnvironment(ctx context.Context, graphSchema model.SchemaEnvironment) error {
	if err := graphSchema.Validate(); err != nil {
		return fmt.Errorf("error validating schema environment: %w", err)
	} else if environment, _, err = o.openGraphSchemaRepository.GetSchemaEnvironmentByName(ctx, model.Filters{"name": []model.Filter{{ // check to see if extension exists
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

	// perform map sync generating actions required for nodes, edges and properties
	// TODO: Do we need ItemsToUpsert? OnUpsert performs on pointers to src and dst structs.
	nodeKindActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(), existingNodeKinds.ToMapKeyedOnName(), convertGraphSchemaNodeKinds)
	edgeKindActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(), existingEdgeKinds.ToMapKeyedOnName(), convertGraphSchemaEdgeKinds)
	propertyActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaProperties.ToMapKeyedOnName(), existingProperties.ToMapKeyedOnName(), convertGraphSchemaProperties)

	if err = o.handleNodeKindDiffActions(ctx, extension.ID, nodeKindActions); err != nil {
		return err
	} else if err = o.handleEdgeKindDiffActions(ctx, extension.ID, edgeKindActions); err != nil {
		return err
	} else if err = o.handlePropertyDiffActions(ctx, extension.ID, propertyActions); err != nil {
		return err
	}

	// commit transaction

	return nil
}

func (o *OpenGraphSchemaService) GetSchemaEnvironmentByName(ctx context.Context) (model.SchemaEnvironment, error) {
	return nil
}

