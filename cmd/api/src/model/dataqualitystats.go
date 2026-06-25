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
	SchemaExtensionID       int32                 `json:"schema_extension_id"`
	SchemaEnvironmentKindID int32                 `json:"schema_environment_kind_id"`
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
	case "id",
		"run_id",
		"schema_extension_id",
		"schema_environment_kind_id",
		"environment_id",
		"metric_type",
		"metric_name",
		"metric_value",
		"kind_id",
		"updated_at",
		"created_at":
		return true
	default:
		return false
	}
}

type DataQualityEnvironmentSelector struct {
	Type                       string `json:"type"`
	Name                       string `json:"name"`
	ObjectID                   string `json:"id"`
	Collected                  bool   `json:"collected"`
	SchemaExtensionID          int32  `json:"schema_extension_id"`
	SchemaExtensionDisplayName string `json:"schema_extension_display_name"`
	IsBuiltin                  bool   `json:"is_builtin"`
	EnvironmentKindID          int32  `json:"environment_kind_id"`
	EnvironmentKind            string `json:"environment_kind"`
	SourceKindID               int32  `json:"source_kind_id"`
	SourceKind                 string `json:"source_kind"`
}

type DataQualityEnvironmentSelectors []DataQualityEnvironmentSelector

func (s DataQualityEnvironmentSelectors) IsSortable(column string) bool {
	switch column {
	case "objectid",
		"name":
		return true
	default:
		return false
	}
}

type DataQualityNodeKindStat struct {
	Serial
	RunID                      string                `json:"run_id"`
	SchemaExtensionID          int32                 `json:"schema_extension_id"`
	SchemaExtensionDisplayName string                `json:"schema_extension_display_name"`
	IsBuiltin                  bool                  `json:"is_builtin"`
	SchemaEnvironmentKindID    int32                 `json:"schema_environment_kind_id"`
	EnvironmentKind            string                `json:"environment_kind"`
	SourceKind                 string                `json:"source_kind"`
	EnvironmentID              string                `json:"environment_id"`
	MetricType                 DataQualityMetricType `json:"metric_type"`
	MetricName                 string                `json:"metric_name"`
	MetricValue                float64               `json:"metric_value"`
	KindID                     null.Int32            `json:"kind_id"`
	KindName                   string                `json:"kind_name"`
}

type DataQualityNodeKindStats []DataQualityNodeKindStat
