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

package job

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
)

const jobActivityTimeout = time.Minute * 20

func updateIngestJobStatus(ctx context.Context, db JobData, job model.IngestJob, status model.JobStatus, message string) error {
	job.Status = status
	job.StatusMessage = message
	job.EndTime = time.Now().UTC()

	return db.UpdateIngestJob(ctx, job)
}

func timeOutIngestJob(ctx context.Context, db JobData, jobID int64, message string) error {
	if job, err := db.GetIngestJob(ctx, jobID); err != nil {
		return err
	} else {
		job.Status = model.JobStatusTimedOut
		job.StatusMessage = message
		job.EndTime = time.Now().UTC()

		return db.UpdateIngestJob(ctx, job)
	}
}

// ProcessStaleIngestJobs fetches all runnings ingest jobs and transitions them to a timed out state if the job has been inactive for too long.
func (s *JobService) ProcessStaleIngestJobs() {
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
		slog.ErrorContext(
			s.ctx,
			"Error getting running jobs",
			attr.Error(err),
		)
	} else {
		for _, job := range jobs {
			if job.LastIngest.Before(threshold) {
				slog.WarnContext(
					s.ctx,
					"Ingest timeout: No ingest activity observed for Job ID. Upload incomplete",
					slog.Int64("job_id", job.ID),
					slog.Float64("timeout_minutes", now.Sub(threshold).Minutes()),
					slog.String("last_ingest_time", job.LastIngest.Format(time.RFC3339)),
				)

				if err := timeOutIngestJob(s.ctx, s.db, job.ID, fmt.Sprintf("Ingest timeout: No ingest activity observed in %f minutes. Upload incomplete.", now.Sub(threshold).Minutes())); err != nil {
					slog.ErrorContext(
						s.ctx,
						"Error marking ingest job as timed out",
						slog.Int64("job_id", job.ID),
						attr.Error(err),
					)
				}
			}
		}
	}
}

func (s *JobService) HasIngestJobsWaitingForAnalysis() (bool, error) {
	if ingestJobsUnderAnalysis, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusAnalyzing); err != nil {
		return false, err
	} else {
		return len(ingestJobsUnderAnalysis) > 0, nil
	}
}

func (s *JobService) FailAnalyzedIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if s.ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(
			s.ctx,
			"Failed to load ingest jobs under analysis",
			attr.Error(err),
		)
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			if err := updateIngestJobStatus(s.ctx, s.db, job, model.JobStatusFailed, "Analysis failed"); err != nil {
				slog.ErrorContext(
					s.ctx,
					"Failed updating ingest job to failed status",
					slog.Int64("job_id", job.ID),
					attr.Error(err),
				)
			}
		}
	}
}

func (s *JobService) PartialCompleteIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if s.ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(
			s.ctx,
			"Failed to load ingest jobs under analysis",
			attr.Error(err),
		)
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			if err := updateIngestJobStatus(s.ctx, s.db, job, model.JobStatusPartiallyComplete, "Partially Completed"); err != nil {
				slog.ErrorContext(
					s.ctx,
					"Failed updating ingest job to partially completed status",
					slog.Int64("job_id", job.ID),
					attr.Error(err),
				)
			}
		}
	}
}

func (s *JobService) CompleteAnalyzedIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if s.ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(
			s.ctx,
			"Failed to load ingest jobs under analysis",
			attr.Error(err),
		)
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
			} else if job.PartialFailedFiles > 0 {
				status = model.JobStatusPartiallyComplete
				message = fmt.Sprintf("%d File(s) partially failed to ingest as JSON Content", job.PartialFailedFiles)
			}

			if err := updateIngestJobStatus(s.ctx, s.db, job, status, message); err != nil {
				slog.ErrorContext(
					s.ctx,
					"Error updating ingest job",
					slog.Int64("job_id", job.ID),
					attr.Error(err),
				)
			}
		}
	}
}

// ProcessFinishedIngestJobs transitions all jobs in an ingesting state to an analyzing state, if there are no further tasks associated with the job in question
func (s *JobService) ProcessFinishedIngestJobs() {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when shutting down.
	if s.ctx.Err() != nil {
		return
	}

	if jobs, err := s.db.GetIngestJobsWithStatus(s.ctx, model.JobStatusIngesting); err != nil {
		slog.ErrorContext(
			s.ctx,
			"Failed to look up finished ingest jobs",
			attr.Error(err),
		)
	} else {
		for _, job := range jobs {
			if remainingIngestTasks, err := s.db.GetIngestTasksForJob(s.ctx, job.ID); err != nil {
				slog.ErrorContext(
					s.ctx,
					"Failed looking up remaining ingest tasks for ingest job",
					slog.Int64("job_id", job.ID),
					attr.Error(err),
				)
			} else if len(remainingIngestTasks) == 0 {
				if err := updateIngestJobStatus(s.ctx, s.db, job, model.JobStatusAnalyzing, "Analyzing"); err != nil {
					slog.ErrorContext(
						s.ctx,
						"Error updating ingest job",
						slog.Int64("job_id", job.ID),
						attr.Error(err),
					)
				}
			}
		}
	}
}
