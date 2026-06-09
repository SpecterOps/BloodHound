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
	"errors"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"gorm.io/gorm"
)

type DataQualityData interface {
	CreateADDataQualityStats(ctx context.Context, stats model.ADDataQualityStats) (model.ADDataQualityStats, error)
	CreateADDataQualityAggregation(ctx context.Context, aggregation model.ADDataQualityAggregation) (model.ADDataQualityAggregation, error)
	CreateAzureDataQualityStats(ctx context.Context, stats model.AzureDataQualityStats) (model.AzureDataQualityStats, error)
	CreateAzureDataQualityAggregation(ctx context.Context, aggregation model.AzureDataQualityAggregation) (model.AzureDataQualityAggregation, error)
	CreateDataQualityObjectCountRun(ctx context.Context, run model.DataQualityObjectCountRun) (model.DataQualityObjectCountRun, error)
	CreateDataQualitySourceObjectCounts(ctx context.Context, counts model.DataQualitySourceObjectCounts) (model.DataQualitySourceObjectCounts, error)
	CreateDataQualityEnvironmentObjectCounts(ctx context.Context, counts model.DataQualityEnvironmentObjectCounts) (model.DataQualityEnvironmentObjectCounts, error)
	GetDataQualitySourceObjectCounts(ctx context.Context, start time.Time, end time.Time, filters model.DataQualitySourceObjectCountFilters, order string, limit int, skip int) (model.DataQualitySourceObjectCounts, int, error)
	GetDataQualitySourceObjectCountSummaries(ctx context.Context, start time.Time, end time.Time, filters model.DataQualitySourceObjectCountFilters, order string, limit int, skip int) (model.DataQualitySourceObjectCountSummaries, int, error)
	GetDataQualityEnvironmentObjectCounts(ctx context.Context, start time.Time, end time.Time, filters model.DataQualityEnvironmentObjectCountFilters, order string, limit int, skip int) (model.DataQualityEnvironmentObjectCounts, int, error)
	GetEnvironments(ctx context.Context) ([]model.SchemaEnvironment, error)
	GetKindsByIDs(ctx context.Context, ids ...int32) ([]model.Kind, error)
	GetPrimaryDisplayKinds(ctx context.Context) (graphschema.PrimaryDisplayKinds, error)
	GetSourceKinds(ctx context.Context) ([]model.SourceKind, error)
}

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

func (s *BloodhoundDB) CreateDataQualityObjectCountRun(ctx context.Context, run model.DataQualityObjectCountRun) (model.DataQualityObjectCountRun, error) {
	result := s.db.WithContext(ctx).Create(&run)
	return run, CheckError(result)
}

func (s *BloodhoundDB) CreateDataQualitySourceObjectCounts(ctx context.Context, counts model.DataQualitySourceObjectCounts) (model.DataQualitySourceObjectCounts, error) {
	result := s.db.WithContext(ctx).Create(&counts)
	return counts, CheckError(result)
}

func (s *BloodhoundDB) CreateDataQualityEnvironmentObjectCounts(ctx context.Context, counts model.DataQualityEnvironmentObjectCounts) (model.DataQualityEnvironmentObjectCounts, error) {
	result := s.db.WithContext(ctx).Create(&counts)
	return counts, CheckError(result)
}

func (s *BloodhoundDB) GetDataQualitySourceObjectCounts(ctx context.Context, start time.Time, end time.Time, filters model.DataQualitySourceObjectCountFilters, order string, limit int, skip int) (model.DataQualitySourceObjectCounts, int, error) {
	var (
		count                   int64
		objectCounts            model.DataQualitySourceObjectCounts
		sourceObjectCountsQuery = func() *gorm.DB {
			return applyDataQualitySourceObjectCountFilters(
				s.db.Model(model.DataQualitySourceObjectCount{}).WithContext(ctx).Where("created_at between ? and ?", start, end),
				filters,
			)
		}
		result *gorm.DB
	)

	if filters.Latest {
		if runID, err := s.getLatestDataQualityObjectCountRunID(ctx, start, end); errors.Is(err, ErrNotFound) {
			return objectCounts, 0, nil
		} else if err != nil {
			return objectCounts, 0, err
		} else {
			filters.RunID = runID
		}
	}

	result = sourceObjectCountsQuery().Count(&count)
	if CheckError(result) != nil {
		return objectCounts, 0, result.Error
	}

	if order == "" {
		order = "created_at desc"
	}

	result = sourceObjectCountsQuery().Scopes(Paginate(skip, limit)).Order(order).Find(&objectCounts)
	if CheckError(result) != nil {
		return objectCounts, 0, result.Error
	}

	return objectCounts, int(count), nil
}

