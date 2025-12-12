// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

const (
	completedTasksTable = "completed_tasks"
)

func (s *BloodhoundDB) CreateCompletedTask(ctx context.Context, task model.CompletedTask) (model.CompletedTask, error) {
	result := s.db.WithContext(ctx).Raw(
		fmt.Sprintf("INSERT INTO %s (ingest_job_id, file_name, parent_file_name,  errors, warnings, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW()) RETURNING id;", completedTasksTable),
		task.IngestJobId, task.FileName, task.ParentFileName, task.Errors, task.Warnings).Scan(&task.ID)
	return task, CheckError(result)
}

func (s *BloodhoundDB) GetCompletedTasks(ctx context.Context, ingestJobId int64) ([]model.CompletedTask, error) {
	var completedTasks []model.CompletedTask
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, ingest_job_id, file_name, parent_file_name, errors, warnings, created_at, updated_at FROM %s WHERE ingest_job_id = ?;", completedTasksTable), ingestJobId).Scan(&completedTasks)

	return completedTasks, CheckError(result)
}
