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

const DuplicateKeyValueErrorString = "duplicate key value violates unique constraint"

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
			if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
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

// GetGraphSchemaExtensions gets all the rows from the extensions table that match the given SQLFilter. It returns a slice of GraphSchemaExtension structs
// populated with the data, as well as an integer giving the total number of rows returned by the query (excluding any given pagination)
func (s *BloodhoundDB) GetGraphSchemaExtensions(ctx context.Context, extensionFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error) {
	var (
		extensions        = model.GraphSchemaExtensions{}
		skipLimitString   string
		whereClauseString string
		totalRowCount     int
		orderSQL          string
	)

	filter, err := buildSQLFilter(extensionFilters)
	if err != nil {
		return extensions, 0, err
	}

	// if no sort specified, default to ID so pagination is consistent
	if len(sort) == 0 {
		sort = append(sort, model.SortItem{Column: "id", Direction: model.AscendingSortDirection})
	}
	orderSQL, err = buildSQLSort(sort)
	if err != nil {
		return extensions, 0, err
	}

	if limit > 0 {
		skipLimitString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		skipLimitString += fmt.Sprintf(" OFFSET %d", skip)
	}

	if filter.sqlString != "" {
		whereClauseString = fmt.Sprintf("WHERE %s", filter.sqlString)
	}

	sqlStr := fmt.Sprintf(`SELECT id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at
								FROM %s %s %s %s`,
		model.GraphSchemaExtension{}.TableName(),
		whereClauseString,
		orderSQL,
		skipLimitString)

	if result := s.db.WithContext(ctx).Raw(sqlStr, filter.params...).Scan(&extensions); result.Error != nil {
		return model.GraphSchemaExtensions{}, 0, CheckError(result)
	} else {
		// we need an overall count of the rows if pagination is supplied
		if limit > 0 || skip > 0 {
			countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`,
				model.GraphSchemaExtension{}.TableName(),
				whereClauseString)

			if err := s.db.WithContext(ctx).Raw(countSqlStr, filter.params...).Scan(&totalRowCount).Error; err != nil {
				return model.GraphSchemaExtensions{}, 0, err
			}
		} else {
			totalRowCount = len(extensions)
		}
	}

	return extensions, totalRowCount, nil
}

// UpdateGraphSchemaExtension updates an existing Graph Schema Extension. Only the `name`, `display_name`, and `version` fields are updatable. It returns the updated extension, or an error if the update violates schema constraints or did not succeed.
func (s *BloodhoundDB) UpdateGraphSchemaExtension(ctx context.Context, extension model.GraphSchemaExtension) (model.GraphSchemaExtension, error) {
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		UPDATE %s 
		SET name = ?, display_name = ?, version = ?, updated_at = NOW()
		WHERE id = ? 
		RETURNING id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at`,
		extension.TableName()), extension.Name, extension.DisplayName, extension.Version, extension.ID).Scan(&extension); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return extension, fmt.Errorf("%w: %v", ErrDuplicateGraphSchemaExtensionName, result.Error)
		}
		return extension, CheckError(result)
	} else if result.RowsAffected == 0 {
		return extension, ErrNotFound
	}
	return extension, nil
}

// DeleteGraphSchemaExtension deletes an existing Graph Schema Extension based on the extension ID. It returns an error if the extension does not exist.
func (s *BloodhoundDB) DeleteGraphSchemaExtension(ctx context.Context, extensionId int32) error {
	var schemaExtension model.GraphSchemaExtension
	if result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, schemaExtension.TableName()), extensionId); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
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
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return model.GraphSchemaNodeKind{}, fmt.Errorf("%w: %v", ErrDuplicateSchemaNodeKindName, result.Error)
		}
		return model.GraphSchemaNodeKind{}, CheckError(result)
	}
	return schemaNodeKind, nil
}

// GetGraphSchemaNodeKinds - returns all rows from the schema_node_kinds table that matches the given model.Filters. It returns a slice of model.GraphSchemaNodeKinds structs
// populated with data, as well as an integer indicating the total number of rows returned by the query (excluding any given pagination).
func (s *BloodhoundDB) GetGraphSchemaNodeKinds(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaNodeKinds, int, error) {
	var (
		schemaNodeKinds   = model.GraphSchemaNodeKinds{}
		skipLimitString   string
		whereClauseString string
		totalRowCount     int
		orderSQL          string
	)

	filter, err := buildSQLFilter(filters)
	if err != nil {
		return schemaNodeKinds, 0, err
	}

	// if no sort specified, default to ID so pagination is consistent
	if len(sort) == 0 {
		sort = append(sort, model.SortItem{Column: "id", Direction: model.AscendingSortDirection})
	}
	orderSQL, err = buildSQLSort(sort)
	if err != nil {
		return schemaNodeKinds, 0, err
	}

	if limit > 0 {
		skipLimitString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		skipLimitString += fmt.Sprintf(" OFFSET %d", skip)
	}

	if filter.sqlString != "" {
		whereClauseString = fmt.Sprintf("WHERE %s", filter.sqlString)
	}

	sqlStr := fmt.Sprintf(`SELECT id, name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at
									FROM %s %s %s %s`,
		model.GraphSchemaNodeKind{}.TableName(),
		whereClauseString,
		orderSQL,
		skipLimitString)

	if result := s.db.WithContext(ctx).Raw(sqlStr, filter.params...).Scan(&schemaNodeKinds); result.Error != nil {
		return nil, 0, CheckError(result)
	} else {
		if limit > 0 || skip > 0 {
			countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, model.GraphSchemaNodeKind{}.TableName(), whereClauseString)
			if countResult := s.db.WithContext(ctx).Raw(countSqlStr, filter.params...).Scan(&totalRowCount); countResult.Error != nil {
				return model.GraphSchemaNodeKinds{}, 0, CheckError(countResult)
			}
		} else {
			totalRowCount = len(schemaNodeKinds)
		}
	}
	return schemaNodeKinds, totalRowCount, nil
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
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
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
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return model.GraphSchemaProperty{}, fmt.Errorf("%w: %v", ErrDuplicateGraphSchemaExtensionPropertyName, result.Error)
		}
		return model.GraphSchemaProperty{}, CheckError(result)
	}

	return extensionProperty, nil
}

