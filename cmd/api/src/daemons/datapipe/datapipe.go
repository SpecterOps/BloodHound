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

	"github.com/specterops/bloodhound/src/model"
	"pkg.specterops.io/bloodhoundad/bhe/database"
)

const (
	pruningInterval = time.Hour * 24
)

type Daemon struct {
	tickInterval time.Duration
	pipeline     Pipeline
	db           database.Database
}

func (s *Daemon) Name() string {
	return "Data Pipe Daemon"
}

func NewDaemon(pipeline Pipeline, tickInterval time.Duration, db database.Database) *Daemon {
	return &Daemon{
		db:           db,
		tickInterval: tickInterval,
		pipeline:     pipeline,
	}
}

// PipelineInterface defines methods that operate on instance state.
// These methods are not static and require a fully initialized pipeline.
type Pipeline interface {
	PruneData(context.Context)
	DeleteData(context.Context)
	IngestTasks(context.Context)
	Analyze(context.Context)
}

func (s *Daemon) Start(ctx context.Context) {
	var (
		datapipeLoopTimer = time.NewTimer(s.tickInterval)
		pruningTicker     = time.NewTicker(pruningInterval)
	)

	defer datapipeLoopTimer.Stop()
	defer pruningTicker.Stop()

	s.WithDatapipeStatus(ctx, model.DatapipeStatusPurging, s.pipeline.PruneData)

	for {
		select {
		case <-pruningTicker.C:

			s.WithDatapipeStatus(ctx, model.DatapipeStatusPurging, s.pipeline.PruneData)

		case <-datapipeLoopTimer.C:
			if s.db.HasCollectedGraphDataDeletionRequest(ctx) {
				s.WithDatapipeStatus(ctx, model.DatapipeStatusPurging, s.pipeline.DeleteData)
			}

			s.WithDatapipeStatus(ctx, model.DatapipeStatusIngesting, s.pipeline.IngestTasks)

			s.WithDatapipeStatus(ctx, model.DatapipeStatusAnalyzing, s.pipeline.Analyze)

			datapipeLoopTimer.Reset(s.tickInterval)

		case <-ctx.Done():
			return
		}
	}
}

func (s *Daemon) Stop(ctx context.Context) error {
	return nil
}

// Any function can be wrapped with a datapipe lock, giving it the status. If everything locks
// the datapipe through this same wrapper, it should always defer the idle status after.
func (s *Daemon) WithDatapipeStatus(ctx context.Context, status model.DatapipeStatus, action func(context.Context)) {

	if err := s.db.SetDatapipeStatus(ctx, status, false); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error setting datapipe status: %v", err))
		return
	}
	defer func() {
		if err := s.db.SetDatapipeStatus(ctx, model.DatapipeStatusIdle, false); err != nil {
			slog.ErrorContext(ctx, "failed to reset datapipe status", "error", err)
		}
	}()

	action(ctx)
}
