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

	"gorm.io/gorm"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type OpenGraphSchema interface {
	CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string, namespace string) (model.GraphSchemaExtension, error)
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

	CreateGraphSchemaRelationshipKind(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) (model.GraphSchemaRelationshipKind, error)
	GetGraphSchemaRelationshipKinds(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaRelationshipKinds, int, error)
	GetGraphSchemaRelationshipKindById(ctx context.Context, schemaRelationshipKindId int32) (model.GraphSchemaRelationshipKind, error)
	UpdateGraphSchemaRelationshipKind(ctx context.Context, schemaRelationshipKind model.GraphSchemaRelationshipKind) (model.GraphSchemaRelationshipKind, error)
	DeleteGraphSchemaRelationshipKind(ctx context.Context, schemaRelationshipKindId int32) error

	GetGraphSchemaRelationshipKindsWithSchemaName(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaRelationshipKindsWithNamedSchema, int, error)

	CreateEnvironment(ctx context.Context, extensionId int32, environmentKindId int32, sourceKindId int32) (model.SchemaEnvironment, error)
	GetEnvironmentByKinds(ctx context.Context, environmentKindId, sourceKindId int32) (model.SchemaEnvironment, error)
	GetEnvironmentById(ctx context.Context, environmentId int32) (model.SchemaEnvironment, error)
	GetEnvironments(ctx context.Context) ([]model.SchemaEnvironment, error)
	DeleteEnvironment(ctx context.Context, environmentId int32) error

	CreateSchemaRelationshipFinding(ctx context.Context, extensionId int32, relationshipKindId int32, environmentId int32, name string, displayName string) (model.SchemaRelationshipFinding, error)
	GetSchemaRelationshipFindingById(ctx context.Context, findingId int32) (model.SchemaRelationshipFinding, error)
	GetSchemaRelationshipFindingByName(ctx context.Context, name string) (model.SchemaRelationshipFinding, error)
	DeleteSchemaRelationshipFinding(ctx context.Context, findingId int32) error

	CreateRemediation(ctx context.Context, findingId int32, shortDescription string, longDescription string, shortRemediation string, longRemediation string) (model.Remediation, error)
	GetRemediationByFindingId(ctx context.Context, findingId int32) (model.Remediation, error)
	GetRemediationByFindingName(ctx context.Context, findingName string) (model.Remediation, error)
	UpdateRemediation(ctx context.Context, findingId int32, shortDescription string, longDescription string, shortRemediation string, longRemediation string) (model.Remediation, error)
	DeleteRemediation(ctx context.Context, findingId int32) error

	CreatePrincipalKind(ctx context.Context, environmentId int32, principalKind int32) (model.SchemaEnvironmentPrincipalKind, error)
	GetPrincipalKindsByEnvironmentId(ctx context.Context, environmentId int32) (model.SchemaEnvironmentPrincipalKinds, error)
	DeletePrincipalKind(ctx context.Context, environmentId int32, principalKind int32) error
}

const (
	DuplicateKeyValueErrorString = "duplicate key value violates unique constraint"
)

type FilterAndPagination struct {
	Filter      sqlFilter
	SkipLimit   string
	WhereClause string
	OrderSql    string
}

