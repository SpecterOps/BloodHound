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

// UpsertOpenGraphExtension - compares then upserts the incoming model.GraphSchema with the one stored in
// the BloodHoundDB.
//
// During development, it was decided to push the logic of how extensions are upserted down to the database
// layer due to difficulties decoupling the database and service layers while still providing transactional
// consistency. The following functions use models intended for the service layer and call the database public
// methods directly, rather than using an interface.
func (s *BloodhoundDB) UpsertOpenGraphExtension(ctx context.Context, graphSchema model.GraphSchema) (bool, error) {
	var (
		err          error
		schemaExists bool

		tx                      = s.db.WithContext(ctx).Begin()
		bloodhoundDBTransaction = BloodhoundDB{db: tx}
	)
	// Check for an immediate error after beginning the transaction
	if err = tx.Error; err != nil {
		return schemaExists, err
	}

	defer func() {
		tx.Rollback() // rollback is a no-op if the tx has already been committed, todo: confirm
	}()

	if graphSchema, schemaExists, err = bloodhoundDBTransaction.upsertGraphSchemaExtension(ctx, graphSchema); err != nil {
		return false, err
	}

	if err = tx.Commit().Error; err != nil {
		return false, err
	}

	return schemaExists, nil
}

// upsertGraphSchemaExtension - upserts the model.GraphSchema portion of a model.GraphExtension. TODO: replace with entire extension model.
func (s *BloodhoundDB) upsertGraphSchemaExtension(ctx context.Context, graphSchema model.GraphSchema) (model.GraphSchema, bool, error) {
	var (
		err          error
		schemaExists bool

		existingGraphSchema = model.GraphSchema{
			GraphSchemaNodeKinds:  make(model.GraphSchemaNodeKinds, 0),
			GraphSchemaProperties: make(model.GraphSchemaProperties, 0),
			GraphSchemaEdgeKinds:  make(model.GraphSchemaEdgeKinds, 0),
		}

		nodeKindActions MapDiffActions[model.GraphSchemaNodeKind]
		propertyActions MapDiffActions[model.GraphSchemaProperty]
		edgeKindActions MapDiffActions[model.GraphSchemaEdgeKind]
	)

	if existingGraphSchema, err = s.GetGraphSchemaByExtensionName(ctx, graphSchema.GraphSchemaExtension.Name); err != nil {
		if !errors.Is(err, ErrNotFound) {
			return graphSchema, schemaExists, err
		} else {
			// extension does not exist so create extension
			if graphSchema.GraphSchemaExtension, err = s.CreateGraphSchemaExtension(ctx, graphSchema.GraphSchemaExtension.Name,
				graphSchema.GraphSchemaExtension.DisplayName, graphSchema.GraphSchemaExtension.Version); err != nil {
				return graphSchema, schemaExists, err
			}
		}
	} else {
		// extension exists, transfer model.Serial and update
		schemaExists = true
		if graphSchema.GraphSchemaExtension.IsBuiltin {
			return graphSchema, schemaExists, fmt.Errorf("cannot modify a built-in graph schema extension")
		}
		graphSchema.GraphSchemaExtension.Serial = existingGraphSchema.GraphSchemaExtension.Serial
		if graphSchema.GraphSchemaExtension, err = s.UpdateGraphSchemaExtension(ctx, graphSchema.GraphSchemaExtension); err != nil {
			return graphSchema, schemaExists, err
		}
	}

	// explicitly assign nodes, properties and edges their extension id
	for idx := range graphSchema.GraphSchemaNodeKinds {
		graphSchema.GraphSchemaNodeKinds[idx].SchemaExtensionId = graphSchema.GraphSchemaExtension.ID
	}
	for idx := range graphSchema.GraphSchemaEdgeKinds {
		graphSchema.GraphSchemaEdgeKinds[idx].SchemaExtensionId = graphSchema.GraphSchemaExtension.ID
	}
	for idx := range graphSchema.GraphSchemaProperties {
		graphSchema.GraphSchemaProperties[idx].SchemaExtensionId = graphSchema.GraphSchemaExtension.ID
	}

	// GenerateMapDiffActions compares the incoming graph schema extension (src) with the one stored
	// in the schema database (dst). It generates actions (inserts, updates and deletes) that
	// HandleMapDiffAction apply to the database to upsert the incoming graph schema. These actions
	// are generated for nodes, edges and properties atm.
	nodeKindActions = GenerateMapDiffActions(graphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(),
		existingGraphSchema.GraphSchemaNodeKinds.ToMapKeyedOnName(), convertGraphSchemaNodeKinds)
	edgeKindActions = GenerateMapDiffActions(graphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(),
		existingGraphSchema.GraphSchemaEdgeKinds.ToMapKeyedOnName(), convertGraphSchemaEdgeKinds)
	propertyActions = GenerateMapDiffActions(graphSchema.GraphSchemaProperties.ToMapKeyedOnName(),
		existingGraphSchema.GraphSchemaProperties.ToMapKeyedOnName(), convertGraphSchemaProperties)

	if graphSchema.GraphSchemaNodeKinds, err = HandleMapDiffAction(ctx, nodeKindActions,
		s.deleteGraphSchemaNodeKind, s.updateGraphSchemaNodeKind, s.createGraphSchemaNodeKind); err != nil {
		return graphSchema, schemaExists, err
	} else if graphSchema.GraphSchemaEdgeKinds, err = HandleMapDiffAction(ctx, edgeKindActions,
		s.deleteGraphSchemaEdgeKind, s.updateGraphSchemaEdgeKind, s.createGraphSchemaEdgeKind); err != nil {
		return graphSchema, schemaExists, err
	} else if graphSchema.GraphSchemaProperties, err = HandleMapDiffAction(ctx, propertyActions,
		s.deleteGraphSchemaProperty, s.updateGraphSchemaProperty, s.createGraphSchemaProperty); err != nil {
		return graphSchema, schemaExists, err
	}
	return graphSchema, schemaExists, nil
}

