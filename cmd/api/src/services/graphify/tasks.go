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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/bomenc"
	"github.com/specterops/bloodhound/packages/go/errorlist"
	"github.com/specterops/dawgs/graph"
)

// UpdateJobFunc is passed to the graphify service to let it tell us about the tasks as they are processed
//
// The datapipe doesn't know or care about tasks, and the graphify service doesn't know or care about jobs.
// Instead, this func is provided as an abstraction for graphify.
type UpdateJobFunc func(jobId int64, fileData []IngestFileData)

// clearFileTask removes a generic ingest task for ingested data.
func (s *GraphifyService) clearFileTask(ingestTask model.IngestTask) {
	if err := s.db.DeleteIngestTask(s.ctx, ingestTask); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error removing ingest task from db: %v", err))
	}
}

type IngestFileData struct {
	Name       string
	ParentFile string
	Path       string
	Errors     []string
}

// extractIngestFiles will take a path and extract zips if necessary, returning the paths for files to process
// along with any errors and the number of failed files (in the case of a zip archive)
func (s *GraphifyService) extractIngestFiles(path string, providedFileName string, fileType model.FileType) ([]IngestFileData, error) {
	if fileType == model.FileTypeJson {
		//If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []IngestFileData{
			{
				Name:   providedFileName,
				Path:   path,
				Errors: []string{},
			},
		}, nil
	} else if archive, err := zip.OpenReader(path); err != nil {
		return []IngestFileData{}, err
	} else {
		var (
			errs     = errorlist.NewBuilder()
			fileData = make([]IngestFileData, 0)
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
				fileData = append(fileData, IngestFileData{
					Name:       f.Name,
					ParentFile: providedFileName,
					Errors:     []string{err.Error()},
				})

				errs.Add(err)
			} else {
				fileData = append(fileData, IngestFileData{
					Name:       f.Name,
					ParentFile: providedFileName,
					Path:       fileName,
				})
			}
		}

		return fileData, errs.Build()
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

// ProcessIngestFile reads the files at the path supplied, and returns the total number of files in the
// archive, the number of files that failed to ingest as JSON, and an error
func (s *GraphifyService) ProcessIngestFile(ctx context.Context, task model.IngestTask, ingestTime time.Time) ([]IngestFileData, error) {
	// Try to pre-process the file. If any of them fail, stop processing and return the error
	if fileData, err := s.extractIngestFiles(task.StoredFileName, task.OriginalFileName, task.FileType); err != nil {
		return []IngestFileData{}, err
	} else if changelogFF, err := s.db.GetFlagByKey(ctx, appcfg.FeatureChangelog); err != nil {
		return []IngestFileData{}, fmt.Errorf("get feature flag: %w", err)
	} else {
		errs := errorlist.NewBuilder()

		return fileData, s.graphdb.BatchOperation(ctx, func(batch graph.Batch) error {
			ingestCtx := s.newIngestContext(ctx, batch, ingestTime, changelogFF.Enabled)

			for i, data := range fileData {
				readOpts := ReadOptions{
					IngestSchema:       s.schema,
					FileType:           task.FileType,
					RegisterSourceKind: s.db.RegisterSourceKind(s.ctx),
				}

				if err := processSingleFile(ctx, data, ingestCtx, readOpts); err != nil {
					var graphifyError errorlist.Error
					if ok := errors.As(err, &graphifyError); ok {
						fileData[i].Errors = append(fileData[i].Errors, graphifyError.AsStrings()...)
					} else {
						fileData[i].Errors = append(fileData[i].Errors, err.Error())
					}
					errs.Add(err) // graphifyErrorBuilder at fn scope
					continue      // keep ingesting the rest
				}
			}

			return errs.Build()
		})
	}
}

func (s *GraphifyService) newIngestContext(ctx context.Context, batch BatchUpdater, ingestTime time.Time, useChangelog bool) *IngestContext {
	opts := []IngestOption{WithIngestTime(ingestTime)}
	if useChangelog {
		opts = append(opts, WithChangeManager(s.changeManager))
	}
	return NewIngestContext(ctx, batch, opts...)
}

func processSingleFile(ctx context.Context, fileData IngestFileData, ingestContext *IngestContext, readOpts ReadOptions) error {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelDebug, "processing single file for ingest", slog.String("filepath", fileData.Path))()

	file, err := os.Open(fileData.Path)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error opening ingest file %s: %v", fileData.Path, err))
		return err
	}

	defer func() {
		file.Close()
		// Always remove the file after attempting to ingest it. Even if it failed
		if err := os.Remove(fileData.Path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			slog.ErrorContext(ctx, fmt.Sprintf("Error removing ingest file %s: %v", fileData.Path, err))
		}
	}()

	if err := ReadFileForIngest(ingestContext, file, readOpts); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error reading ingest file %s: %v", fileData.Path, err))
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
	tasks := s.getAllTasks()
	if len(tasks) == 0 {
		// nothing to do
		return
	}

	start := time.Now()
	slog.InfoContext(s.ctx,
		"ingest run starting",
		"task_count", len(tasks),
	)

	defer func() {
		slog.InfoContext(s.ctx,
			"ingest run finished",
			"duration", time.Since(start),
			"task_count", len(tasks),
		)
	}()

	if s.ctx.Err() != nil {
		return
	}

	if s.cfg.DisableIngest {
		slog.WarnContext(s.ctx, "Skipped processing of ingestTasks due to config flag.")
		return
	}

	for _, task := range tasks {
		fileData, err := s.ProcessIngestFile(s.ctx, task, time.Now().UTC())

		switch {
		case errors.Is(err, fs.ErrNotExist):
			slog.WarnContext(s.ctx,
				"ingest file missing",
				"task_id", task.ID,
				"file", task.OriginalFileName,
				"err", err,
			)
		case err != nil:
			slog.ErrorContext(s.ctx,
				"ingest task failed",
				"task_id", task.ID,
				"file", task.OriginalFileName,
				"err", err,
			)
		default:
			slog.InfoContext(s.ctx,
				"ingest task processed",
				"task_id", task.ID,
				"file", task.OriginalFileName,
			)
		}

		updateJob(task.JobId.ValueOrZero(), fileData)
		s.clearFileTask(task)
	}

	// todo: add guard by lifting ingestCtx constructor to this func
	// if ingestCtx.HasChangelog() {
	// this logs basic metrics for the changelog, how many hits/misses per file
	s.changeManager.FlushStats()
	// }
}
