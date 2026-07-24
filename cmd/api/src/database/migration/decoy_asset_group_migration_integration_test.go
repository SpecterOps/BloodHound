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

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	decoyPreviousMigrationVersion int64 = 20260610180916
	decoyMigrationVersion         int64 = 20260618120000
)

type decoyMigrationTableCounts struct {
	AssetGroups    int64 `gorm:"column:asset_groups"`
	AssetGroupTags int64 `gorm:"column:asset_group_tags"`
	Kinds          int64 `gorm:"column:kinds"`
}

type decoyMigrationTagRecord struct {
	ID                    int  `gorm:"column:id"`
	KindID                int  `gorm:"column:kind_id"`
	AnalysisEnabledIsNull bool `gorm:"column:analysis_enabled_is_null"`
}

func newDecoyMigrationProvider(t *testing.T, testContext gooseTestContext) *goose.Provider {
	var (
		provider *goose.Provider
		err      error
	)

	t.Helper()

	provider, err = goose.NewProvider(
		goose.DialectPostgres,
		testContext.migrator.SqlDB,
		testContext.migrator.GooseFS,
		goose.WithAllowOutofOrder(true),
	)
	require.NoError(t, err)

	return provider
}

func getDecoyMigrationTableCounts(t *testing.T, databaseConnection *gorm.DB) decoyMigrationTableCounts {
	var tableCounts decoyMigrationTableCounts

	t.Helper()

	require.NoError(t, databaseConnection.Raw(`
		SELECT
			(SELECT COUNT(*) FROM asset_groups) AS asset_groups,
			(SELECT COUNT(*) FROM asset_group_tags) AS asset_group_tags,
			(SELECT COUNT(*) FROM kind) AS kinds
	`).Scan(&tableCounts).Error)

	return tableCounts
}

func assertDecoyMigrationRowCount(t *testing.T, databaseConnection *gorm.DB, expectedCount int64, sqlQuery string, parameters ...any) {
	var rowCount int64

	t.Helper()

	require.NoError(t, databaseConnection.Raw(sqlQuery, parameters...).Scan(&rowCount).Error)
	assert.Equal(t, expectedCount, rowCount)
}

func TestMigration_AddDecoyAssetGroupRejectsCollisions(t *testing.T) {
	var testCases = []struct {
		name                 string
		seedSQL              string
		expectedErrorMessage string
	}{
		{
			name: "Legacy asset group name",
			seedSQL: `
				INSERT INTO asset_groups (name, tag, system_group, created_at, updated_at)
				VALUES ('Decoy', 'custom_decoy_name', false, current_timestamp, current_timestamp)
			`,
			expectedErrorMessage: `asset_groups.name "Decoy" is already in use`,
		},
		{
			name: "Legacy asset group tag",
			seedSQL: `
				INSERT INTO asset_groups (name, tag, system_group, created_at, updated_at)
				VALUES ('Custom Decoy', 'decoy', false, current_timestamp, current_timestamp)
			`,
			expectedErrorMessage: `asset_groups.tag "decoy" is already in use`,
		},
		{
			name: "Asset group tag type",
			seedSQL: `
				WITH custom_kind AS (
					INSERT INTO kind (name)
					VALUES ('Tag_Custom_Type_Four')
					RETURNING id
				)
				INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by)
				SELECT id, 4, 'Custom Type Four', 'Custom type collision', current_timestamp, 'test-user', current_timestamp, 'test-user'
				FROM custom_kind
			`,
			expectedErrorMessage: `asset_group_tags.type 4 is already in use`,
		},
		{
			name: "Active asset group tag name",
			seedSQL: `
				WITH custom_kind AS (
					INSERT INTO kind (name)
					VALUES ('Tag_Custom_Decoy_Name')
					RETURNING id
				)
				INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by)
				SELECT id, 2, 'Decoy', 'Custom name collision', current_timestamp, 'test-user', current_timestamp, 'test-user'
				FROM custom_kind
			`,
			expectedErrorMessage: `active asset_group_tags.name "Decoy" is already in use`,
		},
		{
			name: "Asset group tag glyph",
			seedSQL: `
				WITH custom_kind AS (
					INSERT INTO kind (name)
					VALUES ('Tag_Custom_Mask')
					RETURNING id
				)
				INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by, glyph)
				SELECT id, 2, 'Custom Mask', 'Custom glyph collision', current_timestamp, 'test-user', current_timestamp, 'test-user', 'mask'
				FROM custom_kind
			`,
			expectedErrorMessage: `asset_group_tags.glyph "mask" is already in use`,
		},
		{
			name:                 "Graph kind",
			seedSQL:              `INSERT INTO kind (name) VALUES ('Tag_Decoy')`,
			expectedErrorMessage: `kind.name "Tag_Decoy" is already in use`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				testContext  = setupGooseTestContext(t)
				provider     = newDecoyMigrationProvider(t, testContext)
				countsBefore decoyMigrationTableCounts
				countsAfter  decoyMigrationTableCounts
				appliedCount int64
				err          error
			)

			_, err = provider.UpTo(testContext.ctx, decoyPreviousMigrationVersion)
			require.NoError(t, err)
			require.NoError(t, testContext.gormDB.Exec(testCase.seedSQL).Error)
			countsBefore = getDecoyMigrationTableCounts(t, testContext.gormDB)

			_, err = provider.UpTo(testContext.ctx, decoyMigrationVersion)
			require.ErrorContains(t, err, testCase.expectedErrorMessage)

			countsAfter = getDecoyMigrationTableCounts(t, testContext.gormDB)
			assert.Equal(t, countsBefore, countsAfter, "failed migration must not modify protected tables")

			require.NoError(t, testContext.gormDB.Raw(`
				SELECT COUNT(*)
				FROM goose_db_version
				WHERE version_id = ? AND is_applied = true
			`, decoyMigrationVersion).Scan(&appliedCount).Error)
			assert.Zero(t, appliedCount, "failed migration must not be recorded as applied")
		})
	}
}