// CreateGraphSchemaExtension creates a new row in the extensions table. A GraphSchemaExtension struct is returned, populated with the value as it stands in the database.
func (s *BloodhoundDB) CreateGraphSchemaExtension(ctx context.Context, name string, displayName string, version string, namespace string) (model.GraphSchemaExtension, error) {
	var (
		extension = model.GraphSchemaExtension{
			Name:        name,
			DisplayName: displayName,
			Version:     version,
			Namespace:   namespace,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateGraphSchemaExtension,
			Model:  &extension, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		if result := tx.Raw(fmt.Sprintf(`
			INSERT INTO %s (name, display_name, version, is_builtin, namespace, created_at, updated_at)
			VALUES (?, ?, ?, FALSE, ?, NOW(), NOW())
			RETURNING id, name, display_name, version, is_builtin, namespace, created_at, updated_at, deleted_at`,
			extension.TableName()),
			name, displayName, version, namespace).Scan(&extension); result.Error != nil {
			if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
				if strings.Contains(result.Error.Error(), "namespace") {
					return fmt.Errorf("%w: %v", model.ErrDuplicateGraphSchemaExtensionNamespace, namespace)
				} else {
					return fmt.Errorf("%w: %v", model.ErrDuplicateGraphSchemaExtensionName, name)
				}
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
		SELECT id, name, display_name, version, is_builtin, namespace, created_at, updated_at, deleted_at
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
		extensions    = model.GraphSchemaExtensions{}
		totalRowCount int
	)

	if filterAndPagination, err := parseFiltersAndPagination(extensionFilters, sort, skip, limit); err != nil {
		return extensions, 0, err
	} else {

		sqlStr := fmt.Sprintf(`SELECT id, name, display_name, version, is_builtin, namespace, created_at, updated_at, deleted_at
								FROM %s %s %s %s`,
			model.GraphSchemaExtension{}.TableName(),
			filterAndPagination.WhereClause,
			filterAndPagination.OrderSql,
			filterAndPagination.SkipLimit)

		if result := s.db.WithContext(ctx).Raw(sqlStr, filterAndPagination.Filter.params...).Scan(&extensions); result.Error != nil {
			return model.GraphSchemaExtensions{}, 0, CheckError(result)
		} else {
			// we need an overall count of the rows if pagination is supplied
			if limit > 0 || skip > 0 {
				countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`,
					model.GraphSchemaExtension{}.TableName(),
					filterAndPagination.WhereClause)

				if err := s.db.WithContext(ctx).Raw(countSqlStr, filterAndPagination.Filter.params...).Scan(&totalRowCount).Error; err != nil {
					return model.GraphSchemaExtensions{}, 0, err
				}
			} else {
				totalRowCount = len(extensions)
			}
		}

		return extensions, totalRowCount, nil
	}
}

// UpdateGraphSchemaExtension updates an existing Graph Schema Extension. Only the `name`, `display_name`, and `version` fields are updatable. It returns the updated extension, or an error if the update violates schema constraints or did not succeed.
func (s *BloodhoundDB) UpdateGraphSchemaExtension(ctx context.Context, extension model.GraphSchemaExtension) (model.GraphSchemaExtension, error) {
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		UPDATE %s
		SET name = ?, display_name = ?, version = ?, namespace = ?, updated_at = NOW()
		WHERE id = ?
		RETURNING id, name, display_name, version, is_builtin, namespace, created_at, updated_at, deleted_at`,
		extension.TableName()), extension.Name, extension.DisplayName, extension.Version, extension.Namespace, extension.ID).Scan(&extension); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			if strings.Contains(result.Error.Error(), "namespace") {
				return model.GraphSchemaExtension{}, fmt.Errorf("%w: %v", model.ErrDuplicateGraphSchemaExtensionNamespace, extension.Namespace)
			} else {
				return model.GraphSchemaExtension{}, fmt.Errorf("%w: %v", model.ErrDuplicateGraphSchemaExtensionName, extension.Name)
			}
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

// CreateGraphSchemaNodeKind - creates a new row in the schema_node_kinds table. A model.GraphSchemaNodeKind struct is
// returned, populated with the value as it stands in the database. This will also create a kind in the DAWGS kind table
// if the kind does not already exist.
//
// Since this inserts directly into the kinds table, the business logic calling this func
// must also call the DAWGS RefreshKinds function to ensure the kinds are reloaded into the in memory kind map.
func (s *BloodhoundDB) CreateGraphSchemaNodeKind(ctx context.Context, name string, extensionId int32, displayName string, description string, isDisplayKind bool, icon, iconColor string) (model.GraphSchemaNodeKind, error) {
	var schemaNodeKind = model.GraphSchemaNodeKind{}
	if result := s.db.WithContext(ctx).Raw(`
	WITH dawgs_kind (id, name) AS ( SELECT id, name FROM upsert_kind(?)),
    inserted_schema_node AS (
		INSERT INTO schema_node_kinds (kind_id, schema_extension_id, display_name, description, is_display_kind, icon, icon_color)
		SELECT dk.id, ?, ?, ?, ?, ?, ?
		FROM dawgs_kind dk
		RETURNING id, kind_id, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at
	)
	SELECT isn.id, isn.schema_extension_id, dk.name, isn.display_name, isn.description, isn.is_display_kind, isn.icon, isn.icon_color, isn.created_at, isn.updated_at, isn.deleted_at
	FROM inserted_schema_node isn
	JOIN dawgs_kind dk ON isn.kind_id = dk.id;`, name, extensionId, displayName, description,
		isDisplayKind, icon, iconColor).Scan(&schemaNodeKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return model.GraphSchemaNodeKind{}, fmt.Errorf("%w: %s", model.ErrDuplicateSchemaNodeKindName, name)
		}
		return model.GraphSchemaNodeKind{}, CheckError(result)
	}
	return schemaNodeKind, nil
}

// GetGraphSchemaNodeKinds - returns all rows from the schema_node_kinds table that matches the given model.Filters. It returns a slice of model.GraphSchemaNodeKinds structs
// populated with data, as well as an integer indicating the total number of rows returned by the query (excluding any given pagination).
func (s *BloodhoundDB) GetGraphSchemaNodeKinds(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaNodeKinds, int, error) {
	var (
		schemaNodeKinds = model.GraphSchemaNodeKinds{}
		totalRowCount   int
	)

	if filterAndPagination, err := parseFiltersAndPagination(filters, sort, skip, limit); err != nil {
		return schemaNodeKinds, 0, err
	} else {
		sqlStr := fmt.Sprintf(`SELECT nk.id, k.name, nk.schema_extension_id, nk.display_name, nk.description,
									nk.is_display_kind, nk.icon, nk.icon_color, nk.created_at, nk.updated_at, nk.deleted_at
									FROM %s nk
									JOIN %s k ON nk.kind_id = k.id
									%s %s %s`,
			model.GraphSchemaNodeKind{}.TableName(), kindTable, filterAndPagination.WhereClause, filterAndPagination.OrderSql, filterAndPagination.SkipLimit)
		if result := s.db.WithContext(ctx).Raw(sqlStr, filterAndPagination.Filter.params...).Scan(&schemaNodeKinds); result.Error != nil {
			return nil, 0, CheckError(result)
		} else {
			if limit > 0 || skip > 0 {
				countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s nk JOIN %s k ON nk.kind_id = k.id %s`,
					model.GraphSchemaNodeKind{}.TableName(), kindTable, filterAndPagination.WhereClause)
				if countResult := s.db.WithContext(ctx).Raw(countSqlStr, filterAndPagination.Filter.params...).Scan(&totalRowCount); countResult.Error != nil {
					return model.GraphSchemaNodeKinds{}, 0, CheckError(countResult)
				}
			} else {
				totalRowCount = len(schemaNodeKinds)
			}
		}
		return schemaNodeKinds, totalRowCount, nil
	}
}

// GetGraphSchemaNodeKindById - gets a row from the schema_node_kinds table by id. It returns a model.GraphSchemaNodeKind struct populated with the data, or an error if that id does not exist.
func (s *BloodhoundDB) GetGraphSchemaNodeKindById(ctx context.Context, schemaNodeKindId int32) (model.GraphSchemaNodeKind, error) {
	var schemaNodeKind model.GraphSchemaNodeKind
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT %s.id, name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at
		FROM %s JOIN %s ON %s.kind_id = %s.id WHERE %s.id = ?`, schemaNodeKind.TableName(), schemaNodeKind.TableName(), kindTable,
		schemaNodeKind.TableName(), kindTable, schemaNodeKind.TableName()), schemaNodeKindId).First(&schemaNodeKind); result.Error != nil {
		return model.GraphSchemaNodeKind{}, CheckError(result)
	}
	return schemaNodeKind, nil
}

// UpdateGraphSchemaNodeKind - updates a row in the schema_node_kinds table based on the provided id. It will return an
// error if the target schema node kind does not exist or if any of the updates violate the schema constraints.
//
// This function does NOT update the DAWGS name column since the schema_node_kinds table FKs to the DAWGS kind table, and that
// table is append only. A new node kind should be created instead.
func (s *BloodhoundDB) UpdateGraphSchemaNodeKind(ctx context.Context, schemaNodeKind model.GraphSchemaNodeKind) (model.GraphSchemaNodeKind, error) {
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		WITH updated_row AS (
			UPDATE %s
			SET schema_extension_id = ?, display_name = ?, description = ?, is_display_kind = ?, icon = ?, icon_color = ?, updated_at = NOW()
			WHERE id = ?
			RETURNING id, kind_id, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at
		)
		SELECT updated_row.id, %s.name, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at
		FROM updated_row
		JOIN %s ON %s.id = updated_row.kind_id`,
		schemaNodeKind.TableName(), kindTable, kindTable, kindTable), schemaNodeKind.SchemaExtensionId,
		schemaNodeKind.DisplayName, schemaNodeKind.Description, schemaNodeKind.IsDisplayKind, schemaNodeKind.Icon,
		schemaNodeKind.IconColor, schemaNodeKind.ID).Scan(&schemaNodeKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return model.GraphSchemaNodeKind{}, fmt.Errorf("%w: %s", model.ErrDuplicateSchemaNodeKindName, schemaNodeKind.Name)
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
			return model.GraphSchemaProperty{}, fmt.Errorf("%w: %s", model.ErrDuplicateGraphSchemaExtensionPropertyName, name)
		}
		return model.GraphSchemaProperty{}, CheckError(result)
	}

	return extensionProperty, nil
}

// GetGraphSchemaProperties - returns all rows from the schema_properties table that matches the given model.Filters. It returns a slice of model.GraphSchemaProperties structs
// populated with data, as well as an integer indicating the total number of rows returned by the query (excluding any given pagination).
func (s *BloodhoundDB) GetGraphSchemaProperties(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaProperties, int, error) {
	var (
		schemaProperties = model.GraphSchemaProperties{}
		totalRowCount    int
	)

	if filterAndPagination, err := parseFiltersAndPagination(filters, sort, skip, limit); err != nil {
		return schemaProperties, 0, err
	} else {
		sqlStr := fmt.Sprintf(`SELECT id, schema_extension_id, name, display_name, data_type, description, created_at, updated_at, deleted_at
									FROM %s %s %s %s`,
			model.GraphSchemaProperty{}.TableName(),
			filterAndPagination.WhereClause,
			filterAndPagination.OrderSql,
			filterAndPagination.SkipLimit)

		if result := s.db.WithContext(ctx).Raw(sqlStr, filterAndPagination.Filter.params...).Scan(&schemaProperties); result.Error != nil {
			return nil, 0, CheckError(result)
		} else {
			if limit > 0 || skip > 0 {
				countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, model.GraphSchemaProperty{}.TableName(), filterAndPagination.WhereClause)
				if countResult := s.db.WithContext(ctx).Raw(countSqlStr, filterAndPagination.Filter.params...).Scan(&totalRowCount); countResult.Error != nil {
					return model.GraphSchemaProperties{}, 0, CheckError(countResult)
				}
			} else {
				totalRowCount = len(schemaProperties)
			}
		}
		return schemaProperties, totalRowCount, nil
	}

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

// UpdateGraphSchemaProperty - updates a row in the schema_properties table based on the provided id. It will return an
// error if the target property does not exist or if any of the updates violate the schema constraints.
func (s *BloodhoundDB) UpdateGraphSchemaProperty(ctx context.Context, property model.GraphSchemaProperty) (model.GraphSchemaProperty, error) {
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		UPDATE %s SET name = ?, schema_extension_id = ?, display_name = ?, data_type = ?, description = ?, updated_at = NOW() WHERE id = ?
		RETURNING id, schema_extension_id, name, display_name, data_type, description, created_at, updated_at, deleted_at`,
		property.TableName()),
		property.Name, property.SchemaExtensionId, property.DisplayName, property.DataType, property.Description, property.ID).Scan(&property); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return model.GraphSchemaProperty{}, fmt.Errorf("%w: %s", model.ErrDuplicateGraphSchemaExtensionPropertyName, property.Name)
		}
		return model.GraphSchemaProperty{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.GraphSchemaProperty{}, ErrNotFound
	}

	return property, nil
}

// DeleteGraphSchemaProperty - deletes a schema_properties row based on the provided id. It will return an error if that id does not exist.
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

// CreateGraphSchemaRelationshipKind - creates a new row in the schema_relationship_kinds table. A model.GraphSchemaRelationshipKind struct is
// returned, populated with the value as it stands in the database. This will also create a kind in the DAWGS kind table
// if the kind does not already exist.
//
// Since this inserts directly into the kinds table, the business logic calling this func
// must also call the DAWGS RefreshKinds function to ensure the kinds are reloaded into the in memory kind map.
func (s *BloodhoundDB) CreateGraphSchemaRelationshipKind(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) (model.GraphSchemaRelationshipKind, error) {
	var schemaRelationshipKind model.GraphSchemaRelationshipKind

	if result := s.db.WithContext(ctx).Raw(`
	WITH dawgs_kind (id, name) AS ( SELECT id, name FROM upsert_kind(?)),
	inserted_edges AS (
		INSERT INTO schema_relationship_kinds (kind_id, schema_extension_id, description, is_traversable)
		SELECT dk.id, ?, ?, ?
		FROM dawgs_kind dk
		RETURNING id, kind_id, schema_extension_id, description, is_traversable, created_at, updated_at, deleted_at
	)
	SELECT ie.id, ie.schema_extension_id, dk.name, ie.description, ie.is_traversable, ie.created_at, ie.updated_at, ie.deleted_at
	FROM inserted_edges ie
	JOIN dawgs_kind dk ON ie.kind_id = dk.id;`, name, schemaExtensionId, description, isTraversable).Scan(&schemaRelationshipKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return schemaRelationshipKind, fmt.Errorf("%w: %s", model.ErrDuplicateSchemaRelationshipKindName, name)
		}
		return schemaRelationshipKind, CheckError(result)
	}
	return schemaRelationshipKind, nil
}

