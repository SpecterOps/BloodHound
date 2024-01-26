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

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/fileupload"
)

func (s *Daemon) failJobsUnderAnalysis() {
	if fileUploadJobsUnderAnalysis, err := s.db.GetFileUploadJobsWithStatus(model.JobStatusAnalyzing); err != nil {
		log.Errorf("Failed to load file upload jobs under analysis: %v", err)
	} else {
		for _, job := range fileUploadJobsUnderAnalysis {
			if err := fileupload.FailFileUploadJob(s.db, job.ID, "Analysis failed"); err != nil {
				log.Errorf("Failed updating file upload job %d to failed status: %v", job.ID, err)
			}
		}
	}
}

func (s *Daemon) completeJobsUnderAnalysis() {
	if fileUploadJobsUnderAnalysis, err := s.db.GetFileUploadJobsWithStatus(model.JobStatusAnalyzing); err != nil {
		log.Errorf("Failed to load file upload jobs under analysis: %v", err)
	} else {
		for _, job := range fileUploadJobsUnderAnalysis {
			if err := fileupload.UpdateFileUploadJobStatus(s.db, job, model.JobStatusComplete, "Complete"); err != nil {
				log.Errorf("Error updating fileupload job %d: %v", job.ID, err)
			}
		}
	}
}

func (s *Daemon) processIngestedFileUploadJobs() {
	if ingestedFileUploadJobs, err := s.db.GetFileUploadJobsWithStatus(model.JobStatusIngesting); err != nil {
		log.Errorf("Failed to look up finished file upload jobs: %v", err)
	} else {
		for _, ingestedFileUploadJob := range ingestedFileUploadJobs {
			if err := fileupload.UpdateFileUploadJobStatus(s.db, ingestedFileUploadJob, model.JobStatusAnalyzing, "Analyzing"); err != nil {
				log.Errorf("Error updating fileupload job %d: %v", ingestedFileUploadJob.ID, err)
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
			if err := s.ReadWrapper(batch, jsonFile); err != nil {
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
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := s.processIngestFile(ctx, ingestTask.FileName); err != nil {
			log.Errorf("Failed processing ingest task %d with file %s: %v", ingestTask.ID, ingestTask.FileName, err)
		}

		s.clearFileTask(ingestTask)
	}
}

func (s *Daemon) clearTask(ingestTask model.IngestTask) {
	if err := s.db.DeleteIngestTask(ingestTask); err != nil {
		log.Errorf("Error removing task from db: %v", err)
	}
}
