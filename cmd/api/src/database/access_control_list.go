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
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"gorm.io/gorm"
)

const (
	EnvironmentAccessControlTable = "environment_access_control"
)

type EnvironmentAccessControlData interface {
	GetEnvironmentAccessListForUser(ctx context.Context, user model.User) ([]model.EnvironmentAccess, error)
	UpdateEnvironmentListForUser(ctx context.Context, user model.User, environments []string) ([]model.EnvironmentAccess, error)
	DeleteEnvironmentListForUser(ctx context.Context, user model.User) error
}

// GetEnvironmentAccessListForUser given a user's id, this will return all access control list rows for the user
func (s *BloodhoundDB) GetEnvironmentAccessListForUser(ctx context.Context, user model.User) ([]model.EnvironmentAccess, error) {
	var accessControlList []model.EnvironmentAccess

	result := s.db.WithContext(ctx).Table(EnvironmentAccessControlTable).Where("user_id = ?", user.ID.String()).Find(&accessControlList)
	return accessControlList, CheckError(result)
}

// UpdateEnvironmentListForUser will remove all entries in the access control list for a user and add a new entry for each environment provided
// This method will also set all_environments to false for the provided user
func (s *BloodhoundDB) UpdateEnvironmentListForUser(ctx context.Context, user model.User, environments []string) ([]model.EnvironmentAccess, error) {
	var (
		auditData = model.AuditData{
			"userUuid":     user.ID.String(),
			"environments": environments,
		}
		auditEntry, err       = model.NewAuditEntry(model.AuditLogActionUpdateEnvironmentAccessList, model.AuditLogStatusIntent, auditData)
		availableEnvironments = make([]model.EnvironmentAccess, 0, len(environments))
	)

	if err != nil {
		return nil, err
	}

	err = s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {

		// When updating a user's environment list, we will first clear their existing environment list to avoid duplication
		if err := s.DeleteEnvironmentListForUser(ctx, user); err != nil {
			return fmt.Errorf("error deleting user's environment list: %w", err)
		}

		// If there are no environments being set, then simply return after deleting the user's existing environments
		if len(environments) == 0 {
			return nil
		}

		for _, environment := range environments {
			newAccessControl := model.EnvironmentAccess{
				UserID:      user.ID.String(),
				Environment: environment,
			}

			availableEnvironments = append(availableEnvironments, newAccessControl)
		}

		result := tx.WithContext(ctx).Table(EnvironmentAccessControlTable).Create(&availableEnvironments)

		if err := CheckError(result); err != nil {
			return err
		}

		// If a user has a TAC List, then they no longer have access to all environments
		user.AllEnvironments = false
		saveUserResult := tx.WithContext(ctx).Save(user)

		return CheckError(saveUserResult)
	})

	return availableEnvironments, err
}

// DeleteEnvironmentListForUser will remove all rows associated with a user in the environment_access_control table
func (s *BloodhoundDB) DeleteEnvironmentListForUser(ctx context.Context, user model.User) error {
	var (
		auditData = model.AuditData{
			"userUuid": user.ID.String(),
		}
		auditEntry, err = model.NewAuditEntry(model.AuditLogActionDeleteEnvironmentAccessList, model.AuditLogStatusIntent, auditData)
	)
	if err != nil {
		return fmt.Errorf("error creating AuditLogActionDeleteEnvironmentAccessList audit entry: %w", err)
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Delete(&model.EnvironmentAccess{}, "user_id = ?", user.ID.String())
		return CheckError(result)
	})
}