// GetGraphSchemaByExtensionName - returns a graph schema extension with nodes, edges and properties. Will return
// ErrNotFound if the extension does not exist.
func (s *BloodhoundDB) GetGraphSchemaByExtensionName(ctx context.Context, extensionName string) (model.GraphSchema, error) {
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

func (s *BloodhoundDB) updateGraphSchemaNodeKind(ctx context.Context, nodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error) {
	return s.UpdateGraphSchemaNodeKind(ctx, nodeKind)
}

func (s *BloodhoundDB) createGraphSchemaNodeKind(ctx context.Context, nodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error) {
	return s.CreateGraphSchemaNodeKind(ctx, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName,
		nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
}

func (s *BloodhoundDB) deleteGraphSchemaEdgeKind(ctx context.Context, edgeKind model.GraphSchemaEdgeKind) error {
	return s.DeleteGraphSchemaEdgeKind(ctx, edgeKind.ID)
}

func (s *BloodhoundDB) updateGraphSchemaEdgeKind(ctx context.Context, edgeKind model.GraphSchemaEdgeKind) (model.GraphSchemaEdgeKind, error) {
	return s.UpdateGraphSchemaEdgeKind(ctx, edgeKind)
}

func (s *BloodhoundDB) createGraphSchemaEdgeKind(ctx context.Context, edgeKind model.GraphSchemaEdgeKind) (model.GraphSchemaEdgeKind, error) {
	return s.CreateGraphSchemaEdgeKind(ctx, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description,
		edgeKind.IsTraversable)
}

func (s *BloodhoundDB) deleteGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) error {
	return s.DeleteGraphSchemaProperty(ctx, property.ID)
}

func (s *BloodhoundDB) updateGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error) {
	return s.UpdateGraphSchemaProperty(ctx, property)
}

func (s *BloodhoundDB) createGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error) {
	return s.CreateGraphSchemaProperty(ctx, property.SchemaExtensionId, property.Name,
		property.DisplayName, property.DataType, property.Description)
}
