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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/tasker.go -package=mocks . Tasker
package datapipe

import (
	"context"
	"errors"
	"github.com/specterops/bloodhound/src/bootstrap"
	"sync/atomic"
	"time"

	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/services/fileupload"
)

const (
	pruningInterval = time.Hour * 24
)

type Tasker interface {
	RequestAnalysis()
	GetStatus() model.DatapipeStatusWrapper
}

type Daemon struct {
	db                  database.Database
	graphdb             graph.Database
	cache               cache.Cache
	cfg                 config.Configuration
	analysisRequested   *atomic.Bool
	tickInterval        time.Duration
	status              model.DatapipeStatusWrapper
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
		analysisRequested:   &atomic.Bool{},
		orphanedFileSweeper: NewOrphanFileSweeper(NewOSFileOperations(), cfg.TempDirectory()),
		tickInterval:        tickInterval,
		status: model.DatapipeStatusWrapper{
			Status:    model.DatapipeStatusIdle,
			UpdatedAt: time.Now().UTC(),
		},
	}
}

func (s *Daemon) RequestAnalysis() {
	s.setAnalysisRequested(true)
}

func (s *Daemon) GetStatus() model.DatapipeStatusWrapper {
	return s.status
}

func (s *Daemon) getAnalysisRequested() bool {
	return s.analysisRequested.Load()
}

func (s *Daemon) setAnalysisRequested(requested bool) {
	s.analysisRequested.Store(requested)
}

func (s *Daemon) analyze() {
	// Ensure that the user-requested analysis switch is flipped back to false. This is done at the beginning of the
	// function so that any re-analysis requests are caught while analysis is in-progress.
	s.setAnalysisRequested(false)

	if s.cfg.DisableAnalysis {
		return
	}

	s.status.Update(model.DatapipeStatusAnalyzing, false)
	defer log.LogAndMeasure(log.LevelInfo, "Graph Analysis")()

	if err := RunAnalysisOperations(s.ctx, s.db, s.graphdb, s.cfg); err != nil {
		if errors.Is(err, ErrAnalysisFailed) {
			FailAnalyzedFileUploadJobs(s.ctx, s.db)
			s.status.Update(model.DatapipeStatusIdle, false)
		} else if errors.Is(err, ErrAnalysisPartiallyCompleted) {
			PartialCompleteFileUploadJobs(s.ctx, s.db)
			s.status.Update(model.DatapipeStatusIdle, true)
		}
	} else {
		CompleteAnalyzedFileUploadJobs(s.ctx, s.db)

		if entityPanelCachingFlag, err := s.db.GetFlagByKey(appcfg.FeatureEntityPanelCaching); err != nil {
			log.Errorf("Error retrieving entity panel caching flag: %v", err)
		} else {
			resetCache(s.cache, entityPanelCachingFlag.Enabled)
		}

		s.status.Update(model.DatapipeStatusIdle, true)
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
	if ingestTasks, err := s.db.GetAllIngestTasks(); err != nil {
		log.Errorf("Failed fetching available ingest tasks: %v", err)
	} else {
		s.processIngestTasks(s.ctx, ingestTasks)
	}
}

func (s *Daemon) Start() {
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
			// Ingest all available ingest tasks
			s.ingestAvailableTasks()

			// Manage time-out state progression for file upload jobs
			fileupload.ProcessStaleFileUploadJobs(s.ctx, s.db)

			// Manage nominal state transitions for file upload jobs
			ProcessIngestedFileUploadJobs(s.ctx, s.db)

			// If there are completed file upload jobs or if analysis was user-requested, perform analysis.
			if hasJobsWaitingForAnalysis, err := HasFileUploadJobsWaitingForAnalysis(s.db); err != nil {
				log.Errorf("Failed looking up jobs waiting for analysis: %v", err)
			} else if hasJobsWaitingForAnalysis || s.getAnalysisRequested() {
				s.analyze()
			}

			datapipeLoopTimer.Reset(s.tickInterval)

		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Daemon) Stop(ctx context.Context) error {
	return nil
}

func (s *Daemon) clearOrphanedData() {
	if ingestTasks, err := s.db.GetAllIngestTasks(); err != nil {
		log.Errorf("Failed fetching available file upload ingest tasks: %v", err)
	} else {
		expectedFiles := make([]string, len(ingestTasks))

		for idx, ingestTask := range ingestTasks {
			expectedFiles[idx] = ingestTask.FileName
		}

		go s.orphanedFileSweeper.Clear(s.ctx, expectedFiles)
	}
}
