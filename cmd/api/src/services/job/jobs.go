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
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/ingest"
)

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
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to load ingest jobs under analysis: %v", err))
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			if err := ingest.UpdateIngestJobStatus(s.ctx, s.db, job, model.JobStatusFailed, "Analysis failed"); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed updating ingest job %d to failed status: %v", job.ID, err))
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
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to load ingest jobs under analysis: %v", err))
	} else {
		for _, job := range ingestJobsUnderAnalysis {
			if err := ingest.UpdateIngestJobStatus(s.ctx, s.db, job, model.JobStatusPartiallyComplete, "Partially Completed"); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed updating ingest job %d to partially completed status: %v", job.ID, err))
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

			if err := ingest.UpdateIngestJobStatus(s.ctx, s.db, job, status, message); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Error updating ingest job %d: %v", job.ID, err))
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
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to look up finished ingest jobs: %v", err))
	} else {
		for _, job := range jobs {
			if remainingIngestTasks, err := s.db.GetIngestTasksForJob(s.ctx, job.ID); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed looking up remaining ingest tasks for ingest job %d: %v", job.ID, err))
			} else if len(remainingIngestTasks) == 0 {
				if err := ingest.UpdateIngestJobStatus(s.ctx, s.db, job, model.JobStatusAnalyzing, "Analyzing"); err != nil {
					slog.ErrorContext(s.ctx, fmt.Sprintf("Error updating ingest job %d: %v", job.ID, err))
				}
			}
		}
	}
}
