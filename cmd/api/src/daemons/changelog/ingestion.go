// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package changelog

import (
	"context"
	"log/slog"
	"time"

	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/channels"
)

// ingestionCoordinator manages the lifecycle of change ingestion into the graph.
// It runs a reconciliation loop that buffers incoming Change events and periodically
// flushes them into the database. Its main purpose is to apply lastseen timestamps
type ingestionCoordinator struct {
	db            graph.Database
	batchSize     int
	flushInterval time.Duration

	// Channel communication
	writerC chan<- Change
	readerC <-chan Change

	buffer []Change

	// Lifecycle management
	cancel context.CancelFunc
	done   chan struct{}
}

func newIngestionCoordinator(db graph.Database) *ingestionCoordinator {
	return &ingestionCoordinator{
		db: db,
	}
}

func (s *ingestionCoordinator) start(ctx context.Context, batchSize int, flushInterval time.Duration) {
	var cctx context.Context
	cctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})

	s.batchSize = batchSize
	s.buffer = make([]Change, 0, batchSize)

	if flushInterval <= 0 {
		slog.WarnContext(cctx, "Invalid flush interval; defaulting to 5s", slog.Duration("requested_interval", flushInterval))
		flushInterval = 5 * time.Second
	}
	s.flushInterval = flushInterval

	s.writerC, s.readerC = channels.BufferedPipe[Change](cctx)

	go func() {
		defer close(s.done)
		s.runIngestionLoop(cctx)
	}()
}

func (s *ingestionCoordinator) runIngestionLoop(ctx context.Context) {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// final flush on shutdown
			flushCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			s.flushBuffer(flushCtx, true)
			slog.InfoContext(flushCtx, "Ending changelog loop")
			cancel()
			return

		case change, ok := <-s.readerC:
			if !ok {
				return // channel closed
			}

			s.buffer = append(s.buffer, change)

			if len(s.buffer) >= s.batchSize {
				if err := s.flushBuffer(ctx, false); err != nil {
					slog.WarnContext(ctx, "Size-based flush failed", slog.String("err", err.Error()))
				}
			}

		case <-ticker.C:
			if len(s.buffer) > 0 {
				slog.InfoContext(ctx, "Periodic flush")
				if err := s.flushBuffer(ctx, true); err != nil {
					slog.WarnContext(ctx, "Periodic flush failed", slog.String("err", err.Error()))
				}
			}
		}
	}
}

func (s *ingestionCoordinator) flushBuffer(ctx context.Context, force bool) error {
	if len(s.buffer) == 0 {
		return nil
	}

	if !force && len(s.buffer) < s.batchSize {
		return nil // not ready to flush
	}

	err := s.db.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, change := range s.buffer {
			if change == nil {
				slog.DebugContext(ctx, "Skipping nil change")
				continue
			}
			if err := change.Apply(batch); err != nil {
				return err
			}
		}
		return nil
	})

	// Clear buffer regardless of success (same as original logic)
	s.buffer = s.buffer[:0]
	return err
}

// submit enqueues a change to the internal channel.
func (s *ingestionCoordinator) submit(ctx context.Context, change Change) bool {
	return channels.Submit(ctx, s.writerC, change)
}

// stop gracefully shuts down the coordinator.
func (s *ingestionCoordinator) stop(ctx context.Context) error {
	if s.cancel == nil {
		return nil
	}

	s.cancel()

	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
