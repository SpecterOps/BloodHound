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
	"fmt"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/job"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/dawgs/graph"
)

const (
	pruningInterval = time.Hour * 24
)

// Pipeline defines instance methods that operate on pipeline state.
// These methods require a fully initialized Pipeline instance including
// graph and db connections. Whatever is neeeded by the pipe to do the work
type Pipeline interface {
	PruneData(context.Context) bool
	DeleteData(context.Context) bool
	IngestTasks(context.Context) bool
	Analyze(context.Context) bool
	Start(context.Context) bool
	IsActive(context.Context, model.DatapipeStatus) (bool, context.Context)
}

type Daemon struct {
	startDelay   time.Duration
	tickInterval time.Duration
	pipeline     Pipeline
	db           database.Database
}

func (s *Daemon) Name() string {
	return "Data Pipe Daemon"
}

func NewDaemon(pipeline Pipeline, startDelay time.Duration, tickInterval time.Duration, db database.Database) *Daemon {
	return &Daemon{
		db:           db,
		tickInterval: tickInterval,
		pipeline:     pipeline,
		startDelay:   startDelay,
	}
}

func (s *Daemon) Start(ctx context.Context) {
	var (
		datapipeLoopTimer = time.NewTimer(s.startDelay)
		pruningTicker     = time.NewTicker(pruningInterval)
	)

	defer datapipeLoopTimer.Stop()
	defer pruningTicker.Stop()

	s.WithDatapipeStatus(ctx, model.DatapipeStatusStarting, s.pipeline.Start)

	for {
		select {
		case <-pruningTicker.C:

			s.WithDatapipeStatus(ctx, model.DatapipeStatusPruning, s.pipeline.PruneData)

		case <-datapipeLoopTimer.C:
			s.WithDatapipeStatus(ctx, model.DatapipeStatusPurging, s.pipeline.DeleteData)

			s.WithDatapipeStatus(ctx, model.DatapipeStatusIngesting, s.pipeline.IngestTasks)

			s.WithDatapipeStatus(ctx, model.DatapipeStatusAnalyzing, s.pipeline.Analyze)

			datapipeLoopTimer.Reset(s.tickInterval)

		case <-ctx.Done():
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
	defer measure.Measure(slog.LevelInfo, "Purge Graph Data Completed")()

	if err := s.db.SetDatapipeStatus(s.ctx, model.DatapipeStatusPurging, false); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error setting datapipe status: %v", err))
		return
	}

	slog.Info("Begin Purge Graph Data")

	if err := s.db.CancelAllIngestJobs(s.ctx); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error cancelling jobs during data deletion: %v", err))
	} else if err := s.db.DeleteAllIngestTasks(s.ctx); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error deleting ingest tasks during data deletion: %v", err))
	} else if err := DeleteCollectedGraphData(s.ctx, s.graphdb); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error deleting graph data: %v", err))
	}
}

func (s *Daemon) Stop(ctx context.Context) error {
	return nil
}

// Any function can be wrapped with a datapipe lock, giving it the status. If everything locks
// the datapipe through this same wrapper, it should always defer the idle status after.
func (s *Daemon) WithDatapipeStatus(ctx context.Context, status model.DatapipeStatus, action func(context.Context) bool) {
	active, pipelineContext := s.pipeline.IsActive(ctx, status)
	successful := false
	if !active {
		return
	}

	if err := s.db.SetDatapipeStatus(pipelineContext, status, false); err != nil {
		slog.ErrorContext(pipelineContext, fmt.Sprintf("Error setting datapipe status: %v", err))
		return
	}
	defer func() {
		// The set datapipe status has a sneaky update in it's depths. This flag being set true sets the
		// last_analyzed time on the datapipe status. This should only happen after analyzsis (no other steps)
		// and only if the analysis is successful
		flagAsAnalyzedComplete := (status == model.DatapipeStatusAnalyzing && successful)
		if err := s.db.SetDatapipeStatus(pipelineContext, model.DatapipeStatusIdle, flagAsAnalyzedComplete); err != nil {
			slog.ErrorContext(pipelineContext, "failed to reset datapipe status", "error", err)
		}
	}()

	successful = action(pipelineContext)
}