func (s *BloodhoundDB) GetDataQualitySourceObjectCountSummaries(ctx context.Context, start time.Time, end time.Time, filters model.DataQualitySourceObjectCountFilters, order string, limit int, skip int) (model.DataQualitySourceObjectCountSummaries, int, error) {
	var (
		count     int64
		params    = dataQualitySourceObjectCountQueryParams(start, end, filters, limit, skip)
		result    *gorm.DB
		summaries model.DataQualitySourceObjectCountSummaries
	)

	if filters.Latest {
		if runID, err := s.getLatestDataQualityObjectCountRunID(ctx, start, end); errors.Is(err, ErrNotFound) {
			return summaries, 0, nil
		} else if err != nil {
			return summaries, 0, err
		} else {
			filters.RunID = runID
			params = dataQualitySourceObjectCountQueryParams(start, end, filters, limit, skip)
		}
	}

	result = s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM data_quality_object_count_runs AS runs
		WHERE runs.created_at BETWEEN @start AND @end
		  AND (@run_id = '' OR runs.run_id = @run_id);
	`, params).Scan(&count)
	if CheckError(result) != nil {
		return summaries, 0, result.Error
	}

	if order == "" {
		order = "created_at desc"
	}

	result = s.db.WithContext(ctx).Raw(`
		SELECT
			runs.run_id,
			COALESCE(SUM(object_counts.count), 0)::bigint AS count,
			runs.created_at
		FROM data_quality_object_count_runs AS runs
		LEFT JOIN data_quality_source_object_counts AS object_counts
		  ON object_counts.run_id = runs.run_id
		 AND (@source_kind = '' OR object_counts.source_kind = @source_kind)
		 AND (@node_kind = '' OR object_counts.node_kind = @node_kind)
		WHERE runs.created_at BETWEEN @start AND @end
		  AND (@run_id = '' OR runs.run_id = @run_id)
		GROUP BY runs.run_id, runs.created_at
		ORDER BY `+order+`
		LIMIT @limit OFFSET @skip;
	`, params).Scan(&summaries)
	if CheckError(result) != nil {
		return summaries, 0, result.Error
	}

	return summaries, int(count), nil
}

func (s *BloodhoundDB) GetDataQualityEnvironmentObjectCounts(ctx context.Context, start time.Time, end time.Time, filters model.DataQualityEnvironmentObjectCountFilters, order string, limit int, skip int) (model.DataQualityEnvironmentObjectCounts, int, error) {
	var (
		count                        int64
		environmentObjectCountsQuery = func() *gorm.DB {
			return applyDataQualityEnvironmentObjectCountFilters(
				s.db.Model(model.DataQualityEnvironmentObjectCount{}).WithContext(ctx).Where("created_at between ? and ?", start, end),
				filters,
			)
		}
		objectCounts model.DataQualityEnvironmentObjectCounts
		result       *gorm.DB
	)

	if filters.Latest {
		if runID, err := s.getLatestDataQualityObjectCountRunID(ctx, start, end); errors.Is(err, ErrNotFound) {
			return objectCounts, 0, nil
		} else if err != nil {
			return objectCounts, 0, err
		} else {
			filters.RunID = runID
		}
	}

	result = environmentObjectCountsQuery().Count(&count)
	if CheckError(result) != nil {
		return objectCounts, 0, result.Error
	}

	if order == "" {
		order = "created_at desc"
	}

	result = environmentObjectCountsQuery().Scopes(Paginate(skip, limit)).Order(order).Find(&objectCounts)
	if CheckError(result) != nil {
		return objectCounts, 0, result.Error
	}

	return objectCounts, int(count), nil
}

func (s *BloodhoundDB) getLatestDataQualityObjectCountRunID(ctx context.Context, start time.Time, end time.Time) (string, error) {
	var (
		run    model.DataQualityObjectCountRun
		result = s.db.WithContext(ctx).Where("created_at between ? and ?", start, end).Order("created_at desc").First(&run)
	)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", ErrNotFound
	}

	if CheckError(result) != nil {
		return "", result.Error
	}

	return run.RunID, nil
}

func dataQualitySourceObjectCountQueryParams(start time.Time, end time.Time, filters model.DataQualitySourceObjectCountFilters, limit int, skip int) map[string]any {
	return map[string]any{
		"start":       start,
		"end":         end,
		"source_kind": filters.SourceKind,
		"node_kind":   filters.NodeKind,
		"run_id":      filters.RunID,
		"limit":       limit,
		"skip":        skip,
	}
}

func applyDataQualitySourceObjectCountFilters(query *gorm.DB, filters model.DataQualitySourceObjectCountFilters) *gorm.DB {
	if filters.SourceKind != "" {
		query = query.Where("source_kind = ?", filters.SourceKind)
	}

	if filters.NodeKind != "" {
		query = query.Where("node_kind = ?", filters.NodeKind)
	}

	if filters.RunID != "" {
		query = query.Where("run_id = ?", filters.RunID)
	}

	return query
}

func applyDataQualityEnvironmentObjectCountFilters(query *gorm.DB, filters model.DataQualityEnvironmentObjectCountFilters) *gorm.DB {
	if filters.SourceKind != "" {
		query = query.Where("source_kind = ?", filters.SourceKind)
	}

	if filters.EnvironmentKind != "" {
		query = query.Where("environment_kind = ?", filters.EnvironmentKind)
	}

	if filters.EnvironmentID != "" {
		query = query.Where("environment_id = ?", filters.EnvironmentID)
	}

	if filters.NodeKind != "" {
		query = query.Where("node_kind = ?", filters.NodeKind)
	}

	if filters.RunID != "" {
		query = query.Where("run_id = ?", filters.RunID)
	}

	return query
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
	if err := CheckError(
		s.db.WithContext(ctx).Exec("DELETE FROM data_quality_environment_object_counts; DELETE FROM data_quality_source_object_counts; DELETE FROM data_quality_object_count_runs;"),
	); err != nil {
		return err
	}

	return CheckError(
		s.db.WithContext(ctx).Exec("DELETE FROM ad_data_quality_aggregations; DELETE FROM ad_data_quality_stats; DELETE FROM azure_data_quality_aggregations; DELETE FROM azure_data_quality_stats; DELETE FROM data_quality_object_counts;"),
	)
}
