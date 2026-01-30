// Copyright 2025 Specter Ops, Inc.
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
	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
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

var ErrAnalysisDisabled = errors.New("analysis is disabled by configuration")

type BHCEPipeline struct {
	db                  database.Database
	graphdb             graph.Database
	cache               cache.Cache
	cfg                 config.Configuration
	orphanedFileSweeper *OrphanFileSweeper
	ingestSchema        upload.IngestSchema
	jobService          job.JobService
	graphifyService     graphify.GraphifyService
	changelog           *changelog.Changelog
}

func NewPipeline(ctx context.Context, cfg config.Configuration, db database.Database, graphDB graph.Database, cache cache.Cache, ingestSchema upload.IngestSchema, cl *changelog.Changelog) *BHCEPipeline {
	return &BHCEPipeline{
		db:                  db,
		graphdb:             graphDB,
		cache:               cache,
		cfg:                 cfg,
		orphanedFileSweeper: NewOrphanFileSweeper(NewOSFileOperations(), cfg.TempDirectory()),
		ingestSchema:        ingestSchema,
		jobService:          job.NewJobService(ctx, db),
		graphifyService:     graphify.NewGraphifyService(ctx, db, graphDB, cfg, ingestSchema, cl),
		changelog:           cl,
	}
}

func (s *BHCEPipeline) Start(ctx context.Context) error {
	return s.PruneData(ctx)
}

// This handles the deletion of data if the customer requests it
func (s *BHCEPipeline) DeleteData(ctx context.Context) error {
	deleteRequest, ok := s.db.HasCollectedGraphDataDeletionRequest(ctx)
	if !ok {
		return nil
	}
	defer func() {
		_ = s.db.DeleteAnalysisRequest(ctx)
		_ = s.db.RequestAnalysis(ctx, "datapipe")
	}()
	defer measure.LogAndMeasure(slog.LevelInfo, "Purge Graph Data")()

	slog.InfoContext(ctx, "Begin Purge Graph Data")

	if err := s.db.CancelAllIngestJobs(ctx); err != nil {
		return fmt.Errorf("cancelling jobs during data deletion: %v", err)
	} else if err := s.db.DeleteAllIngestTasks(ctx); err != nil {
		return fmt.Errorf("deleting ingest tasks during data deletion: %v", err)
	} else if err := PurgeGraphData(ctx, deleteRequest, s.graphdb, s.db); err != nil {
		return fmt.Errorf("purging graph data failed: %w", err)
	}

	// Clear changelog cache to ensure consistency after graph data deletion
	if s.changelog != nil {
		s.changelog.ClearCache(ctx)
	}

	return nil
}

func PurgeGraphData(
	ctx context.Context,
	deleteRequest model.AnalysisRequest,
	graphdb graph.Database,
	db database.SourceKindsData,
) error {
	sourceKinds, err := db.GetSourceKinds(ctx)
	if err != nil {
		return fmt.Errorf("getting source kinds: %w", err)
	}

	allSourceKinds := extractKindNames(sourceKinds)
	filteredKinds := filterDeletableKinds(deleteRequest.DeleteSourceKinds)

	if err := DeleteCollectedGraphData(ctx, graphdb, deleteRequest, allSourceKinds); err != nil {
		return fmt.Errorf("deleting graph data: %w", err)
	}

	if err := db.DeactivateSourceKindsByName(ctx, filteredKinds); err != nil {
		return fmt.Errorf("deactivating source kinds: %w", err)
	}

	return nil
}

func extractKindNames(sourceKinds []database.SourceKind) graph.Kinds {
	var kinds graph.Kinds
	for _, k := range sourceKinds {
		kinds = append(kinds, k.Name)
	}
	return kinds
}

// if the delete request specifies any source_kinds for deletion we want to remove them from the source_kinds table.
// we want to remove 3rd party source_kinds when requested(e.g. GithubBase, HelloBase), but this ensures that we never remove Base and AZBase.
func filterDeletableKinds(kindsToDelete []string) graph.Kinds {
	var filtered graph.Kinds
	for _, kind := range kindsToDelete {
		k := graph.StringKind(kind)
		if !k.Is(ad.Entity) && !k.Is(azure.Entity) {
			filtered = append(filtered, k)
		}
	}
	return filtered
}

