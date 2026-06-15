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

type OpenGraphDataQualityMetricType string

const (
	OpenGraphDataQualityMetricTypeNode         OpenGraphDataQualityMetricType = "node"
	OpenGraphDataQualityMetricTypeRelationship OpenGraphDataQualityMetricType = "relationship"
)

// OpenGraphDataQualityStat stores one metric for one graph environment ID.
type OpenGraphDataQualityStat struct {
	Serial
	RunID                    string                         `json:"run_id" gorm:"index"`
	SchemaExtensionID        int32                          `json:"schema_extension_id" gorm:"index"`
	SchemaEnvironmentID      int32                          `json:"schema_environment_id" gorm:"index"`
	EnvironmentID            string                         `json:"environment_id" gorm:"index"`
	MetricType               OpenGraphDataQualityMetricType `json:"metric_type" gorm:"index"`
	MetricName               string                         `json:"metric_name" gorm:"index"`
	MetricValue              float64                        `json:"metric_value"`
	SchemaNodeKindID         null.Int32                     `json:"schema_node_kind_id" gorm:"index"`
	SchemaRelationshipKindID null.Int32                     `json:"schema_relationship_kind_id" gorm:"index"`
}

func (OpenGraphDataQualityStat) TableName() string {
	return "data_quality_stats"
}

// OpenGraphDataQualityAggregation stores the same metrics rolled up by schema environment kind.
type OpenGraphDataQualityAggregation struct {
	Serial
	RunID                    string                         `json:"run_id" gorm:"index"`
	SchemaExtensionID        int32                          `json:"schema_extension_id" gorm:"index"`
	SchemaEnvironmentID      int32                          `json:"schema_environment_id" gorm:"index"`
	MetricType               OpenGraphDataQualityMetricType `json:"metric_type" gorm:"index"`
	MetricName               string                         `json:"metric_name" gorm:"index"`
	MetricValue              float64                        `json:"metric_value"`
	SchemaNodeKindID         null.Int32                     `json:"schema_node_kind_id" gorm:"index"`
	SchemaRelationshipKindID null.Int32                     `json:"schema_relationship_kind_id" gorm:"index"`
}

func (OpenGraphDataQualityAggregation) TableName() string {
	return "data_quality_stats_aggregation"
}

type OpenGraphDataQualityStats []OpenGraphDataQualityStat

type OpenGraphDataQualityAggregations []OpenGraphDataQualityAggregation

func (s OpenGraphDataQualityStats) IsSortable(column string) bool {
	switch column {
	case "id",
		"run_id",
		"schema_extension_id",
		"schema_environment_id",
		"environment_id",
		"metric_type",
		"metric_name",
		"metric_value",
		"schema_node_kind_id",
		"schema_relationship_kind_id",
		"updated_at",
		"created_at":
		return true
	default:
		return false
	}
}

func (s OpenGraphDataQualityAggregations) IsSortable(column string) bool {
	switch column {
	case "id",
		"run_id",
		"schema_extension_id",
		"schema_environment_id",
		"metric_type",
		"metric_name",
		"metric_value",
		"schema_node_kind_id",
		"schema_relationship_kind_id",
		"updated_at",
		"created_at":
		return true
	default:
		return false
	}
}
