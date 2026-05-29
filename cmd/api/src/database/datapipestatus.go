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
	"fmt"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type DatapipeStatusData interface {
	SetLastAnalysisStartTime(ctx context.Context) error
	UpdateLastAnalysisCompleteTime(ctx context.Context) error
	SetDatapipeStatus(ctx context.Context, status model.DatapipeStatus) error
	GetDatapipeStatus(ctx context.Context) (model.DatapipeStatusWrapper, error)
	SetNextScheduledAnalysisStartTime(ctx context.Context, time time.Time) error
}

// This should be called at the start of analysis processing (not every datapipe tick, but start of real work)
func (s *BloodhoundDB) SetLastAnalysisStartTime(ctx context.Context) error {
	var datapipeStatus model.DatapipeStatus
	return s.db.WithContext(ctx).Exec(fmt.Sprintf(`UPDATE %s SET updated_at = current_timestamp, last_analysis_run_at = current_timestamp`, datapipeStatus.TableName())).Error
}

// This should be called at the end of a successful analysis run (not always every analysis)
func (s *BloodhoundDB) UpdateLastAnalysisCompleteTime(ctx context.Context) error {
	var datapipeStatus model.DatapipeStatus
	return s.db.WithContext(ctx).Exec(fmt.Sprintf("UPDATE %s SET updated_at = current_timestamp, last_complete_analysis_at = current_timestamp", datapipeStatus.TableName())).Error
}

func (s *BloodhoundDB) SetDatapipeStatus(ctx context.Context, status model.DatapipeStatus) error {
	var datapipeStatus model.DatapipeStatus
	return s.db.WithContext(ctx).Exec(fmt.Sprintf("UPDATE %s SET status = ?, updated_at = current_timestamp", datapipeStatus.TableName()), status).Error
}

func (s *BloodhoundDB) GetDatapipeStatus(ctx context.Context) (model.DatapipeStatusWrapper, error) {
	var (
		datapipeStatusWrapper model.DatapipeStatusWrapper
		datapipeStatus        model.DatapipeStatus
	)

	tx := s.db.WithContext(ctx).Select("status, updated_at, last_complete_analysis_at, last_analysis_run_at, next_scheduled_analysis_at").Table(datapipeStatus.TableName()).First(&datapipeStatusWrapper)

	return datapipeStatusWrapper, CheckError(tx)
}

// No-op in BHCE
func (s *BloodhoundDB) SetNextScheduledAnalysisStartTime(ctx context.Context, time time.Time) error {
	return nil
}
