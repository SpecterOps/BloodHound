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

package ingest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"
	"github.com/specterops/bloodhound/src/utils"
)

const jobActivityTimeout = time.Minute * 20

var ErrInvalidJSON = errors.New("file is not valid json")

// ProcessStaleIngestJobs fetches all runnings ingest jobs and transitions them to a timed out state if the job has been inactive for too long.
func ProcessStaleIngestJobs(ctx context.Context, db IngestData) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when shutting down.
	if ctx.Err() != nil {
		return
	}

	var (
		now       = time.Now().UTC()
		threshold = now.Add(-jobActivityTimeout)
	)

	if jobs, err := db.GetIngestJobsWithStatus(ctx, model.JobStatusRunning); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error getting running jobs: %v", err))
	} else {
		for _, job := range jobs {
			if job.LastIngest.Before(threshold) {
				slog.WarnContext(ctx, fmt.Sprintf("Ingest timeout: No ingest activity observed for Job ID %d in %f minutes (last ingest was %s)). Upload incomplete",
					job.ID,
					now.Sub(threshold).Minutes(),
					job.LastIngest.Format(time.RFC3339)))

				if err := TimeOutIngestJob(ctx, db, job.ID, fmt.Sprintf("Ingest timeout: No ingest activity observed in %f minutes. Upload incomplete.", now.Sub(threshold).Minutes())); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error marking ingest job %d as timed out: %v", job.ID, err))
				}
			}
		}
	}
}

func GetAllIngestJobs(ctx context.Context, db IngestData, skip int, limit int, order string, filter model.SQLFilter) ([]model.IngestJob, int, error) {
	return db.GetAllIngestJobs(ctx, skip, limit, order, filter)
}

func StartIngestJob(ctx context.Context, db IngestData, user model.User) (model.IngestJob, error) {
	job := model.IngestJob{
		UserID:     user.ID,
		User:       user,
		Status:     model.JobStatusRunning,
		StartTime:  time.Now().UTC(),
		LastIngest: time.Now().UTC(),
	}
	return db.CreateIngestJob(ctx, job)
}

func GetIngestJobByID(ctx context.Context, db IngestData, jobID int64) (model.IngestJob, error) {
	return db.GetIngestJob(ctx, jobID)
}

func SaveIngestFile(location string, request *http.Request, validator IngestValidator) (IngestTaskParams, error) {
	fileData := request.Body
	tempFile, err := os.CreateTemp(location, "bh")
	if err != nil {
		return IngestTaskParams{Filename: "", FileType: model.FileTypeJson}, fmt.Errorf("error creating ingest file: %w", err)
	}

	var (
		fileType     model.FileType
		validationFn FileValidator
	)

	switch {
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
		fileType = model.FileTypeJson
		validationFn = validator.WriteAndValidateJSON
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
		fileType = model.FileTypeZip
		validationFn = WriteAndValidateZip
	default:
		return IngestTaskParams{}, fmt.Errorf("invalid content type for ingest file")
	}

	if metadata, err := WriteAndValidateFile(fileData, tempFile, validationFn); err != nil {
		return IngestTaskParams{}, err
	} else {
		isGeneric := false
		if metadata.Type == ingest.DataTypeGeneric {
			isGeneric = true
		}
		return IngestTaskParams{
			Filename:  tempFile.Name(),
			FileType:  fileType,
			IsGeneric: isGeneric,
		}, nil
	}

}

func WriteAndValidateFile(fileData io.ReadCloser, tempFile *os.File, validationFunc FileValidator) (ingest.Metadata, error) {
	if meta, err := validationFunc(fileData, tempFile); err != nil {
		if err := tempFile.Close(); err != nil {
			slog.Error(fmt.Sprintf("Error closing temp file %s with failed validation: %v", tempFile.Name(), err))
		} else if err := os.Remove(tempFile.Name()); err != nil {
			slog.Error(fmt.Sprintf("Error deleting temp file %s: %v", tempFile.Name(), err))
		}
		return meta, err
	} else {
		if err := tempFile.Close(); err != nil {
			slog.Error(fmt.Sprintf("Error closing temp file with successful validation %s: %v", tempFile.Name(), err))
		}
		return meta, nil
	}
}

func TouchIngestJobLastIngest(ctx context.Context, db IngestData, job model.IngestJob) error {
	job.LastIngest = time.Now().UTC()
	return db.UpdateIngestJob(ctx, job)
}

func EndIngestJob(ctx context.Context, db IngestData, job model.IngestJob) error {
	job.Status = model.JobStatusIngesting

	if err := db.UpdateIngestJob(ctx, job); err != nil {
		return fmt.Errorf("error ending ingest job: %w", err)
	}

	return nil
}

func UpdateIngestJobStatus(ctx context.Context, db IngestData, job model.IngestJob, status model.JobStatus, message string) error {
	job.Status = status
	job.StatusMessage = message
	job.EndTime = time.Now().UTC()

	return db.UpdateIngestJob(ctx, job)
}

func TimeOutIngestJob(ctx context.Context, db IngestData, jobID int64, message string) error {
	if job, err := db.GetIngestJob(ctx, jobID); err != nil {
		return err
	} else {
		job.Status = model.JobStatusTimedOut
		job.StatusMessage = message
		job.EndTime = time.Now().UTC()

		return db.UpdateIngestJob(ctx, job)
	}
}
