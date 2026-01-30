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

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type DatapipeStatusData interface {
	UpdateLastAnalysisCompleteTime(ctx context.Context) error
	SetDatapipeStatus(ctx context.Context, status model.DatapipeStatus) error
	GetDatapipeStatus(ctx context.Context) (model.DatapipeStatusWrapper, error)
}

// This should be called at the end of a successful analysis run (not always every analysis)
func (s *BloodhoundDB) UpdateLastAnalysisCompleteTime(ctx context.Context) error {
	now := time.Now().UTC()
	return s.db.WithContext(ctx).Exec("UPDATE datapipe_status SET updated_at = ?, last_complete_analysis_at = ?", now, now).Error
}

func (s *BloodhoundDB) SetDatapipeStatus(ctx context.Context, status model.DatapipeStatus) error {
	now := time.Now().UTC()
	return s.db.WithContext(ctx).Exec("UPDATE datapipe_status SET status = ?, updated_at = ?;", status, now).Error
}

func (s *BloodhoundDB) GetDatapipeStatus(ctx context.Context) (model.DatapipeStatusWrapper, error) {
	var datapipeStatus model.DatapipeStatusWrapper

	tx := s.db.WithContext(ctx).Select("status, updated_at, last_complete_analysis_at, last_analysis_run_at").Table("datapipe_status").First(&datapipeStatus)

	return datapipeStatus, CheckError(tx)
}
