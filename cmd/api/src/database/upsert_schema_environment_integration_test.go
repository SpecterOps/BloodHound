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
		assert        func(t *testing.T, db *database.BloodhoundDB, args args)
		expectedError string
	}{
		{
			name: "Success: Create new environment with principal kinds",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "Base",
				principalKinds:  []string{"Tag_Tier_Zero", "Tag_Owned"},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, args args) {
				t.Helper()

				sourceKind, err := db.GetSourceKindByName(context.Background(), args.sourceKind)
				require.NoError(t, err, "unexpected error occurred when retrieving source kind by name")

				environmentKind, err := db.GetKindByName(context.Background(), args.environmentKind)
				require.NoError(t, err, "unexpected error occurred when retrieving kind by name")

				environment, err := db.GetEnvironmentByKinds(context.Background(), environmentKind.ID, int32(sourceKind.ID))
				require.NoError(t, err, "unexpected error occurred when getting environment by kinds")

				principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), environment.ID)
				require.NoError(t, err, "unexpected error occurred when retrieving principal kinds by env id")
				require.Len(t, principalKinds, len(args.principalKinds))

				expectedKindIDs := make([]int32, 0, len(args.principalKinds))
				for _, name := range args.principalKinds {
					kind, err := db.GetKindByName(context.Background(), name)
					require.NoError(t, err, "unexpected error occurred when retrieving kind by name")

					expectedKindIDs = append(expectedKindIDs, int32(kind.ID))
				}

				actualKindIDs := make([]int32, 0, len(principalKinds))
				for _, pk := range principalKinds {
					assert.Equal(t, environment.ID, pk.EnvironmentId)
					actualKindIDs = append(actualKindIDs, pk.PrincipalKind)
				}

				assert.ElementsMatch(t, expectedKindIDs, actualKindIDs)
			},
		},
		{
			name: "Success: Upsert replaces existing environment",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_2")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
			assert: func(t *testing.T, db *database.BloodhoundDB, args args) {
				t.Helper()

				sourceKind, err := db.GetSourceKindByName(context.Background(), args.sourceKind)
				require.NoError(t, err, "unexpected error occurred when retrieving source kind by name")

				environmentKind, err := db.GetKindByName(context.Background(), args.environmentKind)
				require.NoError(t, err, "unexpected error occurred when retrieving kind by name")

				environment, err := db.GetEnvironmentByKinds(context.Background(), environmentKind.ID, int32(sourceKind.ID))
				require.NoError(t, err, "unexpected error occurred when getting environment by kinds")

				principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), environment.ID)
				require.NoError(t, err, "unexpected error occurred when retrieving principal kinds by env id")
				require.Len(t, principalKinds, 1)

				expectedKind, err := db.GetKindByName(context.Background(), args.principalKinds[0])
				require.NoError(t, err, "unexpected error occurred when retrieving kind by name")

				assert.Equal(t, int32(expectedKind.ID), principalKinds[0].PrincipalKind)
				assert.Equal(t, environment.ID, principalKinds[0].EnvironmentId)
			},
		},
		{
			name: "Success: Source kind auto-registers",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_3")
				require.NoError(t, err, "unexpected error occurred when creating extension")
				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "NewSource_" + t.Name(), // Make unique per test
				principalKinds:  []string{"Tag_Tier_Zero"},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, args args) {
				t.Helper()

				sourceKind, err := db.GetSourceKindByName(context.Background(), args.sourceKind)
				require.NoError(t, err, "unexpected error occurred when retrieving source kind by name")

				assert.Equal(t, graph.StringKind(args.sourceKind), sourceKind.Name)
				assert.True(t, sourceKind.Active)
			},
		},
		{
			name: "Error: Environment kind not found",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_4")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				return ext.ID
			},
			args: args{
				environmentKind: "NonExistent",
				sourceKind:      "Base",
				principalKinds:  []string{},
			},
			expectedError: "environment kind 'NonExistent' not found",
			assert: func(t *testing.T, db *database.BloodhoundDB, args args) {
				t.Helper()

				_, err := db.GetKindByName(context.Background(), args.environmentKind)
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		{
			name: "Error: Principal kind not found",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_5")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "Base",
				principalKinds:  []string{"NonExistent"},
			},
			expectedError: "principal kind 'NonExistent' not found",
			assert: func(t *testing.T, db *database.BloodhoundDB, args args) {
				t.Helper()

				// Verify no environment was created for this extension
				sourceKind, err := db.GetSourceKindByName(context.Background(), args.sourceKind)
				require.NoError(t, err, "unexpected error occurred when retrieving source kind by name")

				environmentKind, err := db.GetKindByName(context.Background(), args.environmentKind)
				require.NoError(t, err, "unexpected error occurred when retrieving kind by name")

				_, err = db.GetEnvironmentByKinds(context.Background(), environmentKind.ID, int32(sourceKind.ID))
				assert.Error(t, err, "Environment should not exist after rollback")
			},
		},
		{
			name: "Rollback: Partial failure on second principal kind",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_6")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				return ext.ID
			},
			args: args{
				environmentKind: "Tag_Tier_Zero",
				sourceKind:      "Base",
				principalKinds:  []string{"Tag_Owned", "NonExistent"},
			},
			expectedError: "principal kind 'NonExistent' not found",
			assert: func(t *testing.T, db *database.BloodhoundDB, args args) {
				t.Helper()

				// Verify no environment was created for this extension
				sourceKind, err := db.GetSourceKindByName(context.Background(), args.sourceKind)
				require.NoError(t, err, "unexpected error occurred when retrieving source kind by name")

				environmentKind, err := db.GetKindByName(context.Background(), args.environmentKind)
				require.NoError(t, err, "unexpected error occurred when retrieving kind by name")

				_, err = db.GetEnvironmentByKinds(context.Background(), environmentKind.ID, int32(sourceKind.ID))
				assert.Error(t, err, "Environment should not exist after rollback")
			},
		},
		{
			name: "Success: Multiple environments with different combinations",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_7")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
			assert: func(t *testing.T, db *database.BloodhoundDB, args args) {
				t.Helper()

				// Count environments created in setup + this test
				// We can't rely on total environment count due to existing data

				// Verify the first environment exists
				envKind1, err := db.GetKindByName(context.Background(), "Tag_Tier_Zero")
				require.NoError(t, err, "unexpected error occurred when retrieving kind 1 by name")

				sourceKind, err := db.GetSourceKindByName(context.Background(), "Base")
				require.NoError(t, err, "unexpected error occurred when retrieving source kind by name")

				env1, err := db.GetEnvironmentByKinds(context.Background(), envKind1.ID, int32(sourceKind.ID))
				require.NoError(t, err, "unexpected error occurred when retrieving environment by kinds")

				principalKinds1, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), env1.ID)
				require.NoError(t, err, "unexpected error occurred when retrieving principal kind 1 by env id")
				require.Len(t, principalKinds1, 1)

				// Verify the second environment exists
				envKind2, err := db.GetKindByName(context.Background(), args.environmentKind)
				require.NoError(t, err, "unexpected error occurred when retrieving kind 2 by name")

				env2, err := db.GetEnvironmentByKinds(context.Background(), envKind2.ID, int32(sourceKind.ID))
				require.NoError(t, err, "unexpected error occurred when getting environment by kinds")

				principalKinds2, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), env2.ID)
				require.NoError(t, err, "unexpected error occurred when retrieving principal kind 2 by env id")
				require.Len(t, principalKinds2, 1)

				// Ensure they are different environments
				assert.NotEqual(t, env1.ID, env2.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			extensionID := tt.setupData(t, testSuite.BHDatabase)

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
					tt.assert(t, testSuite.BHDatabase, tt.args)
				}
				return
			}

			require.NoError(t, err)
			if tt.assert != nil {
				tt.assert(t, testSuite.BHDatabase, tt.args)
			}
		})
	}
}
