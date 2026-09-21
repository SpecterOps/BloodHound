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
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/channels"
)

const (
	pgDeadlockErrorCode = "40P01"

	defaultMaxDeadlockRetries = 3
	defaultDeadlockBaseDelay  = 50 * time.Millisecond
	defaultDeadlockMaxJitter  = 100 * time.Millisecond
)

// ingestionCoordinator manages the lifecycle of change ingestion into the graph.
// It runs a reconciliation loop that buffers incoming Change events and periodically
// flushes them into the database. Its main purpose is to apply lastseen timestamps
type ingestionCoordinator struct {
	db            graph.Database
	batchSize     int
	flushInterval time.Duration

	// Retry configuration for deadlock handling
	maxRetries     int
	retryBaseDelay time.Duration
	retryMaxJitter time.Duration

	// Channel communication
	changeC chan Change
	buffer  []Change

	// Lifecycle management
	cancel context.CancelFunc
	done   chan struct{}
}

func newIngestionCoordinator(db graph.Database, retryConfig RetryConfig) *ingestionCoordinator {
	return &ingestionCoordinator{
		db:             db,
		maxRetries:     retryConfig.MaxRetries,
		retryBaseDelay: retryConfig.BaseDelay,
		retryMaxJitter: retryConfig.MaxJitter,
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
	s.changeC = make(chan Change, batchSize*2)

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

		case change, ok := <-s.changeC:
			if !ok {
				return // channel closed
			}

			s.buffer = append(s.buffer, change)

			if len(s.buffer) >= s.batchSize {
				if err := s.flushBuffer(ctx, false); err != nil {
					slog.WarnContext(ctx, "Size-based flush failed", attr.Error(err))
				}
			}

		case <-ticker.C:
			if len(s.buffer) > 0 {
				slog.InfoContext(ctx, "Periodic flush")
				if err := s.flushBuffer(ctx, true); err != nil {
					slog.WarnContext(ctx, "Periodic flush failed", attr.Error(err))
				}
			}
		}
	}
}

// flushBuffer writes buffered lastseen updates to the database. On deadlock errors,
// it retries with exponential backoff to avoid immediately colliding with the
// competing transaction again.
func (s *ingestionCoordinator) flushBuffer(ctx context.Context, force bool) error {
	var (
		flushErr error
		retried  bool
	)

	if len(s.buffer) == 0 {
		return nil
	}

	if !force && len(s.buffer) < s.batchSize {
		return nil // not ready to flush
	}

	for attempt := range s.maxRetries {
		flushErr = s.db.BatchOperation(ctx, func(batch graph.Batch) error {
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

		if flushErr == nil {
			break
		}

		if !isDeadlockError(flushErr) {
			break
		}

		retried = true
		slog.WarnContext(ctx, "Coordinator flush hit deadlock, retrying",
			slog.Int("attempt", attempt+1),
			slog.Int("max_retries", s.maxRetries),
			slog.Int("buffer_size", len(s.buffer)),
		)

		// Jitter: randomizes the wait between 0 and retryMaxJitter.
		// Provides random periods of time between retries.
		var jitter time.Duration
		if s.retryMaxJitter > 0 {
			jitter = time.Duration(rand.Int64N(int64(s.retryMaxJitter)))
		}

		// Exponential backoff: base delay doubles each attempt (50ms, 100ms, 200ms, ...)
		backoff := s.retryBaseDelay * time.Duration(1<<attempt)
		retryDelay := backoff + jitter
		select {
		case <-ctx.Done():
			s.buffer = s.buffer[:0]
			return ctx.Err()
		case <-time.After(retryDelay):
		}
	}

	if flushErr != nil {
		slog.ErrorContext(ctx, "Coordinator flush failed",
			slog.Int("buffer_size", len(s.buffer)),
			attr.Error(flushErr),
		)
	} else if retried {
		slog.InfoContext(ctx, "Coordinator flush completed after retry",
			slog.Int("buffer_size", len(s.buffer)),
		)
	}

	// Clear buffer regardless of success to prevent unbounded growth
	s.buffer = s.buffer[:0]
	return flushErr
}

// isDeadlockError returns true if the error chain contains a PostgreSQL deadlock error (SQLSTATE 40P01).
func isDeadlockError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgDeadlockErrorCode
}

// submit enqueues a change to the internal channel.
func (s *ingestionCoordinator) submit(ctx context.Context, change Change) bool {
	return channels.Submit(ctx, s.changeC, change)
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
