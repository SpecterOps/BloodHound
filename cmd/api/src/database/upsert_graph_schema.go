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

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// UpsertGraphSchemaProperties - upserts model.GraphSchemaProperties and returns the updated schema properties.
func (s *BloodhoundDB) UpsertGraphSchemaProperties(ctx context.Context, newGraphSchemaProperties, existingGraphSchemaProperties model.GraphSchemaProperties) (model.GraphSchemaProperties, error) {
	var (
		err             error
		propertyActions MapDiffActions[model.GraphSchemaProperty]
	)

	// GenerateMapDiffActions compares the incoming graph node kinds (src) with the one stored
	// in the schema database (dst). It generates actions (inserts, updates and deletes) that
	// HandleMapDiffAction apply to the database to upsert the incoming graph schema.
	propertyActions = GenerateMapDiffActions(newGraphSchemaProperties.ToMapKeyedOnName(),
		existingGraphSchemaProperties.ToMapKeyedOnName(), convertGraphSchemaProperties)

	if newGraphSchemaProperties, err = HandleMapDiffAction(ctx, propertyActions,
		s.deleteGraphSchemaProperty, s.UpdateGraphSchemaProperty, s.createGraphSchemaProperty); err != nil {
		return newGraphSchemaProperties, err
	}
	return newGraphSchemaProperties, nil
}

// UpsertGraphSchemaEdgeKinds - upserts model.GraphSchemaEdgeKinds and returns the updated schema edge kinds.
func (s *BloodhoundDB) UpsertGraphSchemaEdgeKinds(ctx context.Context, newGraphSchemaEdgeKinds, existingGraphSchemaEdgeKinds model.GraphSchemaEdgeKinds) (model.GraphSchemaEdgeKinds, error) {
	var (
		err             error
		edgeKindActions MapDiffActions[model.GraphSchemaEdgeKind]
	)

	// GenerateMapDiffActions compares the incoming graph node kinds (src) with the one stored
	// in the schema database (dst). It generates actions (inserts, updates and deletes) that
	// HandleMapDiffAction apply to the database to upsert the incoming graph schema.
	edgeKindActions = GenerateMapDiffActions(newGraphSchemaEdgeKinds.ToMapKeyedOnName(),
		existingGraphSchemaEdgeKinds.ToMapKeyedOnName(), convertGraphSchemaEdgeKinds)
	if newGraphSchemaEdgeKinds, err = HandleMapDiffAction(ctx, edgeKindActions,
		s.deleteGraphSchemaEdgeKind, s.UpdateGraphSchemaEdgeKind, s.createGraphSchemaEdgeKind); err != nil {
		return newGraphSchemaEdgeKinds, err
	}

	return newGraphSchemaEdgeKinds, nil
}

// UpsertGraphSchemaNodeKinds - upserts model.GraphSchemaNodeKinds and returns the updated schema node kinds.
func (s *BloodhoundDB) UpsertGraphSchemaNodeKinds(ctx context.Context, newGraphSchemaNodeKinds, existingGraphSchemaNodeKinds model.GraphSchemaNodeKinds) (model.GraphSchemaNodeKinds, error) {
	var (
		err             error
		nodeKindActions MapDiffActions[model.GraphSchemaNodeKind]
	)

	// GenerateMapDiffActions compares the incoming graph node kinds (src) with the one stored
	// in the schema database (dst). It generates actions (inserts, updates and deletes) that
	// HandleMapDiffAction apply to the database to upsert the incoming graph schema. The
	// new and existing model.GraphSchemaNodeKinds are converted to maps keyed on name as name must be unique.
	nodeKindActions = GenerateMapDiffActions(newGraphSchemaNodeKinds.ToMapKeyedOnName(),
		existingGraphSchemaNodeKinds.ToMapKeyedOnName(), convertGraphSchemaNodeKinds)
	if newGraphSchemaNodeKinds, err = HandleMapDiffAction(ctx, nodeKindActions,
		s.deleteGraphSchemaNodeKind, s.UpdateGraphSchemaNodeKind, s.createGraphSchemaNodeKind); err != nil {
		return newGraphSchemaNodeKinds, err
	}
	return newGraphSchemaNodeKinds, nil
}

// convertGraphSchemaNodeKinds - reassigns model.Serial data from dst to src if neither is nil.
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

func (s *BloodhoundDB) createGraphSchemaNodeKind(ctx context.Context, nodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error) {
	return s.CreateGraphSchemaNodeKind(ctx, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName,
		nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
}

func (s *BloodhoundDB) deleteGraphSchemaEdgeKind(ctx context.Context, edgeKind model.GraphSchemaEdgeKind) error {
	return s.DeleteGraphSchemaEdgeKind(ctx, edgeKind.ID)
}

func (s *BloodhoundDB) createGraphSchemaEdgeKind(ctx context.Context, edgeKind model.GraphSchemaEdgeKind) (model.GraphSchemaEdgeKind, error) {
	return s.CreateGraphSchemaEdgeKind(ctx, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description,
		edgeKind.IsTraversable)
}

func (s *BloodhoundDB) deleteGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) error {
	return s.DeleteGraphSchemaProperty(ctx, property.ID)
}

func (s *BloodhoundDB) createGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error) {
	return s.CreateGraphSchemaProperty(ctx, property.SchemaExtensionId, property.Name,
		property.DisplayName, property.DataType, property.Description)
}