// GetGraphSchemaRelationshipKinds - returns all rows from the schema_relationship_kinds table that matches the given model.Filters. It returns a slice of model.GraphSchemaRelationshipKinds
// populated with data, as well as an integer indicating the total number of rows returned by the query (excluding any given pagination).
func (s *BloodhoundDB) GetGraphSchemaRelationshipKinds(ctx context.Context, relationshipKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaRelationshipKinds, int, error) {
	var (
		schemaRelationshipKinds = model.GraphSchemaRelationshipKinds{}
		totalRowCount           int
	)

	if filterAndPagination, err := parseFiltersAndPagination(relationshipKindFilters, sort, skip, limit); err != nil {
		return schemaRelationshipKinds, 0, err
	} else {
		sqlStr := fmt.Sprintf(`SELECT ek.id, k.name, ek.schema_extension_id, ek.description, ek.is_traversable,
									ek.created_at, ek.updated_at, ek.deleted_at
									FROM %s ek
									JOIN %s k ON ek.kind_id = k.id
									%s %s %s`,
			model.GraphSchemaRelationshipKind{}.TableName(), kindTable, filterAndPagination.WhereClause,
			filterAndPagination.OrderSql, filterAndPagination.SkipLimit)
		if result := s.db.WithContext(ctx).Raw(sqlStr, filterAndPagination.Filter.params...).Scan(&schemaRelationshipKinds); result.Error != nil {
			return nil, 0, CheckError(result)
		} else {
			if limit > 0 || skip > 0 {
				countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s ek JOIN %s k on ek.kind_id = k.id %s`,
					model.GraphSchemaRelationshipKind{}.TableName(), kindTable, filterAndPagination.WhereClause)
				if countResult := s.db.WithContext(ctx).Raw(countSqlStr, filterAndPagination.Filter.params...).Scan(&totalRowCount); countResult.Error != nil {
					return model.GraphSchemaRelationshipKinds{}, 0, CheckError(countResult)
				}
			} else {
				totalRowCount = len(schemaRelationshipKinds)
			}
		}
		return schemaRelationshipKinds, totalRowCount, nil
	}
}

func (s *BloodhoundDB) GetGraphSchemaRelationshipKindsWithSchemaName(ctx context.Context, relationshipKindFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaRelationshipKindsWithNamedSchema, int, error) {
	var (
		schemaRelationshipKinds = model.GraphSchemaRelationshipKindsWithNamedSchema{}
		totalRowCount           int
	)

	if filterAndPagination, err := parseFiltersAndPagination(relationshipKindFilters, sort, skip, limit); err != nil {
		return schemaRelationshipKinds, 0, err
	} else {
		sqlStr := fmt.Sprintf(`SELECT edge.id, k.name, edge.description, edge.is_traversable, schema.name as schema_name
									FROM %s edge JOIN %s schema ON edge.schema_extension_id = schema.id JOIN %s k ON edge.kind_id = k.id %s %s %s`,
			model.GraphSchemaRelationshipKind{}.TableName(),
			model.GraphSchemaExtension{}.TableName(),
			kindTable,
			filterAndPagination.WhereClause,
			filterAndPagination.OrderSql,
			filterAndPagination.SkipLimit)

		if result := s.db.WithContext(ctx).Raw(sqlStr, filterAndPagination.Filter.params...).Scan(&schemaRelationshipKinds); result.Error != nil {
			return nil, 0, CheckError(result)
		} else {
			if limit > 0 || skip > 0 {
				countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s edge JOIN %s schema ON edge.schema_extension_id = schema.id JOIN %s k ON edge.kind_id = k.id %s`,
					model.GraphSchemaRelationshipKind{}.TableName(), model.GraphSchemaExtension{}.TableName(), kindTable,
					filterAndPagination.WhereClause)
				if countResult := s.db.WithContext(ctx).Raw(countSqlStr, filterAndPagination.Filter.params...).Scan(&totalRowCount); countResult.Error != nil {
					return model.GraphSchemaRelationshipKindsWithNamedSchema{}, 0, CheckError(countResult)
				}
			} else {
				totalRowCount = len(schemaRelationshipKinds)
			}
		}
		return schemaRelationshipKinds, totalRowCount, nil

	}
}

