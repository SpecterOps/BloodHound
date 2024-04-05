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
	"archive/zip"
	"context"
	"fmt"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"io"
	"os"

	"github.com/specterops/bloodhound/src/database"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/fileupload"
)

func HasFileUploadJobsWaitingForAnalysis(ctx context.Context, db database.Database) (bool, error) {
	if fileUploadJobsUnderAnalysis, err := db.GetFileUploadJobsWithStatus(ctx, model.JobStatusAnalyzing); err != nil {
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

	if fileUploadJobsUnderAnalysis, err := db.GetFileUploadJobsWithStatus(ctx, model.JobStatusAnalyzing); err != nil {
		log.Errorf("Failed to load file upload jobs under analysis: %v", err)
	} else {
		for _, job := range fileUploadJobsUnderAnalysis {
			if err := fileupload.UpdateFileUploadJobStatus(ctx, db, job, model.JobStatusFailed, "Analysis failed"); err != nil {
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

	if fileUploadJobsUnderAnalysis, err := db.GetFileUploadJobsWithStatus(ctx, model.JobStatusAnalyzing); err != nil {
		log.Errorf("Failed to load file upload jobs under analysis: %v", err)
	} else {
		for _, job := range fileUploadJobsUnderAnalysis {
			if err := fileupload.UpdateFileUploadJobStatus(ctx, db, job, model.JobStatusPartiallyComplete, "Partially Completed"); err != nil {
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

	if fileUploadJobsUnderAnalysis, err := db.GetFileUploadJobsWithStatus(ctx, model.JobStatusAnalyzing); err != nil {
		log.Errorf("Failed to load file upload jobs under analysis: %v", err)
	} else {
		for _, job := range fileUploadJobsUnderAnalysis {
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

			if err := fileupload.UpdateFileUploadJobStatus(ctx, db, job, status, message); err != nil {
				log.Errorf("Error updating file upload job %d: %v", job.ID, err)
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

	if ingestingFileUploadJobs, err := db.GetFileUploadJobsWithStatus(ctx, model.JobStatusIngesting); err != nil {
		log.Errorf("Failed to look up finished file upload jobs: %v", err)
	} else {
		for _, ingestingFileUploadJob := range ingestingFileUploadJobs {
			if remainingIngestTasks, err := db.GetIngestTasksForJob(ctx, ingestingFileUploadJob.ID); err != nil {
				log.Errorf("Failed looking up remaining ingest tasks for file upload job %d: %v", ingestingFileUploadJob.ID, err)
			} else if len(remainingIngestTasks) == 0 {
				if err := fileupload.UpdateFileUploadJobStatus(ctx, db, ingestingFileUploadJob, model.JobStatusAnalyzing, "Analyzing"); err != nil {
					log.Errorf("Error updating fileupload job %d: %v", ingestingFileUploadJob.ID, err)
				}
			}
		}
	}
}

// clearFileTask removes a generic file upload task for ingested data.
func (s *Daemon) clearFileTask(ingestTask model.IngestTask) {
	if err := s.db.DeleteIngestTask(s.ctx, ingestTask); err != nil {
		log.Errorf("Error removing file upload task from db: %v", err)
	}
}

// preProcessIngestFile will take a path and extract zips if necessary, returning the paths for files to process
func (s *Daemon) preProcessIngestFile(path string, fileType model.FileType) ([]string, error) {
	if fileType == model.FileTypeJson {
		//If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []string{path}, nil
	} else if archive, err := zip.OpenReader(path); err != nil {
		return []string{}, err
	} else {
		filePaths := make([]string, len(archive.File))
		for i, f := range archive.File {
			//skip directories
			if f.FileInfo().IsDir() {
				continue
			}
			//Break out on an error creating files
			if tempFile, err := os.CreateTemp(s.cfg.TempDirectory(), "bh"); err != nil {
				return []string{}, err
			} else if srcFile, err := f.Open(); err != nil {
				log.Errorf("Error opening file %s in archive %s: %v", f.Name, path, err)
			} else if _, err := io.Copy(tempFile, srcFile); err != nil {
				log.Errorf("Error extracting file %s in archive %s: %v", f.Name, path, err)
			} else if err := tempFile.Close(); err != nil {
				log.Errorf("Error closing temp file %s: %v", f.Name, err)
			} else {
				filePaths[i] = tempFile.Name()
			}
		}

		//Close the archive and delete it
		if err := archive.Close(); err != nil {
			log.Errorf("Error closing archive %s: %v", path, err)
		} else if err := os.Remove(path); err != nil {
			log.Errorf("Error deleting archive %s: %v", path, err)
		}

		return filePaths, nil
	}
}

// processIngestFile reads the files at the path supplied, and returns the total number of files in the
// archive, the number of files that failed to ingest as JSON, and an error
func (s *Daemon) processIngestFile(ctx context.Context, path string, fileType model.FileType) (int, int, error) {
	if adcsEnabled, err := s.db.GetFlagByKey(ctx, appcfg.FeatureAdcs); err != nil {
		return 0, 0, fmt.Errorf("error getting adcs feature flag: %w", err)
	} else if paths, err := s.preProcessIngestFile(path, fileType); err != nil {
		return 0, 0, err
	} else {
		failed := 0

		return len(paths), failed, s.graphdb.BatchOperation(ctx, func(batch graph.Batch) error {
			for _, filePath := range paths {
				file, err := os.Open(filePath)
				if err != nil {
					failed++
					return err
				} else if err := ReadFileForIngest(batch, file, adcsEnabled.Enabled); err != nil {
					failed++
					log.Errorf("Error reading ingest file %s: %v", filePath, err)
				}

				if err := file.Close(); err != nil {
					log.Errorf("Error closing ingest file %s: %v", filePath, err)
				} else if err := os.Remove(filePath); err != nil {
					log.Errorf("Error removing ingest file %s: %v", filePath, err)
				}
			}

			return nil
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

		if s.cfg.DisableIngest {
			log.Warnf("Skipped processing of ingestTasks due to config flag.")
			return
		}

		if job, err := s.db.GetFileUploadJob(ctx, ingestTask.TaskID.ValueOrZero()); err != nil {
			log.Errorf("Failed to fetch job for ingest task %d: %v", ingestTask.ID, err)
		} else if total, failed, err := s.processIngestFile(ctx, ingestTask.FileName, ingestTask.FileType); err != nil {
			log.Errorf("Failed processing ingest task %d with file %s: %v", ingestTask.ID, ingestTask.FileName, err)
		} else {
			job.TotalFiles = total
			job.FailedFiles += failed
			if err = s.db.UpdateFileUploadJob(ctx, job); err != nil {
				log.Errorf("Failed to update number of failed files for file upload job ID %s: %v", job.ID, err)
			}
		}

		s.clearFileTask(ingestTask)
	}
}
