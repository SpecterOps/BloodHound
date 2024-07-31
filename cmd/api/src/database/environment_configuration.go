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
	"encoding/json"

	"github.com/specterops/bloodhound/src/model"
	// "gorm.io/gorm"
)

type EnvironmentConfigurationsData interface {
	CreateEnvironmentConfiguration(ctx context.Context, name string, data json.RawMessage) (model.EnvironmentConfiguration, error)
	GetEnvironmentConfiguration(ctx context.Context, name string) (model.EnvironmentConfiguration, error)
	ListEnvironmentConfigurations(ctx context.Context, order string, filter model.SQLFilter, skip, limit int) (model.EnvironmentConfigurations, int, error)
}

func (s *BloodhoundDB) CreateEnvironmentConfiguration(ctx context.Context, name string, data json.RawMessage) (model.EnvironmentConfiguration, error) {
	envConfig := model.EnvironmentConfiguration{
		Name: name,
		Data: data,
	}

	result := s.db.WithContext(ctx).Create(&envConfig)
	return envConfig, CheckError(result)
}

func (s *BloodhoundDB) GetEnvironmentConfiguration(ctx context.Context, name string) (model.EnvironmentConfiguration, error) {
	var envConfig model.EnvironmentConfiguration
	result := s.db.WithContext(ctx).Where("name = ?", name).First(&envConfig)
	return envConfig, CheckError(result)
}

func (s *BloodhoundDB) ListEnvironmentConfigurations(ctx context.Context, order string, filter model.SQLFilter, skip, limit int) (model.EnvironmentConfigurations, int, error) {
	var (
		envConfigs model.EnvironmentConfigurations
		count      int64
		cursor     = s.Scope(Paginate(skip, limit)).WithContext(ctx)
	)

	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params)
		result := s.db.Model(&envConfigs).WithContext(ctx).Where(filter.SQLString, filter.Params).Count(&count)
		if result.Error != nil {
			return envConfigs, 0, result.Error
		}
	} else {
		result := s.db.Model(&envConfigs).WithContext(ctx).Count(&count)
		if result.Error != nil {
			return envConfigs, 0, result.Error
		}
	}

	if order != "" {
		cursor = cursor.Order(order)
	}
	result := cursor.Find(&envConfigs)

	return envConfigs, int(count), CheckError(result)
}