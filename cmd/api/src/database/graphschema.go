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

func (s *BloodhoundDB) GetGraphSchemaExtensionsFilteredAndPaginated(ctx context.Context, extensionSqlFilter model.SQLFilter, sort model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error) {
	var (
		extensions      = model.GraphSchemaExtensions{}
		skipLimitString string
		totalRowCount   int
		orderSQL        string
	)

	var extensionSqlFilterStr string
	if extensionSqlFilter.SQLString != "" {
		extensionSqlFilterStr = " WHERE " + extensionSqlFilter.SQLString
	}

	if len(sort) == 0 {
		sort = append(sort, model.SortItem{Column: "id", Direction: model.AscendingSortDirection})
	}

	var sortColumns []string
	for _, item := range sort {
		dirString := "ASC"
		if item.Direction == model.DescendingSortDirection {
			dirString = "DESC"
		}
		sortColumns = append(sortColumns, fmt.Sprintf("%s %s", item.Column, dirString))
	}
	orderSQL = "ORDER BY " + strings.Join(sortColumns, ", ")

	if limit > 0 {
		skipLimitString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		skipLimitString += fmt.Sprintf(" OFFSET %d", skip)
	}

	var (
		sqlStr = fmt.Sprintf(`SELECT id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at
								FROM %s %s %s %s`,
			model.GraphSchemaExtension{}.TableName(),
			extensionSqlFilterStr,
			orderSQL,
			skipLimitString)
	)

	if result := s.db.WithContext(ctx).Raw(sqlStr, extensionSqlFilter.Params...).Scan(&extensions); result.Error != nil {
		return model.GraphSchemaExtensions{}, 0, CheckError(result)
	} else {
		// we need an overall count of the rows if pagination is supplied
		if limit > 0 || skip > 0 {
			countSqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`,
				model.GraphSchemaExtension{}.TableName(),
				extensionSqlFilterStr)

			if err := s.db.WithContext(ctx).Raw(countSqlStr, extensionSqlFilter.Params...).Scan(&totalRowCount).Error; err != nil {
				return model.GraphSchemaExtensions{}, 0, err
			}
		} else {
			totalRowCount = len(extensions)
		}
	}

	return extensions, totalRowCount, nil
}
