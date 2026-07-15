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
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/errorlist"
	"github.com/specterops/bloodhound/packages/go/metrics"
	"github.com/specterops/bloodhound/packages/go/storage"
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

func processSingleFile(ctx context.Context, fileService storage.FileService, tempDirectory string, fileData IngestFileData, ingestContext *IngestContext, readOpts ReadOptions) error {
	defer measure.ContextLogAndMeasureWithThreshold(ctx, slog.LevelDebug, "processing single file for ingest", slog.String("filepath", fileData.Path))()

	file, scratchPath, err := OpenScratchReadSeeker(ctx, tempDirectory, fileService, fileData.Path)
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
		if err := file.Close(); err != nil {
			slog.WarnContext(
				ctx,
				"Error closing ingest scratch file",
				slog.String("scratch_path", scratchPath),
				attr.Error(err),
			)
		}

		// Always remove the file after attempting to ingest it. Even if it failed.
		if err := os.Remove(scratchPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			slog.WarnContext(
				ctx,
				"Error removing ingest scratch file",
				slog.String("scratch_path", scratchPath),
				attr.Error(err),
			)
		}

		if err := fileService.DeleteFile(ctx, fileData.Path); err != nil {
			slog.WarnContext(
				ctx,
				"Error removing ingest file",
				slog.String("storage_path", fileData.Path),
				attr.Error(err),
			)
		}
	}()

	if err := ReadFileForIngest(ingestContext, file, readOpts); err != nil {
		slog.ErrorContext(
			ctx,
			"Error reading ingest file",
			slog.String("storage_path", fileData.Path),
			attr.Error(err),
		)
		return err
	}

	return nil
}

// ProcessIngestFile reads the files at the path supplied, and returns the total number of files in the
// archive, the number of files that failed to ingest as JSON, and an error
func (s *GraphifyService) ProcessIngestFile(ic *IngestContext, fileService storage.FileService, task model.IngestTask) ([]IngestFileData, error) {
	// Try to pre-process the file. If any of them fail, stop processing and return the error
	if fileData, err := ExtractIngestFiles(ic.Ctx, s.cfg.ScratchDirectory(), fileService, task.StoredFileName, task.OriginalFileName, task.FileType, fmt.Sprintf("file_upload_job_%d_", ic.JobId)); err != nil {
		return fileData, err
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

				if len(data.Errors) > 0 || data.Path == "" {
					continue
				}

				if err := processSingleFile(ic.Ctx, fileService, s.cfg.ScratchDirectory(), data, ic, readOpts); err != nil {
					var (
						graphifyError errorlist.Error
						resolutionErr endpoint.ResolutionError
					)

					if errors.As(err, &graphifyError) {
						for _, graphifyErr := range graphifyError.Errors {
							if ok := errors.As(graphifyErr, &resolutionErr); ok {
								// Resolution errors are data quality issues. They are surfaced to the user via
								// UserDataErrs but must not trigger a batch rollback.
								fileData[i].UserDataErrs = append(fileData[i].UserDataErrs, resolutionErr.Error())
							} else {
								fileData[i].Errors = append(fileData[i].Errors, graphifyErr.Error())
								errs.Add(graphifyErr)
							}
						}
					} else {
						fileData[i].Errors = append(fileData[i].Errors, err.Error())
						errs.Add(err)
					}
				}
			}
			return errs.Build()
		})
	}
}

func (s *GraphifyService) NewIngestContext(ctx context.Context, ingestTime time.Time, useChangelog bool, jobId int64, useRawObjectIDs bool) *IngestContext {
	opts := []IngestOption{
		WithIngestTime(ingestTime),
		WithEndpointResolver(s.endpointResolver),
		WithNodeKindRegistrar(s.RegisterNodeKind(ctx)),
		WithUseRawObjectIDs(useRawObjectIDs),
	}

	if useChangelog {
		opts = append(opts, WithChangeManager(s.changeManager))
	}

	if jobId > 0 {
		opts = append(opts, WithJobId(jobId))
	}

	return NewIngestContext(ctx, opts...)
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

	ingestFileService, err := s.fileServiceResolver.Resolve(storage.FileServiceIngest)
	if err != nil {
		slog.ErrorContext(s.ctx, "Error resolving ingest file service", attr.Error(err))
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

	// Lookup feature flag once per run. dont fail ingest on flag lookup, just default to false
	flagUseRawObjectIDsEnabled := appcfg.GetUseRawObjectIDsEnabled(s.ctx, s.db)

	for _, task := range tasks {
		// Record task latency metric: time from when task was created until picked up for processing
		metrics.RecordIngestTaskQueueLatency(task.CreatedAt, metrics.IngestSourceFile)

		ingestCtx := s.NewIngestContext(s.ctx, time.Now().UTC(), flagChangeLogEnabled, task.JobId.ValueOrZero(), flagUseRawObjectIDsEnabled)
		fileData, err := s.ProcessIngestFile(ingestCtx, ingestFileService, task)

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

func (s *GraphifyService) RegisterNodeKind(ctx context.Context) func(kind graph.Kind) error {
	return func(kind graph.Kind) error {
		if kind == nil || kind == graph.EmptyKind {
			return nil
		}

		if err := s.db.EnsureStubbedCustomNodeKindForIngest(ctx, kind.String()); err != nil {
			return err
		}

		return s.graphdb.RefreshKinds(ctx)
	}
}