// GetGraphSchemaRelationshipKindById - retrieves a row from the schema_relationship_kinds table
func (s *BloodhoundDB) GetGraphSchemaRelationshipKindById(ctx context.Context, schemaRelationshipKindId int32) (model.GraphSchemaRelationshipKind, error) {
	var schemaRelationshipKind model.GraphSchemaRelationshipKind
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
	SELECT %s.id, name, schema_extension_id, description, is_traversable, created_at, updated_at, deleted_at
	FROM %s JOIN %s ON %s.kind_id = %s.id WHERE %s.id = ?`, schemaRelationshipKind.TableName(), schemaRelationshipKind.TableName(), kindTable,
		schemaRelationshipKind.TableName(), kindTable, schemaRelationshipKind.TableName()), schemaRelationshipKindId).First(&schemaRelationshipKind); result.Error != nil {
		return schemaRelationshipKind, CheckError(result)
	}
	return schemaRelationshipKind, nil
}

// UpdateGraphSchemaRelationshipKind - updates a row in the schema_relationship_kinds table based on the provided id. It will return an
// error if the target schema edge kind does not exist or if any of the updates violate the schema constraints.
//
// This function does NOT update the DAWGS name column since the schema_relationship_kinds table FKs to the DAWGS kind table, and that
// table is append only. A new edge kind should be created instead.
func (s *BloodhoundDB) UpdateGraphSchemaRelationshipKind(ctx context.Context, schemaRelationshipKind model.GraphSchemaRelationshipKind) (model.GraphSchemaRelationshipKind, error) {
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		WITH updated_row as (
			UPDATE %s
			SET schema_extension_id = ?, description = ?, is_traversable = ?, updated_at = NOW()
			WHERE id = ?
			RETURNING id, kind_id, schema_extension_id, description, is_traversable, created_at, updated_at, deleted_at
		)
		SELECT updated_row.id, %s.name, schema_extension_id, description, is_traversable, created_at, updated_at, deleted_at
		FROM updated_row
		JOIN %s ON %s.id = updated_row.kind_id`,
		schemaRelationshipKind.TableName(), kindTable, kindTable, kindTable),
		schemaRelationshipKind.SchemaExtensionId, schemaRelationshipKind.Description, schemaRelationshipKind.IsTraversable,
		schemaRelationshipKind.ID).Scan(&schemaRelationshipKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return schemaRelationshipKind, fmt.Errorf("%w: %v", model.ErrDuplicateSchemaRelationshipKindName, schemaRelationshipKind.Name)
		}
		return model.GraphSchemaRelationshipKind{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.GraphSchemaRelationshipKind{}, ErrNotFound
	}
	return schemaRelationshipKind, nil
}

