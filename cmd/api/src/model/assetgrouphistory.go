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
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
)

type AssetGroupHistoryAction string

const (
	AssetGroupHistoryActionCreateTag AssetGroupHistoryAction = "CreateTag"
	AssetGroupHistoryActionUpdateTag AssetGroupHistoryAction = "UpdateTag"
	AssetGroupHistoryActionDeleteTag AssetGroupHistoryAction = "DeleteTag"

	AssetGroupHistoryActionAnalysisEnabledTag  AssetGroupHistoryAction = "AnalysisEnabledTag"
	AssetGroupHistoryActionAnalysisDisabledTag AssetGroupHistoryAction = "AnalysisDisabledTag"

	AssetGroupHistoryActionCreateSelector AssetGroupHistoryAction = "CreateSelector"
	AssetGroupHistoryActionUpdateSelector AssetGroupHistoryAction = "UpdateSelector"
	AssetGroupHistoryActionDeleteSelector AssetGroupHistoryAction = "DeleteSelector"
)

// AssetGroupHistory is the record of CRUD changes associated with v2 of the asset groups feature
type AssetGroupHistory struct {
	ID              int64                   `json:"id" gorm:"primaryKey"`
	CreatedAt       time.Time               `json:"created_at"`
	Actor           string                  `json:"actor"`
	Email           null.String             `json:"email"`
	Action          AssetGroupHistoryAction `json:"action"`
	Target          string                  `json:"target"`
	AssetGroupTagId int                     `json:"asset_group_tag_id"`
	EnvironmentId   null.String             `json:"environment_id"`
	Note            null.String             `json:"note"`
}

func (AssetGroupHistory) TableName() string {
	return "asset_group_history"
}

func (s AssetGroupHistory) AuditData() AuditData {
	return AuditData{
		"id":                 s.ID,
		"created_at":         s.CreatedAt,
		"actor":              s.Actor,
		"email":              s.Email,
		"action":             s.Action,
		"target":             s.Target,
		"asset_group_tag_id": s.AssetGroupTagId,
		"environment_id":     s.EnvironmentId,
		"note":               s.Note,
	}
}
