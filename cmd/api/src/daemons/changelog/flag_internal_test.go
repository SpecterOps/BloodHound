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

	"github.com/specterops/bloodhound/cmd/api/src/daemons/ha"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagManager(t *testing.T) {
	t.Run("enable creates cache", func(t *testing.T) {
		var (
			manager = newFeatureFlagManager(nil, time.Second, ha.NewDummyHA())
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
			manager = newFeatureFlagManager(nil, time.Second, ha.NewDummyHA())
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
			manager       = newFeatureFlagManager(nil, time.Second, ha.NewDummyHA())
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
			manager = newFeatureFlagManager(flagGetter, 10*time.Millisecond, ha.NewDummyHA())
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
			manager = newFeatureFlagManager(flagGetter, 10*time.Millisecond, ha.NewDummyHA())
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
			manager = newFeatureFlagManager(flagGetter, 10*time.Millisecond, ha.NewDummyHA())
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
			manager = newFeatureFlagManager(flagGetter, 10*time.Millisecond, ha.NewDummyHA())
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

// mockHA implements the HA interface for testing
type mockHA struct {
	isPrimary bool
	lockError error
}

func (m *mockHA) TryLock() (ha.LockResult, error) {
	if m.lockError != nil {
		return ha.LockResult{}, m.lockError
	}
	return ha.LockResult{
		Context:   context.Background(),
		IsPrimary: m.isPrimary,
	}, nil
}

func TestFeatureFlag_HA(t *testing.T) {
	t.Run("Primary instance enables cache", func(t *testing.T) {
		var (
			flagEnabled = false
			mockHAMutex = &mockHA{isPrimary: true}
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				return flagEnabled, 1000, nil
			}
			manager     = newFeatureFlagManager(flagGetter, 50*time.Millisecond, mockHAMutex)
			ctx, cancel = context.WithCancel(context.Background())
		)
		defer cancel()

		// Start the manager
		manager.start(ctx)

		// Initially cache should be nil
		require.Nil(t, manager.getCache())

		// Enable the flag
		flagEnabled = true

		// Wait for polling cycle
		time.Sleep(100 * time.Millisecond)

		// Cache should now be enabled
		cache := manager.getCache()
		require.NotNil(t, cache)
		require.Equal(t, 1000, cache.getCapacity())
	})

	t.Run("Non-primary instance skips cache operations", func(t *testing.T) {
		var (
			flagEnabled = true
			mockHAMutex = &mockHA{isPrimary: false}
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				return flagEnabled, 1000, nil
			}
			manager     = newFeatureFlagManager(flagGetter, 50*time.Millisecond, mockHAMutex)
			ctx, cancel = context.WithCancel(context.Background())
		)
		defer cancel()

		// Start the manager
		manager.start(ctx)

		// Wait for polling cycle
		time.Sleep(100 * time.Millisecond)

		// Cache should remain nil since we're not primary
		require.Nil(t, manager.getCache())
	})

	t.Run("Nil HA mutex defaults to primary", func(t *testing.T) {
		var (
			flagEnabled = false
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				return flagEnabled, 1000, nil
			}
			manager     = newFeatureFlagManager(flagGetter, 50*time.Millisecond, nil)
			ctx, cancel = context.WithCancel(context.Background())
		)
		defer cancel()

		// Start the manager
		manager.start(ctx)

		// Initially cache should be nil
		require.Nil(t, manager.getCache())

		// Enable the flag
		flagEnabled = true

		// Wait for polling cycle
		time.Sleep(100 * time.Millisecond)

		// Cache should be enabled since nil HA mutex defaults to primary
		cache := manager.getCache()
		require.NotNil(t, cache)
		require.Equal(t, 1000, cache.getCapacity())
	})
}

func TestClearCache_HA(t *testing.T) {
	t.Run("Primary instance clears cache", func(t *testing.T) {
		var (
			mockHAMutex = &mockHA{isPrimary: true}
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				return true, 1000, nil
			}
			manager = newFeatureFlagManager(flagGetter, time.Hour, mockHAMutex)
			ctx     = context.Background()
		)

		// Enable cache first
		manager.enable(ctx, 1000)
		require.NotNil(t, manager.getCache())

		// Clear cache - should work since we're primary
		manager.clearCache(ctx)

		// Cache should still exist but be empty (clear() doesn't nil the cache)
		cache := manager.getCache()
		require.NotNil(t, cache)
		require.Equal(t, 0, len(cache.data))
	})

	t.Run("Non-primary instance skips cache clear", func(t *testing.T) {
		var (
			mockHAMutex = &mockHA{isPrimary: false}
			flagGetter  = func(ctx context.Context) (bool, int, error) {
				return true, 1000, nil
			}
			manager = newFeatureFlagManager(flagGetter, time.Hour, mockHAMutex)
			ctx     = context.Background()
		)

		// Enable cache first (simulate primary enabling it)
		manager.enable(ctx, 1000)
		cache := manager.getCache()
		require.NotNil(t, cache)

		// Add some data to cache
		cache.data[23] = 45
		require.Equal(t, 1, len(cache.data))

		// Clear cache - should be skipped since we're not primary
		manager.clearCache(ctx)

		// Cache should still have data since clear was skipped
		require.Equal(t, 1, len(cache.data))
	})
}

func TestDummyHA(t *testing.T) {
	dummy := ha.NewDummyHA()

	lockResult, err := dummy.TryLock()
	require.NoError(t, err)
	require.True(t, lockResult.IsPrimary)
	require.NotNil(t, lockResult.Context)
}
