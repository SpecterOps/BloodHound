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
	"time"

	"github.com/specterops/bloodhound/src/model"
)

type DatapipeStatusData interface {
	SetDatapipeStatus(ctx context.Context, status model.DatapipeStatus, updateAnalysisTime bool) error
	GetDatapipeStatus(ctx context.Context) (model.DatapipeStatusWrapper, error)
}

func (s *BloodhoundDB) SetDatapipeStatus(ctx context.Context, status model.DatapipeStatus, updateAnalysisTime bool) error {
	now := time.Now().UTC()
	// All queries will update the status and table update time
	updateSql := "UPDATE datapipe_status SET status = ?, updated_at = ?"

	if status == model.DatapipeStatusAnalyzing {
		// Updates last run anytime we start analysis
		updateSql += ", last_analysis_run_at = ?;"
		return s.db.WithContext(ctx).Exec(updateSql, status, now, now).Error
	} else if updateAnalysisTime {
		// Updates last completed when analysis is set to complete
		updateSql += ", last_complete_analysis_at = ?;"
		return s.db.WithContext(ctx).Exec(updateSql, status, now, now).Error
	} else {
		// Otherwise, only update status and last update to the table
		updateSql += ";"
		return s.db.WithContext(ctx).Exec(updateSql, status, now).Error
	}
}

func (s *BloodhoundDB) GetDatapipeStatus(ctx context.Context) (model.DatapipeStatusWrapper, error) {
	var datapipeStatus model.DatapipeStatusWrapper

	tx := s.db.WithContext(ctx).Select("status, updated_at, last_complete_analysis_at, last_analysis_run_at").Table("datapipe_status").First(&datapipeStatus)

	return datapipeStatus, CheckError(tx)
}
