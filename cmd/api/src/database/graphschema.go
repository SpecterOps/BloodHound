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
	"gorm.io/gorm"
)

type OpenGraphSchema interface {
	CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string) (model.GraphSchemaExtension, error)
	GetGraphSchemaExtensionById(ctx context.Context, extensionId int32) (model.GraphSchemaExtension, error)

	CreateGraphSchemaNodeKind(ctx context.Context, name string, extensionId int32, displayName string, description string, isDisplayKind bool, icon, iconColor string) (model.GraphSchemaNodeKind, error)
	GetGraphSchemaNodeKindById(ctx context.Context, schemaNodeKindID int32) (model.GraphSchemaNodeKind, error)
	GetGraphSchemaNodeKinds(ctx context.Context, nodeKindSQLFilter model.SQLFilter, sort model.Sort, skip, limit int) (model.GraphSchemaNodeKinds, int, error)
	UpdateGraphSchemaNodeKind(ctx context.Context, schemaNodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error)
	DeleteGraphSchemaNodeKind(ctx context.Context, schemaNodeKindId int32) error

	CreateGraphSchemaProperty(ctx context.Context, extensionId int32, name string, displayName string, dataType string, description string) (model.GraphSchemaProperty, error)
	GetGraphSchemaPropertyById(ctx context.Context, extensionPropertyId int32) (model.GraphSchemaProperty, error)
	UpdateGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error)
	DeleteGraphSchemaProperty(ctx context.Context, propertyID int32) error

	CreateGraphSchemaEdgeKind(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) (model.GraphSchemaEdgeKind, error)
	GetGraphSchemaEdgeKindById(ctx context.Context, schemaEdgeKindId int32) (model.GraphSchemaEdgeKind, error)
	UpdateGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKind model.GraphSchemaEdgeKind) (model.GraphSchemaEdgeKind, error)
	DeleteGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKindId int32) error
}

// CreateGraphSchemaExtension creates a new row in the extensions table. A GraphSchemaExtension struct is returned, populated with the value as it stands in the database.
func (s *BloodhoundDB) CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string) (model.GraphSchemaExtension, error) {
	var (
		extension = model.GraphSchemaExtension{
			Name:        name,
			DisplayName: displayName,
			Version:     version,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateGraphSchemaExtension,
			Model:  &extension, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		if result := tx.Raw(fmt.Sprintf(`
			INSERT INTO %s (name, display_name, version, is_builtin, created_at, updated_at)
			VALUES (?, ?, ?, FALSE, NOW(), NOW())
			RETURNING id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at`,
			extension.TableName()),
			name, displayName, version).Scan(&extension); result.Error != nil {
			if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
				return fmt.Errorf("%w: %v", ErrDuplicateGraphSchemaExtensionName, result.Error)
			}
			return CheckError(result)
		}
		return nil
	}); err != nil {
		return model.GraphSchemaExtension{}, err
	}

	return extension, nil
}

// GetGraphSchemaExtensionById gets a row from the extensions table by id. It returns a GraphSchemaExtension struct populated with the data, or an error if that id does not exist.
func (s *BloodhoundDB) GetGraphSchemaExtensionById(ctx context.Context, extensionId int32) (model.GraphSchemaExtension, error) {
	var extension model.GraphSchemaExtension

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at
		FROM %s WHERE id = ?`,
		extension.TableName()),
		extensionId).First(&extension); result.Error != nil {
		return model.GraphSchemaExtension{}, CheckError(result)
	}

	return extension, nil
}

// CreateGraphSchemaNodeKind - creates a new row in the schema_node_kinds table. A model.GraphSchemaNodeKind struct is returned, populated with the value as it stands in the database.
func (s *BloodhoundDB) CreateGraphSchemaNodeKind(ctx context.Context, name string, extensionId int32, displayName string, description string, isDisplayKind bool, icon, iconColor string) (model.GraphSchemaNodeKind, error) {
	schemaNodeKind := model.GraphSchemaNodeKind{}

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
			INSERT INTO %s (name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			RETURNING id, name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at`,
		schemaNodeKind.TableName()),
		name, extensionId, displayName, description, isDisplayKind, icon, iconColor).Scan(&schemaNodeKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return model.GraphSchemaNodeKind{}, fmt.Errorf("%w: %v", ErrDuplicateSchemaNodeKindName, result.Error)
		}
		return model.GraphSchemaNodeKind{}, CheckError(result)
	}
	return schemaNodeKind, nil
}