// DeleteGraphSchemaRelationshipKind - deletes a schema_relationship_kind row based on the provided id. It will return an error if that id does not exist.
func (s *BloodhoundDB) DeleteGraphSchemaRelationshipKind(ctx context.Context, schemaRelationshipKindId int32) error {
	var schemaRelationshipKind model.GraphSchemaRelationshipKind
	if result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, schemaRelationshipKind.TableName()), schemaRelationshipKindId); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateEnvironment - creates a new schema_environment.
func (s *BloodhoundDB) CreateEnvironment(ctx context.Context, extensionId int32, environmentKindId int32, sourceKindId int32) (model.SchemaEnvironment, error) {
	var schemaEnvironment model.SchemaEnvironment

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		INSERT INTO %s (schema_extension_id, environment_kind_id, source_kind_id, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
		RETURNING id, schema_extension_id, environment_kind_id, source_kind_id, created_at, updated_at, deleted_at`,
		schemaEnvironment.TableName()),
		extensionId, environmentKindId, sourceKindId).Scan(&schemaEnvironment); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return model.SchemaEnvironment{}, fmt.Errorf("%w", model.ErrDuplicateSchemaEnvironment)
		}
		return model.SchemaEnvironment{}, CheckError(result)
	}
	return schemaEnvironment, nil
}

// GetEnvironments - retrieves list of schema environments.
func (s *BloodhoundDB) GetEnvironments(ctx context.Context) ([]model.SchemaEnvironment, error) {
	var result []model.SchemaEnvironment

	query := `
		SELECT
			se.id,
			se.schema_extension_id,
			ext.display_name as schema_extension_display_name,
			se.environment_kind_id,
			k.name as environment_kind_name,
			se.source_kind_id,
			se.created_at,
			se.updated_at,
			se.deleted_at
		FROM schema_environments se
		INNER JOIN kind k ON se.environment_kind_id = k.id
		INNER JOIN schema_extensions ext ON se.schema_extension_id = ext.id
		ORDER BY se.id`

	if err := CheckError(s.db.WithContext(ctx).Raw(query).Scan(&result)); err != nil {
		return nil, err
	}

	if result == nil {
		result = []model.SchemaEnvironment{}
	}

	return result, nil
}

// GetEnvironmentsByExtensionId - retrieves a slice of model.SchemaEnvironment by extension id.
func (s *BloodhoundDB) GetEnvironmentsByExtensionId(ctx context.Context, extensionId int32) ([]model.SchemaEnvironment, error) {
	var (
		environments = make([]model.SchemaEnvironment, 0)
	)

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
	SELECT e.id, e.schema_extension_id, e.environment_kind_id, k.name as "environment_kind_name", e.source_kind_id, e.created_at, e.updated_at, e.deleted_at
	FROM %s e
	JOIN %s k ON e.environment_kind_id = k.id
	WHERE schema_extension_id = ?
	ORDER BY id`,
		model.SchemaEnvironment{}.TableName(), kindTable), extensionId).Scan(&environments); result.Error != nil {
		return nil, CheckError(result)
	}

	return environments, nil

}

