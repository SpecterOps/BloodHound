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
	"time"

	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

func (s *BloodhoundDB) CreateADDataQualityStats(stats model.ADDataQualityStats) (model.ADDataQualityStats, error) {
	result := s.db.Create(&stats)
	return stats, CheckError(result)
}

func (s *BloodhoundDB) GetADDataQualityStats(domainSid string, start time.Time, end time.Time, order string, limit int, skip int) (model.ADDataQualityStats, int, error) {
	const (
		defaultOrder = "created_at desc"
		defaultWhere = "domain_sid = ? and (created_at between ? and ?)"
	)

	var (
		adDataQualityStats model.ADDataQualityStats
		count              int64
		result             *gorm.DB
	)

	if order != "" {
		result = s.Scope(Paginate(skip, limit)).Where(defaultWhere, domainSid, start, end).Order(order).Find(&adDataQualityStats)
		if CheckError(result) != nil {
			return adDataQualityStats, 0, result.Error
		}
		result = s.db.Model(model.ADDataQualityStats{}).Where(defaultWhere, domainSid, start, end).Order(order).Count(&count)
		if CheckError(result) != nil {
			return adDataQualityStats, 0, result.Error
		}
	} else {
		result = s.Scope(Paginate(skip, limit)).Where(defaultWhere, domainSid, start, end).Order(defaultOrder).Find(&adDataQualityStats)
		if CheckError(result) != nil {
			return adDataQualityStats, 0, result.Error
		}
		result = s.db.Model(model.ADDataQualityStats{}).Where(defaultWhere, domainSid, start, end).Order(defaultOrder).Model(model.ADDataQualityStats{}).Count(&count)
		if CheckError(result) != nil {
			return adDataQualityStats, 0, result.Error
		}
	}

	return adDataQualityStats, int(count), nil
}

func (s *BloodhoundDB) CreateADDataQualityAggregation(aggregation model.ADDataQualityAggregation) (model.ADDataQualityAggregation, error) {
	result := s.db.Create(&aggregation)
	return aggregation, CheckError(result)
}

func (s *BloodhoundDB) GetADDataQualityAggregations(start time.Time, end time.Time, order string, limit int, skip int) (model.ADDataQualityAggregations, int, error) {
	const (
		defaultOrder = "created_at desc"
		defaultWhere = "created_at between ? and ?"
	)

	var (
		adDataQualityAggregations model.ADDataQualityAggregations
		count                     int64
		result                    *gorm.DB
	)

	if order != "" {
		result = s.Scope(Paginate(skip, limit)).Where(defaultWhere, start, end).Order(order).Find(&adDataQualityAggregations)
		if CheckError(result) != nil {
			return adDataQualityAggregations, 0, result.Error
		}
		result = s.db.Model(model.ADDataQualityAggregations{}).Where(defaultWhere, start, end).Order(order).Count(&count)
		if CheckError(result) != nil {
			return adDataQualityAggregations, 0, result.Error
		}
	} else {
		result = s.Scope(Paginate(skip, limit)).Where(defaultWhere, start, end).Order(defaultOrder).Find(&adDataQualityAggregations)
		if CheckError(result) != nil {
			return adDataQualityAggregations, 0, result.Error
		}
		result = s.db.Model(model.ADDataQualityAggregations{}).Where(defaultWhere, start, end).Order(defaultOrder).Count(&count)
		if CheckError(result) != nil {
			return adDataQualityAggregations, 0, result.Error
		}
	}

	return adDataQualityAggregations, int(count), nil
}

func (s *BloodhoundDB) CreateAzureDataQualityStats(stats model.AzureDataQualityStats) (model.AzureDataQualityStats, error) {
	result := s.db.Create(&stats)
	return stats, CheckError(result)
}

func (s *BloodhoundDB) GetAzureDataQualityStats(tenantId string, start time.Time, end time.Time, order string, limit int, skip int) (model.AzureDataQualityStats, int, error) {
	const (
		defaultOrder = "created_at desc"
		defaultWhere = "tenant_id = ? and (created_at between ? and ?)"
	)

	var (
		azureDataQualityStats model.AzureDataQualityStats
		count                 int64
		result                *gorm.DB
	)

	if order != "" {
		result = s.Scope(Paginate(skip, limit)).Where(defaultWhere, tenantId, start, end).Order(order).Find(&azureDataQualityStats)
		if CheckError(result) != nil {
			return azureDataQualityStats, 0, result.Error
		}
		result = s.db.Model(model.AzureDataQualityStats{}).Where(defaultWhere, tenantId, start, end).Order(order).Count(&count)
		if CheckError(result) != nil {
			return azureDataQualityStats, 0, result.Error
		}
	} else {
		result = s.Scope(Paginate(skip, limit)).Where(defaultWhere, tenantId, start, end).Order(defaultOrder).Find(&azureDataQualityStats)
		if CheckError(result) != nil {
			return azureDataQualityStats, 0, result.Error
		}
		result = s.db.Model(model.AzureDataQualityStats{}).Where(defaultWhere, tenantId, start, end).Order(defaultOrder).Count(&count)
		if CheckError(result) != nil {
			return azureDataQualityStats, 0, result.Error
		}
	}

	return azureDataQualityStats, int(count), nil
}

func (s *BloodhoundDB) CreateAzureDataQualityAggregation(aggregation model.AzureDataQualityAggregation) (model.AzureDataQualityAggregation, error) {
	result := s.db.Create(&aggregation)
	return aggregation, CheckError(result)
}

func (s *BloodhoundDB) GetAzureDataQualityAggregations(start time.Time, end time.Time, order string, limit int, skip int) (model.AzureDataQualityAggregations, int, error) {
	const (
		defaultOrder = "created_at desc"
		defaultWhere = "created_at between ? and ?"
	)

	var (
		azureDataQualityAggregations model.AzureDataQualityAggregations
		count                        int64
		result                       *gorm.DB
	)

	if order != "" {
		result = s.Scope(Paginate(skip, limit)).Where(defaultWhere, start, end).Order(order).Find(&azureDataQualityAggregations)
		if CheckError(result) != nil {
			return azureDataQualityAggregations, 0, result.Error
		}
		result = s.db.Model(model.AzureDataQualityAggregations{}).Where(defaultWhere, start, end).Order(order).Count(&count)
		if CheckError(result) != nil {
			return azureDataQualityAggregations, 0, result.Error
		}
	} else {
		result = s.Scope(Paginate(skip, limit)).Where(defaultWhere, start, end).Order(defaultOrder).Find(&azureDataQualityAggregations)
		if CheckError(result) != nil {
			return azureDataQualityAggregations, 0, result.Error
		}
		result = s.db.Model(model.AzureDataQualityAggregations{}).Where(defaultWhere, start, end).Order(defaultOrder).Count(&count)
		if CheckError(result) != nil {
			return azureDataQualityAggregations, 0, result.Error
		}
	}

	return azureDataQualityAggregations, int(count), nil
}
