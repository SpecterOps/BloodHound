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

	"github.com/specterops/bloodhound/cmd/api/src/daemons/ha"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/dawgs/graph"
)

// featureFlagManager handles feature flag polling and cache lifecycle management.
type featureFlagManager struct {
	flagGetter   func(context.Context) (bool, int, error)
	pollInterval time.Duration

	mu    sync.RWMutex
	cache *cache

	haMutex ha.HAMutex
}

// isPrimary checks if this instance is the primary API instance using the HA mutex.
// Returns true and the primary context if this instance is primary, false otherwise.
func (s *featureFlagManager) isPrimary(ctx context.Context) (bool, context.Context) {
	// Safety check: if no HA mutex is configured, assume we're always primary (single instance)
	if s.haMutex == nil {
		return true, ctx
	}

	// Try to get the HA lock to determine if we're primary
	if lockResult, err := s.haMutex.TryLock(); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed to validate HA election status: %v", err))
		return false, ctx
	} else if lockResult.IsPrimary {
		// If we are primary, return the primary context
		return true, lockResult.Context
	} else {
		// If we're not primary, we don't perform cache operations
		return false, ctx
	}
}

func newFeatureFlagManager(flagGetter func(context.Context) (bool, int, error), pollInterval time.Duration, haMutex ha.HAMutex) *featureFlagManager {
	return &featureFlagManager{
		flagGetter:   flagGetter,
		pollInterval: pollInterval,
		haMutex:      haMutex,
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
			// don't perform the following actions if we aren't the primary instance
			active, primaryCtx := s.isPrimary(ctx)
			if !active {
				continue
			}

			flagEnabled, size, err := s.flagGetter(ctx)
			if err != nil {
				slog.WarnContext(ctx, "feature flag check failed", "err", err)
				continue
			}

			switch {
			case flagEnabled && !isEnabled:
				s.enable(primaryCtx, size)
				isEnabled = true
			case !flagEnabled && isEnabled:
				s.disable(primaryCtx)
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

// clearCache forcibly clears the cache regardless of feature flag state.
// This is used during graph data deletion to ensure cache consistency.
// Only the primary instance should clear the cache in HA deployments.
func (s *featureFlagManager) clearCache(ctx context.Context) {
	// Check if we're the primary instance before clearing cache
	isPrimary, primaryCtx := s.isPrimary(ctx)
	if !isPrimary {
		slog.InfoContext(ctx, "skipping cache clear - not primary instance")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache != nil {
		slog.InfoContext(primaryCtx, "forcibly clearing changelog cache due to graph data deletion")
		s.cache.clear()
	} else {
		slog.InfoContext(primaryCtx, "changelog cache already cleared (feature disabled)")
	}
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
