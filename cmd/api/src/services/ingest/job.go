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

// kpom-todo: These are called by the API, so we'll want to move the handler to use the service model as well
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

	if metadata, err := writeAndValidateFile(fileData, tempFile, validationFn); err != nil {
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

func writeAndValidateFile(fileData io.ReadCloser, tempFile *os.File, validationFunc FileValidator) (ingest.Metadata, error) {
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

// Datapipe calls the IngestService below to create/manage jobs in all of their different states

func (s *IngestService) HasIngestJobsWaitingForAnalysis() (bool, error) {
	if ingestJobsUnderAnalysis, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusAnalyzing); err != nil {
		return false, err
	} else {
		return len(ingestJobsUnderAnalysis) > 0, nil
	}
}

func (s *IngestService) FailAnalyzedIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if s.ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to load ingest jobs under analysis: %v", err))
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			if err := updateIngestJobStatus(s.ctx, s.db, job, model.JobStatusFailed, "Analysis failed"); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed updating ingest job %d to failed status: %v", job.ID, err))
			}
		}
	}
}

func (s *IngestService) PartialCompleteIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if s.ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to load ingest jobs under analysis: %v", err))
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			if err := updateIngestJobStatus(s.ctx, s.db, job, model.JobStatusPartiallyComplete, "Partially Completed"); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed updating ingest job %d to partially completed status: %v", job.ID, err))
			}
		}
	}
}

func (s *IngestService) CompleteAnalyzedIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if s.ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to load ingest jobs under analysis: %v", err))
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			var (
				status  = model.JobStatusComplete
				message = "Complete"
			)

			if job.FailedFiles > 0 {
				if job.FailedFiles < job.TotalFiles {
					status = model.JobStatusPartiallyComplete
					message = fmt.Sprintf("%d File(s) failed to ingest as JSON Content", job.FailedFiles)
				} else {
					status = model.JobStatusFailed
					message = "All files failed to ingest as JSON Content"
				}
			}

			if err := updateIngestJobStatus(s.ctx, s.db, job, status, message); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Error updating ingest job %d: %v", job.ID, err))
			}
		}
	}
}

// ProcessFinishedIngestJobs transitions all jobs in an ingesting state to an analyzing state, if there are no further tasks associated with the job in question
func (s *IngestService) ProcessFinishedIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when shutting down.
	if s.ctx.Err() != nil {
		return
	}

	if jobs, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusIngesting); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to look up finished ingest jobs: %v", err))
	} else {
		for _, job := range jobs {
			if remainingIngestTasks, err := s.db.GetIngestTasksForJob(s.ctx, job.ID); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed looking up remaining ingest tasks for ingest job %d: %v", job.ID, err))
			} else if len(remainingIngestTasks) == 0 {
				if err := updateIngestJobStatus(s.ctx, s.db, job, model.JobStatusAnalyzing, "Analyzing"); err != nil {
					slog.ErrorContext(s.ctx, fmt.Sprintf("Error updating ingest job %d: %v", job.ID, err))
				}
			}
		}
	}
}

// ProcessStaleIngestJobs fetches all runnings ingest jobs and transitions them to a timed out state if the job has been inactive for too long.
func (s *IngestService) ProcessStaleIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when shutting down.
	if s.ctx.Err() != nil {
		return
	}

	var (
		now       = time.Now().UTC()
		threshold = now.Add(-jobActivityTimeout)
	)

	if jobs, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusRunning); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error getting running jobs: %v", err))
	} else {
		for _, job := range jobs {
			if job.LastIngest.Before(threshold) {
				slog.WarnContext(s.ctx, fmt.Sprintf("Ingest timeout: No ingest activity observed for Job ID %d in %f minutes (last ingest was %s)). Upload incomplete",
					job.ID,
					now.Sub(threshold).Minutes(),
					job.LastIngest.Format(time.RFC3339)))

				if err := timeOutIngestJob(s.ctx, s.db, job.ID, fmt.Sprintf("Ingest timeout: No ingest activity observed in %f minutes. Upload incomplete.", now.Sub(threshold).Minutes())); err != nil {
					slog.ErrorContext(s.ctx, fmt.Sprintf("Error marking ingest job %d as timed out: %v", job.ID, err))
				}
			}
		}
	}
}

func timeOutIngestJob(ctx context.Context, db IngestData, jobID int64, message string) error {
	if job, err := db.GetIngestJob(ctx, jobID); err != nil {
		return err
	} else {
		job.Status = model.JobStatusTimedOut
		job.StatusMessage = message
		job.EndTime = time.Now().UTC()

		return db.UpdateIngestJob(ctx, job)
	}
}

func updateIngestJobStatus(ctx context.Context, db IngestData, job model.IngestJob, status model.JobStatus, message string) error {
	job.Status = status
	job.StatusMessage = message
	job.EndTime = time.Now().UTC()

	return db.UpdateIngestJob(ctx, job)
}
