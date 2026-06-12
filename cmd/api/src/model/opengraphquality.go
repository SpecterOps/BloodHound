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

type OpenGraphDataQualityStat struct {
	Serial
	RunID               string `json:"run_id" gorm:"index"`
	SchemaExtensionID   int32  `json:"schema_extension_id" gorm:"index"`
	SchemaEnvironmentID int32  `json:"schema_environment_id" gorm:"index"`
	EnvironmentID       string `json:"environment_id" gorm:"index"`
	SchemaNodeKindID    int32  `json:"schema_node_kind_id" gorm:"index"`
	NodeKind            string `json:"node_kind"`
	NodeCount           int    `json:"node_count"`
}

func (OpenGraphDataQualityStat) TableName() string {
	return "data_quality_stats"
}

type OpenGraphDataQualityAggregation struct {
	Serial
	RunID             string `json:"run_id" gorm:"index"`
	SchemaExtensionID int32  `json:"schema_extension_id" gorm:"index"`
	SchemaNodeKindID  int32  `json:"schema_node_kind_id" gorm:"index"`
	NodeKind          string `json:"node_kind"`
	NodeCount         int    `json:"node_count"`
}

func (OpenGraphDataQualityAggregation) TableName() string {
	return "data_quality_stats_aggregation"
}

type OpenGraphDataQualityStats []OpenGraphDataQualityStat

type OpenGraphDataQualityAggregations []OpenGraphDataQualityAggregation
