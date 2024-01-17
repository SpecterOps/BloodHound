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
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

func (s *BloodhoundDB) UpdateFileUploadJob(job model.FileUploadJob) error {
	result := s.db.Save(&job)
	return CheckError(result)
}

func (s *BloodhoundDB) CreateFileUploadJob(job model.FileUploadJob) (model.FileUploadJob, error) {
	result := s.db.Create(&job)
	return job, CheckError(result)
}

func (s *BloodhoundDB) GetFileUploadJob(id int64) (model.FileUploadJob, error) {
	var job model.FileUploadJob
	if result := s.db.Preload("User").First(&job, id); result.Error != nil {
		return job, CheckError(result)
	} else {
		return job, nil
	}
}

func (s *BloodhoundDB) GetFileUploadJobsWithStatus(status model.JobStatus) ([]model.FileUploadJob, error) {
	var jobs model.FileUploadJobs
	result := s.db.Where("status = ?", status).Find(&jobs)

	return jobs, CheckError(result)
}

func (s *BloodhoundDB) GetAllFileUploadJobs(skip int, limit int, order string, filter model.SQLFilter) ([]model.FileUploadJob, int, error) {
	var (
		jobs   []model.FileUploadJob
		result *gorm.DB
		count  int64
	)

	if order == "" {
		order = "end_time desc"
	}

	if filter.SQLString != "" {
		result = s.db.Model(model.FileUploadJob{}).Where(filter.SQLString, filter.Params).Count(&count)
	} else {
		result = s.db.Model(model.FileUploadJob{}).Count(&count)
	}

	if result.Error != nil {
		return nil, 0, CheckError(result)
	}

	if filter.SQLString != "" {
		result = s.Scope(Paginate(skip, limit)).Preload("User").Where(filter.SQLString, filter.Params).Order(order).Find(&jobs)
	} else {
		result = s.Scope(Paginate(skip, limit)).Preload("User").Order(order).Find(&jobs)
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
