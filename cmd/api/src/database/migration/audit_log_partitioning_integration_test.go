// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
//go:build integration

package migration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// versionBeforeAuditPartitioning is the goose version applied immediately before
// the audit_logs partitioning migration. Migrating up to it leaves the original,
// non-partitioned audit_logs table (from the init baseline) in place so the test
// can seed legacy rows before the partitioning migration converts the table.
const (
	versionBeforeAuditPartitioning int64 = 20260702124154
	auditPartitioningVersion       int64 = 20260707000001
)

// scalarInt64 runs a query expected to return a single int64 and returns it.
func scalarInt64(t *testing.T, db *gorm.DB, query string, args ...any) int64 {
	t.Helper()

	var value int64
	require.NoError(t, db.Raw(query, args...).Scan(&value).Error)
	return value
}

// isRangePartitioned reports whether the named table is a range-partitioned
// parent table.
func isRangePartitioned(t *testing.T, db *gorm.DB, tableName string) bool {
	t.Helper()

	var partitioned bool
	err := db.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM pg_partitioned_table pt
			JOIN pg_class c ON c.oid = pt.partrelid
			WHERE c.relname = ? AND pt.partstrat = 'r'
		)
	`, tableName).Scan(&partitioned).Error
	require.NoError(t, err)
	return partitioned
}

// TestMigrator_AuditLogPartitioning validates the batched, single-idempotent-pass
// audit_logs partitioning migration (20260707000001): it seeds the legacy
// audit_logs table across the backfill batch boundary, runs the partitioning
// migration, and asserts the data was moved into the correct partitions, the table
// is now range-partitioned, the id sequence advanced past the copied rows, and
// re-running the migration is a no-op. The migration is not restart-resumable (the
// backfill is one transaction with no committed partial progress); it is instead
// re-run-safe, converging from the start on any failure via idempotency.
func TestMigrator_AuditLogPartitioning(t *testing.T) {
	testContext := setupGooseTestContext(t)

	provider := testContext.migrator.GooseProvider
	db := testContext.gormDB

	// Migrate up to just before the partitioning migration. This leaves the
	// original, non-partitioned audit_logs table from the init baseline in place.
	_, err := provider.UpTo(testContext.ctx, versionBeforeAuditPartitioning)
	require.NoError(t, err)

	require.False(t, isRangePartitioned(t, db, "audit_logs"),
		"audit_logs should not be partitioned before the migration")

	// Seed legacy rows across the batch boundary (batch_size in the migration is
	// 50000) so the backfill loop iterates more than once, plus rows that route to
	// a monthly partition and rows that route to the default partition (a
	// created_at outside the 2024-01..2026-08 partition range, and a NULL that the
	// migration coalesces to 2020-01-01).
	const seededRows = 120000
	require.NoError(t, db.Exec(`
		INSERT INTO audit_logs (created_at, action, actor_id, actor_name, status)
		SELECT
			'2025-06-15T00:00:00Z'::timestamptz,
			'seed_action',
			'actor-' || g,
			'Seed Actor',
			'success'
		FROM generate_series(1, ?) AS g
	`, seededRows).Error)

	// A row whose created_at predates the earliest monthly partition -> default.
	require.NoError(t, db.Exec(`
		INSERT INTO audit_logs (created_at, action, status)
		VALUES ('2019-01-01T00:00:00Z'::timestamptz, 'old_action', 'success')
	`).Error)

	// A NULL created_at -> coalesced to 2020-01-01 by the migration -> default.
	require.NoError(t, db.Exec(`
		INSERT INTO audit_logs (created_at, action, status)
		VALUES (NULL, 'null_action', 'success')
	`).Error)

	var (
		totalSeeded = int64(seededRows + 2)
		sourceCount = scalarInt64(t, db, `SELECT count(*) FROM audit_logs`)
		sourceMaxID = scalarInt64(t, db, `SELECT MAX(id) FROM audit_logs`)
	)
	require.Equal(t, totalSeeded, sourceCount, "sanity: seeded row count")

	// Run the partitioning migration.
	_, err = provider.UpTo(testContext.ctx, auditPartitioningVersion)
	require.NoError(t, err)

	// audit_logs is now the range-partitioned parent table under its original name.
	assert.True(t, isRangePartitioned(t, db, "audit_logs"),
		"audit_logs should be range-partitioned after the migration")

	// All rows were moved and none were lost.
	assert.Equal(t, totalSeeded, scalarInt64(t, db, `SELECT count(*) FROM audit_logs`),
		"row count must be preserved through the partition conversion")

	// The 2025-06 rows landed in the matching monthly partition.
	assert.Equal(t, int64(seededRows),
		scalarInt64(t, db, `SELECT count(*) FROM audit_logs_2025_06`),
		"seeded 2025-06 rows should live in the 2025_06 partition")

	// The out-of-range and NULL-coalesced rows landed in the default partition.
	assert.Equal(t, int64(2),
		scalarInt64(t, db, `SELECT count(*) FROM audit_logs_default`),
		"out-of-range and NULL created_at rows should live in the default partition")

	// The id sequence advanced past the largest copied id so new inserts do not
	// collide with backfilled rows.
	assert.Equal(t, sourceMaxID,
		scalarInt64(t, db, `SELECT last_value FROM audit_logs_id_seq`),
		"audit_logs_id_seq should be advanced to the max copied id")

	// Re-running the partitioning migration must be a safe no-op: its version is
	// already applied, so migrating up to it again applies nothing. Targeting the
	// version explicitly (rather than Up) keeps this assertion scoped to the
	// partitioning migration and unaffected by any later migrations in the tree.
	results, err := provider.UpTo(testContext.ctx, auditPartitioningVersion)
	require.NoError(t, err)
	assert.Empty(t, results, "re-running the partitioning migration should apply nothing")

	// Data is unchanged after the re-run.
	assert.Equal(t, totalSeeded, scalarInt64(t, db, `SELECT count(*) FROM audit_logs`),
		"row count must be unchanged after re-running migrations")
}

// TestMigrator_AuditLogPartitioning_EmptyTableAdvancedSequence guards against id
// reuse when the partitioning migration runs against an empty audit_logs table
// whose sequence has already issued ids (rows were inserted and later deleted).
// The migration must be forward-only: it must not rewind audit_logs_id_seq, so the
// next insert cannot reuse an id that was previously handed out.
func TestMigrator_AuditLogPartitioning_EmptyTableAdvancedSequence(t *testing.T) {
	testContext := setupGooseTestContext(t)

	provider := testContext.migrator.GooseProvider
	db := testContext.gormDB

	// Migrate up to just before the partitioning migration, leaving the original,
	// non-partitioned audit_logs table in place.
	_, err := provider.UpTo(testContext.ctx, versionBeforeAuditPartitioning)
	require.NoError(t, err)

	require.False(t, isRangePartitioned(t, db, "audit_logs"),
		"audit_logs should not be partitioned before the migration")

	// Issue ids by inserting rows, then delete them so the table is empty while the
	// sequence has advanced past the ids it already handed out.
	require.NoError(t, db.Exec(`
		INSERT INTO audit_logs (created_at, action, actor_id, actor_name, status)
		SELECT
			'2025-06-15T00:00:00Z'::timestamptz,
			'seed_action',
			'actor-' || g,
			'Seed Actor',
			'success'
		FROM generate_series(1, 5) AS g
	`).Error)

	advancedSeq := scalarInt64(t, db, `SELECT last_value FROM audit_logs_id_seq`)
	require.Equal(t, int64(5), advancedSeq,
		"sanity: sequence should have issued ids 1..5")

	require.NoError(t, db.Exec(`DELETE FROM audit_logs`).Error)
	require.Equal(t, int64(0), scalarInt64(t, db, `SELECT count(*) FROM audit_logs`),
		"sanity: audit_logs should be empty before the migration")

	// Run the partitioning migration against the now-empty table.
	_, err = provider.UpTo(testContext.ctx, auditPartitioningVersion)
	require.NoError(t, err)

	assert.True(t, isRangePartitioned(t, db, "audit_logs"),
		"audit_logs should be range-partitioned after the migration")

	// The sequence must not have been rewound; its position is preserved so the
	// next insert cannot reuse a previously issued id.
	assert.Equal(t, advancedSeq,
		scalarInt64(t, db, `SELECT last_value FROM audit_logs_id_seq`),
		"audit_logs_id_seq must not be rewound on an empty table")

	// The next inserted row must receive an id beyond the previously issued range,
	// confirming no id reuse.
	var nextID int64
	require.NoError(t, db.Raw(`
		INSERT INTO audit_logs (created_at, action, actor_id, actor_name, status)
		VALUES ('2025-06-15T00:00:00Z'::timestamptz, 'post_action', 'actor-post', 'Post Actor', 'success')
		RETURNING id
	`).Scan(&nextID).Error)
	assert.Greater(t, nextID, advancedSeq,
		"the first insert after migration must not reuse a previously issued id")
}
