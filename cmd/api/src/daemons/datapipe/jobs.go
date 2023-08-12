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
	"os"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/fileupload"
)

func (s *Daemon) numAvailableCompletedFileUploadJobs() int {
	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.completedFileUploadJobIDs)
}

func (s *Daemon) failJobsUnderAnalysis() {
	for _, jobID := range s.fileUploadJobIDsUnderAnalysis {
		if err := fileupload.FailFileUploadJob(s.db, jobID, "Analysis failed"); err != nil {
			log.Errorf("Failed updating job %d to failed status: %v", jobID, err)
		}
	}

	s.clearJobsFromAnalysis()
}

func (s *Daemon) clearJobsFromAnalysis() {
	s.lock.Lock()
	s.fileUploadJobIDsUnderAnalysis = s.fileUploadJobIDsUnderAnalysis[:0]
	s.lock.Unlock()
}

func (s *Daemon) processCompletedFileUploadJobs() {
	completedJobIDs := s.getAndTransitionCompletedJobIDs()

	for _, id := range completedJobIDs {
		if ingestTasks, err := s.db.GetIngestTasksForJob(id); err != nil {
			log.Errorf("Failed fetching available ingest tasks: %v", err)
		} else {
			s.processIngestTasks(ingestTasks)
		}

		if err := fileupload.UpdateFileUploadJobStatus(s.db, id, model.JobStatusComplete, "Complete"); err != nil {
			log.Errorf("Error updating fileupload job %d: %v", id, err)
		}
	}
}

func (s *Daemon) getAndTransitionCompletedJobIDs() []int64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	// transition completed jobs to analysis
	s.fileUploadJobIDsUnderAnalysis = append(s.fileUploadJobIDsUnderAnalysis, s.completedFileUploadJobIDs...)
	s.completedFileUploadJobIDs = s.completedFileUploadJobIDs[:0]

	return s.fileUploadJobIDsUnderAnalysis
}

func (s *Daemon) processIngestTasks(ingestTasks model.IngestTasks) {
	s.status.Update(model.DatapipeStatusIngesting, false)
	defer s.status.Update(model.DatapipeStatusIdle, false)

	for _, ingestTask := range ingestTasks {
		jsonFile, err := os.Open(ingestTask.FileName)
		if err != nil {
			log.Errorf("Error reading file for ingest task %v: %v", ingestTask.ID, err)
		}

		if err = s.graphdb.BatchOperation(s.ctx, func(batch graph.Batch) error {
			if err := s.ReadWrapper(batch, jsonFile); err != nil {
				return err
			} else {
				return nil
			}
		}); err != nil {
			log.Errorf("Error processing ingest task %v: %v", ingestTask.ID, err)
		}

		s.clearTask(ingestTask)
		jsonFile.Close()
	}
}

func (s *Daemon) clearTask(ingestTask model.IngestTask) {
	if err := s.db.DeleteIngestTask(ingestTask); err != nil {
		log.Errorf("Error removing task from db: %v", err)
	}
}

func (s *Daemon) NotifyOfFileUploadJobStatus(job model.FileUploadJob) {
	if job.Status == model.JobStatusIngesting {
		s.lock.Lock()
		s.completedFileUploadJobIDs = append(s.completedFileUploadJobIDs, job.ID)
		s.lock.Unlock()
	}
}
