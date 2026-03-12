// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"gorm.io/gorm"
)

const (
	customNodeKindTable = "custom_node_kinds"
)

type CustomNodeKindData interface {
	CreateCustomNodeKinds(ctx context.Context, customNodeKind model.CustomNodeKinds) (model.CustomNodeKinds, error)
	GetCustomNodeKinds(ctx context.Context, filters model.Filters) ([]model.CustomNodeKind, error)
	GetCustomNodeKindsMap(ctx context.Context) (model.CustomNodeKindMap, error)
	GetCustomNodeKind(ctx context.Context, kindName string) (model.CustomNodeKind, error)
	UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error)
	DeleteCustomNodeKind(ctx context.Context, kindName string) error
}

func (s *BloodhoundDB) CreateCustomNodeKinds(ctx context.Context, customNodeKinds model.CustomNodeKinds) (model.CustomNodeKinds, error) {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateCustomNodeKind,
			Model:  &customNodeKinds,
		}
	)

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		err := tx.Create(&customNodeKinds).Error

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"custom_node_kinds_kind_name_key\"") {
				return fmt.Errorf("%w: %v", ErrDuplicateCustomNodeKindName, err)
			}
		}

		return err
	})

	return customNodeKinds, err
}

func (s *BloodhoundDB) GetCustomNodeKinds(ctx context.Context, filters model.Filters) ([]model.CustomNodeKind, error) {
	var customNodeKinds []model.CustomNodeKind

	sqlFilter, err := buildSQLFilter(filters)
	if err != nil {
		return nil, err
	}

	whereClause := ""
	if sqlFilter.sqlString != "" {
		whereClause = fmt.Sprintf("WHERE %s", sqlFilter.sqlString)
	}
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, kind_name, config FROM %s %s;", customNodeKindTable, whereClause)).Scan(&customNodeKinds)

	return customNodeKinds, CheckError(result)
}

func (s *BloodhoundDB) GetCustomNodeKind(ctx context.Context, kindName string) (model.CustomNodeKind, error) {
	var customNodeKind model.CustomNodeKind
	if results, err := s.GetCustomNodeKinds(ctx, model.Filters{"kind_name": []model.Filter{{Value: kindName, Operator: model.Equals}}}); err != nil {
		return customNodeKind, err
	} else if len(results) == 0 {
		return customNodeKind, ErrNotFound
	} else {
		return results[0], nil
	}
}

func (s *BloodhoundDB) GetCustomNodeKindsMap(ctx context.Context) (model.CustomNodeKindMap, error) {
	if openGraphSearchFeatureFlag, err := s.GetFlagByKey(ctx, appcfg.FeatureOpenGraphSearch); err != nil {
		return nil, err
	} else if !openGraphSearchFeatureFlag.Enabled {
		return nil, nil
	} else if customNodeKinds, err := s.GetCustomNodeKinds(ctx, nil); err != nil {
		return nil, err
	} else {
		customNodeKindMap := make(model.CustomNodeKindMap, len(customNodeKinds))
		for _, kind := range customNodeKinds {
			customNodeKindMap[kind.KindName] = kind.Config
		}
		return customNodeKindMap, nil
	}
}

func (s *BloodhoundDB) UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error) {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionUpdateCustomNodeKind,
			Model:  &customNodeKind,
		}
	)

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
			bhdb := NewBloodhoundDB(tx, s.idResolver, s.config)
		if result := tx.Raw(fmt.Sprintf("UPDATE %s SET schema_node_kind_id = COALESCE(?, schema_node_kind_id), config = ?, updated_at = NOW() WHERE kind_name = ? RETURNING id;", customNodeKindTable), customNodeKind.SchemaNodeKindId, customNodeKind.Config, customNodeKind.KindName).
			Scan(&customNodeKind.ID); result.RowsAffected == 0 {
			return ErrNotFound
		} else if result.Error != nil {
			return CheckError(result)
		} else if customNodeKind.SchemaNodeKindId != nil {
			// Update the icon in the schema_node_kinds table to match the new icon, if a schema_node_kind_id exists
			if _, err := bhdb.UpdateGraphSchemaNodeKindIconById(ctx, *customNodeKind.SchemaNodeKindId, customNodeKind.Config.Icon); err != nil {
				return err
			}
		}

		return nil

	})

	return customNodeKind, err
}

func (s *BloodhoundDB) DeleteCustomNodeKind(ctx context.Context, kindName string) error {
	var (
		customNodeKind = model.CustomNodeKind{KindName: kindName}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionDeleteCustomNodeKind,
			Model:  &customNodeKind,
		}
	)

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		if err := tx.Raw(fmt.Sprintf("DELETE FROM %s WHERE kind_name = ? RETURNING id, config;", customNodeKindTable), kindName).
			Row().Scan(&customNodeKind.ID, &customNodeKind.Config); errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else {
			return err
		}
	})

	return err
}
