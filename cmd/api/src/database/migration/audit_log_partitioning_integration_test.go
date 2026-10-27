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

const (
	versionBeforeAuditPartitioning int64 = 20261014222926
	auditPartitioningVersion       int64 = 20261022130000
)

func scalarInt64(t *testing.T, database *gorm.DB, query string, args ...any) int64 {
	t.Helper()

	var value int64
	require.NoError(t, database.Raw(query, args...).Scan(&value).Error)
	return value
}

func scalarString(t *testing.T, database *gorm.DB, query string, args ...any) string {
	t.Helper()

	var value string
	require.NoError(t, database.Raw(query, args...).Scan(&value).Error)
	return value
}

func isRangePartitioned(t *testing.T, database *gorm.DB, tableName string) bool {
	t.Helper()

	var partitioned bool
	require.NoError(t, database.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM pg_partitioned_table partitioned_table
			JOIN pg_class table_class ON table_class.oid = partitioned_table.partrelid
			WHERE table_class.relname = ? AND partitioned_table.partstrat = 'r'
		)
	`, tableName).Scan(&partitioned).Error)
	return partitioned
}

func TestMigrator_AuditLogPartitioning(t *testing.T) {
	var (
		testContext = setupGooseTestContext(t)
		provider    = testContext.migrator.GooseProvider
		database    = testContext.gormDB
	)

	_, err := provider.UpTo(testContext.ctx, versionBeforeAuditPartitioning)
	require.NoError(t, err)
	require.False(t, isRangePartitioned(t, database, "audit_logs"))

	require.NoError(t, database.Exec(`
		INSERT INTO audit_logs (created_at, action, status)
		VALUES
			('2019-01-01T00:00:00Z'::timestamptz, 'old_action', 'success'),
			(NULL, 'null_action', 'success'),
			(CURRENT_TIMESTAMP, 'current_action', 'success')
	`).Error)

	var (
		totalSeeded = int64(3)
		sourceMaxID = scalarInt64(t, database, `SELECT MAX(id) FROM audit_logs`)
	)

	_, err = provider.UpTo(testContext.ctx, auditPartitioningVersion)
	require.NoError(t, err)

	assert.True(t, isRangePartitioned(t, database, "audit_logs"))
	assert.Zero(t, scalarInt64(t, database, `SELECT count(*) FROM audit_logs`))
	assert.Zero(t, scalarInt64(t, database, `SELECT count(*) FROM audit_logs_default`))
	assert.Equal(t, sourceMaxID, scalarInt64(t, database, `SELECT last_value FROM audit_logs_id_seq`))
	assert.False(t, isRangePartitioned(t, database, "audit_logs_old"))
	assert.Equal(t, totalSeeded, scalarInt64(t, database, `SELECT count(*) FROM audit_logs_old`))
	assert.Equal(t, "audit_logs", scalarString(t, database, `
		SELECT tablename FROM pg_indexes WHERE indexname = 'idx_audit_logs_created_at'
	`))

	var nextID int64
	require.NoError(t, database.Raw(`
		INSERT INTO audit_logs (action, status)
		VALUES ('post_migration_action', 'success')
		RETURNING id
	`).Scan(&nextID).Error)
	assert.Greater(t, nextID, sourceMaxID)
	assert.Equal(t, int64(1), scalarInt64(t, database, `SELECT count(*) FROM audit_logs`))
	assert.Equal(t, totalSeeded, scalarInt64(t, database, `SELECT count(*) FROM audit_logs_old`))
	assert.Equal(t,
		scalarString(t, database, `SELECT 'audit_logs_' || to_char(CURRENT_TIMESTAMP, 'YYYY_MM')`),
		scalarString(t, database, `SELECT tableoid::regclass::text FROM audit_logs WHERE action = 'post_migration_action'`),
	)

	results, err := provider.UpTo(testContext.ctx, auditPartitioningVersion)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestMigrator_AuditLogPartitioningPreservesAdvancedSequence(t *testing.T) {
	var (
		testContext = setupGooseTestContext(t)
		provider    = testContext.migrator.GooseProvider
		database    = testContext.gormDB
	)

	_, err := provider.UpTo(testContext.ctx, versionBeforeAuditPartitioning)
	require.NoError(t, err)

	require.NoError(t, database.Exec(`
		INSERT INTO audit_logs (created_at, action, status)
		SELECT '2025-06-15T00:00:00Z'::timestamptz, 'seed_action', 'success'
		FROM generate_series(1, 5)
	`).Error)
	advancedSequence := scalarInt64(t, database, `SELECT last_value FROM audit_logs_id_seq`)
	require.NoError(t, database.Exec(`DELETE FROM audit_logs`).Error)

	_, err = provider.UpTo(testContext.ctx, auditPartitioningVersion)
	require.NoError(t, err)
	assert.Equal(t, advancedSequence, scalarInt64(t, database, `SELECT last_value FROM audit_logs_id_seq`))

	var nextID int64
	require.NoError(t, database.Raw(`
		INSERT INTO audit_logs (created_at, action, status)
		VALUES ('2025-06-15T00:00:00Z'::timestamptz, 'post_action', 'success')
		RETURNING id
	`).Scan(&nextID).Error)
	assert.Greater(t, nextID, advancedSequence)
}
