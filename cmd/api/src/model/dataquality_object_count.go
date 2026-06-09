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

import "time"

type DataQualityObjectCountRun struct {
	RunID     string    `json:"run_id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s DataQualityObjectCountRun) TableName() string {
	return "data_quality_object_count_runs"
}

type DataQualitySourceObjectCount struct {
	SourceKind string `json:"source_kind" gorm:"index:idx_dq_source_obj_counts_source_node_kind"`
	NodeKind   string `json:"node_kind" gorm:"index:idx_dq_source_obj_counts_source_node_kind"`
	Count      int64  `json:"count"`
	RunID      string `json:"run_id" gorm:"index"`

	Serial
}

type DataQualitySourceObjectCounts []DataQualitySourceObjectCount

type DataQualitySourceObjectCountFilters struct {
	SourceKind string
	NodeKind   string
	RunID      string
	Latest     bool
}

type DataQualitySourceObjectCountSummary struct {
	RunID     string    `json:"run_id"`
	Count     int64     `json:"count"`
	CreatedAt time.Time `json:"created_at"`
}

type DataQualitySourceObjectCountSummaries []DataQualitySourceObjectCountSummary

func (s DataQualitySourceObjectCountSummaries) IsSortable(column string) bool {
	switch column {
	case "count",
		"run_id",
		"created_at":
		return true
	default:
		return false
	}
}

func (s DataQualitySourceObjectCounts) IsSortable(column string) bool {
	switch column {
	case "source_kind",
		"node_kind",
		"count",
		"run_id",
		"updated_at",
		"created_at":
		return true
	default:
		return false
	}
}

type DataQualityEnvironmentObjectCount struct {
	SourceKind      string `json:"source_kind" gorm:"index:idx_dq_env_obj_counts_source_env_node_kind"`
	EnvironmentKind string `json:"environment_kind" gorm:"index:idx_dq_env_obj_counts_source_env_node_kind"`
	EnvironmentID   string `json:"environment_id" gorm:"index:idx_dq_env_obj_counts_source_env_node_kind"`
	NodeKind        string `json:"node_kind" gorm:"index:idx_dq_env_obj_counts_source_env_node_kind"`
	Count           int64  `json:"count"`
	RunID           string `json:"run_id" gorm:"index"`

	Serial
}

type DataQualityEnvironmentObjectCounts []DataQualityEnvironmentObjectCount

type DataQualityEnvironmentObjectCountFilters struct {
	SourceKind      string
	EnvironmentKind string
	EnvironmentID   string
	NodeKind        string
	RunID           string
	Latest          bool
}

func (s DataQualityEnvironmentObjectCounts) IsSortable(column string) bool {
	switch column {
	case "source_kind",
		"environment_kind",
		"environment_id",
		"node_kind",
		"count",
		"run_id",
		"updated_at",
		"created_at":
		return true
	default:
		return false
	}
}