// GetGraphSchemaNodeKindById - gets a row from the schema_node_kinds table by id. It returns a model.GraphSchemaNodeKind struct populated with the data, or an error if that id does not exist.
func (s *BloodhoundDB) GetGraphSchemaNodeKindById(ctx context.Context, schemaNodeKindId int32) (model.GraphSchemaNodeKind, error) {
	var schemaNodeKind model.GraphSchemaNodeKind
	return schemaNodeKind, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at
		FROM %s WHERE id = ?`, schemaNodeKind.TableName()), schemaNodeKindId).First(&schemaNodeKind))
}

// UpdateGraphSchemaNodeKind - updates a row in the schema_node_kinds table based on the provided id. It will return an error if the target schema node kind does not exist or if any of the updates violate the schema constraints.
func (s *BloodhoundDB) UpdateGraphSchemaNodeKind(ctx context.Context, schemaNodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error) {
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		UPDATE %s 
		SET name = ?, schema_extension_id = ?, display_name = ?, description = ?, is_display_kind = ?, icon = ?, icon_color = ?, updated_at = NOW()
    	WHERE id = ?
    	RETURNING id, name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at`,
		schemaNodeKind.TableName()), schemaNodeKind.Name, schemaNodeKind.SchemaExtensionId, schemaNodeKind.DisplayName, schemaNodeKind.Description, schemaNodeKind.IsDisplayKind, schemaNodeKind.Icon, schemaNodeKind.IconColor, schemaNodeKind.ID).Scan(&schemaNodeKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return model.GraphSchemaNodeKind{}, fmt.Errorf("%w: %v", ErrDuplicateSchemaNodeKindName, result.Error)
		}
		return model.GraphSchemaNodeKind{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.GraphSchemaNodeKind{}, ErrNotFound
	}
	return schemaNodeKind, nil
}

// DeleteGraphSchemaNodeKind - deletes a schema_node_kinds row based on the provided id. Will return an error if that id does not exist.
func (s *BloodhoundDB) DeleteGraphSchemaNodeKind(ctx context.Context, schemaNodeKindId int32) error {
	var schemaNodeKind model.GraphSchemaNodeKind

	if result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, schemaNodeKind.TableName()), schemaNodeKindId); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateGraphSchemaProperty creates a new row in the schema_properties table. A GraphSchemaProperty struct is returned, populated with the value as it stands in the database.
func (s *BloodhoundDB) CreateGraphSchemaProperty(ctx context.Context, extensionId int32, name string, displayName string, dataType string, description string) (model.GraphSchemaProperty, error) {
	var extensionProperty model.GraphSchemaProperty

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
			INSERT INTO %s (schema_extension_id, name, display_name, data_type, description)
			VALUES (?, ?, ?, ?, ?)
			RETURNING id, schema_extension_id, name, display_name, data_type, description, created_at, updated_at, deleted_at`,
		extensionProperty.TableName()),
		extensionId, name, displayName, dataType, description).Scan(&extensionProperty); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return model.GraphSchemaProperty{}, fmt.Errorf("%w: %v", ErrDuplicateGraphSchemaExtensionPropertyName, result.Error)
		}
		return model.GraphSchemaProperty{}, CheckError(result)
	}

	return extensionProperty, nil
}

// GetGraphSchemaPropertyById gets a row from the schema_properties table by id. It returns a GraphSchemaProperty struct populated with the data, or an error if that id does not exist.
func (s *BloodhoundDB) GetGraphSchemaPropertyById(ctx context.Context, extensionPropertyId int32) (model.GraphSchemaProperty, error) {
	var extensionProperty model.GraphSchemaProperty

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, schema_extension_id, name, display_name, data_type, description, created_at, updated_at, deleted_at
			FROM %s WHERE id = ?`,
		extensionProperty.TableName()),
		extensionPropertyId).First(&extensionProperty); result.Error != nil {
		return model.GraphSchemaProperty{}, CheckError(result)
	}

	return extensionProperty, nil
}

