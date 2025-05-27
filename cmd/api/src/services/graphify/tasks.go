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

package graphify

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/specterops/bloodhound/bomenc"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

// clearFileTask removes a generic ingest task for ingested data.
func (s *GraphifyService) clearFileTask(ingestTask model.IngestTask) {
	if err := s.db.DeleteIngestTask(s.ctx, ingestTask); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error removing ingest task from db: %v", err))
	}
}

// preProcessIngestFile will take a path and extract zips if necessary, returning the paths for files to process
// along with any errors and the number of failed files (in the case of a zip archive)
func (s *GraphifyService) preProcessIngestFile(path string, fileType model.FileType) ([]string, int, error) {
	if fileType == model.FileTypeJson {
		//If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []string{path}, 0, nil
	} else if archive, err := zip.OpenReader(path); err != nil {
		return []string{}, 0, err
	} else {
		var (
			errs      = util.NewErrorCollector()
			failed    = 0
			filePaths = make([]string, len(archive.File))
		)

		for i, f := range archive.File {
			//skip directories
			if f.FileInfo().IsDir() {
				continue
			}
			// Break out if temp file creation fails
			// Collect errors for other failures within the archive
			if tempFile, err := os.CreateTemp(s.cfg.TempDirectory(), "bh"); err != nil {
				return []string{}, 0, err
			} else if srcFile, err := f.Open(); err != nil {
				errs.Add(fmt.Errorf("error opening file %s in archive %s: %v", f.Name, path, err))
				failed++
			} else if normFile, err := bomenc.NormalizeToUTF8(srcFile); err != nil {
				errs.Add(fmt.Errorf("error normalizing file %s to UTF8 in archive %s: %v", f.Name, path, err))
				failed++
			} else if _, err := io.Copy(tempFile, normFile); err != nil {
				errs.Add(fmt.Errorf("error extracting file %s in archive %s: %v", f.Name, path, err))
				failed++
			} else if err := tempFile.Close(); err != nil {
				errs.Add(fmt.Errorf("error closing temp file %s: %v", f.Name, err))
				failed++
			} else {
				filePaths[i] = tempFile.Name()
			}
		}

		//Close the archive and delete it
		if err := archive.Close(); err != nil {
			slog.ErrorContext(s.ctx, fmt.Sprintf("Error closing archive %s: %v", path, err))
		} else if err := os.Remove(path); err != nil {
			slog.ErrorContext(s.ctx, fmt.Sprintf("Error deleting archive %s: %v", path, err))
		}

		return filePaths, failed, errs.Combined()
	}
}

// processIngestFile reads the files at the path supplied, and returns the total number of files in the
// archive, the number of files that failed to ingest as JSON, and an error
func (s *GraphifyService) processIngestFile(ctx context.Context, task model.IngestTask, ingestTime time.Time) (int, int, error) {
	adcsEnabled := false
	if adcsFlag, err := s.db.GetFlagByKey(ctx, appcfg.FeatureAdcs); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error getting ADCS flag: %v", err))
	} else {
		adcsEnabled = adcsFlag.Enabled
	}
	if paths, failed, err := s.preProcessIngestFile(task.FileName, task.FileType); err != nil {
		return 0, failed, err
	} else {
		failed = 0

		return len(paths), failed, s.graphdb.BatchOperation(ctx, func(batch graph.Batch) error {
			timestampedBatch := NewTimestampedBatch(batch, ingestTime)

			for _, filePath := range paths {
				file, err := os.Open(filePath)
				if err != nil {
					failed++
					return err
				} else if err := ReadFileForIngest(timestampedBatch, file, ReadOptions{IngestSchema: s.schema, FileType: task.FileType, ADCSEnabled: adcsEnabled}); err != nil {
					failed++
					slog.ErrorContext(ctx, fmt.Sprintf("Error reading ingest file %s: %v", filePath, err))
				}

				if err := file.Close(); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error closing ingest file %s: %v", filePath, err))
				} else if err := os.Remove(filePath); errors.Is(err, fs.ErrNotExist) {
					slog.WarnContext(ctx, fmt.Sprintf("Removing ingest file %s: %v", filePath, err))
				} else if err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error removing ingest file %s: %v", filePath, err))
				}
			}

			return nil
		})
	}
}

// processIngestTasks covers the generic ingest case for ingested data.
func (s *GraphifyService) ProcessTasks() {
	if tasks, err := s.db.GetAllIngestTasks(s.ctx); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed fetching available ingest tasks: %v", err))
	} else {
		for _, task := range tasks {
			// Check the context to see if we should continue processing ingest tasks. This has to be explicit since error
			// handling assumes that all failures should be logged and not returned.
			if s.ctx.Err() != nil {
				return
			}

			if s.cfg.DisableIngest {
				slog.WarnContext(s.ctx, "Skipped processing of ingestTasks due to config flag.")
				return
			}
			total, failed, err := s.processIngestFile(s.ctx, task, time.Now().UTC())

			if errors.Is(err, fs.ErrNotExist) {
				slog.WarnContext(s.ctx, fmt.Sprintf("Did not process ingest task %d with file %s: %v", task.ID, task.FileName, err))
			} else if err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed processing ingest task %d with file %s: %v", task.ID, task.FileName, err))
			} else if job, err := s.db.GetIngestJob(s.ctx, task.TaskID.ValueOrZero()); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to fetch job for ingest task %d: %v", task.ID, err))
			} else {
				job.TotalFiles += total
				job.FailedFiles += failed

				if err = s.db.UpdateIngestJob(s.ctx, job); err != nil {
					slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to update number of failed files for ingest job ID %d: %v", job.ID, err))
				}
			}

			s.clearFileTask(task)
		}
	}
}
