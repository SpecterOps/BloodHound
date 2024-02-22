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
	"os"

	"github.com/specterops/bloodhound/src/database"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/fileupload"
)

func HasFileUploadJobsWaitingForAnalysis(db database.Database) (bool, error) {
	if fileUploadJobsUnderAnalysis, err := db.GetFileUploadJobsWithStatus(model.JobStatusAnalyzing); err != nil {
		return false, err
	} else {
		return len(fileUploadJobsUnderAnalysis) > 0, nil
	}
}

func FailAnalyzedFileUploadJobs(ctx context.Context, db database.Database) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if ctx.Err() != nil {
		return
	}

	if fileUploadJobsUnderAnalysis, err := db.GetFileUploadJobsWithStatus(model.JobStatusAnalyzing); err != nil {
		log.Errorf("Failed to load file upload jobs under analysis: %v", err)
	} else {
		for _, job := range fileUploadJobsUnderAnalysis {
			if err := fileupload.UpdateFileUploadJobStatus(db, job, model.JobStatusFailed, "Analysis failed"); err != nil {
				log.Errorf("Failed updating file upload job %d to failed status: %v", job.ID, err)
			}
		}
	}
}

func PartialCompleteFileUploadJobs(ctx context.Context, db database.Database) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if ctx.Err() != nil {
		return
	}

	if fileUploadJobsUnderAnalysis, err := db.GetFileUploadJobsWithStatus(model.JobStatusAnalyzing); err != nil {
		log.Errorf("Failed to load file upload jobs under analysis: %v", err)
	} else {
		for _, job := range fileUploadJobsUnderAnalysis {
			if err := fileupload.UpdateFileUploadJobStatus(db, job, model.JobStatusPartiallyComplete, "Partially Completed"); err != nil {
				log.Errorf("Failed updating file upload job %d to partially completed status: %v", job.ID, err)
			}
		}
	}
}

func CompleteAnalyzedFileUploadJobs(ctx context.Context, db database.Database) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when we are shutting down.
	if ctx.Err() != nil {
		return
	}

	if fileUploadJobsUnderAnalysis, err := db.GetFileUploadJobsWithStatus(model.JobStatusAnalyzing); err != nil {
		log.Errorf("Failed to load file upload jobs under analysis: %v", err)
	} else {
		for _, job := range fileUploadJobsUnderAnalysis {
			if err := fileupload.UpdateFileUploadJobStatus(db, job, model.JobStatusComplete, "Complete"); err != nil {
				log.Errorf("Error updating fileupload job %d: %v", job.ID, err)
			}
		}
	}
}

func ProcessIngestedFileUploadJobs(ctx context.Context, db database.Database) {
	// Because our database interfaces do not yet accept contexts this is a best-effort check to ensure that we do not
	// commit state transitions when shutting down.
	if ctx.Err() != nil {
		return
	}

	if ingestingFileUploadJobs, err := db.GetFileUploadJobsWithStatus(model.JobStatusIngesting); err != nil {
		log.Errorf("Failed to look up finished file upload jobs: %v", err)
	} else {
		for _, ingestingFileUploadJob := range ingestingFileUploadJobs {
			if remainingIngestTasks, err := db.GetIngestTasksForJob(ingestingFileUploadJob.ID); err != nil {
				log.Errorf("Failed looking up remaining ingest tasks for file upload job %d: %v", ingestingFileUploadJob.ID, err)
			} else if len(remainingIngestTasks) == 0 {
				if err := fileupload.UpdateFileUploadJobStatus(db, ingestingFileUploadJob, model.JobStatusAnalyzing, "Analyzing"); err != nil {
					log.Errorf("Error updating fileupload job %d: %v", ingestingFileUploadJob.ID, err)
				}
			}
		}
	}
}

// clearFileTask removes a generic file upload task for ingested data.
func (s *Daemon) clearFileTask(ingestTask model.IngestTask) {
	if err := s.db.DeleteIngestTask(ingestTask); err != nil {
		log.Errorf("Error removing file upload task from db: %v", err)
	}
}

func (s *Daemon) processIngestFile(ctx context.Context, path string) error {
	if jsonFile, err := os.Open(path); err != nil {
		return err
	} else {
		defer func() {
			if err := jsonFile.Close(); err != nil {
				log.Errorf("Failed closing ingest file %s: %v", path, err)
			}
		}()

		return s.graphdb.BatchOperation(ctx, func(batch graph.Batch) error {
			if err := ReadFileForIngest(batch, jsonFile); err != nil {
				return err
			} else {
				return nil
			}
		})
	}
}

// processIngestTasks covers the generic file upload case for ingested data.
func (s *Daemon) processIngestTasks(ctx context.Context, ingestTasks model.IngestTasks) {
	s.status.Update(model.DatapipeStatusIngesting, false)
	defer s.status.Update(model.DatapipeStatusIdle, false)

	for _, ingestTask := range ingestTasks {
		// Check the context to see if we should continue processing ingest tasks. This has to be explicit since error
		// handling assumes that all failures should be logged and not returned.
		if ctx.Err() != nil {
			return
		}

		if err := s.processIngestFile(ctx, ingestTask.FileName); err != nil {
			log.Errorf("Failed processing ingest task %d with file %s: %v", ingestTask.ID, ingestTask.FileName, err)
		}

		s.clearFileTask(ingestTask)
	}
}
