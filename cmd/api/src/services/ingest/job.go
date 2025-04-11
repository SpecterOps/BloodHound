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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/mock.go -package=mocks . IngestData
package ingest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/specterops/bloodhound/bomenc"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"
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

// TODO: Integration test here
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

func WriteAndValidateZip(src io.Reader, dst io.Writer) error {
	tr := io.TeeReader(src, dst)
	return ValidateZipFile(tr)
}

func WriteAndValidateJSON(src io.Reader, dst io.Writer) error {
	normalizedContent, err := bomenc.NormalizeToUTF8(src)
	if err != nil {
		return err
	}
	tr := io.TeeReader(normalizedContent, dst)
	_, err = ValidateMetaTag(tr, true)
	return err
}

// TODO: Integration test here
func SaveIngestFile(location string, contentType string, fileData io.ReadCloser) (string, model.FileType, error) {
	tempFile, err := os.CreateTemp(location, "bh")
	if err != nil {
		return "", model.FileTypeJson, fmt.Errorf("error creating ingest file: %w", err)
	}

	if contentType == mediatypes.ApplicationJson.String() {
		return tempFile.Name(), model.FileTypeJson, WriteAndValidateFile(fileData, tempFile, WriteAndValidateJSON)

	} else if slices.Contains(ingest.AllowedZipFileUploadTypes, contentType) {
		return tempFile.Name(), model.FileTypeZip, WriteAndValidateFile(fileData, tempFile, WriteAndValidateZip)
	} else {
		// We should never get here since this is checked a level above
		return "", model.FileTypeJson, fmt.Errorf("invalid content type for ingest file")
	}
}

type FileValidator func(src io.Reader, dst io.Writer) error

func WriteAndValidateFile(fileData io.ReadCloser, tempFile *os.File, validationFunc FileValidator) error {
	if err := validationFunc(fileData, tempFile); err != nil {
		if err := tempFile.Close(); err != nil {
			slog.Error(fmt.Sprintf("Error closing temp file %s with failed validation: %v", tempFile.Name(), err))
		} else if err := os.Remove(tempFile.Name()); err != nil {
			slog.Error(fmt.Sprintf("Error deleting temp file %s: %v", tempFile.Name(), err))
		}
		return err
	} else {
		if err := tempFile.Close(); err != nil {
			slog.Error(fmt.Sprintf("Error closing temp file with successful validation %s: %v", tempFile.Name(), err))
		}
		return nil
	}
}

func TouchIngestJobLastIngest(ctx context.Context, db IngestData, job model.IngestJob) error {
	job.LastIngest = time.Now().UTC()
	return db.UpdateIngestJob(ctx, job)
}

// TODO: Integration test here
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
