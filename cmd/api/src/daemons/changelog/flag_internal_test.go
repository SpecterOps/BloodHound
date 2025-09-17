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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFeatureFlagManager(t *testing.T) {
	t.Run("enable creates cache", func(t *testing.T) {
		var (
			manager = newFeatureFlagManager(nil, time.Second)
			ctx     = context.Background()
		)

		manager.enable(ctx, 1000)

		cache := manager.getCache()
		require.NotNil(t, cache)
		require.NotNil(t, cache.data)
		require.Equal(t, 0, len(cache.data)) // Should start empty
	})

	t.Run("disable clears cache", func(t *testing.T) {
		var (
			manager = newFeatureFlagManager(nil, time.Second)
			ctx     = context.Background()
		)

		// First enable to create cache
		manager.enable(ctx, 1000)
		require.NotNil(t, manager.getCache())

		// Then disable
		manager.disable(ctx)
		require.Nil(t, manager.getCache())
	})

	t.Run("getCache is thread-safe", func(t *testing.T) {
		var (
			manager       = newFeatureFlagManager(nil, time.Second)
			ctx           = context.Background()
			wg            sync.WaitGroup
			numGoroutines = 10
		)

		// Start multiple goroutines that enable/disable cache
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				if id%2 == 0 {
					manager.enable(ctx, 1000)
				} else {
					manager.disable(ctx)
				}
			}(i)
		}

		// Start multiple goroutines that read cache
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// This should not panic or race
				_ = manager.getCache()
			}()
		}

		wg.Wait()
		// Test passes if no race conditions or panics occur
	})
}

func TestFeatureFlagManagerPoller(t *testing.T) {
	t.Run("runPoller enables cache when flag becomes enabled", func(t *testing.T) {
		var (
			ctx, cancel = context.WithCancel(context.Background())
			enabled     = false
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				return enabled, 10, nil
			}
			manager = newFeatureFlagManager(flagGetter, 10*time.Millisecond)
		)
		defer cancel()

		// Start poller
		go manager.runPoller(ctx)

		// Initially cache should be nil
		require.Nil(t, manager.getCache())

		// Enable flag
		enabled = true

		// Wait for poller to pick up the change
		require.Eventually(t, func() bool {
			return manager.getCache() != nil
		}, 100*time.Millisecond, 5*time.Millisecond)

		cache := manager.getCache()
		require.NotNil(t, cache)
		require.Equal(t, 0, len(cache.data)) // Should start empty
	})

	t.Run("runPoller disables cache when flag becomes disabled", func(t *testing.T) {
		var (
			ctx, cancel = context.WithCancel(context.Background())
			enabled     = true
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				return enabled, 1500, nil
			}
			manager = newFeatureFlagManager(flagGetter, 10*time.Millisecond)
		)
		defer cancel()

		// Start poller
		go manager.runPoller(ctx)

		// Wait for initial enable
		require.Eventually(t, func() bool {
			return manager.getCache() != nil
		}, 100*time.Millisecond, 5*time.Millisecond)

		// Disable flag
		enabled = false

		// Wait for poller to pick up the change
		require.Eventually(t, func() bool {
			return manager.getCache() == nil
		}, 100*time.Millisecond, 5*time.Millisecond)
	})

	t.Run("runPoller handles flag getter errors gracefully", func(t *testing.T) {
		var (
			ctx, cancel = context.WithCancel(context.Background())
			callCount   = 0
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				callCount++
				if callCount <= 2 {
					return false, 0, errors.New("flag service unavailable")
				}
				return true, 3000, nil
			}
			manager = newFeatureFlagManager(flagGetter, 10*time.Millisecond)
		)
		defer cancel()

		// Start poller
		go manager.runPoller(ctx)

		// Should eventually succeed after errors
		require.Eventually(t, func() bool {
			return manager.getCache() != nil
		}, 200*time.Millisecond, 10*time.Millisecond)

		require.True(t, callCount >= 3) // Should have retried after errors
	})

	t.Run("start method launches poller in background", func(t *testing.T) {
		var (
			ctx, cancel = context.WithCancel(context.Background())
			enabled     = false
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				return enabled, 1000, nil
			}
			manager = newFeatureFlagManager(flagGetter, 10*time.Millisecond)
		)
		defer cancel()

		// Start should return immediately
		manager.start(ctx)

		// Initially cache should be nil
		require.Nil(t, manager.getCache())

		// Enable flag
		enabled = true

		// Poller should pick up the change in background
		require.Eventually(t, func() bool {
			return manager.getCache() != nil
		}, 100*time.Millisecond, 5*time.Millisecond)
	})
}