// GetEnvironmentByKinds - retrieves an environment by its environment kind and source kind.
func (s *BloodhoundDB) GetEnvironmentByKinds(ctx context.Context, environmentKindId, sourceKindId int32) (model.SchemaEnvironment, error) {
	var env model.SchemaEnvironment

	if result := s.db.WithContext(ctx).Raw(
		"SELECT * FROM schema_environments WHERE environment_kind_id = ? AND source_kind_id = ? AND deleted_at IS NULL",
		environmentKindId, sourceKindId,
	).Scan(&env); result.Error != nil {
		return model.SchemaEnvironment{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.SchemaEnvironment{}, ErrNotFound
	}

	return env, nil
}

// GetEnvironmentById - retrieves a schema environment by id.
func (s *BloodhoundDB) GetEnvironmentById(ctx context.Context, environmentId int32) (model.SchemaEnvironment, error) {
	var schemaEnvironment model.SchemaEnvironment

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, schema_extension_id, environment_kind_id, source_kind_id, created_at, updated_at, deleted_at
		FROM %s WHERE id = ?`,
		schemaEnvironment.TableName()),
		environmentId).Scan(&schemaEnvironment); result.Error != nil {
		return model.SchemaEnvironment{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.SchemaEnvironment{}, ErrNotFound
	}

	return schemaEnvironment, nil
}

// DeleteEnvironment - deletes a schema environment by id.
func (s *BloodhoundDB) DeleteEnvironment(ctx context.Context, environmentId int32) error {
	var schemaEnvironment model.SchemaEnvironment

	if result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, schemaEnvironment.TableName()), environmentId); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// CreateSchemaRelationshipFinding - creates a new schema relationship finding.
func (s *BloodhoundDB) CreateSchemaRelationshipFinding(ctx context.Context, extensionId int32, relationshipKindId int32, environmentId int32, name string, displayName string) (model.SchemaRelationshipFinding, error) {
	var finding model.SchemaRelationshipFinding

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		INSERT INTO %s (schema_extension_id, relationship_kind_id, environment_id, name, display_name, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
		RETURNING id, schema_extension_id, relationship_kind_id, environment_id, name, display_name, created_at`,
		finding.TableName()),
		extensionId, relationshipKindId, environmentId, name, displayName).Scan(&finding); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return model.SchemaRelationshipFinding{}, fmt.Errorf("%w: %s", model.ErrDuplicateSchemaRelationshipFindingName, name)
		}
		return model.SchemaRelationshipFinding{}, CheckError(result)
	}
	return finding, nil
}

