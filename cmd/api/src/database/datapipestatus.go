// Copyright 2024 Specter Ops, Inc.
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
	"database/sql"
	"time"

	"github.com/specterops/bloodhound/src/model"
)

func (s *BloodhoundDB) SetDatapipeStatus(ctx context.Context, status model.DatapipeStatus, updateAnalysisTime bool) error {

	updateSql := "UPDATE datapipe_status SET status = ?, updated_at = ?"
	now := time.Now().UTC()

	if updateAnalysisTime {
		updateSql += ", last_complete_analysis_at = ?;"
		return s.db.WithContext(ctx).Exec(updateSql, status, now, now).Error
	} else {
		updateSql += ";"
		return s.db.WithContext(ctx).Exec(updateSql, status, now).Error
	}

}

func (s *BloodhoundDB) GetDatapipeStatus(ctx context.Context) (model.DatapipeStatusWrapper, error) {
	var datapipeStatus model.DatapipeStatusWrapper

	if tx := s.db.WithContext(ctx).Raw("SELECT status, updated_at, last_complete_analysis_at FROM datapipe_status LIMIT 1;").Scan(&datapipeStatus); tx.RowsAffected == 0 {
		return datapipeStatus, sql.ErrNoRows
	}

	return datapipeStatus, nil
}
