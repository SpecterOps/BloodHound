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

package database

import (
	"strings"

	"github.com/specterops/bloodhound/src/model/appcfg"
	"gorm.io/gorm/clause"
)

func (s *BloodhoundDB) GetFlag(id int32) (appcfg.FeatureFlag, error) {
	var flag appcfg.FeatureFlag
	return flag, CheckError(s.db.Find(&flag, id))
}

func (s *BloodhoundDB) GetFlagByKey(key string) (appcfg.FeatureFlag, error) {
	var flag appcfg.FeatureFlag
	return flag, CheckError(s.db.Where("key = ?", key).First(&flag))
}

func (s *BloodhoundDB) GetAllFlags() ([]appcfg.FeatureFlag, error) {
	var flags []appcfg.FeatureFlag
	return flags, CheckError(s.db.Find(&flags))
}

func (s *BloodhoundDB) SetFlag(flag appcfg.FeatureFlag) error {
	return CheckError(s.db.Save(&flag))
}

func (s *BloodhoundDB) GetAllConfigurationParameters() (appcfg.Parameters, error) {
	var appConfig appcfg.Parameters
	return appConfig, CheckError(s.db.Find(&appConfig))
}

func (s *BloodhoundDB) GetConfigurationParameter(parameterKey string) (appcfg.Parameter, error) {
	var parameter appcfg.Parameter
	return parameter, CheckError(s.db.First(&parameter, "key = ?", parameterKey))
}

func (s *BloodhoundDB) GetConfigurationParametersByPrefix(parameterKeyPrefix string) (appcfg.Parameters, error) {
	const matchAllWildcard = "%"

	if !strings.HasSuffix(parameterKeyPrefix, matchAllWildcard) {
		parameterKeyPrefix += matchAllWildcard
	}

	var parameters appcfg.Parameters
	return parameters, CheckError(s.db.First(&parameters, "key like ?", parameterKeyPrefix))
}

func (s *BloodhoundDB) SetConfigurationParameter(parameter appcfg.Parameter) error {
	return CheckError(s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&parameter))
}