func TestMigration_AddDecoyAssetGroupUpAndDown(t *testing.T) {
	var (
		testContext     = setupGooseTestContext(t)
		provider        = newDecoyMigrationProvider(t, testContext)
		decoyTag        decoyMigrationTagRecord
		unrelatedKindID int
		unrelatedTagID  int
		decoySelectorID int
		migrationError  error
	)

	_, migrationError = provider.UpTo(testContext.ctx, decoyPreviousMigrationVersion)
	require.NoError(t, migrationError)

	require.NoError(t, testContext.gormDB.Raw(`
		INSERT INTO kind (name)
		VALUES ('Tag_Unrelated_Migration_Test')
		RETURNING id
	`).Scan(&unrelatedKindID).Error)
	require.NoError(t, testContext.gormDB.Raw(`
		INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by, glyph)
		VALUES (?, 2, 'Unrelated Migration Test', 'Must survive Decoy rollback', current_timestamp, 'test-user', current_timestamp, 'test-user', 'unrelated-migration-test')
		RETURNING id
	`, unrelatedKindID).Scan(&unrelatedTagID).Error)
	require.NoError(t, testContext.gormDB.Exec(`
		INSERT INTO asset_groups (name, tag, system_group, created_at, updated_at)
		VALUES ('Unrelated Migration Test', 'unrelated_migration_test', false, current_timestamp, current_timestamp)
	`).Error)

	_, migrationError = provider.UpTo(testContext.ctx, decoyMigrationVersion)
	require.NoError(t, migrationError)

	require.NoError(t, testContext.gormDB.Raw(`
		SELECT
			asset_group_tags.id,
			asset_group_tags.kind_id,
			asset_group_tags.analysis_enabled IS NULL AS analysis_enabled_is_null
		FROM asset_group_tags
		JOIN kind ON kind.id = asset_group_tags.kind_id
		WHERE asset_group_tags.type = 4
			AND asset_group_tags.name = 'Decoy'
			AND asset_group_tags.created_by = 'BloodHound'
			AND kind.name = 'Tag_Decoy'
	`).Scan(&decoyTag).Error)
	require.NotZero(t, decoyTag.ID)
	require.NotZero(t, decoyTag.KindID)
	assert.True(t, decoyTag.AnalysisEnabledIsNull)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 1, `
		SELECT COUNT(*)
		FROM asset_groups
		WHERE name = 'Decoy'
			AND tag = 'decoy'
			AND system_group = true
	`)

	require.NoError(t, testContext.gormDB.Exec(`
		UPDATE asset_group_tags
		SET description = 'Updated after migration',
			updated_by = 'test-user'
		WHERE id = ?
	`, decoyTag.ID).Error)
	require.NoError(t, testContext.gormDB.Raw(`
		INSERT INTO asset_group_tag_selectors (
			asset_group_tag_id,
			created_at,
			created_by,
			updated_at,
			updated_by,
			name
		)
		VALUES (?, current_timestamp, 'test-user', current_timestamp, 'test-user', 'Decoy rollback selector')
		RETURNING id
	`, decoyTag.ID).Scan(&decoySelectorID).Error)
	require.NoError(t, testContext.gormDB.Exec(`
		INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
		VALUES (?, 1, 'S-1-5-21-Decoy-Migration-Test')
	`, decoySelectorID).Error)
	require.NoError(t, testContext.gormDB.Exec(`
		INSERT INTO asset_group_tag_selector_nodes (selector_id, node_id, certified, created_at, updated_at)
		VALUES (?, 1, 0, current_timestamp, current_timestamp)
	`, decoySelectorID).Error)
	require.NoError(t, testContext.gormDB.Exec(`
		INSERT INTO asset_group_history (actor, action, target, asset_group_tag_id, created_at)
		VALUES ('test-user', 'UpdateTag', 'Decoy', ?, current_timestamp)
	`, decoyTag.ID).Error)

	_, migrationError = provider.DownTo(testContext.ctx, decoyPreviousMigrationVersion)
	require.NoError(t, migrationError)

	assertDecoyMigrationRowCount(t, testContext.gormDB, 0, `SELECT COUNT(*) FROM asset_group_tags WHERE id = ?`, decoyTag.ID)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 0, `SELECT COUNT(*) FROM kind WHERE id = ?`, decoyTag.KindID)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 0, `SELECT COUNT(*) FROM asset_group_tag_selectors WHERE id = ?`, decoySelectorID)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 0, `SELECT COUNT(*) FROM asset_group_tag_selector_seeds WHERE selector_id = ?`, decoySelectorID)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 0, `SELECT COUNT(*) FROM asset_group_tag_selector_nodes WHERE selector_id = ?`, decoySelectorID)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 0, `SELECT COUNT(*) FROM asset_group_history WHERE asset_group_tag_id = ?`, decoyTag.ID)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 0, `SELECT COUNT(*) FROM asset_groups WHERE tag = 'decoy'`)

	assertDecoyMigrationRowCount(t, testContext.gormDB, 1, `SELECT COUNT(*) FROM asset_group_tags WHERE id = ?`, unrelatedTagID)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 1, `SELECT COUNT(*) FROM kind WHERE id = ?`, unrelatedKindID)
	assertDecoyMigrationRowCount(t, testContext.gormDB, 1, `
		SELECT COUNT(*)
		FROM asset_groups
		WHERE name = 'Unrelated Migration Test'
			AND tag = 'unrelated_migration_test'
			AND system_group = false
	`)
}
