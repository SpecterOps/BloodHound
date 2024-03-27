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
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/model/ingest"
	"github.com/specterops/bloodhound/src/utils"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
)

const jobActivityTimeout = time.Minute * 20

const (
	UTF8BOM1 = 0xef
	UTF8BOM2 = 0xbb
	UTF8BMO3 = 0xbf
)

var ErrInvalidJSON = errors.New("file is not valid json")

type FileUploadData interface {
	CreateFileUploadJob(ctx context.Context, job model.FileUploadJob) (model.FileUploadJob, error)
	UpdateFileUploadJob(ctx context.Context, job model.FileUploadJob) error
	GetFileUploadJob(ctx context.Context, id int64) (model.FileUploadJob, error)
	GetAllFileUploadJobs(ctx context.Context, skip int, limit int, order string, filter model.SQLFilter) ([]model.FileUploadJob, int, error)
	GetFileUploadJobsWithStatus(ctx context.Context, status model.JobStatus) ([]model.FileUploadJob, error)
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

	if jobs, err := db.GetFileUploadJobsWithStatus(ctx, model.JobStatusRunning); err != nil {
		log.Errorf("Error getting running jobs: %v", err)
	} else {
		for _, job := range jobs {
			if job.LastIngest.Before(threshold) {
				log.Warnf("Ingest timeout: No ingest activity observed for Job ID %d in %f minutes (last ingest was %s). Upload incomplete",
					job.ID,
					now.Sub(threshold).Minutes(),
					job.LastIngest.Format(time.RFC3339))

				if err := TimeOutUploadJob(ctx, db, job.ID, fmt.Sprintf("Ingest timeout: No ingest activity observed in %f minutes. Upload incomplete.", now.Sub(threshold).Minutes())); err != nil {
					log.Errorf("Error marking file upload job %d as timed out: %v", job.ID, err)
				}
			}
		}
	}
}

func GetAllFileUploadJobs(ctx context.Context, db FileUploadData, skip int, limit int, order string, filter model.SQLFilter) ([]model.FileUploadJob, int, error) {
	return db.GetAllFileUploadJobs(ctx, skip, limit, order, filter)
}

func StartFileUploadJob(ctx context.Context, db FileUploadData, user model.User) (model.FileUploadJob, error) {
	job := model.FileUploadJob{
		UserID:     user.ID,
		User:       user,
		Status:     model.JobStatusRunning,
		StartTime:  time.Now().UTC(),
		LastIngest: time.Now().UTC(),
	}
	return db.CreateFileUploadJob(ctx, job)
}

func GetFileUploadJobByID(ctx context.Context, db FileUploadData, jobID int64) (model.FileUploadJob, error) {
	return db.GetFileUploadJob(ctx, jobID)
}

func WriteAndValidateZip(src io.Reader, dst io.Writer) error {
	tr := io.TeeReader(src, dst)
	return ValidateZipFile(tr)
}

func WriteAndValidateJSON(src io.Reader, dst io.Writer) error {
	tr := io.TeeReader(src, dst)
	bufReader := bufio.NewReader(tr)
	if b, err := bufReader.Peek(3); err != nil {
		return err
	} else {
		if b[0] == UTF8BOM1 && b[1] == UTF8BOM2 && b[2] == UTF8BMO3 {
			if _, err := bufReader.Discard(3); err != nil {
				return err
			}
		}
	}
	_, err := ValidateMetaTag(bufReader, true)
	return err
}

func SaveIngestFile(location string, request *http.Request) (string, model.FileType, error) {
	fileData := request.Body
	tempFile, err := os.CreateTemp(location, "bh")
	if err != nil {
		return "", model.FileTypeJson, fmt.Errorf("error creating ingest file: %w", err)
	}

	if utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()) {
		return tempFile.Name(), model.FileTypeJson, WriteAndValidateFile(fileData, tempFile, WriteAndValidateJSON)
	} else if utils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...) {
		return tempFile.Name(), model.FileTypeZip, WriteAndValidateFile(fileData, tempFile, WriteAndValidateZip)
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

func TouchFileUploadJobLastIngest(ctx context.Context, db FileUploadData, fileUploadJob model.FileUploadJob) error {
	fileUploadJob.LastIngest = time.Now().UTC()
	return db.UpdateFileUploadJob(ctx, fileUploadJob)
}

func EndFileUploadJob(ctx context.Context, db FileUploadData, job model.FileUploadJob) error {
	job.Status = model.JobStatusIngesting

	if err := db.UpdateFileUploadJob(ctx, job); err != nil {
		return fmt.Errorf("error ending file upload job: %w", err)
	}

	return nil
}

func UpdateFileUploadJobStatus(ctx context.Context, db FileUploadData, fileUploadJob model.FileUploadJob, status model.JobStatus, message string) error {
	fileUploadJob.Status = status
	fileUploadJob.StatusMessage = message
	fileUploadJob.EndTime = time.Now().UTC()

	return db.UpdateFileUploadJob(ctx, fileUploadJob)
}

func TimeOutUploadJob(ctx context.Context, db FileUploadData, jobID int64, message string) error {
	if job, err := db.GetFileUploadJob(ctx, jobID); err != nil {
		return err
	} else {
		job.Status = model.JobStatusTimedOut
		job.StatusMessage = message
		job.EndTime = time.Now().UTC()

		return db.UpdateFileUploadJob(ctx, job)
	}
}
