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

import (
	"context"

	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

func (s *BloodhoundDB) UpdateIngestJob(ctx context.Context, job model.IngestJob) error {
	result := s.db.WithContext(ctx).Save(&job)
	return CheckError(result)
}

func (s *BloodhoundDB) CreateIngestJob(ctx context.Context, job model.IngestJob) (model.IngestJob, error) {
	result := s.db.WithContext(ctx).Create(&job)
	return job, CheckError(result)
}

func (s *BloodhoundDB) GetIngestJob(ctx context.Context, id int64) (model.IngestJob, error) {
	var job model.IngestJob
	if result := s.db.Preload("User").WithContext(ctx).First(&job, id); result.Error != nil {
		return job, CheckError(result)
	} else {
		return job, nil
	}
}

func (s *BloodhoundDB) GetIngestJobsWithStatus(ctx context.Context, status model.JobStatus) ([]model.IngestJob, error) {
	var jobs model.IngestJobs
	result := s.db.WithContext(ctx).Where("status = ?", status).Find(&jobs)

	return jobs, CheckError(result)
}

func (s *BloodhoundDB) CancelAllIngestJobs(ctx context.Context) error {
	runningStates := []model.JobStatus{model.JobStatusAnalyzing, model.JobStatusRunning, model.JobStatusIngesting}
	return CheckError(s.db.Model(model.IngestJob{}).WithContext(ctx).Where("status in ?", runningStates).Update("status", model.JobStatusCanceled))
}

func (s *BloodhoundDB) GetAllIngestJobs(ctx context.Context, skip int, limit int, order string, filter model.SQLFilter) ([]model.IngestJob, int, error) {
	var (
		jobs   []model.IngestJob
		result *gorm.DB
		count  int64
	)

	if order == "" {
		order = "end_time desc"
	}

	if filter.SQLString != "" {
		result = s.db.Model(model.IngestJob{}).WithContext(ctx).Where(filter.SQLString, filter.Params...).Count(&count)
	} else {
		result = s.db.Model(model.IngestJob{}).WithContext(ctx).Count(&count)
	}

	if result.Error != nil {
		return nil, 0, CheckError(result)
	}

	if filter.SQLString != "" {
		result = s.Scope(Paginate(skip, limit)).WithContext(ctx).Preload("User").Where(filter.SQLString, filter.Params...).Order(order).Find(&jobs)
	} else {
		result = s.Scope(Paginate(skip, limit)).WithContext(ctx).Preload("User").Order(order).Find(&jobs)
	}

	if result.Error != nil {
		return nil, int(count), CheckError(result)
	} else {
		for idx := range jobs {
			jobs[idx].UserEmailAddress = jobs[idx].User.EmailAddress
		}
		return jobs, int(count), nil
	}
}

func (s *BloodhoundDB) DeleteAllIngestJobs(ctx context.Context) error {
	return CheckError(
		s.db.WithContext(ctx).Exec("DELETE FROM ingest_jobs"),
	)
}

func (s *BloodhoundDB) DeleteAllIngestTasks(ctx context.Context) error {
	return CheckError(s.db.WithContext(ctx).Exec("DELETE FROM ingest_tasks"))
}
