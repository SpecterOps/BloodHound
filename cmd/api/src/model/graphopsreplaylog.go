// Copyright 2025 Specter Ops, Inc.
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

import (
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
)

// ChangeType represents the type of operation performed on a graph object
type ChangeType string

const (
	ChangeTypeCreate        ChangeType = "create"
	ChangeTypeUpdate        ChangeType = "update" // Reserved for future use
	ChangeTypeDelete        ChangeType = "delete"
	ChangeTypeAnalysisStart ChangeType = "analysis_start"
	ChangeTypeAnalysisEnd   ChangeType = "analysis_end"
)

// ObjectType represents the type of graph object being modified
type ObjectType string

const (
	ObjectTypeNode     ObjectType = "node"
	ObjectTypeEdge     ObjectType = "edge"
	ObjectTypeAnalysis ObjectType = "analysis"
)

// GraphOperationReplayLogEntry represents a single entry in the graph operations replay log.
// This tracks all modifications (creates, updates, deletes) to nodes and edges in the graph
// database for replay and audit purposes.
type GraphOperationReplayLogEntry struct {
	Serial
	ChangeType     ChangeType  `json:"change_type" gorm:"column:change_type"`
	ObjectType     ObjectType  `json:"object_type" gorm:"column:object_type"`
	ObjectID       string      `json:"object_id" gorm:"column:object_id"`
	Labels         []byte      `json:"labels,omitempty" gorm:"type:jsonb;column:labels"`
	SourceObjectID null.String `json:"source_object_id,omitempty" gorm:"column:source_object_id"`
	TargetObjectID null.String `json:"target_object_id,omitempty" gorm:"column:target_object_id"`
	Properties     []byte      `json:"properties" gorm:"type:jsonb;column:properties"`
	RolledBackAt   null.Time   `json:"rolled_back_at,omitempty" gorm:"column:rolled_back_at"`
}

func (GraphOperationReplayLogEntry) TableName() string {
	return "graph_operations_replay_log"
}
