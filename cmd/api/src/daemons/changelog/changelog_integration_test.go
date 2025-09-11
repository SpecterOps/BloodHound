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

//go:build integration

package changelog

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

// testChangelogConfig provides consistent changelog configuration for tests
type testChangelogConfig struct {
	BatchSize     int
	FlushInterval time.Duration
	PollInterval  time.Duration
}

func defaultTestConfig() testChangelogConfig {
	return testChangelogConfig{
		BatchSize:     3,
		FlushInterval: 100 * time.Millisecond,
		PollInterval:  50 * time.Millisecond,
	}
}

// changelogHarness encapsulates common changelog test setup and teardown
type changelogHarness struct {
	suite     *IntegrationTestSuite
	changelog *Changelog
	ctx       context.Context
	t         *testing.T
}

func setupChangelogTest(t *testing.T, config testChangelogConfig) *changelogHarness {
	suite := setupIntegrationTest(t)

	changelog := NewChangelog(suite.GraphDB, suite.BloodhoundDB, Options{
		BatchSize:     config.BatchSize,
		FlushInterval: config.FlushInterval,
		PollInterval:  config.PollInterval,
	})

	return &changelogHarness{
		suite:     &suite,
		changelog: changelog,
		ctx:       suite.Context,
		t:         t,
	}
}

func (s *changelogHarness) enableAndStart() {
	s.suite.enableChangelog(s.t)
	s.changelog.Start(s.ctx)
}

func (s *changelogHarness) waitForCacheInit() {
	time.Sleep(200 * time.Millisecond)
}

func (s *changelogHarness) enableChangelog() {
	s.suite.enableChangelog(s.t)
}

func (s *changelogHarness) disableChangelog() {
	s.suite.disableChangelog(s.t)
}

func (s *changelogHarness) close() {
	s.changelog.Stop(context.Background())
	teardownIntegrationTest(s.t, s.suite)
}

// createTestChange creates a standardized test change
func createTestChange(objectID string, kind string, extraProps map[string]any) Change {
	props := map[string]any{
		"objectid": objectID,
		"lastseen": time.Now(),
	}

	// Add any extra properties
	for k, v := range extraProps {
		props[k] = v
	}

	return NewNodeChange(objectID, graph.Kinds{graph.StringKind(kind)},
		graph.NewProperties().SetAll(props))
}

// assertChangeSubmission tests a change and verifies the expected submission result
func (s *changelogHarness) assertChangeSubmission(change Change, expectedSubmit bool, description string) {
	shouldSubmit, err := s.changelog.ResolveChange(change)
	require.NoError(s.t, err, "ResolveChange failed for %s", description)
	require.Equal(s.t, expectedSubmit, shouldSubmit, "Unexpected submission result for %s", description)
}

// submitChange submits a change and verifies it was accepted
func (s *changelogHarness) submitChange(change Change) {
	submitted := s.changelog.Submit(s.ctx, change)
	require.True(s.t, submitted, "Change submission was rejected")
}

// assertNodesExistInDB verifies that nodes with the given objectIDs exist in the database
func (s *changelogHarness) assertNodesExistInDB(objectIDs []string, expectedCount int) {
	var nodeCount int
	err := s.suite.GraphDB.ReadTransaction(s.ctx, func(tx graph.Transaction) error {
		criteria := make([]graph.Criteria, len(objectIDs))
		for i, id := range objectIDs {
			criteria[i] = query.Equals(query.NodeProperty("objectid"), id)
		}

		return tx.Nodes().Filterf(func() graph.Criteria {
			return query.Or(criteria...)
		}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for range cursor.Chan() {
				nodeCount++
			}
			return cursor.Error()
		})
	})
	require.NoError(s.t, err, "Database query failed")
	require.Equal(s.t, expectedCount, nodeCount, "Unexpected number of nodes in database")
}

