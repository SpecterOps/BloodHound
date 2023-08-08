// Copyright 2023 Specter Ops, Inc.
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

package migration

import (
	"fmt"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"gorm.io/gorm"
)

func ListBHModels() []any {
	return []any{
		// NOTE: We are not letting Gorm re-order the model list to determine relational dependency because
		// it seems to get it wrong. Please make sure models are ordered correctly.

		// Runtime configuration parameters
		&appcfg.Parameter{},
		&appcfg.FeatureFlag{},

		// Audit log model
		&model.AuditLog{},

		// Asset group isolation model
		&model.AssetGroup{},
		&model.AssetGroupSelector{},
		&model.AssetGroupCollection{},
		&model.AssetGroupCollectionEntry{},

		// Auth model
		&model.Installation{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.AuthSecret{},
		&model.AuthToken{},
		&model.SAMLProvider{},
		&model.UserSession{},

		// Ingest model
		&model.IngestTask{},

		// Database stats
		&model.ADDataQualityStat{},
		&model.ADDataQualityAggregation{},
		&model.AzureDataQualityStat{},
		&model.AzureDataQualityAggregation{},
		&model.DomainCollectionResult{},

		&model.FileUploadJob{},
	}
}

func (s *Migrator) gormAutoMigrate(models []any) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, currentModel := range models {
			if result := tx.Exec(`alter table asset_group_selectors drop constraint if exists idx_asset_group_selectors_name;`); result.Error != nil {
				return result.Error
			}

			if err := tx.Migrator().AutoMigrate(currentModel); err != nil {
				return fmt.Errorf("failed to migrate model %T: %w", currentModel, err)
			}
		}
		return nil
	})
}
