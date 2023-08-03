// Copyright 2023 Specter Ops, Inc.
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

import "github.com/specterops/bloodhound/src/model"

func (s *BloodhoundDB) CreateIngestTask(ingestTask model.IngestTask) (model.IngestTask, error) {
	result := s.db.Create(&ingestTask)

	return ingestTask, CheckError(result)
}

func (s *BloodhoundDB) GetAllIngestTasks() (model.IngestTasks, error) {
	var ingestTasks model.IngestTasks
	result := s.db.Find(&ingestTasks)

	return ingestTasks, CheckError(result)
}

func (s *BloodhoundDB) DeleteIngestTask(ingestTask model.IngestTask) error {
	result := s.db.Delete(&ingestTask)
	return CheckError(result)
}

func (s *BloodhoundDB) GetIngestTasksForJob(jobID int64) (model.IngestTasks, error) {
	var ingestTasks model.IngestTasks
	result := s.db.Where("task_id=?", jobID).Find(&ingestTasks)

	return ingestTasks, CheckError(result)
}

func (s *BloodhoundDB) GetUnfinishedIngestIDs() ([]int64, error) {
	var ids []int64
	result := s.db.Model(&model.IngestTask{}).Distinct("task_id").Pluck("task_id", &ids)

	return ids, CheckError(result)
}
