// Copyright 2024 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *BloodhoundDB) GetAllConfigurationParameters(ctx context.Context) (appcfg.Parameters, error) {
	var appConfig appcfg.Parameters
	return appConfig, CheckError(s.db.WithContext(ctx).Find(&appConfig))
}

func (s *BloodhoundDB) GetConfigurationParameter(ctx context.Context, parameterKey string) (appcfg.Parameter, error) {
	var parameter appcfg.Parameter
	return parameter, CheckError(s.db.WithContext(ctx).First(&parameter, "key = ?", parameterKey))
}

func (s *BloodhoundDB) SetConfigurationParameter(ctx context.Context, parameter appcfg.Parameter) error {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionUpdateParameter,
		Model:  &parameter, // Pointer is required to ensure success log contains updated fields after transaction
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(s.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(&parameter))
	})
}
