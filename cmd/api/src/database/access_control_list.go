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

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"gorm.io/gorm"
)

const (
	EnvironmentAccessControlTable = "environment_access_control"
)

type EnvironmentAccessControlData interface {
	GetEnvironmentAccessListForUser(ctx context.Context, userUuid uuid.UUID) (EnvironmentAccessList, error)
	UpdateEnvironmentListForUser(ctx context.Context, userUuid uuid.UUID, environments ...string) error
}

// EnvironmentAccess defines the model for a row in the environment_access_control table
type EnvironmentAccess struct {
	UserID      string `json:"user_id"`
	Environment string `json:"environment"`
	model.BigSerial
}

// EnvironmentAccessList is a slice of EnvironmentAccess that provides additional helper methods
type EnvironmentAccessList []EnvironmentAccess

// GetEnvironmentAccessListForUser given a user's id, this will return all access control list rows for the user
func (s *BloodhoundDB) GetEnvironmentAccessListForUser(ctx context.Context, userUuid uuid.UUID) (EnvironmentAccessList, error) {
	var accessControlList []EnvironmentAccess

	result := s.db.WithContext(ctx).Table(EnvironmentAccessControlTable).Select("environment").Where("user_id = ?", userUuid.String()).Scan(&accessControlList)
	return accessControlList, CheckError(result)
}

// UpdateEnvironmentListForUser will remove all entries in the access control list for a user and add a new entry for each environment provided
func (s *BloodhoundDB) UpdateEnvironmentListForUser(ctx context.Context, userUuid uuid.UUID, environments ...string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := CheckError(tx.WithContext(ctx).Table(EnvironmentAccessControlTable).Where("user_id = ?", userUuid).Delete(&EnvironmentAccess{})); err != nil {
			return err
		} else {

			availableEnvironments := make([]EnvironmentAccess, 0, 10)

			for _, environment := range environments {
				newAccessControl := EnvironmentAccess{
					UserID:      userUuid.String(),
					Environment: environment,
				}

				availableEnvironments = append(availableEnvironments, newAccessControl)
			}

			result := tx.WithContext(ctx).Table(EnvironmentAccessControlTable).Create(&availableEnvironments)

			if err := CheckError(result); err != nil {
				return err
			}
		}
		return nil
	})
}
