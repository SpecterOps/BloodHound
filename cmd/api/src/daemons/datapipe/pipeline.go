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
	"errors"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/services/graphify"
	"github.com/specterops/bloodhound/src/services/job"
	"github.com/specterops/bloodhound/src/services/upload"
)

// PipelineInterface defines methods that operate on instance state.
// These methods are not static and require a fully initialized pipeline.
type Pipeline interface {
	PruneData(context.Context)
	DeleteData(context.Context)
	IngestTasks(context.Context)
	Analyze(context.Context)
	Start(context.Context)
}

type BHCEPipeline struct {
	db                  database.Database
	graphdb             graph.Database
	cache               cache.Cache
	cfg                 config.Configuration
	orphanedFileSweeper *OrphanFileSweeper
	ingestSchema        upload.IngestSchema
	jobService          job.JobService
	graphifyService     graphify.GraphifyService
}

func NewPipeline(ctx context.Context, cfg config.Configuration, db database.Database, graphDB graph.Database, cache cache.Cache, ingestSchema upload.IngestSchema) *BHCEPipeline {
	return &BHCEPipeline{
		db:                  db,
		graphdb:             graphDB,
		cache:               cache,
		cfg:                 cfg,
		orphanedFileSweeper: NewOrphanFileSweeper(NewOSFileOperations(), cfg.TempDirectory()),
		ingestSchema:        ingestSchema,
		jobService:          job.NewJobService(ctx, db),
		graphifyService:     graphify.NewGraphifyService(ctx, db, graphDB, cfg, ingestSchema),
	}
}

func (s *BHCEPipeline) Start(ctx context.Context) {
	s.PruneData(ctx)
}

// This handles the deletion of data if the customer requests it. I'm not sure why the defers
// are stacked the way the are, but this should keep all current functionality.
func (s *BHCEPipeline) DeleteData(ctx context.Context) {

	defer func() {
		_ = s.db.DeleteAnalysisRequest(ctx)
		_ = s.db.RequestAnalysis(ctx, "datapie")
		measure.Measure(slog.LevelInfo, "Purge Graph Data Completed")()
	}()

	slog.Info("Begin Purge Graph Data")

	if err := s.db.CancelAllIngestJobs(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error cancelling jobs during data deletion: %v", err))
	} else if err := s.db.DeleteAllIngestTasks(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error deleting ingest tasks during data deletion: %v", err))
	} else if err := DeleteCollectedGraphData(ctx, s.graphdb); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error deleting graph data: %v", err))
	}
}

// This is called on Daemon start. We get a list of all filenames we know/expect and delete any
// other files. Would love to move this out of datapipe entirely eventually and into... somewhere
func (s *BHCEPipeline) PruneData(ctx context.Context) {
	if ingestTasks, err := s.db.GetAllIngestTasks(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed fetching available ingest tasks: %v", err))
	} else {
		expectedFiles := make([]string, len(ingestTasks))

		for idx, ingestTask := range ingestTasks {
			expectedFiles[idx] = ingestTask.FileName
		}

		go s.orphanedFileSweeper.Clear(ctx, expectedFiles)
	}
}

// This is currently public to support as a first class testing seam, but with some refactoring may be split away from the
// Daemon object enough to be self standing and pulled to an internal package namespace
func (s *BHCEPipeline) IngestTasks(ctx context.Context) {
	// Ingest all available ingest tasks
	s.graphifyService.ProcessTasks(updateJobFunc(ctx, s.db))

	// Manage time-out state progression for ingest jobs
	s.jobService.ProcessStaleIngestJobs()

	// Manage nominal state transitions for ingest jobs
	s.jobService.ProcessFinishedIngestJobs()
}

// updateJobFunc generates a valid graphify.UpdateJobFunc by injecting the parent context and database interface
// Only used as a callback, so not exposed
func updateJobFunc(ctx context.Context, db database.Database) graphify.UpdateJobFunc {
	return func(jobID int64, totalFiles int, totalFailed int) {
		if job, err := db.GetIngestJob(ctx, jobID); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed to fetch job for ingest task %d: %v", jobID, err))
		} else {
			job.TotalFiles += totalFiles
			job.FailedFiles += totalFailed

			if err = db.UpdateIngestJob(ctx, job); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed to update number of failed files for ingest job ID %d: %v", job.ID, err))
			}
		}
	}
}

func (s *BHCEPipeline) Analyze(ctx context.Context) {

	// If there are completed ingest jobs or if analysis was user-requested, perform analysis.
	if hasJobsWaitingForAnalysis, err := s.jobService.HasIngestJobsWaitingForAnalysis(); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed looking up jobs waiting for analysis: %v", err))
	} else if hasJobsWaitingForAnalysis || s.db.HasAnalysisRequest(ctx) {

		// Ensure that the user-requested analysis switch is deleted. This is done at the beginning of the
		// function so that any re-analysis requests are caught while analysis is in-progress.
		if err := s.db.DeleteAnalysisRequest(ctx); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error deleting analysis request: %v", err))
			return
		}

		if s.cfg.DisableAnalysis {
			return
		}

		defer measure.LogAndMeasure(slog.LevelInfo, "Graph Analysis")()

		if err := RunAnalysisOperations(ctx, s.db, s.graphdb, s.cfg); err != nil {
			if errors.Is(err, ErrAnalysisFailed) {
				s.jobService.FailAnalyzedIngestJobs()

			} else if errors.Is(err, ErrAnalysisPartiallyCompleted) {
				s.jobService.PartialCompleteIngestJobs()
			}
		} else {
			s.jobService.CompleteAnalyzedIngestJobs()

			if _, err := s.db.GetFlagByKey(ctx, appcfg.FeatureEntityPanelCaching); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error retrieving entity panel caching flag: %v", err))
			} else {
				if err := s.cache.Reset(); err != nil {
					slog.Error(fmt.Sprintf("Error while resetting the cache: %v", err))
				} else {
					slog.Info("Cache successfully reset by datapipe daemon")
				}
			}
		}
	}
}
