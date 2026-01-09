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
package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// UpsertGraphSchemaExtension - upserts the provided graph schema. During development, it was decided to push the
// logic of how extensions are upserted down to the database layer due to difficulties decoupling the database and service
// layers while still providing transactional consistency. As such this and corresponding functions use models intended
// for the service layer and call the database public methods directly.
func (s *BloodhoundDB) UpsertGraphSchemaExtension(ctx context.Context, graphSchema model.GraphSchema) (bool, error) {
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

		tx                      = s.db.WithContext(ctx).Begin()
		bloodhoundDBTransaction = BloodhoundDB{db: tx}
	)

	defer func() {
		tx.Rollback() // rollback is a no-op if the tx has already been committed
	}()

	if existingGraphSchema, err = bloodhoundDBTransaction.getGraphSchemaByExtensionName(ctx, graphSchema.GraphSchemaExtension.Name); err != nil {
		if !errors.Is(err, ErrNotFound) {
			return schemaExists, err
		} else {
			// extension does not exist so create extension
			if extension, err = bloodhoundDBTransaction.CreateGraphSchemaExtension(ctx, graphSchema.GraphSchemaExtension.Name,
				graphSchema.GraphSchemaExtension.DisplayName, graphSchema.GraphSchemaExtension.Version); err != nil {
				return schemaExists, err
			}
		}
	} else {
		// extension exists, transfer model.Serial and update
		extension = graphSchema.GraphSchemaExtension
		schemaExists = true
		if extension.IsBuiltin {
			return schemaExists, fmt.Errorf("cannot modify a built-in graph schema extension")
		}
		extension.Serial = existingGraphSchema.GraphSchemaExtension.Serial
		if extension, err = bloodhoundDBTransaction.UpdateGraphSchemaExtension(ctx, extension); err != nil {
			return schemaExists, err
		}
	}

	// explicitly assign nodes, properties and edges to their extension id
	for _, nodeKind := range graphSchema.GraphSchemaNodeKinds {
		nodeKind.SchemaExtensionId = extension.ID
	}
	for _, edgeKind := range graphSchema.GraphSchemaEdgeKinds {
		edgeKind.SchemaExtensionId = extension.ID
	}
	for _, property := range graphSchema.GraphSchemaProperties {
		property.SchemaExtensionId = extension.ID
	}

	// GenerateMapSynchronizationDiffActions is used to compare the incoming graph schema extension with the one stored
	// in the schema database. It generates the actions needed to reconcile the database with
	// the incoming schema. These actions can be applied to nodes, edges and properties atm.
	nodeKindActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(),
		existingGraphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(), convertGraphSchemaNodeKinds)
	edgeKindActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(),
		existingGraphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(), convertGraphSchemaEdgeKinds)
	propertyActions = GenerateMapSynchronizationDiffActions(graphSchema.GraphSchemaProperties.ToMapKeyedOnName(),
		existingGraphSchema.GraphSchemaProperties.ToMapKeyedOnName(), convertGraphSchemaProperties)

	if err = HandleMapDiffAction(ctx, nodeKindActions, s.deleteGraphSchemaNodeKind, s.updateGraphSchemaNodeKind, s.createGraphSchemaNodeKind); err != nil {
		return schemaExists, err
	} else if err = HandleMapDiffAction(ctx, edgeKindActions, s.deleteGraphSchemaEdgeKind, s.deleteGraphSchemaEdgeKind, s.createGraphSchemaEdgeKind); err != nil {
		return schemaExists, err
	} else if err = HandleMapDiffAction(ctx, propertyActions, s.deleteGraphSchemaProperty, s.deleteGraphSchemaProperty, s.createGraphSchemaProperty); err != nil {
		return schemaExists, err
	}

	tx.Commit()

	return schemaExists, nil
}

