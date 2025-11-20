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
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// CreateSchemaNodeKind - creates a new row in the schema_node_kinds table. A model.SchemaNodeKind struct is returned, populated with the value as it stands in the database.
func (s *BloodhoundDB) CreateSchemaNodeKind(ctx context.Context, name string, extensionID int32, displayName string, description string, isDisplayKind bool, icon, iconColor string) (model.SchemaNodeKind, error) {
	schemaNodeKind := model.SchemaNodeKind{
		Name:              name,
		SchemaExtensionId: extensionID,
		DisplayName:       displayName,
		Description:       description,
		IsDisplayKind:     isDisplayKind,
		Icon:              icon,
		IconColor:         iconColor,
	}

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
			INSERT INTO %s (name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			RETURNING id, name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at`,
		schemaNodeKind.TableName()),
		name, extensionID, displayName, description, isDisplayKind, icon, iconColor).Scan(&schemaNodeKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return model.SchemaNodeKind{}, fmt.Errorf("%w: %v", ErrDuplicateSchemaNodeKindName, result.Error)
		}
		return model.SchemaNodeKind{}, CheckError(result)
	}
	return schemaNodeKind, nil
}

// GetSchemaNodeKindByID - gets a row from the schema_node_kinds table by id. It returns a model.SchemaNodeKind struct populated with the data, or an error if that id does not exist.
func (s *BloodhoundDB) GetSchemaNodeKindByID(ctx context.Context, schemaNodeKindID int32) (model.SchemaNodeKind, error) {
	var schemaNodeKind model.SchemaNodeKind
	return schemaNodeKind, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at
			FROM %s WHERE id = ?`, schemaNodeKind.TableName()), schemaNodeKindID).First(&schemaNodeKind))
}
