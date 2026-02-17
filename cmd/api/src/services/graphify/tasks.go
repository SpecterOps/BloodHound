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
	"io"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
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
		slog.ErrorContext(s.ctx, "Error removing ingest task from db", attr.Error(err))
	}
}

type IngestFileData struct {
	Name         string
	ParentFile   string
	Path         string
	Errors       []string
	UserDataErrs []string
}

// extractIngestFiles will take a path and extract zips if necessary, returning the paths for files to process
// along with any errors and the number of failed files (in the case of a zip archive)
func (s *GraphifyService) extractIngestFiles(path string, providedFileName string, fileType model.FileType) ([]IngestFileData, error) {
	if fileType == model.FileTypeJson {
		// If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
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
				slog.ErrorContext(
					s.ctx,
					"Error closing archive",
					slog.String("path", path),
					attr.Error(err),
				)
			}
			if err := os.Remove(path); err != nil {
				slog.ErrorContext(
					s.ctx,
					"Error deleting archive",
					slog.String("path", path),
					attr.Error(err),
				)
			}
		}()

		for _, f := range archive.File {
			// skip directories
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
func (s *GraphifyService) ProcessIngestFile(ic *IngestContext, task model.IngestTask) ([]IngestFileData, error) {
	// Try to pre-process the file. If any of them fail, stop processing and return the error
	if fileData, err := s.extractIngestFiles(task.StoredFileName, task.OriginalFileName, task.FileType); err != nil {
		return []IngestFileData{}, err
	} else {
		errs := errorlist.NewBuilder()

		return fileData, s.graphdb.BatchOperation(ic.Ctx, func(batch graph.Batch) error {
			// bind batch to ingest context now that its in scope.
			ic.BindBatchUpdater(batch)
			for i, data := range fileData {
				readOpts := ReadOptions{
					IngestSchema:       s.schema,
					FileType:           task.FileType,
					RegisterSourceKind: s.RegisterSourceKind(s.ctx),
				}

				if err := processSingleFile(ic.Ctx, data, ic, readOpts); err != nil {
					var graphifyError errorlist.Error

					if ok := errors.As(err, &graphifyError); ok {
						var userDataErr IngestUserDataError
						for _, graphifyErr := range graphifyError.Errors {
							if ok := errors.As(graphifyErr, &userDataErr); ok {
								fileData[i].UserDataErrs = append(fileData[i].UserDataErrs, userDataErr.Error())
							} else {
								fileData[i].Errors = append(fileData[i].Errors, graphifyErr.Error())
							}
						}
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

func (s *GraphifyService) NewIngestContext(ctx context.Context, ingestTime time.Time, useChangelog bool) *IngestContext {
	opts := []IngestOption{WithIngestTime(ingestTime)}

	if useChangelog {
		opts = append(opts, WithChangeManager(s.changeManager))
	}

	return NewIngestContext(ctx, opts...)
}

func processSingleFile(ctx context.Context, fileData IngestFileData, ingestContext *IngestContext, readOpts ReadOptions) error {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelDebug, "processing single file for ingest", slog.String("filepath", fileData.Path))()

	file, err := os.Open(fileData.Path)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Error opening ingest file",
			slog.String("filepath", fileData.Path),
			attr.Error(err),
		)
		return err
	}

	defer func() {
		file.Close()

		// Always remove the file after attempting to ingest it. Even if it failed
		if err := os.Remove(fileData.Path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			slog.ErrorContext(
				ctx,
				"Error removing ingest file",
				slog.String("filepath", fileData.Path),
				attr.Error(err),
			)
		}
	}()

	if err := ReadFileForIngest(ingestContext, file, readOpts); err != nil {
		slog.ErrorContext(
			ctx,
			"Error reading ingest file",
			slog.String("filepath", fileData.Path),
			attr.Error(err),
		)
		return err
	}

	return nil
}

func (s *GraphifyService) getAllTasks() model.IngestTasks {
	tasks, err := s.db.GetAllIngestTasks(s.ctx)
	if err != nil {
		slog.ErrorContext(s.ctx, "Failed fetching available ingest tasks", attr.Error(err))
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
		"Ingest run starting",
		slog.Int("task_count", len(tasks)),
	)

	if s.ctx.Err() != nil {
		return
	}

	if s.cfg.DisableIngest {
		slog.WarnContext(s.ctx, "Skipped processing of ingestTasks due to config flag.")
		return
	}

	// Lookup feature flag once per run. dont fail ingest on flag lookup, just default to false
	flagChangeLogEnabled := false
	if changelogFF, err := s.db.GetFlagByKey(s.ctx, appcfg.FeatureChangelog); err != nil {
		slog.WarnContext(s.ctx, "Get changelog feature flag failed", attr.Error(err))
	} else {
		flagChangeLogEnabled = changelogFF.Enabled
	}

	for _, task := range tasks {
		ingestCtx := s.NewIngestContext(s.ctx, time.Now().UTC(), flagChangeLogEnabled)
		fileData, err := s.ProcessIngestFile(ingestCtx, task)

		switch {
		case errors.Is(err, fs.ErrNotExist):
			slog.WarnContext(s.ctx,
				"Ingest file missing",
				slog.Int64("task_id", task.ID),
				slog.String("file", task.OriginalFileName),
				attr.Error(err),
			)
		case err != nil:
			slog.ErrorContext(s.ctx,
				"Ingest task failed",
				slog.Int64("task_id", task.ID),
				slog.String("file", task.OriginalFileName),
				attr.Error(err),
			)
		default:
			slog.InfoContext(s.ctx,
				"Ingest task processed",
				slog.Int64("task_id", task.ID),
				slog.String("file", task.OriginalFileName),
			)
		}

		updateJob(task.JobId.ValueOrZero(), fileData)
		s.clearFileTask(task)
	}

	slog.InfoContext(s.ctx,
		"Ingest run finished",
		slog.Duration("duration", time.Since(start)),
		slog.Int("task_count", len(tasks)),
	)

	if flagChangeLogEnabled {
		s.changeManager.FlushStats()
	}
}

// RegisterSourceKind - returns a function that will register a source kind and then refresh the in-memory DAWGS kind map
func (s *GraphifyService) RegisterSourceKind(ctx context.Context) func(kind graph.Kind) error {
	return func(kind graph.Kind) error {
		if err := s.db.RegisterSourceKind(ctx)(kind); err != nil {
			return err
		} else {
			return s.graphdb.RefreshKinds(ctx)
		}
	}
}
