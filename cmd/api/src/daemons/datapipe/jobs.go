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

package datapipe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/ingest"
)

const jobActivityTimeout = time.Minute * 20

func HasIngestJobsWaitingForAnalysis(ctx context.Context, db database.Database) (bool, error) {
	if ingestJobsUnderAnalysis, err := db.GetIngestJobsWithStatus(ctx, model.JobStatusAnalyzing); err != nil {
		return false, err
	} else {
		return len(ingestJobsUnderAnalysis) > 0, nil
	}
}

func FailAnalyzedIngestJobs(ctx context.Context, db database.Database) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := db.GetIngestJobsWithStatus(ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed to load ingest jobs under analysis: %v", err))
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			if err := ingest.UpdateIngestJobStatus(ctx, db, job, model.JobStatusFailed, "Analysis failed"); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed updating ingest job %d to failed status: %v", job.ID, err))
			}
		}
	}
}

func PartialCompleteIngestJobs(ctx context.Context, db database.Database) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := db.GetIngestJobsWithStatus(ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed to load ingest jobs under analysis: %v", err))
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			if err := ingest.UpdateIngestJobStatus(ctx, db, job, model.JobStatusPartiallyComplete, "Partially Completed"); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed updating ingest job %d to partially completed status: %v", job.ID, err))
			}
		}
	}
}

func CompleteAnalyzedIngestJobs(ctx context.Context, db database.Database) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if ctx.Err() != nil {
		return
	}

	if ingestJobsUnderAnalysis, err := db.GetIngestJobsWithStatus(ctx, model.JobStatusAnalyzing); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed to load ingest jobs under analysis: %v", err))
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

			if err := ingest.UpdateIngestJobStatus(ctx, db, job, status, message); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error updating ingest job %d: %v", job.ID, err))
			}
		}
	}
}

// ProcessFinishedIngestJobs transitions all jobs in an ingesting state to an analyzing state, if there are no further tasks associated with the job in question
func ProcessFinishedIngestJobs(ctx context.Context, db database.Database) error {
	var errs = make([]error, 0)

	jobs, err := db.GetIngestJobsWithStatus(ctx, model.JobStatusIngesting)
	if err != nil {
		return fmt.Errorf("look up finished ingest jobs: %w", err)
	}

	for _, job := range jobs {
		// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
		// commit state transitions when shutting down.
		if ctx.Err() != nil {
			errs = append(errs, fmt.Errorf("context error during finished ingest jobs handling: %w", ctx.Err()))
			return errors.Join(errs...)
		}

		if remainingIngestTasks, err := db.GetIngestTasksForJob(ctx, job.ID); err != nil {
			errs = append(errs, fmt.Errorf("look up remaining ingest tasks for job id %d: %w", job.ID, err))
		} else if len(remainingIngestTasks) == 0 {
			if err := ingest.UpdateIngestJobStatus(ctx, db, job, model.JobStatusAnalyzing, "Analyzing"); err != nil {
				errs = append(errs, fmt.Errorf("update ingest job id %d: %w", job.ID, err))
			}
		}
	}

	return errors.Join(errs...)
}

// ProcessStaleIngestJobs fetches all runnings ingest jobs and transitions them to a timed out state if the job has been inactive for too long.
func ProcessStaleIngestJobs(ctx context.Context, db database.Database) error {
	var (
		now       = time.Now().UTC()
		threshold = now.Add(-jobActivityTimeout)
		errs      = make([]error, 0)
	)

	jobs, err := db.GetIngestJobsWithStatus(ctx, model.JobStatusRunning)
	if err != nil {
		return fmt.Errorf("could not get running jobs: %w", err)
	}

	for _, job := range jobs {
		// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
		// commit state transitions when shutting down.
		if ctx.Err() != nil {
			errs := append(errs, fmt.Errorf("context error during stale ingest task handling: %w", ctx.Err()))
			return errors.Join(errs...)
		}

		if job.LastIngest.Before(threshold) {
			slog.WarnContext(
				ctx,
				"Ingest timeout, upload incomplete. See job_id, last_ingest, and minutes_since_last_activity values",
				slog.Int64("job_id", job.ID),
				slog.String("last_ingest", job.LastIngest.Format(time.RFC3339)),
				slog.Float64("minutes_since_last_activity", now.Sub(job.LastIngest).Minutes()),
			)

			if err := ingest.TimeOutIngestJob(ctx, db, job.ID, fmt.Sprintf("Ingest timeout: No ingest activity observed in %f minutes. Upload incomplete.", now.Sub(threshold).Minutes())); err != nil {
				errs = append(errs, fmt.Errorf("could not mark ingest job id %d as timed out: %w", job.ID, err))
			}
		}
	}

	return errors.Join(errs...)
}
