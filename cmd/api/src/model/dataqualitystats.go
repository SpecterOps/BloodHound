// Copyright 2026 Specter Ops, Inc.
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

package model

import "github.com/specterops/bloodhound/cmd/api/src/database/types/null"

type DataQualityMetricType string

const (
	DataQualityMetricTypeNode         DataQualityMetricType = "node"
	DataQualityMetricTypeRelationship DataQualityMetricType = "relationship"
)

type DataQualityStat struct {
	Serial
	RunID                   string                `json:"run_id"`
	SchemaExtensionID       int32                 `json:"extension_id"`
	SchemaEnvironmentKindID int32                 `json:"environment_kind_id"`
	EnvironmentID           string                `json:"environment_id"`
	MetricType              DataQualityMetricType `json:"metric_type"`
	MetricName              string                `json:"metric_name"`
	MetricValue             float64               `json:"metric_value"`
	KindID                  null.Int32            `json:"kind_id"`
}

func (DataQualityStat) TableName() string {
	return "data_quality_stats"
}

type DataQualityStats []DataQualityStat

func (s DataQualityStats) IsSortable(column string) bool {
	switch column {
	case
		"updated_at",
		"created_at":
		return true
	default:
		return false
	}
}

// DataQualityAggregation represent a single aggregated data quality metric for a custom OpenGraph schema,
// combined across environment instances for one collection run.
type DataQualityAggregation struct {
	RunID                   string                `json:"run_id"`
	SchemaExtensionID       int32                 `json:"schema_extension_id"`
	SchemaEnvironmentKindID int32                 `json:"schema_environment_kind_id"`
	MetricType              DataQualityMetricType `json:"metric_type"`
	MetricName              string                `json:"metric_name"`
	MetricValue             float64               `json:"metric_value"`
	KindID                  null.Int32            `json:"kind_id"`

	Serial
}

func (s DataQualityAggregation) TableName() string {
	return "data_quality_aggregations"
}

type DataQualityAggregations []DataQualityAggregation

func (s DataQualityAggregations) IsSortable(column string) bool {
	switch column {
	case "run_id",
		"schema_extension_id",
		"schema_environment_kind_id",
		"metric_type",
		"metric_name",
		"metric_value",
		"kind_id",
		"id",
		"created_at",
		"updated_at",
		"deleted_at":
		return true
	default:
		return false
	}
}
