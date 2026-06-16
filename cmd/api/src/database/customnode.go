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
	"errors"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"gorm.io/gorm"
)

var CustomNodeKindStubConfig = model.CustomNodeKindConfig{
	Icon: graphschema.DisplayNodeIcon{
		Name:  "question",
		Type:  graphschema.DisplayNodeTypeFontAwesome,
		Color: "#FFFFFF",
	},
}

type CustomNodeKindData interface {
	CreateCustomNodeKinds(ctx context.Context, customNodeKind model.CustomNodeKinds) (model.CustomNodeKinds, error)
	EnsureStubbedCustomNodeKindForIngest(ctx context.Context, name string) error
	GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error)
	GetCustomNodeKind(ctx context.Context, kindName string) (model.CustomNodeKind, error)
	UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error)
}

func (s *BloodhoundDB) EnsureStubbedCustomNodeKindForIngest(ctx context.Context, name string) error {
	if name == "" || model.IsExtendedNodeKind(graph.StringKind(name)) {
		return errors.New("invalid kind name")
	}

	if result := s.db.WithContext(ctx).Exec("SELECT ensure_stubbed_custom_node_kind_for_ingest(?, ?);", name, CustomNodeKindStubConfig); result.Error != nil {
		return fmt.Errorf("failed to ensure custom node kind stub %q: %w", name, result.Error)
	}

	return nil
}

func (s *BloodhoundDB) CreateCustomNodeKinds(ctx context.Context, customNodeKinds model.CustomNodeKinds) (model.CustomNodeKinds, error) {
	var auditEntry = model.AuditEntry{
		Action: model.AuditLogActionCreateCustomNodeKind,
		Model:  &customNodeKinds,
	}

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.pool, s.idResolver, s.config)

		// Upsert each kind name into the kind table and capture the assigned ID.
		for i := range customNodeKinds {
			upsertedKind, err := bhdb.UpsertKind(ctx, customNodeKinds[i].KindName)
			if err != nil {
				return fmt.Errorf("failed to upsert kind %q: %w", customNodeKinds[i].KindName, err)
			}

			customNodeKinds[i].KindId = int16(upsertedKind.ID)
		}

		// Insert each custom_node_kinds row individually so we can capture the returned id and detect the unique constraint violation on kind_id.
		for i := range customNodeKinds {
			var newID int32

			result := tx.Raw(
				fmt.Sprintf("INSERT INTO %s (config, schema_node_kind_id, kind_id) VALUES (?, ?, ?) RETURNING id", model.CustomNodeKind{}.TableName()),
				customNodeKinds[i].Config, customNodeKinds[i].SchemaNodeKindId, customNodeKinds[i].KindId,
			).Scan(&newID)

			if result.Error != nil {
				if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint \"custom_node_kinds_kind_id_key\"") {
					return fmt.Errorf("%w: %v", ErrDuplicateCustomNodeKindName, result.Error)
				}

				return result.Error
			}

			customNodeKinds[i].ID = newID
		}

		return nil
	})

	return customNodeKinds, err
}

const customNodeKindsSelectQuery = `
	SELECT cnk.id, k.name AS kind_name, cnk.kind_id, cnk.schema_node_kind_id, cnk.config
	FROM %s cnk
	JOIN %s k ON k.id = cnk.kind_id`

func (s *BloodhoundDB) GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error) {
	var customNodeKinds []model.CustomNodeKind

	result := s.db.WithContext(ctx).Raw(
		fmt.Sprintf(customNodeKindsSelectQuery+" ORDER BY cnk.id;", model.CustomNodeKind{}.TableName(), model.Kind{}.TableName()),
	).Scan(&customNodeKinds)

	return customNodeKinds, CheckError(result)
}

func (s *BloodhoundDB) GetCustomNodeKind(ctx context.Context, kindName string) (model.CustomNodeKind, error) {
	var customNodeKind model.CustomNodeKind

	result := s.db.WithContext(ctx).Raw(
		fmt.Sprintf(customNodeKindsSelectQuery+" WHERE k.name = ?;", model.CustomNodeKind{}.TableName(), model.Kind{}.TableName()),
		kindName,
	).Scan(&customNodeKind)

	if result.Error != nil {
		return customNodeKind, CheckError(result)
	}

	if result.RowsAffected == 0 {
		return customNodeKind, ErrNotFound
	}

	return customNodeKind, nil
}

func (s *BloodhoundDB) UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error) {
	var auditEntry = model.AuditEntry{
		Action: model.AuditLogActionUpdateCustomNodeKind,
		Model:  &customNodeKind,
	}

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.pool, s.idResolver, s.config)

		result := tx.Raw(
			fmt.Sprintf(`UPDATE %s cnk
				SET schema_node_kind_id = COALESCE(?, schema_node_kind_id), config = ?, updated_at = NOW()
				FROM kind k
				WHERE k.id = cnk.kind_id AND k.name = ?
				RETURNING cnk.id, k.name AS kind_name, cnk.kind_id, cnk.schema_node_kind_id, cnk.config`,
				model.CustomNodeKind{}.TableName()),
			customNodeKind.SchemaNodeKindId, customNodeKind.Config, customNodeKind.KindName,
		).Scan(&customNodeKind)

		if result.Error != nil {
			return CheckError(result)
		} else if result.RowsAffected == 0 {
			return ErrNotFound
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
