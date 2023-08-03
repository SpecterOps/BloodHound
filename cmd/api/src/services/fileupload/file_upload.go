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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/mock.go -package=mocks . FileUploadData
package fileupload

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
)

const (
	jobActivityTimeout = time.Minute * 20
)

type FileUploadData interface {
	CreateFileUploadJob(job model.FileUploadJob) (model.FileUploadJob, error)
	UpdateFileUploadJob(job model.FileUploadJob) error
	GetFileUploadJob(id int64) (model.FileUploadJob, error)
	GetAllFileUploadJobs(skip int, limit int, order string, filter model.SQLFilter) ([]model.FileUploadJob, int, error)
	GetFileUploadJobsWithStatus(status model.JobStatus) ([]model.FileUploadJob, error)
}

func ProcessStaleFileUploadJobs(db FileUploadData) {
	var (
		now       = time.Now().UTC()
		threshold = now.Add(-jobActivityTimeout)
	)

	if jobs, err := db.GetFileUploadJobsWithStatus(model.JobStatusRunning); err != nil {
		log.Errorf("Error getting running jobs: %v", err)
	} else {
		for _, job := range jobs {
			if job.LastIngest.Before(threshold) {
				log.Warnf("Ingest timeout: No ingest activity observed for Job ID %d in %f minutes (last ingest was %s). Upload incomplete",
					job.ID,
					now.Sub(threshold).Minutes(),
					job.LastIngest.Format(time.RFC3339))

				if err := TimeOutUploadJob(db, job.ID, fmt.Sprintf("Ingest timeout: No ingest activity observed in %f minutes. Upload incomplete.", now.Sub(threshold).Minutes())); err != nil {
					log.Errorf("Error marking file upload job %d as timed out: %v", job.ID, err)
				}
			}
		}
	}
}

func GetAllFileUploadJobs(db FileUploadData, skip int, limit int, order string, filter model.SQLFilter) ([]model.FileUploadJob, int, error) {
	return db.GetAllFileUploadJobs(skip, limit, order, filter)
}

func StartFileUploadJob(db FileUploadData, user model.User) (model.FileUploadJob, error) {
	job := model.FileUploadJob{
		UserID:     user.ID,
		User:       user,
		Status:     model.JobStatusRunning,
		StartTime:  time.Now().UTC(),
		LastIngest: time.Now().UTC(),
	}
	return db.CreateFileUploadJob(job)
}

func GetFileUploadJobByID(db FileUploadData, jobID int64) (model.FileUploadJob, error) {
	return db.GetFileUploadJob(jobID)
}

func SaveToTempFile(location string, fileData io.ReadCloser) (string, error) {
	if tempFile, err := os.CreateTemp(location, "bh"); err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	} else if _, err := io.Copy(tempFile, fileData); err != nil {
		return "", fmt.Errorf("error copying temp file: %w", err)
	} else {
		tempFile.Close()
		return tempFile.Name(), nil
	}
}

func TouchFileUploadJobLastIngest(db FileUploadData, fileUploadJob model.FileUploadJob) error {
	fileUploadJob.LastIngest = time.Now().UTC()
	return db.UpdateFileUploadJob(fileUploadJob)
}

func EndFileUploadJob(db FileUploadData, job model.FileUploadJob) (model.FileUploadJob, error) {
	job.Status = model.JobStatusIngesting
	if err := db.UpdateFileUploadJob(job); err != nil {
		return job, fmt.Errorf("error ending file upload job: %w", err)
	} else {
		return job, nil
	}
}

func UpdateFileUploadJobStatus(db FileUploadData, jobID int64, status model.JobStatus, message string) error {
	if job, err := db.GetFileUploadJob(jobID); err != nil {
		return err
	} else {
		job.Status = status
		job.StatusMessage = message
		job.EndTime = time.Now().UTC()

		return db.UpdateFileUploadJob(job)
	}
}

func CompleteFileUploadJob(db FileUploadData, jobID int64) (model.FileUploadJob, error) {
	return model.FileUploadJob{}, nil
}

func TimeOutUploadJob(db FileUploadData, jobID int64, message string) error {
	if job, err := db.GetFileUploadJob(jobID); err != nil {
		return err
	} else {
		job.Status = model.JobStatusTimedOut
		job.StatusMessage = message
		job.EndTime = time.Now().UTC()

		return db.UpdateFileUploadJob(job)
	}
}

func FailFileUploadJob(db FileUploadData, jobID int64, message string) error {
	if job, err := db.GetFileUploadJob(jobID); err != nil {
		return err
	} else {
		job.Status = model.JobStatusFailed
		job.StatusMessage = message
		job.EndTime = time.Now().UTC()

		return db.UpdateFileUploadJob(job)
	}
}
