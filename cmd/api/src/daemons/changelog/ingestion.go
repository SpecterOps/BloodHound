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

// manages buffering and flushing of changes to graph
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
	s.flushInterval = flushInterval
	s.buffer = make([]Change, 0, batchSize)

	s.writerC, s.readerC = channels.BufferedPipe[Change](cctx)

	go func() {
		defer close(s.done)
		s.runIngestionLoop(cctx)
	}()
}

func (s *ingestionCoordinator) runIngestionLoop(ctx context.Context) {
	idle := time.NewTimer(s.flushInterval)
	idle.Stop()

	defer func() {
		idle.Stop()
		// Final flush on shutdown - use fresh context since original may be cancelled
		flushCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.flushBuffer(flushCtx, true)
		slog.InfoContext(flushCtx, "ending changelog loop")
	}()

	slog.InfoContext(ctx, "starting changelog loop")

	for {
		select {
		case <-ctx.Done():
			return

		case change, ok := <-s.readerC:
			if !ok {
				return // channel closed
			}

			s.buffer = append(s.buffer, change)

			// Size-based flush
			if len(s.buffer) >= s.batchSize {
				if err := s.flushBuffer(ctx, false); err != nil {
					slog.WarnContext(ctx, "size-based flush failed", "err", err)
				}
			}

			// Reset idle timer
			idle.Reset(s.flushInterval)

		case <-idle.C:
			slog.InfoContext(ctx, "idle flush", "timestamp", time.Now())
			if err := s.flushBuffer(ctx, true); err != nil { // force flush on idle
				slog.WarnContext(ctx, "idle flush failed", "err", err)
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
				slog.DebugContext(ctx, "skipping nil change")
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
