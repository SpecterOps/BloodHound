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
)

const (
	pruningInterval = time.Hour * 24
)

// Pipeline defines instance methods that operate on pipeline state.
// These methods require a fully initialized Pipeline instance including
// graph and db connections. Whatever is needed by the pipe to do the work
type Pipeline interface {
	// Start provides an entrypoint into the pipeline
	Start(context.Context) error
	// IsPrimary provides a way to detect if the current instance of the pipeline is in control
	IsPrimary(context.Context, model.DatapipeStatus) (bool, context.Context)
	// PruneData provides a way to remove outdated/invalid ingest files
	PruneData(context.Context) error
	// DeleteData provides a way to handle requests for database table/graph deletion
	DeleteData(context.Context) error
	// IngestTasks provides a way to ingest previously uploaded files
	IngestTasks(context.Context) error
	// Analyze provides a way to analyze and enhance graph data, including post processing
	Analyze(context.Context) error
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

func (s *Daemon) Stop(ctx context.Context) error {
	return nil
}

// Any function can be wrapped with a datapipe lock, giving it the status. If everything locks
// the datapipe through this same wrapper, it should always defer the idle status after.
func (s *Daemon) WithDatapipeStatus(ctx context.Context, status model.DatapipeStatus, action func(context.Context) error) {

	active, pipelineContext := s.pipeline.IsPrimary(ctx, status)
	if !active {
		return
	}

	defer func() {
		if err := s.db.SetDatapipeStatus(pipelineContext, model.DatapipeStatusIdle); err != nil {
			slog.ErrorContext(pipelineContext, "Error setting datapipe status to idle", slog.String("err", err.Error()))
		}
	}()

	if err := s.db.SetDatapipeStatus(pipelineContext, status); err != nil {
		slog.ErrorContext(pipelineContext, fmt.Sprintf("Error setting datapipe status: %v", err))
		return
	}

	if err := action(pipelineContext); err != nil {
		slog.ErrorContext(pipelineContext, "Datapipe action failed", slog.String("err", err.Error()))
	}
}
