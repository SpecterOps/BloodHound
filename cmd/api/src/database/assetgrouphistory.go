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

package database

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
)

// AssetGroupHistoryData defines the methods required to interact with the asset_group_history table
type AssetGroupHistoryData interface {
	CreateAssetGroupHistoryRecord(ctx context.Context, actor, target string, action model.AssetGroupHistoryAction, assetGroupTagId int, environmentId, note null.String) error
	GetAssetGroupHistoryRecords(ctx context.Context) ([]model.AssetGroupHistory, error)
}

func (s *BloodhoundDB) CreateAssetGroupHistoryRecord(ctx context.Context, actor, target string, action model.AssetGroupHistoryAction, assetGroupTagId int, environmentId, note null.String) error {
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("INSERT INTO %s (actor, target, action, asset_group_tag_id, environment_id, note, created_at) VALUES (?, ?, ?, ?, ?, ?, NOW())", (model.AssetGroupHistory{}).TableName()),
		actor, target, action, assetGroupTagId, environmentId, note))
}

func (s *BloodhoundDB) GetAssetGroupHistoryRecords(ctx context.Context) ([]model.AssetGroupHistory, error) {
	var result []model.AssetGroupHistory
	return result, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, actor, target, action, asset_group_tag_id, environment_id, note, created_at FROM %s", (model.AssetGroupHistory{}).TableName())).Find(&result))
}
