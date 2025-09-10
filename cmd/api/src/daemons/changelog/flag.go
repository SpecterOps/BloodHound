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
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/dawgs/graph"
)

// featureFlagManager handles feature flag polling and cache lifecycle management.
type featureFlagManager struct {
	flagGetter   func(context.Context) (bool, int, error)
	pollInterval time.Duration

	mu    sync.RWMutex
	cache *cache
}

func newFeatureFlagManager(flagGetter func(context.Context) (bool, int, error), pollInterval time.Duration) *featureFlagManager {
	return &featureFlagManager{
		flagGetter:   flagGetter,
		pollInterval: pollInterval,
	}
}

func (s *featureFlagManager) start(ctx context.Context) {
	go s.runPoller(ctx)
}

// runPoller periodically checks the feature flag and sizes the cache accordingly.
func (s *featureFlagManager) runPoller(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	var isEnabled bool // track last seen state

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			flagEnabled, size, err := s.flagGetter(ctx)
			if err != nil {
				slog.WarnContext(ctx, "feature flag check failed", "err", err)
				continue
			}

			switch {
			case flagEnabled && !isEnabled:
				s.enable(ctx, size)
				isEnabled = true
			case !flagEnabled && isEnabled:
				s.disable(ctx)
				isEnabled = false
			}
		}
	}
}

func (s *featureFlagManager) enable(ctx context.Context, size int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.InfoContext(ctx, "enabling changelog", "cache size", size)
	cache := newCache(size)
	s.cache = cache
}

// disable resets the cache to free memory.
func (s *featureFlagManager) disable(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.InfoContext(ctx, "disabling changelog, clearing cache")
	s.cache = nil
}

// getCache returns the current cache instance, which may be nil if disabled.
func (s *featureFlagManager) getCache() *cache {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache
}

// flagGetter returns a closure that checks whether the changelog feature
// is enabled and, if so, reports the current graph size (nodes + edges).
// This allows the changelog to size its in-memory cache relative to the graph.
func flagGetter(dawgsDB graph.Database, flagProvider appcfg.GetFlagByKeyer) func(context.Context) (bool, int, error) {
	return func(ctx context.Context) (bool, int, error) {
		flag, err := flagProvider.GetFlagByKey(ctx, appcfg.FeatureChangelog)
		if err != nil {
			return false, 0, fmt.Errorf("getting changelog flag: %w", err)
		}

		if !flag.Enabled {
			return false, 0, nil
		}

		var graphSize int64
		if err := dawgsDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			if nodeCount, err := tx.Nodes().Count(); err != nil {
				return err
			} else if edgeCount, err := tx.Relationships().Count(); err != nil {
				return err
			} else {
				graphSize = nodeCount + edgeCount
				return nil
			}
		}); err != nil {
			return false, 0, fmt.Errorf("counting nodes and relationships in graph: %w ", err)
		}

		return true, int(graphSize), nil
	}
}
