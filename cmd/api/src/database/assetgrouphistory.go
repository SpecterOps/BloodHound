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
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// AssetGroupHistoryData defines the methods required to interact with the asset_group_history table
type AssetGroupHistoryData interface {
	CreateAssetGroupHistoryRecord(ctx context.Context, actorId, email string, target string, action model.AssetGroupHistoryAction, assetGroupTagId int, environmentId, note null.String) error
	GetAssetGroupHistoryRecords(ctx context.Context, sqlFilter model.SQLFilter, sortItems model.Sort, skip, limit int) ([]model.AssetGroupHistory, int, error)
}

func (s *BloodhoundDB) CreateAssetGroupHistoryRecord(ctx context.Context, actorId, emailAddress string, target string, action model.AssetGroupHistoryAction, assetGroupTagId int, environmentId, note null.String) error {
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("INSERT INTO %s (actor, email, target, action, asset_group_tag_id, environment_id, note, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, NOW())", (model.AssetGroupHistory{}).TableName()),
		actorId, emailAddress, target, action, assetGroupTagId, environmentId, note))
}

func (s *BloodhoundDB) GetAssetGroupHistoryRecords(ctx context.Context, sqlFilter model.SQLFilter, sortItems model.Sort, skip, limit int) ([]model.AssetGroupHistory, int, error) {
	var (
		historyRecs     []model.AssetGroupHistory
		skipLimitString string
		rowCount        int
		sortColumns     []string
	)

	if sqlFilter.SQLString != "" {
		sqlFilter.SQLString = " WHERE " + sqlFilter.SQLString
	}

	for _, item := range sortItems {
		dirString := "ASC"
		if item.Direction == model.DescendingSortDirection {
			dirString = "DESC"
		}
		sortColumns = append(sortColumns, fmt.Sprintf("%s %s", item.Column, dirString))
	}
	sortString := "ORDER BY " + strings.Join(sortColumns, ", ")

	if limit > 0 {
		skipLimitString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		skipLimitString += fmt.Sprintf(" OFFSET %d", skip)
	}

	sqlStr := fmt.Sprintf(
		"SELECT id, actor, email, target, action, asset_group_tag_id, environment_id, note, created_at FROM %s%s %s %s",
		(model.AssetGroupHistory{}).TableName(),
		sqlFilter.SQLString,
		sortString,
		skipLimitString)

	sqlCountStr := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s%s",
		(model.AssetGroupHistory{}).TableName(),
		sqlFilter.SQLString)

	if result := s.db.WithContext(ctx).Raw(sqlStr, sqlFilter.Params...).Find(&historyRecs); result.Error != nil {
		return []model.AssetGroupHistory{{}}, 0, CheckError(result)
	}

	if limit > 0 || skip > 0 {
		if result := s.db.WithContext(ctx).Raw(sqlCountStr, sqlFilter.Params...).Scan(&rowCount); result.Error != nil {
			return []model.AssetGroupHistory{{}}, 0, CheckError(result)
		}
	} else {
		rowCount = len(historyRecs)
	}

	return historyRecs, rowCount, nil
}
