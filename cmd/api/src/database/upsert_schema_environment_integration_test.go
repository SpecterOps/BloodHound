// Copyright 2026 Specter Ops, Inc.
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

package database_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertSchemaEnvironmentWithPrincipalKinds(t *testing.T) {
	type args struct {
		environmentKind string
		sourceKind      string
		principalKinds  []string
	}
	tests := []struct {
		name          string
		setupData     func(t *testing.T, db *database.BloodhoundDB) int32
		args          args
		assert        func(t *testing.T, db *database.BloodhoundDB)
		expectedError string
	}{
		{
			name: "Success: Create new environment with principal kinds",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "Base",
				principalKinds:  []string{"Tag_Tier_Zero", "Tag_Owned"},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB) {
				t.Helper()

				expectedPrincipalKindNames := []string{"Tag_Tier_Zero", "Tag_Owned"}

				environments, err := db.GetEnvironments(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, 1, len(environments))

				principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), environments[0].ID)
				assert.NoError(t, err)
				assert.Equal(t, len(expectedPrincipalKindNames), len(principalKinds))

				expectedKindIDs := make([]int32, len(expectedPrincipalKindNames))
				for i, name := range expectedPrincipalKindNames {
					kind, err := db.GetKindByName(context.Background(), name)
					assert.NoError(t, err)
					expectedKindIDs[i] = int32(kind.ID)
				}

				actualKindIDs := make([]int32, len(principalKinds))
				for i, pk := range principalKinds {
					assert.Equal(t, environments[0].ID, pk.EnvironmentId)
					actualKindIDs[i] = pk.PrincipalKind
				}

				assert.ElementsMatch(t, expectedKindIDs, actualKindIDs)
			},
		},
		{
			name: "Success: Upsert replaces existing environment",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				err = db.Transaction(context.Background(), func(tx *database.BloodhoundDB) error {
					return tx.UpsertSchemaEnvironmentWithPrincipalKinds(
						context.Background(),
						ext.ID,
						"Tag_Tier_Zero",
						"Base",
						[]string{"Tag_Owned"},
					)
				})
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "Base",
				principalKinds:  []string{"Tag_Tier_Zero"},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB) {
				t.Helper()

				expectedPrincipalKindNames := []string{"Tag_Tier_Zero"}

				environments, err := db.GetEnvironments(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, 1, len(environments), "Should only have one environment (old one deleted)")

				principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), environments[0].ID)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(principalKinds))

				expectedKind, err := db.GetKindByName(context.Background(), expectedPrincipalKindNames[0])
				assert.NoError(t, err)

				assert.Equal(t, int32(expectedKind.ID), principalKinds[0].PrincipalKind)
				assert.Equal(t, environments[0].ID, principalKinds[0].EnvironmentId)
			},
		},
		{
			name: "Success: Source kind auto-registers",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "NewSource",
				principalKinds:  []string{"Tag_Tier_Zero"},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB) {
				t.Helper()

				sourceKind, err := db.GetSourceKindByName(context.Background(), "NewSource")
				assert.NoError(t, err)
				assert.Equal(t, graph.StringKind("NewSource"), sourceKind.Name)

				environments, err := db.GetEnvironments(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, 1, len(environments))
				assert.Equal(t, int32(sourceKind.ID), environments[0].SourceKindId)

				principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), environments[0].ID)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(principalKinds))
			},
		},
		{
			name: "Error: Environment kind not found",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environmentKind: "NonExistent",
				sourceKind:      "Base",
				principalKinds:  []string{},
			},
			expectedError: "environment kind 'NonExistent' not found",
			assert: func(t *testing.T, db *database.BloodhoundDB) {
				t.Helper()

				// Verify transaction rolled back - no environment created
				environments, err := db.GetEnvironments(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, 0, len(environments), "No environment should exist after rollback")
			},
		},
		{
			name: "Error: Principal kind not found",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "Base",
				principalKinds:  []string{"NonExistent"},
			},
			expectedError: "principal kind 'NonExistent' not found",
			assert: func(t *testing.T, db *database.BloodhoundDB) {
				t.Helper()

				// Verify transaction rolled back - no environment created
				environments, err := db.GetEnvironments(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, 0, len(environments), "No environment should exist after rollback")
			},
		},
		{
			name: "Rollback: Partial failure on second principal kind",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "Base",
				principalKinds:  []string{"Tag_Owned", "NonExistent"},
			},
			expectedError: "principal kind 'NonExistent' not found",
			assert: func(t *testing.T, db *database.BloodhoundDB) {
				t.Helper()

				// Verify transaction rolled back - no environment created
				environments, err := db.GetEnvironments(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, 0, len(environments), "No environment should exist after rollback")
			},
		},
		{
			name: "Success: Multiple environments with different combinations",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				err = db.Transaction(context.Background(), func(tx *database.BloodhoundDB) error {
					return tx.UpsertSchemaEnvironmentWithPrincipalKinds(
						context.Background(),
						ext.ID,
						"Tag_Tier_Zero",
						"Base",
						[]string{"Tag_Tier_Zero"},
					)
				})
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Owned",
				sourceKind:      "Base",
				principalKinds:  []string{"Tag_Owned"},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB) {
				t.Helper()

				environments, err := db.GetEnvironments(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, 2, len(environments), "Should have two different environments")

				for _, env := range environments {
					principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), env.ID)
					assert.NoError(t, err)
					assert.Equal(t, 1, len(principalKinds), "Each environment should have one principal kind")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			extensionID := tt.setupData(t, testSuite.BHDatabase)

			// Wrap the call in a transaction
			err := testSuite.BHDatabase.Transaction(context.Background(), func(tx *database.BloodhoundDB) error {
				return tx.UpsertSchemaEnvironmentWithPrincipalKinds(
					context.Background(),
					extensionID,
					tt.args.environmentKind,
					tt.args.sourceKind,
					tt.args.principalKinds,
				)
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				if tt.assert != nil {
					tt.assert(t, testSuite.BHDatabase)
				}
			} else {
				require.NoError(t, err)
				if tt.assert != nil {
					tt.assert(t, testSuite.BHDatabase)
				}
			}
		})
	}
}
