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

	"github.com/specterops/bloodhound/cmd/api/src/daemons/ha"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/dawgs/graph"
)

// Changelog is a long-running daemon that manages change deduplication
// and buffering for graph ingestion. It coordinates between feature flag
// management and ingestion processing.
type Changelog struct {
	options     Options
	flagManager *featureFlagManager
	coordinator *ingestionCoordinator
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
	// Use dummy HA implementation for BHCE (always primary)
	flagManager := newFeatureFlagManager(flagGetter(dawgsDB, flagProvider), opts.PollInterval, ha.NewDummyHA())
	coordinator := newIngestionCoordinator(dawgsDB)

	return &Changelog{
		flagManager: flagManager,
		coordinator: coordinator,
		options:     opts,
	}
}

// NewChangelogWithHA creates a changelog with a real HA implementation for high-availability deployments.
func NewChangelogWithHA(dawgsDB graph.Database, flagProvider appcfg.GetFlagByKeyer, opts Options, haMutex ha.HAMutex) *Changelog {
	flagManager := newFeatureFlagManager(flagGetter(dawgsDB, flagProvider), opts.PollInterval, haMutex)
	coordinator := newIngestionCoordinator(dawgsDB)

	return &Changelog{
		flagManager: flagManager,
		coordinator: coordinator,
		options:     opts,
	}
}

// Start begins a long-running loop that buffers and flushes node/edge updates
func (s *Changelog) Start(ctx context.Context) {
	// Start ingestion coordination
	s.coordinator.start(ctx, s.options.BatchSize, s.options.FlushInterval)

	// Start feature flag polling
	s.flagManager.start(ctx)
}

func (s *Changelog) Stop(ctx context.Context) error {
	return s.coordinator.stop(ctx)
}

func (s *Changelog) Name() string {
	return "Changelog Daemon"
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
	return s.coordinator.submit(ctx, change)
}

// ClearCache forcibly clears the changelog cache, typically called during
// graph data deletion to ensure cache consistency. This is safe to call
// whether the changelog is enabled or disabled.
func (s *Changelog) ClearCache(ctx context.Context) {
	s.flagManager.clearCache(ctx)
}
