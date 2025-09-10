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

	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/channels"
)

// Changelog is a long-running daemon that manages change deduplication
// and buffering for graph ingestion.
type Changelog struct {
	db      graph.Database
	loop    loop
	options Options

	// Feature flag management
	flagManager *featureFlagManager

	// clean shutdown
	cancel context.CancelFunc
	done   chan struct{}
}

// Cache provides backward compatibility access to the cache instance.
// This property delegates to the featureFlagManager's cache.
func (s *Changelog) Cache() *cache {
	return s.flagManager.getCache()
}

type Options struct {
	BatchSize     int
	FlushInterval time.Duration // interval for force flush when ingest is quiet (clears out leftovers in buffer)
	PollInterval  time.Duration // interval for Feature Flag check
}

func DefaultOptions() Options {
	return Options{
		BatchSize:     1_000,
		FlushInterval: 5 * time.Second,
		PollInterval:  10 * time.Second,
	}
}

func NewChangelog(dawgsDB graph.Database, flagProvider appcfg.GetFlagByKeyer, opts Options) *Changelog {
	flagManager := newFeatureFlagManager(flagGetter(dawgsDB, flagProvider), opts.PollInterval)
	return &Changelog{
		flagManager: flagManager,
		options:     opts,
		db:          dawgsDB,
	}
}

// Start begins a long-running loop that buffers and flushes node/edge updates
func (s *Changelog) Start(ctx context.Context) {
	var cctx context.Context
	cctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})

	s.loop = newLoop(cctx, newDBFlusher(s.db), s.options.BatchSize, s.options.FlushInterval)

	go func() {
		defer close(s.done)
		// this loop owns updating the lastseen property
		s.runLoop(cctx)
	}()

	// Start feature flag polling
	s.flagManager.start(cctx)
}

// runLoop owns the changelog’s inner ingestion loop.
func (s *Changelog) runLoop(ctx context.Context) {
	if err := s.loop.start(ctx); err != nil {
		slog.ErrorContext(ctx, "changelog loop exited with error", "err", err)
	}
}

func (s *Changelog) Stop(ctx context.Context) error {
	if s.cancel == nil {
		return nil // never started
	}

	// tell loop to stop
	s.cancel()

	// wait until loop exits or context times out
	select {
	case <-s.done:
		slog.Info("changelog shutdown complete")
		return nil
	case <-ctx.Done():
		return ctx.Err() // caller's timeout
	}
}

func (s *Changelog) Name() string {
	return "Changelog Daemon"
}

// InitCacheForTest initializes the cache for testing purposes without feature flag polling.
func (s *Changelog) InitCacheForTest(ctx context.Context) {
	s.flagManager.initCacheForTest(ctx)
}

func (s *Changelog) GetStats() cacheStats {
	c := s.flagManager.getCache()
	if c == nil { // cache may be nil when feature is disabled.
		return cacheStats{}
	}
	return c.getStats()
}

func (s *Changelog) FlushStats() {
	c := s.flagManager.getCache()
	if c == nil { // cache may be nil when feature is disabled.
		return
	}

	stats := c.resetStats()
	slog.Info("changelog metrics",
		"hits", stats.Hits,
		"misses", stats.Misses,
	)
}

func (s *Changelog) ResolveChange(change Change) (bool, error) {
	c := s.flagManager.getCache()
	if c == nil { // treat as pass-through when disabled.
		return true, nil
	}

	return c.shouldSubmit(change)
}

func (s *Changelog) Submit(ctx context.Context, change Change) bool {
	return channels.Submit(ctx, s.loop.writerC, change)
}