// This is called on Daemon start. We get a list of all filenames we know/expect and delete any
// other files. Would love to move this out of datapipe entirely eventually and into... somewhere
func (s *BHCEPipeline) PruneData(ctx context.Context) error {
	if ingestTasks, err := s.db.GetAllIngestTasks(ctx); err != nil {
		return fmt.Errorf("fetching available ingest tasks: %v", err)
	} else {
		expectedFiles := make([]string, len(ingestTasks))

		for idx, ingestTask := range ingestTasks {
			expectedFiles[idx] = ingestTask.StoredFileName
		}

		go s.orphanedFileSweeper.Clear(ctx, expectedFiles)
	}
	return nil
}

// This is currently public to support as a first class testing seam, but with some refactoring may be split away from the
// Daemon object enough to be self standing and pulled to an internal package namespace
func (s *BHCEPipeline) IngestTasks(ctx context.Context) error {
	// Ingest all available ingest tasks
	s.graphifyService.ProcessTasks(updateJobFunc(ctx, s.db))

	// Manage time-out state progression for ingest jobs
	s.jobService.ProcessStaleIngestJobs()

	// Manage nominal state transitions for ingest jobs
	s.jobService.ProcessFinishedIngestJobs()
	return nil
}

// updateJobFunc generates a valid graphify.UpdateJobFunc by injecting the parent context and database interface
// Only used as a callback, so not exposed
func updateJobFunc(ctx context.Context, db database.Database) graphify.UpdateJobFunc {
	return func(jobID int64, fileData []graphify.IngestFileData) {
		if job, err := db.GetIngestJob(ctx, jobID); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed to fetch job for ingest task %d: %v", jobID, err))
		} else {
			for _, file := range fileData {
				job.TotalFiles += 1
				completedTask := model.CompletedTask{
					IngestJobId:    job.ID,
					FileName:       file.Name,
					ParentFileName: file.ParentFile,
					Errors:         []string{},
					Warnings:       []string{},
				}

				if len(file.Errors) > 0 {
					job.FailedFiles += 1
					completedTask.Errors = file.Errors
				}
				if len(file.UserDataErrs) > 0 {
					job.PartialFailedFiles += 1
					completedTask.Warnings = file.UserDataErrs
				}

				if _, err = db.CreateCompletedTask(ctx, completedTask); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Failed to create completed task for ingest task %d: %v", job.ID, err))
				}
			}

			if err = db.UpdateIngestJob(ctx, job); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed to update number of failed files for ingest job ID %d: %v", job.ID, err))
			}
		}
	}
}

// If the pipeline needs to do anything to the context, this is called before each other pipeline stage
func (s *BHCEPipeline) IsPrimary(ctx context.Context, status model.DatapipeStatus) (bool, context.Context) {
	return true, ctx
}

func (s *BHCEPipeline) Analyze(ctx context.Context) error {
	// If there are completed ingest jobs or if analysis was user-requested, perform analysis.
	if hasJobsWaitingForAnalysis, err := s.jobService.HasIngestJobsWaitingForAnalysis(); err != nil {
		return fmt.Errorf("looking up jobs for analysis: %v", err)
	} else if hasJobsWaitingForAnalysis || s.db.HasAnalysisRequest(ctx) {
		// Ensure that the user-requested analysis switch is deleted. This is done at the beginning of the
		// function so that any re-analysis requests are caught while analysis is in-progress.
		if err := s.db.DeleteAnalysisRequest(ctx); err != nil {
			return fmt.Errorf("clearing analysis request: %v", err)
		}

		if s.cfg.DisableAnalysis {
			return ErrAnalysisDisabled
		}

		defer measure.LogAndMeasure(slog.LevelInfo, "Graph Analysis")()

		if err := RunAnalysisOperations(ctx, s.db, s.graphdb, s.cfg); err != nil {
			if errors.Is(err, ErrAnalysisFailed) {
				s.jobService.FailAnalyzedIngestJobs()
			} else if errors.Is(err, ErrAnalysisPartiallyCompleted) {
				s.jobService.PartialCompleteIngestJobs()
			}
			return fmt.Errorf("analysis failure: %v", err)
		} else if err := s.db.UpdateLastAnalysisCompleteTime(ctx); err != nil {
			return fmt.Errorf("update last analysis completion time: %v", err)
		} else {
			s.jobService.CompleteAnalyzedIngestJobs()

			// This is cacheclearing. The analysis is still successful here
			if _, err := s.db.GetFlagByKey(ctx, appcfg.FeatureEntityPanelCaching); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error retrieving entity panel caching flag: %v", err))
			} else if err := s.cache.Reset(); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error while resetting the cache: %v", err))
			} else {
				slog.InfoContext(ctx, "Cache successfully reset by datapipe daemon")
			}

			return nil
		}
	} else {
		return nil
	}
}
