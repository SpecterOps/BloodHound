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

	"github.com/stretchr/testify/require"
)

const uniqueUserRoleMigrationVersion int64 = 20261028122028

type userRoleAssignment struct {
	UserID string
	RoleID int32
}

func setupBeforeUniqueUserRoleMigration(t *testing.T) gooseTestContext {
	var (
		testContext     = setupGooseTestContext(t)
		previousVersion int64
		err             error
	)
	t.Helper()

	for _, source := range testContext.migrator.GooseProvider.ListSources() {
		if source.Version < uniqueUserRoleMigrationVersion && source.Version > previousVersion {
			previousVersion = source.Version
		}
	}
	require.Positive(t, previousVersion)
	_, err = testContext.migrator.GooseProvider.UpTo(testContext.ctx, previousVersion)
	require.NoError(t, err)
	return testContext
}

func TestMigration_UniqueUserRoleCleanup(t *testing.T) {
	var (
		testContext = setupBeforeUniqueUserRoleMigration(t)
		roleRows    []struct {
			ID   int32
			Name string
		}
		roleIDs             = make(map[string]int32)
		userIDs             []string
		expectedAssignments []userRoleAssignment
		actualAssignments   []userRoleAssignment
		userCount           int64
		err                 error
		testCases           = []struct {
			userID       string
			roles        []string
			expectedRole string
		}{
			{userID: "pair-read-upload", roles: []string{"Read-Only", "Upload-Only"}, expectedRole: "Read-Only"},
			{userID: "pair-upload-user", roles: []string{"Upload-Only", "User"}, expectedRole: "Upload-Only"},
			{userID: "pair-user-auditor", roles: []string{"User", "Auditor"}, expectedRole: "User"},
			{userID: "pair-auditor-power", roles: []string{"Auditor", "Power User"}, expectedRole: "Auditor"},
			{userID: "pair-power-admin", roles: []string{"Power User", "Administrator"}, expectedRole: "Power User"},
			{userID: "all-roles", roles: []string{"Administrator", "Power User", "Auditor", "User", "Upload-Only", "Read-Only"}, expectedRole: "Read-Only"},
			{userID: "single-role", roles: []string{"Auditor"}, expectedRole: "Auditor"},
			{userID: "roleless"},
		}
	)
	require.NoError(t, testContext.gormDB.Table("roles").Select("id, name").Scan(&roleRows).Error)
	for _, role := range roleRows {
		roleIDs[role.Name] = role.ID
	}

	// Seed every case before one migration run, so we also check that cleanup
	// chooses a survivor independently for each user.
	for _, testCase := range testCases {
		require.NoError(t, testContext.gormDB.Exec(
			"INSERT INTO users (id, principal_name) VALUES (?, ?)", testCase.userID, testCase.userID,
		).Error)
		userIDs = append(userIDs, testCase.userID)
		for _, roleName := range testCase.roles {
			require.Contains(t, roleIDs, roleName)
			require.NoError(t, testContext.gormDB.Exec(
				"INSERT INTO users_roles (user_id, role_id) VALUES (?, ?)", testCase.userID, roleIDs[roleName],
			).Error)
		}
		if testCase.expectedRole != "" {
			expectedAssignments = append(expectedAssignments, userRoleAssignment{
				UserID: testCase.userID,
				RoleID: roleIDs[testCase.expectedRole],
			})
		}
	}

	_, err = testContext.migrator.GooseProvider.UpTo(testContext.ctx, uniqueUserRoleMigrationVersion)
	require.NoError(t, err)
	require.NoError(t, testContext.gormDB.Table("users_roles").Where("user_id IN ?", userIDs).Scan(&actualAssignments).Error)
	require.ElementsMatch(t, expectedAssignments, actualAssignments)
	require.NoError(t, testContext.gormDB.Table("users").Where("id IN ?", userIDs).Count(&userCount).Error)
	require.EqualValues(t, len(testCases), userCount, "cleanup must preserve all user accounts, including the roleless user")

	// Multiple users already share surviving roles above. Only a second role
	// for the same user must be rejected by the new constraint.
	require.ErrorContains(t, testContext.gormDB.Exec(
		"INSERT INTO users_roles (user_id, role_id) VALUES (?, ?)", "pair-read-upload", roleIDs["Administrator"],
	).Error, "users_roles_user_id_key")
}

func TestMigration_UniqueUserRoleRejectsUnknownRoles(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		roleName      any
		expectedLabel string
	}{
		{name: "unrecognized name", roleName: "Unranked Role", expectedLabel: "Unranked Role"},
		{name: "null name", expectedLabel: "<NULL>"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				testContext         = setupBeforeUniqueUserRoleMigration(t)
				originalAssignments []userRoleAssignment
				actualAssignments   []userRoleAssignment
				appliedCount        int64
				constraintCount     int64
				userCount           int64
				unknownRoleID       int32
				err                 error
			)
			require.NoError(t, testContext.gormDB.Raw(
				"INSERT INTO roles (name) VALUES (?) RETURNING id", testCase.roleName,
			).Scan(&unknownRoleID).Error)
			require.NoError(t, testContext.gormDB.Exec(`
                INSERT INTO users (id, principal_name) VALUES
                    ('known-duplicate', 'known-duplicate'), ('unknown-duplicate', 'unknown-duplicate');
                INSERT INTO users_roles (user_id, role_id)
                SELECT 'known-duplicate', id FROM roles WHERE name IN ('Administrator', 'User');
                INSERT INTO users_roles (user_id, role_id)
                SELECT 'unknown-duplicate', id FROM roles WHERE name = 'User';
            `).Error)
			require.NoError(t, testContext.gormDB.Exec(
				"INSERT INTO users_roles (user_id, role_id) VALUES (?, ?)", "unknown-duplicate", unknownRoleID,
			).Error)
			require.NoError(t, testContext.gormDB.Table("users_roles").Scan(&originalAssignments).Error)
			require.Len(t, originalAssignments, 4)

			_, err = testContext.migrator.GooseProvider.UpTo(testContext.ctx, uniqueUserRoleMigrationVersion)
			require.ErrorContains(t, err, "duplicate assignments contain unrecognized roles")
			require.ErrorContains(t, err, testCase.expectedLabel)
			require.NoError(t, testContext.gormDB.Table("users_roles").Scan(&actualAssignments).Error)
			require.ElementsMatch(t, originalAssignments, actualAssignments, "all original assignments must survive the failed migration")
			require.NoError(t, testContext.gormDB.Table("users").Where("id IN ?", []string{"known-duplicate", "unknown-duplicate"}).Count(&userCount).Error)
			require.EqualValues(t, 2, userCount)
			require.NoError(t, testContext.gormDB.Table("goose_db_version").Where("version_id = ? AND is_applied", uniqueUserRoleMigrationVersion).Count(&appliedCount).Error)
			require.Zero(t, appliedCount)
			require.NoError(t, testContext.gormDB.Raw(
				"SELECT count(*) FROM pg_constraint WHERE conrelid = 'users_roles'::regclass AND conname = 'users_roles_user_id_key'",
			).Scan(&constraintCount).Error)
			require.Zero(t, constraintCount)
		})
	}
}