// GetSchemaRelationshipFindingById - retrieves a schema relationship finding by id.
func (s *BloodhoundDB) GetSchemaRelationshipFindingById(ctx context.Context, findingId int32) (model.SchemaRelationshipFinding, error) {
	var finding model.SchemaRelationshipFinding

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, schema_extension_id, relationship_kind_id, environment_id, name, display_name, created_at
		FROM %s WHERE id = ?`,
		finding.TableName()),
		findingId).Scan(&finding); result.Error != nil {
		return model.SchemaRelationshipFinding{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.SchemaRelationshipFinding{}, ErrNotFound
	}

	return finding, nil
}

// GetSchemaRelationshipFindingByName - retrieves a schema relationship finding by finding name.
func (s *BloodhoundDB) GetSchemaRelationshipFindingByName(ctx context.Context, name string) (model.SchemaRelationshipFinding, error) {
	var finding model.SchemaRelationshipFinding

	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, schema_extension_id, relationship_kind_id, environment_id, name, display_name, created_at
		FROM %s WHERE name = ?`,
		finding.TableName()),
		name).Scan(&finding); result.Error != nil {
		return model.SchemaRelationshipFinding{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.SchemaRelationshipFinding{}, ErrNotFound
	}

	return finding, nil
}

// DeleteSchemaRelationshipFinding - deletes a schema relationship finding by id.
func (s *BloodhoundDB) DeleteSchemaRelationshipFinding(ctx context.Context, findingId int32) error {
	var finding model.SchemaRelationshipFinding

	if result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, finding.TableName()), findingId); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetSchemaRelationshipFindingsBySchemaExtensionId - returns all findings by extension id.