func TestChangelogIntegration(t *testing.T) {
	t.Parallel()

	t.Run("coordinator and flag manager startup coordination", func(t *testing.T) {
		harness := setupChangelogTest(t, defaultTestConfig())
		defer harness.close()

		harness.enableAndStart()

		// Initially cache should be nil until flag manager polls
		change1 := createTestChange("test1", "User", nil)
		harness.assertChangeSubmission(change1, true, "initial change before cache init")

		// Wait for flag manager to initialize cache
		harness.waitForCacheInit()

		// Now cache should be available and working
		change2 := createTestChange("test2", "User", nil)
		harness.assertChangeSubmission(change2, true, "first submission after cache init")

		// Same change should be deduplicated
		harness.assertChangeSubmission(change2, false, "duplicate change")
	})

	t.Run("feature flag enable/disable affects cache behavior", func(t *testing.T) {
		harness := setupChangelogTest(t, defaultTestConfig())
		defer harness.close()

		harness.enableAndStart()
		harness.waitForCacheInit()

		change := createTestChange("toggle-test", "User", nil)

		// With cache enabled, first submission should work
		harness.assertChangeSubmission(change, true, "first submission with cache enabled")

		// Second submission should be deduplicated
		harness.assertChangeSubmission(change, false, "duplicate with cache enabled")

		// Disable feature flag
		harness.disableChangelog()
		harness.waitForCacheInit() // Wait for flag manager to poll and disable cache

		// With cache disabled, should pass-through (always true)
		harness.assertChangeSubmission(change, true, "submission with cache disabled")

		// Re-enable feature flag
		harness.enableChangelog()
		harness.waitForCacheInit() // Wait for flag manager to poll and re-enable cache

		// Cache should be fresh after re-enabling, so same change should be submittable again
		harness.assertChangeSubmission(change, true, "submission after cache re-enabled")
	})

	t.Run("end-to-end flow with real database operations", func(t *testing.T) {
		config := testChangelogConfig{
			BatchSize:     2,
			FlushInterval: 50 * time.Millisecond,
			PollInterval:  50 * time.Millisecond,
		}
		harness := setupChangelogTest(t, config)
		defer harness.close()

		harness.enableAndStart()
		harness.waitForCacheInit()

		// Create changes with different node kinds
		change1 := createTestChange("e2e-1", "NK1", map[string]any{"name": "End to End User 1"})
		change2 := createTestChange("e2e-2", "NK2", map[string]any{"name": "End to End Computer 1"})

		// Test full flow: ResolveChange -> Submit -> Database
		harness.assertChangeSubmission(change1, true, "first change resolution")
		harness.submitChange(change1)

		harness.assertChangeSubmission(change2, true, "second change resolution")
		harness.submitChange(change2)

		// Wait for batch processing
		harness.waitForCacheInit()

		// Verify nodes were created in database
		harness.assertNodesExistInDB([]string{"e2e-1", "e2e-2"}, 2)

		// Test deduplication - same changes should not be submitted again
		harness.assertChangeSubmission(change1, false, "first change deduplication")
		harness.assertChangeSubmission(change2, false, "second change deduplication")
	})
}

func TestCacheClearFunctionality(t *testing.T) {
	t.Parallel()

	t.Run("ClearCache clears active cache", func(t *testing.T) {
		harness := setupChangelogTest(t, defaultTestConfig())
		defer harness.close()

		harness.enableAndStart()
		harness.waitForCacheInit()

		// Add some data to the cache
		change1 := createTestChange("clear-test-1", "User", nil)
		change2 := createTestChange("clear-test-2", "User", nil)

		// First submissions should be new
		harness.assertChangeSubmission(change1, true, "first change before clear")
		harness.assertChangeSubmission(change2, true, "second change before clear")

		// Second submissions should be deduplicated
		harness.assertChangeSubmission(change1, false, "first change duplicate before clear")
		harness.assertChangeSubmission(change2, false, "second change duplicate before clear")

		// Clear the cache
		harness.changelog.ClearCache(harness.ctx)

		// After clearing, same changes should be submittable again
		harness.assertChangeSubmission(change1, true, "first change after clear")
		harness.assertChangeSubmission(change2, true, "second change after clear")

		// And they should be deduplicated again
		harness.assertChangeSubmission(change1, false, "first change duplicate after clear")
		harness.assertChangeSubmission(change2, false, "second change duplicate after clear")
	})

	t.Run("ClearCache is safe when changelog disabled", func(t *testing.T) {
		harness := setupChangelogTest(t, defaultTestConfig())
		defer harness.close()

		// Don't enable changelog - it should remain disabled
		harness.changelog.Start(harness.ctx)
		harness.waitForCacheInit()

		// Clearing cache when disabled should not panic or error
		require.NotPanics(t, func() {
			harness.changelog.ClearCache(harness.ctx)
		})

		// Changes should still pass through (always return true when disabled)
		change := createTestChange("disabled-test", "User", nil)
		harness.assertChangeSubmission(change, true, "change with disabled changelog")
		harness.assertChangeSubmission(change, true, "same change with disabled changelog")
	})

	t.Run("Cache maintains capacity after clear", func(t *testing.T) {
		config := testChangelogConfig{
			BatchSize:     3,
			FlushInterval: 100 * time.Millisecond,
			PollInterval:  50 * time.Millisecond,
		}
		harness := setupChangelogTest(t, config)
		defer harness.close()

		harness.enableAndStart()
		harness.waitForCacheInit()

		// Get the cache and verify it has the expected capacity
		cache := harness.changelog.flagManager.getCache()
		require.NotNil(t, cache)

		originalCapacity := cache.getCapacity()
		require.Greater(t, originalCapacity, 0)

		// Add some data
		change := createTestChange("capacity-test", "User", nil)
		harness.assertChangeSubmission(change, true, "change before clear")

		// Clear the cache
		harness.changelog.ClearCache(harness.ctx)

		// Verify capacity is maintained
		clearedCache := harness.changelog.flagManager.getCache()
		require.NotNil(t, clearedCache)
		require.Equal(t, originalCapacity, clearedCache.getCapacity())

		// Verify cache still works after clear
		harness.assertChangeSubmission(change, true, "change after clear")
		harness.assertChangeSubmission(change, false, "duplicate after clear")
	})
}
