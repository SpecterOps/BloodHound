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
	"time"

	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/services/fileupload"
)

const (
	pruningInterval = time.Hour * 24
)

type Daemon struct {
	db                  database.Database
	graphdb             graph.Database
	cache               cache.Cache
	cfg                 config.Configuration
	tickInterval        time.Duration
	ctx                 context.Context
	orphanedFileSweeper *OrphanFileSweeper
}

func (s *Daemon) Name() string {
	return "Data Pipe Daemon"
}

func NewDaemon(ctx context.Context, cfg config.Configuration, connections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch], cache cache.Cache, tickInterval time.Duration) *Daemon {
	return &Daemon{
		db:                  connections.RDMS,
		graphdb:             connections.Graph,
		cache:               cache,
		cfg:                 cfg,
		ctx:                 ctx,
		orphanedFileSweeper: NewOrphanFileSweeper(NewOSFileOperations(), cfg.TempDirectory()),
		tickInterval:        tickInterval,
	}
}

func (s *Daemon) analyze() {
	// Ensure that the user-requested analysis switch is deleted. This is done at the beginning of the
	// function so that any re-analysis requests are caught while analysis is in-progress.
	if err := s.db.DeleteAnalysisRequest(s.ctx); err != nil {
		log.Errorf("Error deleting analysis request: %v", err)
		return
	}

	if s.cfg.DisableAnalysis {
		return
	}

	if err := s.db.SetDatapipeStatus(s.ctx, model.DatapipeStatusAnalyzing, false); err != nil {
		log.Errorf("Error setting datapipe status: %v", err)
		return
	}

	defer log.LogAndMeasure(log.LevelInfo, "Graph Analysis")()

	if err := RunAnalysisOperations(s.ctx, s.db, s.graphdb, s.cfg); err != nil {
		if errors.Is(err, ErrAnalysisFailed) {
			FailAnalyzedFileUploadJobs(s.ctx, s.db)
			if err := s.db.SetDatapipeStatus(s.ctx, model.DatapipeStatusIdle, false); err != nil {
				log.Errorf("Error setting datapipe status: %v", err)
				return
			}

		} else if errors.Is(err, ErrAnalysisPartiallyCompleted) {
			PartialCompleteFileUploadJobs(s.ctx, s.db)
			if err := s.db.SetDatapipeStatus(s.ctx, model.DatapipeStatusIdle, true); err != nil {
				log.Errorf("Error setting datapipe status: %v", err)
				return
			}
		}
	} else {
		CompleteAnalyzedFileUploadJobs(s.ctx, s.db)

		if entityPanelCachingFlag, err := s.db.GetFlagByKey(s.ctx, appcfg.FeatureEntityPanelCaching); err != nil {
			log.Errorf("Error retrieving entity panel caching flag: %v", err)
		} else {
			resetCache(s.cache, entityPanelCachingFlag.Enabled)
		}

		if err := s.db.SetDatapipeStatus(s.ctx, model.DatapipeStatusIdle, true); err != nil {
			log.Errorf("Error setting datapipe status: %v", err)
			return
		}
	}
}

func resetCache(cacher cache.Cache, cacheEnabled bool) {
	if err := cacher.Reset(); err != nil {
		log.Errorf("Error while resetting the cache: %v", err)
	} else {
		log.Infof("Cache successfully reset by datapipe daemon")
	}
}

func (s *Daemon) ingestAvailableTasks() {
	if ingestTasks, err := s.db.GetAllIngestTasks(s.ctx); err != nil {
		log.Errorf("Failed fetching available ingest tasks: %v", err)
	} else {
		s.processIngestTasks(s.ctx, ingestTasks)
	}
}

func (s *Daemon) Start(ctx context.Context) {
	var (
		datapipeLoopTimer = time.NewTimer(s.tickInterval)
		pruningTicker     = time.NewTicker(pruningInterval)
	)

	defer datapipeLoopTimer.Stop()
	defer pruningTicker.Stop()

	s.clearOrphanedData()

	for {
		select {
		case <-pruningTicker.C:
			s.clearOrphanedData()

		case <-datapipeLoopTimer.C:
			if s.db.HasCollectedGraphDataDeletionRequest(s.ctx) {
				s.deleteData()
			}

			// Ingest all available ingest tasks
			s.ingestAvailableTasks()

			// Manage time-out state progression for file upload jobs
			fileupload.ProcessStaleFileUploadJobs(s.ctx, s.db)

			// Manage nominal state transitions for file upload jobs
			ProcessIngestedFileUploadJobs(s.ctx, s.db)

			// If there are completed file upload jobs or if analysis was user-requested, perform analysis.
			if hasJobsWaitingForAnalysis, err := HasFileUploadJobsWaitingForAnalysis(s.ctx, s.db); err != nil {
				log.Errorf("Failed looking up jobs waiting for analysis: %v", err)
			} else if hasJobsWaitingForAnalysis || s.db.HasAnalysisRequest(s.ctx) {
				s.analyze()
			}

			datapipeLoopTimer.Reset(s.tickInterval)

		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Daemon) deleteData() {
	defer func() {
		_ = s.db.SetDatapipeStatus(s.ctx, model.DatapipeStatusIdle, false)
		_ = s.db.DeleteAnalysisRequest(s.ctx)
		_ = s.db.RequestAnalysis(s.ctx, "datapie")
	}()
	defer log.Measure(log.LevelInfo, "Purge Graph Data Completed")()

	if err := s.db.SetDatapipeStatus(s.ctx, model.DatapipeStatusPurging, false); err != nil {
		log.Errorf("Error setting datapipe status: %v", err)
		return
	}

	log.Infof("Begin Purge Graph Data")

	if err := s.db.CancelAllFileUploads(s.ctx); err != nil {
		log.Errorf("Error cancelling jobs during data deletion: %v", err)
	} else if err := s.db.DeleteAllIngestTasks(s.ctx); err != nil {
		log.Errorf("Error deleting ingest tasks during data deletion: %v", err)
	} else if err := DeleteCollectedGraphData(s.ctx, s.graphdb); err != nil {
		log.Errorf("Error deleting graph data: %v", err)
	}
}

func (s *Daemon) Stop(ctx context.Context) error {
	return nil
}

func (s *Daemon) clearOrphanedData() {
	if ingestTasks, err := s.db.GetAllIngestTasks(s.ctx); err != nil {
		log.Errorf("Failed fetching available file upload ingest tasks: %v", err)
	} else {
		expectedFiles := make([]string, len(ingestTasks))

		for idx, ingestTask := range ingestTasks {
			expectedFiles[idx] = ingestTask.FileName
		}

		go s.orphanedFileSweeper.Clear(s.ctx, expectedFiles)
	}
}
