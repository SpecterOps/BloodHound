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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"gorm.io/gorm"
)

const (
	EnvironmentAccessControlTable       = "environment_access_control"
	EnvironmentAccessControlFeatureFlag = "targeted_access_control"
)

type EnvironmentAccessControlData interface {
	GetEnvironmentAccessListForUser(ctx context.Context, user model.User) ([]EnvironmentAccess, error)
	UpdateEnvironmentListForUser(ctx context.Context, user model.User, environments ...string) error
}

// EnvironmentAccess defines the model for a row in the environment_access_control table
type EnvironmentAccess struct {
	UserID      string `json:"user_id"`
	Environment string `json:"environment"`
	model.BigSerial
}

func (EnvironmentAccess) TableName() string {
	return EnvironmentAccessControlTable
}

// GetEnvironmentAccessListForUser given a user's id, this will return all access control list rows for the user
func (s *BloodhoundDB) GetEnvironmentAccessListForUser(ctx context.Context, user model.User) ([]EnvironmentAccess, error) {
	var accessControlList []EnvironmentAccess

	result := s.db.WithContext(ctx).Table(EnvironmentAccessControlTable).Where("user_id = ?", user.ID.String()).Find(&accessControlList)
	return accessControlList, CheckError(result)
}

// UpdateEnvironmentListForUser will remove all entries in the access control list for a user and add a new entry for each environment provided
func (s *BloodhoundDB) UpdateEnvironmentListForUser(ctx context.Context, user model.User, environments ...string) error {
	var (
		auditData = model.AuditData{
			"userUuid":     user.ID.String(),
			"environments": environments,
		}
		auditEntry, err = model.NewAuditEntry(model.AuditLogActionUpdateEnvironmentAccessList, model.AuditLogStatusIntent, auditData)
	)

	if err != nil {
		return err
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {

		err := CheckError(tx.WithContext(ctx).Table(EnvironmentAccessControlTable).Where("user_id = ?", user.ID.String()).Delete(&EnvironmentAccess{}))

		if err != nil {
			return err
		}

		availableEnvironments := make([]EnvironmentAccess, 0, len(environments))

		for _, environment := range environments {
			newAccessControl := EnvironmentAccess{
				UserID:      user.ID.String(),
				Environment: environment,
			}

			availableEnvironments = append(availableEnvironments, newAccessControl)
		}

		result := tx.WithContext(ctx).Table(EnvironmentAccessControlTable).Create(&availableEnvironments)

		if err := CheckError(result); err != nil {
			return err
		}

		user.AllEnvironments = false

		err = s.UpdateUser(ctx, user)

		if err != nil {
			return err
		}

		return nil
	})
}
