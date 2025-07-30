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
	"gorm.io/gorm"
)

const (
	customNodeKindTable = "custom_node_kinds"
)

type CustomNodeKindData interface {
	CreateCustomNodeKinds(ctx context.Context, customNodeKind model.CustomNodeKinds) (model.CustomNodeKinds, error)
	GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error)
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

func (s *BloodhoundDB) GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error) {
	var customNodeKinds []model.CustomNodeKind
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, kind_name, config FROM %s;", customNodeKindTable)).Scan(&customNodeKinds)

	return customNodeKinds, CheckError(result)
}

func (s *BloodhoundDB) GetCustomNodeKind(ctx context.Context, kindName string) (model.CustomNodeKind, error) {
	var customNodeKind model.CustomNodeKind
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, kind_name, config FROM %s WHERE kind_name = ?;", customNodeKindTable), kindName).Scan(&customNodeKind)
	if result.RowsAffected == 0 {
		return customNodeKind, ErrNotFound
	}

	return customNodeKind, CheckError(result)
}

func (s *BloodhoundDB) UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error) {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionUpdateCustomNodeKind,
			Model:  &customNodeKind,
		}
	)

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		result := tx.Raw(fmt.Sprintf("UPDATE %s SET config = ?, updated_at = NOW() WHERE kind_name = ? RETURNING id;", customNodeKindTable), customNodeKind.Config, customNodeKind.KindName).
			Scan(&customNodeKind.ID)
		if result.RowsAffected == 0 {
			return ErrNotFound
		}

		return CheckError(result)
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