func (s *BloodhoundDB) GetSchemaRelationshipFindingsBySchemaExtensionId(ctx context.Context, extensionId int32) ([]model.SchemaRelationshipFinding, error) {
	var findings = make([]model.SchemaRelationshipFinding, 0)
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, schema_extension_id, relationship_kind_id, environment_id, name, display_name, created_at
		FROM %s WHERE schema_extension_id = ? ORDER BY id`, model.SchemaRelationshipFinding{}.TableName()), extensionId).Scan(&findings); result.Error != nil {
		return findings, CheckError(result)
	}
	return findings, nil
}

func (s *BloodhoundDB) CreateRemediation(ctx context.Context, findingId int32, shortDescription string, longDescription string, shortRemediation string, longRemediation string) (model.Remediation, error) {
	var remediation model.Remediation

	if result := s.db.WithContext(ctx).Raw(`
		WITH inserted AS (
			INSERT INTO schema_remediations (finding_id, content_type, content)
			VALUES
				(?, 'short_description', ?),
				(?, 'long_description', ?),
				(?, 'short_remediation', ?),
				(?, 'long_remediation', ?)
			RETURNING finding_id, content_type, content
		)
		SELECT
			finding_id,
			MAX(content) FILTER (WHERE content_type = 'short_description') as short_description,
			MAX(content) FILTER (WHERE content_type = 'long_description') as long_description,
			MAX(content) FILTER (WHERE content_type = 'short_remediation') as short_remediation,
			MAX(content) FILTER (WHERE content_type = 'long_remediation') as long_remediation
		FROM inserted
		GROUP BY finding_id`,
		findingId, shortDescription,
		findingId, longDescription,
		findingId, shortRemediation,
		findingId, longRemediation).Scan(&remediation); result.Error != nil {
		return model.Remediation{}, CheckError(result)
	}

	return remediation, nil
}

func (s *BloodhoundDB) GetRemediationByFindingId(ctx context.Context, findingId int32) (model.Remediation, error) {
	var remediation model.Remediation

	if result := s.db.WithContext(ctx).Raw(`
		SELECT
			finding_id,
			MAX(content) FILTER (WHERE content_type = 'short_description') as short_description,
			MAX(content) FILTER (WHERE content_type = 'long_description') as long_description,
			MAX(content) FILTER (WHERE content_type = 'short_remediation') as short_remediation,
			MAX(content) FILTER (WHERE content_type = 'long_remediation') as long_remediation
		FROM schema_remediations
		WHERE finding_id = ?
		GROUP BY finding_id`,
		findingId).Scan(&remediation); result.Error != nil {
		return model.Remediation{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.Remediation{}, ErrNotFound
	}

	return remediation, nil
}

func (s *BloodhoundDB) GetRemediationByFindingName(ctx context.Context, findingName string) (model.Remediation, error) {
	var remediation model.Remediation

	if result := s.db.WithContext(ctx).Raw(`
		SELECT
			sr.finding_id,
			srf.display_name,
			MAX(sr.content) FILTER (WHERE sr.content_type = 'short_description') as short_description,
			MAX(sr.content) FILTER (WHERE sr.content_type = 'long_description') as long_description,
			MAX(sr.content) FILTER (WHERE sr.content_type = 'short_remediation') as short_remediation,
			MAX(sr.content) FILTER (WHERE sr.content_type = 'long_remediation') as long_remediation
		FROM schema_remediations sr
		JOIN schema_relationship_findings srf ON sr.finding_id = srf.id
		WHERE srf.name = ?
		GROUP BY sr.finding_id, srf.display_name`,
		findingName).Scan(&remediation); result.Error != nil {
		return model.Remediation{}, CheckError(result)
	} else if result.RowsAffected == 0 {
		return model.Remediation{}, ErrNotFound
	}

	return remediation, nil
}

func (s *BloodhoundDB) UpdateRemediation(ctx context.Context, findingId int32, shortDescription string, longDescription string, shortRemediation string, longRemediation string) (model.Remediation, error) {
	var remediation model.Remediation

	if result := s.db.WithContext(ctx).Raw(`
		WITH upserted AS (
			INSERT INTO schema_remediations (finding_id, content_type, content)
			VALUES
				(?, 'short_description', ?),
				(?, 'long_description', ?),
				(?, 'short_remediation', ?),
				(?, 'long_remediation', ?)
			ON CONFLICT (finding_id, content_type) DO UPDATE SET content = EXCLUDED.content
			RETURNING finding_id, content_type, content
		)
		SELECT
			finding_id,
			MAX(content) FILTER (WHERE content_type = 'short_description') as short_description,
			MAX(content) FILTER (WHERE content_type = 'long_description') as long_description,
			MAX(content) FILTER (WHERE content_type = 'short_remediation') as short_remediation,
			MAX(content) FILTER (WHERE content_type = 'long_remediation') as long_remediation
		FROM upserted
		GROUP BY finding_id`,
		findingId, shortDescription,
		findingId, longDescription,
		findingId, shortRemediation,
		findingId, longRemediation).Scan(&remediation); result.Error != nil {
		return model.Remediation{}, CheckError(result)
	}

	return remediation, nil
}

func (s *BloodhoundDB) DeleteRemediation(ctx context.Context, findingId int32) error {
	if result := s.db.WithContext(ctx).Exec(`DELETE FROM schema_remediations WHERE finding_id = ?`, findingId); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *BloodhoundDB) CreatePrincipalKind(ctx context.Context, environmentId int32, principalKind int32) (model.SchemaEnvironmentPrincipalKind, error) {
	var envPrincipalKind model.SchemaEnvironmentPrincipalKind

	if result := s.db.WithContext(ctx).Raw(`
		INSERT INTO schema_environments_principal_kinds (environment_id, principal_kind, created_at)
		VALUES (?, ?, NOW())
		RETURNING environment_id, principal_kind, created_at`,
		environmentId, principalKind).Scan(&envPrincipalKind); result.Error != nil {
		if strings.Contains(result.Error.Error(), DuplicateKeyValueErrorString) {
			return model.SchemaEnvironmentPrincipalKind{}, fmt.Errorf("%w", model.ErrDuplicatePrincipalKind)
		}
		return model.SchemaEnvironmentPrincipalKind{}, CheckError(result)
	}

	return envPrincipalKind, nil
}

// GetPrincipalKindsByEnvironmentID - retrieves a schema environments principal kind by environment id.
func (s *BloodhoundDB) GetPrincipalKindsByEnvironmentId(ctx context.Context, environmentId int32) (model.SchemaEnvironmentPrincipalKinds, error) {
	var envPrincipalKinds model.SchemaEnvironmentPrincipalKinds

	if result := s.db.WithContext(ctx).Raw(`
		SELECT environment_id, principal_kind, created_at
		FROM schema_environments_principal_kinds
		WHERE environment_id = ?`,
		environmentId).Scan(&envPrincipalKinds); result.Error != nil {
		return nil, CheckError(result)
	}

	return envPrincipalKinds, nil
}

func (s *BloodhoundDB) DeletePrincipalKind(ctx context.Context, environmentId int32, principalKind int32) error {
	if result := s.db.WithContext(ctx).Exec(`
		DELETE FROM schema_environments_principal_kinds
		WHERE environment_id = ? AND principal_kind = ?`,
		environmentId, principalKind); result.Error != nil {
		return CheckError(result)
	} else if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func parseFiltersAndPagination(filters model.Filters, sort model.Sort, skip, limit int) (FilterAndPagination, error) {
	var (
		filtersAndPagination FilterAndPagination
		err                  error
	)
	filtersAndPagination.Filter, err = buildSQLFilter(filters)
	if err != nil {
		return filtersAndPagination, err
	}

	// if no sort specified, default to ID so pagination is consistent
	if len(sort) == 0 {
		sort = append(sort, model.SortItem{Column: "id", Direction: model.AscendingSortDirection})
	}
	filtersAndPagination.OrderSql, err = buildSQLSort(sort)
	if err != nil {
		return filtersAndPagination, err
	}

	if limit > 0 {
		filtersAndPagination.SkipLimit += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		filtersAndPagination.SkipLimit += fmt.Sprintf(" OFFSET %d", skip)
	}

	if filtersAndPagination.Filter.sqlString != "" {
		filtersAndPagination.WhereClause = fmt.Sprintf("WHERE %s", filtersAndPagination.Filter.sqlString)
	}
	return filtersAndPagination, nil
}
