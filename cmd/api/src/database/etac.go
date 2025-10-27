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
	ETACTable = "environment_targeted_access_control"
)

type EnvironmentTargetedAccessControlData interface {
	GetEnvironmentTargetedAccessControlForUser(ctx context.Context, user model.User) ([]model.EnvironmentTargetedAccessControl, error)
	DeleteEnvironmentTargetedAccessControlForUser(ctx context.Context, user model.User) error
}

// GetEnvironmentTargetedAccessControlForUser given a user's id, this will return all access control list rows for the user
func (s *BloodhoundDB) GetEnvironmentTargetedAccessControlForUser(ctx context.Context, user model.User) ([]model.EnvironmentTargetedAccessControl, error) {
	var accessControlList []model.EnvironmentTargetedAccessControl

	result := s.db.WithContext(ctx).Table(ETACTable).Where("user_id = ?", user.ID.String()).Find(&accessControlList)
	return accessControlList, CheckError(result)
}

// DeleteEnvironmentTargetedAccessControlForUser will remove all rows associated with a user in the environment_targeted_access_control table
func (s *BloodhoundDB) DeleteEnvironmentTargetedAccessControlForUser(ctx context.Context, user model.User) error {
	var (
		auditData = model.AuditData{
			"userUuid": user.ID.String(),
		}
		auditEntry, err = model.NewAuditEntry(model.AuditLogActionDeleteETACList, model.AuditLogStatusIntent, auditData)
	)
	if err != nil {
		return fmt.Errorf("error creating AuditLogActionDeleteETACList audit entry: %w", err)
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Exec("DELETE FROM environment_targeted_access_control WHERE user_id = ?", user.ID.String())
		return CheckError(result)
	})
}
