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

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/job"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
)

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

func (s *BHCEPipeline) Start(ctx context.Context) bool {
	return s.PruneData(ctx)
}

// This handles the deletion of data if the customer requests it
func (s *BHCEPipeline) DeleteData(ctx context.Context) bool {
	deleteRequest, ok := s.db.HasCollectedGraphDataDeletionRequest()
	if !ok {
		return false
	}

	defer func() {
		_ = s.db.DeleteAnalysisRequest(ctx)
		_ = s.db.RequestAnalysis(ctx, "datapipe")
	}()
	defer measure.Measure(slog.LevelInfo, "Purge Graph Data Completed")()

	if err := s.db.SetDatapipeStatus(ctx, model.DatapipeStatusPurging, false); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error setting datapipe status: %v", err))
		return false
	}

	slog.Info("Begin Purge Graph Data")

	if err := s.db.CancelAllIngestJobs(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error cancelling jobs during data deletion: %v", err))
	} else if err := s.db.DeleteAllIngestTasks(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error deleting ingest tasks during data deletion: %v", err))
	} else if sourceKinds, err := s.db.GetSourceKinds(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error getting source kinds during data deletion: %v", err))
	} else {
		var (
			kinds         graph.Kinds
			filteredKinds graph.Kinds
		)
		for _, k := range sourceKinds {
			kinds = append(kinds, k.Name)
		}
		// Filter out reserved kinds before removing records from source_kinds table
		for _, kind := range kinds {
			if !kind.Is(ad.Entity) && !kind.Is(azure.Entity) {
				filteredKinds = append(filteredKinds, kind)
			}
		}

		if err := DeleteCollectedGraphData(ctx, s.graphdb, deleteRequest, kinds); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error deleting graph data: %v", err))
		} else if err := s.db.DeleteSourceKindsByName(ctx, filteredKinds); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error deleting source kinds: %v", err))
		}
	}
}

// This is called on Daemon start. We get a list of all filenames we know/expect and delete any
// other files. Would love to move this out of datapipe entirely eventually and into... somewhere
func (s *BHCEPipeline) PruneData(ctx context.Context) bool {
	if ingestTasks, err := s.db.GetAllIngestTasks(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed fetching available ingest tasks: %v", err))
		return false
	} else {
		expectedFiles := make([]string, len(ingestTasks))

		for idx, ingestTask := range ingestTasks {
			expectedFiles[idx] = ingestTask.FileName
		}

		go s.orphanedFileSweeper.Clear(ctx, expectedFiles)
	}
	return true
}

// This is currently public to support as a first class testing seam, but with some refactoring may be split away from the
// Daemon object enough to be self standing and pulled to an internal package namespace
func (s *BHCEPipeline) IngestTasks(ctx context.Context) bool {
	// Ingest all available ingest tasks
	s.graphifyService.ProcessTasks(updateJobFunc(ctx, s.db))

	// Manage time-out state progression for ingest jobs
	s.jobService.ProcessStaleIngestJobs()

	// Manage nominal state transitions for ingest jobs
	s.jobService.ProcessFinishedIngestJobs()
	return true
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

// If the pipeline needs to do anything to the context, this is called before each other pipeline stage
func (s *BHCEPipeline) IsActive(ctx context.Context, status model.DatapipeStatus) (bool, context.Context) {
	return true, ctx
}

func (s *BHCEPipeline) Analyze(ctx context.Context) bool {

	// If there are completed ingest jobs or if analysis was user-requested, perform analysis.
	if hasJobsWaitingForAnalysis, err := s.jobService.HasIngestJobsWaitingForAnalysis(); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed looking up jobs waiting for analysis: %v", err))
		return false
	} else if hasJobsWaitingForAnalysis || s.db.HasAnalysisRequest(ctx) {

		// Ensure that the user-requested analysis switch is deleted. This is done at the beginning of the
		// function so that any re-analysis requests are caught while analysis is in-progress.
		if err := s.db.DeleteAnalysisRequest(ctx); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error deleting analysis request: %v", err))
			return false
		}

		if s.cfg.DisableAnalysis {
			return false
		}

		defer measure.LogAndMeasure(slog.LevelInfo, "Graph Analysis")()

		if err := RunAnalysisOperations(ctx, s.db, s.graphdb, s.cfg); err != nil {
			if errors.Is(err, ErrAnalysisFailed) {
				s.jobService.FailAnalyzedIngestJobs()

			} else if errors.Is(err, ErrAnalysisPartiallyCompleted) {
				s.jobService.PartialCompleteIngestJobs()
			}
			return false
		} else {
			s.jobService.CompleteAnalyzedIngestJobs()

			// This is cacheclearing. The analysis is still successful here
			if _, err := s.db.GetFlagByKey(ctx, appcfg.FeatureEntityPanelCaching); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error retrieving entity panel caching flag: %v", err))
			} else {
				if err := s.cache.Reset(); err != nil {
					slog.Error(fmt.Sprintf("Error while resetting the cache: %v", err))
				} else {
					slog.Info("Cache successfully reset by datapipe daemon")
				}
			}
			return true
		}
	}
	return false
}