// getGraphSchemaByExtensionName - returns a graph schema extension with nodes, edges and properties. Will return ErrNotFound
// if the extension does not exist.
func (s *BloodhoundDB) getGraphSchemaByExtensionName(ctx context.Context, extensionName string) (model.GraphSchema, error) {
	var graphSchema = model.GraphSchema{
		GraphSchemaProperties: make(model.GraphSchemaProperties, 0),
		GraphSchemaEdgeKinds:  make(model.GraphSchemaEdgeKinds, 0),
		GraphSchemaNodeKinds:  make(model.GraphSchemaNodeKinds, 0),
	}

	if extensions, totalRecords, err := s.GetGraphSchemaExtensions(ctx,
		model.Filters{"name": []model.Filter{{ // check to see if extension exists
			Operator:    model.Equals,
			Value:       extensionName,
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 1); err != nil && !errors.Is(err, ErrNotFound) {
		return model.GraphSchema{}, err
	} else if totalRecords == 0 || errors.Is(err, ErrNotFound) {
		return model.GraphSchema{}, ErrNotFound
	} else {
		graphSchema.GraphSchemaExtension = extensions[0]
		if graphSchema.GraphSchemaNodeKinds, _, err = s.GetGraphSchemaNodeKinds(ctx,
			model.Filters{"schema_extension_id": []model.Filter{{
				Operator:    model.Equals,
				Value:       strconv.FormatInt(int64(graphSchema.GraphSchemaExtension.ID), 10),
				SetOperator: model.FilterAnd,
			}}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, ErrNotFound) {
			return model.GraphSchema{}, err
		} else if graphSchema.GraphSchemaEdgeKinds, _, err = s.GetGraphSchemaEdgeKinds(ctx,
			model.Filters{"schema_extension_id": []model.Filter{{
				Operator:    model.Equals,
				Value:       strconv.FormatInt(int64(graphSchema.GraphSchemaExtension.ID), 10),
				SetOperator: model.FilterAnd,
			}}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, ErrNotFound) {
			return model.GraphSchema{}, err
		} else if graphSchema.GraphSchemaProperties, _, err = s.GetGraphSchemaProperties(ctx,
			model.Filters{"schema_extension_id": []model.Filter{{
				Operator:    model.Equals,
				Value:       strconv.FormatInt(int64(graphSchema.GraphSchemaExtension.ID), 10),
				SetOperator: model.FilterAnd,
			}}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, ErrNotFound) {
			return model.GraphSchema{}, err
		}
		return graphSchema, nil
	}
}

// TODO: Consolidate the functions below
// convertGraphSchemaNodeKinds - reassigns model.Serial and SchemaExtensionId data from dst to src if neither is nil.
func convertGraphSchemaNodeKinds(src, dst *model.GraphSchemaNodeKind) {
	if dst == nil || src == nil {
		return
	}
	src.Serial = dst.Serial
}

func convertGraphSchemaEdgeKinds(src, dst *model.GraphSchemaEdgeKind) {
	if dst == nil || src == nil {
		return
	}
	src.Serial = dst.Serial
}

func convertGraphSchemaProperties(src, dst *model.GraphSchemaProperty) {
	if dst == nil || src == nil {
		return
	}
	src.Serial = dst.Serial
}

func (s *BloodhoundDB) deleteGraphSchemaNodeKind(ctx context.Context, nodeKind model.GraphSchemaNodeKind) error {
	return s.DeleteGraphSchemaNodeKind(ctx, nodeKind.ID)
}

func (s *BloodhoundDB) updateGraphSchemaNodeKind(ctx context.Context, nodeKind model.GraphSchemaNodeKind) error {
	_, err := s.UpdateGraphSchemaNodeKind(ctx, nodeKind)
	return err
}

func (s *BloodhoundDB) createGraphSchemaNodeKind(ctx context.Context, nodeKind model.GraphSchemaNodeKind) error {
	_, err := s.CreateGraphSchemaNodeKind(ctx, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName,
		nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
	return err
}

func (s *BloodhoundDB) deleteGraphSchemaEdgeKind(ctx context.Context, edgeKind model.GraphSchemaEdgeKind) error {
	return s.DeleteGraphSchemaEdgeKind(ctx, edgeKind.ID)
}

func (s *BloodhoundDB) updateGraphSchemaEdgeKind(ctx context.Context, edgeKind model.GraphSchemaEdgeKind) error {
	_, err := s.UpdateGraphSchemaEdgeKind(ctx, edgeKind)
	return err
}

func (s *BloodhoundDB) createGraphSchemaEdgeKind(ctx context.Context, edgeKind model.GraphSchemaEdgeKind) error {
	_, err := s.CreateGraphSchemaEdgeKind(ctx, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description,
		edgeKind.IsTraversable)
	return err
}

func (s *BloodhoundDB) deleteGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) error {
	return s.DeleteGraphSchemaProperty(ctx, property.ID)
}

func (s *BloodhoundDB) updateGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) error {
	_, err := s.UpdateGraphSchemaProperty(ctx, property)
	return err
}

func (s *BloodhoundDB) createGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) error {
	_, err := s.CreateGraphSchemaProperty(ctx, property.SchemaExtensionId, property.Name,
		property.DisplayName, property.DataType, property.Description)
	return err
}