// GetGraphSchemaProperties - returns all rows from the schema_properties table that matches the given model.Filters. It returns a slice of model.GraphSchemaProperties structs
// // populated with data, as well as an integer indicating the total number of rows returned by the query (excluding any given pagination).
func (s *BloodhoundDB) GetGraphSchemaProperties(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaProperties, int, error) {
	var (
		schemaProperties  = model.GraphSchemaProperties{}
		skipLimitString   string
		whereClauseString string
		totalRowCount     int
		orderSQL          string
	)

	filter, err := buildSQLFilter(filters)
	if err != nil {
		return schemaProperties, 0, err
	}

	// if no sort specified, default to ID so pagination is consistent
	if len(sort) == 0 {
		sort = append(sort, model.SortItem{Column: "id", Direction: model.AscendingSortDirection})
	}
	orderSQL, err = buildSQLSort(sort)
	if err != nil {
		return schemaProperties, 0, err
	}

	if limit > 0 {
		skipLimitString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		skipLimitString += fmt.Sprintf(" OFFSET %d", skip)
	}

	if filter.sqlString != "" {
		whereClauseString = fmt.Sprintf("WHERE %s", filter.sqlString)
	}

	sqlStr := fmt.Sprintf(`SELECT id, schema_extension_id, name, display_name, data_type, description, created_at, updated_at, deleted_at
									FROM %s %s %s %s`,
		model.GraphSchemaProperty{}.TableName(),
		whereClauseString,
		orderSQL,
		skipLimitString)

	if result := s.db.WithContext(ctx).Raw(sqlStr, filter.params...).Scan(&schemaProperties); result.Error != nil {
		return nil, 0, CheckError(result)
	} else {
		if limit > 0 || skip > 0 {
			countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, model.GraphSchemaProperty{}.TableName(), whereClauseString)
			if countResult := s.db.WithContext(ctx).Raw(countSqlStr, filter.params...).Scan(&totalRowCount); countResult.Error != nil {
				return model.GraphSchemaProperties{}, 0, CheckError(countResult)
			}
		} else {
			totalRowCount = len(schemaProperties)
		}
	}
	return schemaProperties, totalRowCount, nil
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
		UPDATE %s SET name = ?, schema_extension_id = ?, display_name = ?, data_type = ?, description = ?, updated_at = NOW() WHERE id = ? 
		RETURNING id, schema_extension_id, name, display_name, data_type, description, created_at, updated_at, deleted_at`,
		property.TableName()),
		property.Name, property.SchemaExtensionID, property.DisplayName, property.DataType, property.Description, property.ID).Scan(&property); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
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
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return schemaEdgeKind, fmt.Errorf("%w: %v", ErrDuplicateSchemaEdgeKindName, result.Error)
		}
		return schemaEdgeKind, CheckError(result)
	}
	return schemaEdgeKind, nil
}

func (s *BloodhoundDB) GetGraphSchemaEdgeKinds(ctx context.Context, edgeKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaEdgeKinds, int, error) {
	var (
		schemaEdgeKinds   = model.GraphSchemaEdgeKinds{}
		skipLimitString   string
		whereClauseString string
		totalRowCount     int
		orderSQL          string
	)

	filter, err := buildSQLFilter(edgeKindFilters)
	if err != nil {
		return schemaEdgeKinds, 0, err
	}

	// if no sort specified, default to ID so pagination is consistent
	if len(sort) == 0 {
		sort = append(sort, model.SortItem{Column: "id", Direction: model.AscendingSortDirection})
	}
	orderSQL, err = buildSQLSort(sort)
	if err != nil {
		return schemaEdgeKinds, 0, err
	}

	if limit > 0 {
		skipLimitString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		skipLimitString += fmt.Sprintf(" OFFSET %d", skip)
	}

	if filter.sqlString != "" {
		whereClauseString = fmt.Sprintf("WHERE %s", filter.sqlString)
	}

	sqlStr := fmt.Sprintf(`SELECT id, name, schema_extension_id, description, is_traversable, created_at, updated_at, deleted_at
									FROM %s %s %s %s`,
		model.GraphSchemaEdgeKind{}.TableName(),
		whereClauseString,
		orderSQL,
		skipLimitString)

	if result := s.db.WithContext(ctx).Raw(sqlStr, filter.params...).Scan(&schemaEdgeKinds); result.Error != nil {
		return nil, 0, CheckError(result)
	} else {
		if limit > 0 || skip > 0 {
			countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, model.GraphSchemaEdgeKind{}.TableName(), whereClauseString)
			if countResult := s.db.WithContext(ctx).Raw(countSqlStr, filter.params...).Scan(&totalRowCount); countResult.Error != nil {
				return model.GraphSchemaEdgeKinds{}, 0, CheckError(countResult)
			}
		} else {
			totalRowCount = len(schemaEdgeKinds)
		}
	}
	return schemaEdgeKinds, totalRowCount, nil
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
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
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
