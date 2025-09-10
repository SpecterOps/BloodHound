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

func TestChangelogIntegration(t *testing.T) {
	t.Parallel()

	t.Run("coordinator and flag manager startup coordination", func(t *testing.T) {
		suite := setupIntegrationTest(t)
		defer teardownIntegrationTest(t, &suite)

		var (
			changelog = NewChangelog(suite.GraphDB, suite.BloodhoundDB, Options{
				BatchSize:     3,
				FlushInterval: 100 * time.Millisecond,
				PollInterval:  50 * time.Millisecond,
			})
			ctx = suite.Context
		)

		// enable changelog by getting existing flag and setting it
		suite.enableChangelog(t)

		// Start changelog - both coordinator and flag manager should start
		changelog.Start(ctx)
		defer changelog.Stop(context.Background())

		// Initially cache should be nil until flag manager polls
		shouldSubmit, err := changelog.ResolveChange(
			NewNodeChange("test1", graph.Kinds{graph.StringKind("User")},
				graph.NewProperties().SetAll(map[string]any{
					"objectid": "test1",
					"lastseen": time.Now(),
				})))
		require.NoError(t, err)
		require.True(t, shouldSubmit) // Should pass-through when cache is nil

		// Wait for flag manager to initialize cache
		time.Sleep(200 * time.Millisecond)

		// Now cache should be available and working
		change := NewNodeChange("test2", graph.Kinds{graph.StringKind("User")},
			graph.NewProperties().SetAll(map[string]any{
				"objectid": "test2",
				"lastseen": time.Now(),
			}))

		shouldSubmit, err = changelog.ResolveChange(change)
		require.NoError(t, err)
		require.True(t, shouldSubmit) // First time should be submitted

		// Same change should be deduplicated
		shouldSubmit, err = changelog.ResolveChange(change)
		require.NoError(t, err)
		require.False(t, shouldSubmit) // Should be deduplicated
	})

	t.Run("feature flag enable/disable affects cache behavior", func(t *testing.T) {
		suite := setupIntegrationTest(t)
		defer teardownIntegrationTest(t, &suite)

		var (
			changelog = NewChangelog(suite.GraphDB, suite.BloodhoundDB, Options{
				BatchSize:     3,
				FlushInterval: 100 * time.Millisecond,
				PollInterval:  50 * time.Millisecond,
			})
			ctx = suite.Context
		)

		// enable changelog
		suite.enableChangelog(t)

		changelog.Start(ctx)
		defer changelog.Stop(context.Background())

		// Wait for initial cache setup
		time.Sleep(200 * time.Millisecond)

		change := NewNodeChange("toggle-test", graph.Kinds{graph.StringKind("User")},
			graph.NewProperties().SetAll(map[string]any{
				"objectid": "toggle-test",
				"lastseen": time.Now(),
			}))

		// With cache enabled, first submission should work
		shouldSubmit, err := changelog.ResolveChange(change)
		require.NoError(t, err)
		require.True(t, shouldSubmit)

		// Second submission should be deduplicated
		shouldSubmit, err = changelog.ResolveChange(change)
		require.NoError(t, err)
		require.False(t, shouldSubmit)

		// Disable feature flag
		suite.disableChangelog(t)

		time.Sleep(200 * time.Millisecond) // Wait for flag manager to poll and disable cache

		// With cache disabled, should pass-through (always true)
		shouldSubmit, err = changelog.ResolveChange(change)
		require.NoError(t, err)
		require.True(t, shouldSubmit) // Pass-through when disabled

		// Re-enable feature flag
		suite.enableChangelog(t)

		time.Sleep(200 * time.Millisecond) // Wait for flag manager to poll and re-enable cache

		// Cache should be fresh after re-enabling, so same change should be submittable again
		shouldSubmit, err = changelog.ResolveChange(change)
		require.NoError(t, err)
		require.True(t, shouldSubmit) // Fresh cache after re-enable
	})

	t.Run("end-to-end flow with real database operations", func(t *testing.T) {
		suite := setupIntegrationTest(t)
		defer teardownIntegrationTest(t, &suite)

		var (
			changelog = NewChangelog(suite.GraphDB, suite.BloodhoundDB, Options{
				BatchSize:     2,
				FlushInterval: 50 * time.Millisecond,
				PollInterval:  50 * time.Millisecond,
			})
			ctx = suite.Context
		)

		suite.enableChangelog(t)

		changelog.Start(ctx)
		defer changelog.Stop(context.Background())

		// Wait for cache initialization
		time.Sleep(200 * time.Millisecond)

		// Create changes
		change1 := NewNodeChange("e2e-1", graph.Kinds{graph.StringKind("NK1")},
			graph.NewProperties().SetAll(map[string]any{
				"objectid": "e2e-1",
				"lastseen": time.Now(),
				"name":     "End to End User 1",
			}))

		change2 := NewNodeChange("e2e-2", graph.Kinds{graph.StringKind("NK2")},
			graph.NewProperties().SetAll(map[string]any{
				"objectid": "e2e-2",
				"lastseen": time.Now(),
				"name":     "End to End Computer 1",
			}))

		// Test full flow: ResolveChange -> Submit -> Database
		shouldSubmit, err := changelog.ResolveChange(change1)
		require.NoError(t, err)
		require.True(t, shouldSubmit)

		submitted := changelog.Submit(ctx, change1)
		require.True(t, submitted)

		shouldSubmit, err = changelog.ResolveChange(change2)
		require.NoError(t, err)
		require.True(t, shouldSubmit)

		submitted = changelog.Submit(ctx, change2)
		require.True(t, submitted)

		// Wait for batch processing
		time.Sleep(200 * time.Millisecond)

		// Verify nodes were created in database
		var nodeCount int
		err = suite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			return tx.Nodes().Filterf(func() graph.Criteria {
				return query.Or(
					query.Equals(query.NodeProperty("objectid"), "e2e-1"),
					query.Equals(query.NodeProperty("objectid"), "e2e-2"),
				)
			}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for range cursor.Chan() {
					nodeCount++
				}
				return cursor.Error()
			})
		})
		require.NoError(t, err)
		require.Equal(t, 2, nodeCount)

		// Test deduplication - same changes should not be submitted again
		shouldSubmit, err = changelog.ResolveChange(change1)
		require.NoError(t, err)
		require.False(t, shouldSubmit) // Should be deduplicated

		shouldSubmit, err = changelog.ResolveChange(change2)
		require.NoError(t, err)
		require.False(t, shouldSubmit) // Should be deduplicated
	})
}