func (s *BloodhoundDB) UpdateGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error) {
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		UPDATE %s SET name = ?, display_name = ?, data_type = ?, description = ?, updated_at = NOW() WHERE id = ? 
		RETURNING id, schema_extension_id, name, display_name, data_type, description, created_at, updated_at, deleted_at`,
		property.TableName()),
		property.Name, property.DisplayName, property.DataType, property.Description, property.ID).Scan(&property); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return model.GraphSchemaProperty{}, fmt.Errorf("%w: %v", ErrDuplicateGraphSchemaExtensionPropertyName, result.Error)
		}
		return model.GraphSchemaProperty{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.GraphSchemaProperty{}, ErrNotFound
	}

	return property, nil
}

func (s *BloodhoundDB) DeleteGraphSchemaProperty(ctx context.Context, propertyID int32) error {
	var property model.GraphSchemaProperty

	if result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`
		DELETE FROM %s WHERE id = ?`,
		property.TableName()), propertyID); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// CreateGraphSchemaEdgeKind - creates a new row in the schema_edge_kinds table. A model.GraphSchemaEdgeKind struct is returned, populated with the value as it stands in the database.
func (s *BloodhoundDB) CreateGraphSchemaEdgeKind(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) (model.GraphSchemaEdgeKind, error) {
	var schemaEdgeKind model.GraphSchemaEdgeKind

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
	INSERT INTO %s (name, schema_extension_id, description, is_traversable)
    VALUES (?, ?, ?, ?)
    RETURNING id, name, schema_extension_id, description, is_traversable, created_at, updated_at, deleted_at`, schemaEdgeKind.TableName()),
		name, schemaExtensionId, description, isTraversable).Scan(&schemaEdgeKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return schemaEdgeKind, fmt.Errorf("%w: %v", ErrDuplicateSchemaEdgeKindName, result.Error)
		}
		return schemaEdgeKind, CheckError(result)
	}
	return schemaEdgeKind, nil
}

// GetGraphSchemaEdgeKindById - retrieves a row from the schema_edge_kinds table
func (s *BloodhoundDB) GetGraphSchemaEdgeKindById(ctx context.Context, schemaEdgeKindId int32) (model.GraphSchemaEdgeKind, error) {
	var schemaEdgeKind model.GraphSchemaEdgeKind
	return schemaEdgeKind, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf(`
	SELECT id, name, schema_extension_id, description, is_traversable, created_at, updated_at, deleted_at
	FROM %s WHERE id = ?`, schemaEdgeKind.TableName()), schemaEdgeKindId).First(&schemaEdgeKind))
}

// UpdateGraphSchemaEdgeKind - updates a row in the schema_edge_kinds table based on the provided id. It will return an error if the target schema edge kind does not exist or if any of the updates violate the schema constraints.
func (s *BloodhoundDB) UpdateGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKind model.GraphSchemaEdgeKind) (model.GraphSchemaEdgeKind, error) {
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		UPDATE %s
		SET name = ?, schema_extension_id = ?, description = ?, is_traversable = ?, updated_at = NOW()
		WHERE id = ?
		RETURNING id, name, schema_extension_id, description, is_traversable, created_at, updated_at, deleted_at`, schemaEdgeKind.TableName()),
		schemaEdgeKind.Name, schemaEdgeKind.SchemaExtensionId, schemaEdgeKind.Description, schemaEdgeKind.IsTraversable,
		schemaEdgeKind.ID).Scan(&schemaEdgeKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return schemaEdgeKind, fmt.Errorf("%w: %v", ErrDuplicateSchemaEdgeKindName, result.Error)
		}
		return model.GraphSchemaEdgeKind{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.GraphSchemaEdgeKind{}, ErrNotFound
	}
	return schemaEdgeKind, nil
}

// DeleteGraphSchemaEdgeKind - deletes a schema_edge_kind row based on the provided id. It will return an error if that id does not exist.
func (s *BloodhoundDB) DeleteGraphSchemaEdgeKind(ctx context.Context, schemaEdgeKindId int32) error {
	var schemaEdgeKind model.GraphSchemaEdgeKind
	if result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, schemaEdgeKind.TableName()), schemaEdgeKindId); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
