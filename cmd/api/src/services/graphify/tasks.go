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

	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/bomenc"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util"
)

// UpdateJobFunc is passed to the graphify service to let it tell us about the tasks as they are processed
//
// The datapipe doesn't know or care about tasks, and the graphify service doesn't know or care about jobs.
// Instead, this func is provided as an abstraction for graphify.
type UpdateJobFunc func(jobId int64, totalFiles int, totalFailed int)

// clearFileTask removes a generic ingest task for ingested data.
func (s *GraphifyService) clearFileTask(ingestTask model.IngestTask) {
	if err := s.db.DeleteIngestTask(s.ctx, ingestTask); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error removing ingest task from db: %v", err))
	}
}

// extractIngestFiles will take a path and extract zips if necessary, returning the paths for files to process
// along with any errors and the number of failed files (in the case of a zip archive)
func (s *GraphifyService) extractIngestFiles(path string, fileType model.FileType) ([]string, int, error) {
	if fileType == model.FileTypeJson {
		//If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []string{path}, 0, nil
	} else if archive, err := zip.OpenReader(path); err != nil {
		return []string{}, 0, err
	} else {
		var (
			errs      = util.NewErrorCollector()
			failed    = 0
			filePaths = make([]string, 0, len(archive.File))
		)

		defer func() {
			if err := archive.Close(); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Error closing archive %s: %v", path, err))
			}
			if err := os.Remove(path); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Error deleting archive %s: %v", path, err))
			}
		}()

		for _, f := range archive.File {
			//skip directories
			if f.FileInfo().IsDir() {
				continue
			}

			fileName, err := s.extractToTempFile(f)
			if err != nil {
				failed++
				errs.Add(err)
			} else {
				filePaths = append(filePaths, fileName)
			}
		}

		return filePaths, failed, errs.Combined()
	}
}

func (s *GraphifyService) extractToTempFile(f *zip.File) (string, error) {
	// Given a single artifact in an archive, extract it out to a temporary file
	tempFile, err := os.CreateTemp(s.cfg.TempDirectory(), "bh")
	if err != nil {
		return "", err
	}

	success := false
	defer func() {
		// Always close the tempFile, but...
		tempFile.Close()
		if !success {
			// ... only delete if it wasn't successful. Otherwise we leave it around to be processed
			os.Remove(tempFile.Name())
		}
	}()

	srcFile, err := f.Open()
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	// this creates a normalized file to feed to the copy
	if normFile, err := bomenc.NormalizeToUTF8(srcFile); err != nil {
		return "", err
		// and this is what actually copies it to disk
	} else if _, err := io.Copy(tempFile, normFile); err != nil {
		return "", err
	} else {
		// let the deferred method above know we shouldn't delete it and return the filename
		success = true
		return tempFile.Name(), nil
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

	// Try to pre-process the file. If any of them fail, stop processing and return the error
	if paths, failedExtracting, err := s.extractIngestFiles(task.FileName, task.FileType); err != nil {
		return 0, failedExtracting, err
	} else {

		failedIngestion := 0

		errs := util.NewErrorCollector()
		return len(paths), failedIngestion, s.graphdb.BatchOperation(ctx, func(batch graph.Batch) error {
			timestampedBatch := NewTimestampedBatch(batch, ingestTime)

			for _, filePath := range paths {
				readOpts := ReadOptions{IngestSchema: s.schema, FileType: task.FileType, ADCSEnabled: adcsEnabled}

				if err := processSingleFile(ctx, filePath, timestampedBatch, readOpts); err != nil {
					failedIngestion++
					errs.Add(err) // util.NewErrorCollector at fn scope
					continue      // keep ingesting the rest
				}
			}

			return errs.Combined()
		})
	}
}

func processSingleFile(ctx context.Context, filePath string, batch *TimestampedBatch, readOpts ReadOptions) error {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelDebug, "processing single file for ingest", slog.String("filepath", filePath))()

	file, err := os.Open(filePath)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error opening ingest file %s: %v", filePath, err))
		return err
	}

	defer func() {
		file.Close()
		// Always remove the file after attempting to ingest it. Even if it failed
		if err := os.Remove(filePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			slog.ErrorContext(ctx, fmt.Sprintf("Error removing ingest file %s: %v", filePath, err))
		}
	}()

	if err := ReadFileForIngest(batch, file, readOpts); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error reading ingest file %s: %v", filePath, err))
		return err
	}

	return nil
}

func (s *GraphifyService) getAllTasks() model.IngestTasks {
	tasks, err := s.db.GetAllIngestTasks(s.ctx)
	if err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed fetching available ingest tasks: %v", err))
		return model.IngestTasks{}
	}
	return tasks
}

func (s *GraphifyService) ProcessTasks(updateJob UpdateJobFunc) {

	for _, task := range s.getAllTasks() {
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
		}

		updateJob(task.JobId.ValueOrZero(), total, failed)
		s.clearFileTask(task)
	}
}
