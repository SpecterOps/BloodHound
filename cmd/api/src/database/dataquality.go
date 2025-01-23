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

// GetAggregateADDataQualityStats will aggregate AD Quality stats by
// summing the maximum asset counts per environment per day. Due to
// session and group completeness being percentages, it will return
// the single maximum value of all environments per day.
func (s *BloodhoundDB) GetAggregateADDataQualityStats(ctx context.Context, domainSIDs []string, start time.Time, end time.Time) (model.ADDataQualityStats, error) {
	var (
		adDataQualityStats model.ADDataQualityStats
		params             = map[string]any{
			"ids":   domainSIDs,
			"start": start,
			"end":   end,
		}
	)

	const aggregateAdDataQualityStatsSql = `
WITH aggregated_quality_stats AS (
    SELECT
        DATE_TRUNC('day', created_at) AS created_date,
        MAX(users) AS max_users,
        MAX(groups) AS max_groups,
        MAX(computers) AS max_computers,
        MAX(ous) AS max_ous,
        MAX(containers) AS max_containers,
        MAX(gpos) AS max_gpos,
        MAX(acls) AS max_acls,
        MAX(sessions) AS max_sessions,
        MAX(relationships) AS max_relationships,
        MAX(aiacas) AS max_aiacas,
        MAX(rootcas) AS max_rootcas,
        MAX(enterprisecas) AS max_enterprisecas,
        MAX(ntauthstores) AS max_ntauthstores,
        MAX(certtemplates) AS max_certtemplates,
        MAX(issuancepolicies) AS max_issuancepolicies,
<<<<<<< HEAD
=======
        -- these fields appear to be a mistake
        -- MAX(o_us) AS max_o_us,
        -- MAX(ac_ls) AS max_ac_ls,
        -- MAX(gp_os) AS max_gp_os,
>>>>>>> b32ec4e0 (feat: BED-5132 - aggregate posture and risk stats)
        MAX(session_completeness) AS max_session_completeness,
        MAX(local_group_completeness) AS max_local_group_completeness
    FROM ad_data_quality_stats
    WHERE domain_sid IN @ids
    AND created_at BETWEEN @start AND @end
    GROUP BY domain_sid, created_date
	)
SELECT
    created_date AS created_at,
    SUM(max_users) AS users,
    SUM(max_groups) AS groups,
    SUM(max_computers) AS computers,
    SUM(max_ous) AS ous,
    SUM(max_containers) AS containers,
    SUM(max_gpos) AS gpos,
    SUM(max_acls) AS acls,
    SUM(max_sessions) AS sessions,
    SUM(max_relationships) AS relationships,
    SUM(max_aiacas) AS aiacas,
    SUM(max_rootcas) AS rootcas,
    SUM(max_enterprisecas) AS enterprisecas,
    SUM(max_ntauthstores) AS ntauthstores,
    SUM(max_certtemplates) AS certtemplates,
    SUM(max_issuancepolicies) AS issuancepolicies,
    -- these fields appear to be a mistake
    -- SUM(max_o_us) AS o_us,
    -- SUM(max_ac_ls) AS ac_ls,
    -- SUM(max_gp_os) AS gp_os,
    MAX(max_session_completeness) AS session_completeness,
    MAX(max_local_group_completeness) AS local_group_completeness
FROM aggregated_quality_stats
GROUP BY created_at
ORDER BY created_at;`

	result := s.db.WithContext(ctx).Raw(aggregateAdDataQualityStatsSql, params).Scan(&adDataQualityStats)

	return adDataQualityStats, CheckError(result)
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
