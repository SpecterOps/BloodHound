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

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"gorm.io/gorm"
)

func (s *Migrator) setAppConfigDefaults() error {
	if err := s.setParameterDefaults(); err != nil {
		return err
	}

	return s.setFeatureFlagDefaults()
}

func (s *Migrator) setFeatureFlagDefaults() error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		for flagKey, availableFlag := range appcfg.AvailableFlags() {
			count := int64(0)

			if result := tx.Model(&appcfg.FeatureFlag{}).Where("key = ?", flagKey).Count(&count); count == 0 {
				if result := tx.Create(&availableFlag); result.Error != nil {
					return fmt.Errorf("error creating feature flag %s: %w", flagKey, result.Error)
				}

				log.Infof("Feature flag %s created", flagKey)
			} else if result.Error != nil {
				return fmt.Errorf("error looking up existing feature flag %s: %w", flagKey, result.Error)
			}
		}

		return nil
	})
}

func (s *Migrator) setParameterDefaults() error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if availParams, err := appcfg.AvailableParameters(); err != nil {
			return fmt.Errorf("error checking AvailableParameters: %w", err)
		} else {
			for parameterKey, availableParameter := range availParams {
				count := int64(0)

				if result := tx.Model(&appcfg.Parameter{}).Where("key = ?", parameterKey).Count(&count); count == 0 {
					if result := tx.Create(&availableParameter); result.Error != nil {
						return fmt.Errorf("error setting configuration parameter %s(%s): %w", parameterKey, availableParameter.Name, result.Error)
					}

					log.Infof("Configuration parameter %s created", parameterKey)
				} else if result.Error != nil {
					return fmt.Errorf("error looking up existing feature flag %s: %w", parameterKey, result.Error)
				}
			}

			return nil
		}
	})
}
