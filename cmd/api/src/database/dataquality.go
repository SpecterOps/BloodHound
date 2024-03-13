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
	"context"
	"time"

	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

func (s *BloodhoundDB) CreateADDataQualityStats(ctx context.Context, stats model.ADDataQualityStats) (model.ADDataQualityStats, error) {
	result := s.db.WithContext(ctx).Create(&stats)
	return stats, CheckError(result)
}

func (s *BloodhoundDB) GetADDataQualityStats(ctx context.Context, domainSid string, start time.Time, end time.Time, order string, limit int, skip int) (model.ADDataQualityStats, int, error) {
	const (
		defaultWhere = "domain_sid = ? and (created_at between ? and ?)"
	)

	var (
		adDataQualityStats model.ADDataQualityStats
		count              int64
		result             *gorm.DB
	)

	result = s.db.Model(model.ADDataQualityStats{}).WithContext(ctx).Where(defaultWhere, domainSid, start, end).Count(&count)
	if CheckError(result) != nil {
		return adDataQualityStats, 0, result.Error
	}

	if order == "" {
		order = "created_at desc"
	}

	result = s.Scope(Paginate(skip, limit)).WithContext(ctx).Where(defaultWhere, domainSid, start, end).Order(order).Find(&adDataQualityStats)
	if CheckError(result) != nil {
		return adDataQualityStats, 0, result.Error
	}

	return adDataQualityStats, int(count), nil
}

func (s *BloodhoundDB) CreateADDataQualityAggregation(ctx context.Context, aggregation model.ADDataQualityAggregation) (model.ADDataQualityAggregation, error) {
	result := s.db.WithContext(ctx).Create(&aggregation)
	return aggregation, CheckError(result)
}

func (s *BloodhoundDB) GetADDataQualityAggregations(ctx context.Context, start time.Time, end time.Time, order string, limit int, skip int) (model.ADDataQualityAggregations, int, error) {
	const (
		defaultWhere = "created_at between ? and ?"
	)

	var (
		adDataQualityAggregations model.ADDataQualityAggregations
		count                     int64
		result                    *gorm.DB
	)

	result = s.db.Model(model.ADDataQualityAggregations{}).WithContext(ctx).Where(defaultWhere, start, end).Count(&count)
	if CheckError(result) != nil {
		return adDataQualityAggregations, 0, result.Error
	}

	if order == "" {
		order = "created_at desc"
	}

	result = s.Scope(Paginate(skip, limit)).WithContext(ctx).Where(defaultWhere, start, end).Order(order).Find(&adDataQualityAggregations)
	if CheckError(result) != nil {
		return adDataQualityAggregations, 0, result.Error
	}

	return adDataQualityAggregations, int(count), nil
}

func (s *BloodhoundDB) CreateAzureDataQualityStats(ctx context.Context, stats model.AzureDataQualityStats) (model.AzureDataQualityStats, error) {
	result := s.db.WithContext(ctx).Create(&stats)
	return stats, CheckError(result)
}

func (s *BloodhoundDB) GetAzureDataQualityStats(ctx context.Context, tenantId string, start time.Time, end time.Time, order string, limit int, skip int) (model.AzureDataQualityStats, int, error) {
	const (
		defaultWhere = "tenant_id = ? and (created_at between ? and ?)"
	)

	var (
		azureDataQualityStats model.AzureDataQualityStats
		count                 int64
		result                *gorm.DB
	)

	result = s.db.Model(model.AzureDataQualityStats{}).WithContext(ctx).Where(defaultWhere, tenantId, start, end).Count(&count)
	if CheckError(result) != nil {
		return azureDataQualityStats, 0, result.Error
	}

	if order == "" {
		order = "created_at desc"
	}

	result = s.Scope(Paginate(skip, limit)).WithContext(ctx).Where(defaultWhere, tenantId, start, end).Order(order).Find(&azureDataQualityStats)
	if CheckError(result) != nil {
		return azureDataQualityStats, 0, result.Error
	}

	return azureDataQualityStats, int(count), nil
}

func (s *BloodhoundDB) CreateAzureDataQualityAggregation(ctx context.Context, aggregation model.AzureDataQualityAggregation) (model.AzureDataQualityAggregation, error) {
	result := s.db.WithContext(ctx).Create(&aggregation)
	return aggregation, CheckError(result)
}

func (s *BloodhoundDB) GetAzureDataQualityAggregations(ctx context.Context, start time.Time, end time.Time, order string, limit int, skip int) (model.AzureDataQualityAggregations, int, error) {
	const (
		defaultWhere = "created_at between ? and ?"
	)

	var (
		azureDataQualityAggregations model.AzureDataQualityAggregations
		count                        int64
		result                       *gorm.DB
	)

	result = s.db.Model(model.AzureDataQualityAggregations{}).WithContext(ctx).Where(defaultWhere, start, end).Count(&count)
	if CheckError(result) != nil {
		return azureDataQualityAggregations, 0, result.Error
	}

	if order == "" {
		order = "created_at desc"
	}

	result = s.Scope(Paginate(skip, limit)).WithContext(ctx).Where(defaultWhere, start, end).Order(order).Find(&azureDataQualityAggregations)
	if CheckError(result) != nil {
		return azureDataQualityAggregations, 0, result.Error
	}

	return azureDataQualityAggregations, int(count), nil
}

func (s *BloodhoundDB) DeleteAllDataQuality(ctx context.Context) error {
	return CheckError(
		s.db.WithContext(ctx).Exec("DELETE FROM ad_data_quality_aggregations; DELETE FROM ad_data_quality_stats; DELETE FROM azure_data_quality_aggregations; DELETE FROM azure_data_quality_stats;"),
	)
}
