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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/mock.go -package=mocks . Tasker
package datapipe

import (
	"context"
	"github.com/specterops/bloodhound/src/bootstrap"
	"os"
	"path/filepath"
	"sync"
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
	db                     database.Database
	graphdb                graph.Database
	cache                  cache.Cache
	cfg                    config.Configuration
	analysisRequested      *atomic.Bool
	tickInterval           time.Duration
	status                 model.DatapipeStatusWrapper
	ctx                    context.Context
	clearOrphanedFilesLock *sync.Mutex
}

func (s *Daemon) Name() string {
	return "Data Pipe Daemon"
}

func NewDaemon(ctx context.Context, cfg config.Configuration, connections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch], cache cache.Cache, tickInterval time.Duration) *Daemon {
	return &Daemon{
		db:                     connections.RDMS,
		graphdb:                connections.Graph,
		cache:                  cache,
		cfg:                    cfg,
		ctx:                    ctx,
		analysisRequested:      &atomic.Bool{},
		clearOrphanedFilesLock: &sync.Mutex{},
		tickInterval:           tickInterval,
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
	log.LogAndMeasure(log.LevelInfo, "Graph Analysis")()

	if err := RunAnalysisOperations(s.ctx, s.db, s.graphdb, s.cfg); err != nil {
		log.Errorf("Graph analysis failed: %v", err)
		s.failJobsUnderAnalysis()

		s.status.Update(model.DatapipeStatusIdle, false)
	} else {
		s.completeJobsUnderAnalysis()

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

func (s *Daemon) getNumJobsWaitingForAnalysis() (int, error) {
	numJobsWaitingForAnalysis := 0

	if fileUploadJobsUnderAnalysis, err := s.db.GetFileUploadJobsWithStatus(model.JobStatusAnalyzing); err != nil {
		return 0, err
	} else {
		numJobsWaitingForAnalysis += len(fileUploadJobsUnderAnalysis)
	}

	return numJobsWaitingForAnalysis, nil
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
			fileupload.ProcessStaleFileUploadJobs(s.db)

			// Manage nominal state transitions for file upload jobs
			s.processIngestedFileUploadJobs()

			// If there are completed file upload jobs or if analysis was user-requested, perform analysis.
			if numJobsWaitingForAnalysis, err := s.getNumJobsWaitingForAnalysis(); err != nil {
				log.Errorf("Failed looking up jobs waiting for analysis: %v", err)
			} else if numJobsWaitingForAnalysis > 0 || s.getAnalysisRequested() {
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
	// Only allow one background thread to run for clearing orphaned data
	if !s.clearOrphanedFilesLock.TryLock() {
		return
	}

	// Release the lock once finished
	defer s.clearOrphanedFilesLock.Unlock()

	relativeTmpDir := s.cfg.TempDirectory()

	if orphanFiles, err := os.ReadDir(s.cfg.TempDirectory()); err != nil {
		log.Errorf("Failed fetching available files: %v", err)
	} else if ingestTasks, err := s.db.GetAllIngestTasks(); err != nil {
		log.Errorf("Failed fetching available ingest tasks: %v", err)
	} else {
		for _, ingestTask := range ingestTasks {
			for idx, orphanFile := range orphanFiles {
				if ingestTask.FileName == filepath.Join(relativeTmpDir, orphanFile.Name()) {
					orphanFiles = append(orphanFiles[:idx], orphanFiles[idx+1:]...)
				}
			}
		}

		for _, orphanFile := range orphanFiles {
			fullPath := filepath.Join(relativeTmpDir, orphanFile.Name())

			if err := os.RemoveAll(fullPath); err != nil {
				log.Errorf("Failed removing file: %s", fullPath)
			}

			// Check to see if we need to shutdown after every file deletion
			select {
			case <-s.ctx.Done():
				return
			default:
			}
		}

		if len(orphanFiles) > 0 {
			log.Infof("Finished removing %d orphaned ingest files", len(orphanFiles))
		}
	}
}
