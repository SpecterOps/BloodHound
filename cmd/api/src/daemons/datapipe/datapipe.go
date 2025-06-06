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

	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/upload"
	"github.com/specterops/dawgs/graph"
)

const (
	pruningInterval = time.Hour * 24
)

type Daemon struct {
	db           database.Database
	tickInterval time.Duration
	pipeline     Pipeline
	ctx          context.Context
}

func (s *Daemon) Name() string {
	return "Data Pipe Daemon"
}

func NewDaemon(ctx context.Context, cfg config.Configuration, db database.Database, graphdb graph.Database, cache cache.Cache, tickInterval time.Duration, ingestSchema upload.IngestSchema) *Daemon {
	return &Daemon{
		tickInterval: tickInterval,
		pipeline:     *NewPipeline(ctx, cfg, db, graphdb, cache, ingestSchema),
	}
}

func (s *Daemon) Start(ctx context.Context) {
	var (
		datapipeLoopTimer = time.NewTimer(s.tickInterval)
		pruningTicker     = time.NewTicker(pruningInterval)
	)

	defer datapipeLoopTimer.Stop()
	defer pruningTicker.Stop()

	s.WithDatapipeStatus(model.DatapipeStatusPurging, s.pipeline.clearOrphanedData)

	for {
		select {
		case <-pruningTicker.C:

			s.WithDatapipeStatus(model.DatapipeStatusPurging, s.pipeline.clearOrphanedData)

		case <-datapipeLoopTimer.C:
			if s.db.HasCollectedGraphDataDeletionRequest(s.ctx) {
				s.WithDatapipeStatus(model.DatapipeStatusPurging, s.pipeline.deleteData)
			}

			s.WithDatapipeStatus(model.DatapipeStatusIngesting, s.pipeline.IngestTasks)

			s.WithDatapipeStatus(model.DatapipeStatusAnalyzing, s.pipeline.analyze)

			datapipeLoopTimer.Reset(s.tickInterval)

		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Daemon) Stop(ctx context.Context) error {
	return nil
}

// Any function can be wrapped with a datapipe lock, giving it the status. If everything locks
// the datapipe through this same wrapper, it should always defer the idle status after.
func (s *Daemon) WithDatapipeStatus(status model.DatapipeStatus, action func()) {

	if err := s.db.SetDatapipeStatus(s.ctx, status, false); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error setting datapipe status: %v", err))
		return
	}
	defer func() {
		if err := s.db.SetDatapipeStatus(s.ctx, model.DatapipeStatusIdle, false); err != nil {
			slog.ErrorContext(s.ctx, "failed to reset datapipe status", "error", err)
		}
	}()

	action()
}
