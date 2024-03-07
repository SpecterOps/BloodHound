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
	"context"
	"errors"
	"fmt"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
)

const jobActivityTimeout = time.Minute * 20

var ErrInvalidJSON = errors.New("file is not valid json")

type FileUploadData interface {
	CreateFileUploadJob(job model.FileUploadJob) (model.FileUploadJob, error)
	UpdateFileUploadJob(job model.FileUploadJob) error
	GetFileUploadJob(id int64) (model.FileUploadJob, error)
	GetAllFileUploadJobs(skip int, limit int, order string, filter model.SQLFilter) ([]model.FileUploadJob, int, error)
	GetFileUploadJobsWithStatus(status model.JobStatus) ([]model.FileUploadJob, error)
	DeleteAllFileUploads(ctx context.Context) error
}

func ProcessStaleFileUploadJobs(ctx context.Context, db FileUploadData) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when shutting down.
	if ctx.Err() != nil {
		return
	}

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

func WriteAndValidateZip(src io.Reader, dst io.Writer) error {
	tr := io.TeeReader(src, dst)
	return ValidateZipFile(tr)
}

func WriteAndValidateJSON(src io.Reader, dst io.Writer) error {
	tr := io.TeeReader(src, dst)
	_, err := ValidateMetaTag(tr, true)
	return err
}

func SaveIngestFile(location string, request *http.Request) (string, model.FileType, error) {
	fileData := request.Body
	tempFile, err := os.CreateTemp(location, "bh")
	if err != nil {
		return "", model.FileTypeJson, fmt.Errorf("error creating ingest file: %w", err)
	}

	if api.HeaderMatches(headers.ContentType.String(), mediatypes.ApplicationJson.String(), request.Header) {
		if err := WriteAndValidateFile(fileData, tempFile, WriteAndValidateJSON); err != nil {
			return "", model.FileTypeJson, err
		} else {
			return tempFile.Name(), model.FileTypeJson, nil
		}
	} else if api.HeaderMatches(headers.ContentType.String(), mediatypes.ApplicationZip.String(), request.Header) {
		if err := WriteAndValidateFile(fileData, tempFile, WriteAndValidateZip); err != nil {
			return "", model.FileTypeZip, err
		} else {
			return tempFile.Name(), model.FileTypeZip, nil
		}
	} else {
		//We should never get here since this is checked a level above
		return "", model.FileTypeJson, fmt.Errorf("invalid content type for ingest file")
	}
}

type FileValidator func(src io.Reader, dst io.Writer) error

func WriteAndValidateFile(fileData io.ReadCloser, tempFile *os.File, validationFunc FileValidator) error {
	if err := validationFunc(fileData, tempFile); err != nil {
		if err := tempFile.Close(); err != nil {
			log.Errorf("Error closing temp file %s with failed validation: %v", tempFile.Name(), err)
		} else if err := os.Remove(tempFile.Name()); err != nil {
			log.Errorf("Error deleting temp file %s: %v", tempFile.Name(), err)
		}
		return err
	} else {
		if err := tempFile.Close(); err != nil {
			log.Errorf("Error closing temp file with successful validation %s: %v", tempFile.Name(), err)
		}
		return nil
	}
}

func TouchFileUploadJobLastIngest(db FileUploadData, fileUploadJob model.FileUploadJob) error {
	fileUploadJob.LastIngest = time.Now().UTC()
	return db.UpdateFileUploadJob(fileUploadJob)
}

func EndFileUploadJob(db FileUploadData, job model.FileUploadJob) error {
	job.Status = model.JobStatusIngesting

	if err := db.UpdateFileUploadJob(job); err != nil {
		return fmt.Errorf("error ending file upload job: %w", err)
	}

	return nil
}

func UpdateFileUploadJobStatus(db FileUploadData, fileUploadJob model.FileUploadJob, status model.JobStatus, message string) error {
	fileUploadJob.Status = status
	fileUploadJob.StatusMessage = message
	fileUploadJob.EndTime = time.Now().UTC()

	return db.UpdateFileUploadJob(fileUploadJob)
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
