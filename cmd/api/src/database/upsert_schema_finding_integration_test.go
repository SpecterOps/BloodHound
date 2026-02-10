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

package database_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertFinding(t *testing.T) {
	type args struct {
		sourceKindName, relationshipKindName, environmentKind, name, displayName string
	}
	tests := []struct {
		name          string
		setupData     func(t *testing.T, db *database.BloodhoundDB) int32 // Returns extensionId
		args          args
		assert        func(t *testing.T, db *database.BloodhoundDB, extensionId int32)
		expectedError string
	}{
		{
			name: "Success: Update existing finding - delete and re-create",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				env, err := db.CreateEnvironment(context.Background(), ext.ID, 1, 1)
				require.NoError(t, err)

				// Create finding
				_, err = db.CreateSchemaRelationshipFinding(context.Background(), ext.ID, 1, env.ID, "Finding Name", "Finding Display Name")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				sourceKindName:       "Base",
				relationshipKindName: "Tag_Tier_Zero",
				environmentKind:      "Tag_Tier_Zero",
				// Name triggers upsert so this needs to match the finding's name that we want to update
				name:        "Finding Name",
				displayName: "Updated Display Name",
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
				t.Helper()

				finding, err := db.GetSchemaRelationshipFindingByName(context.Background(), "Finding Name")
				require.NoError(t, err)

				assert.Equal(t, extensionId, finding.SchemaExtensionId)
				assert.Equal(t, "Finding Name", finding.Name)
				assert.Equal(t, "Updated Display Name", finding.DisplayName)
			},
		},
		{
			name: "Success: Create finding when none exists",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt2", "Test2", "v1.0.0", "test_namespace_2")
				require.NoError(t, err)

				_, err = db.CreateEnvironment(context.Background(), ext.ID, 1, 1)
				require.NoError(t, err)

				// No finding created since we're testing the creation workflow
				return ext.ID
			},
			args: args{
				sourceKindName:       "Base",
				relationshipKindName: "Tag_Tier_Zero",
				environmentKind:      "Tag_Tier_Zero",
				name:                 "Finding",
				displayName:          "Finding Display Name",
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
				t.Helper()

				finding, err := db.GetSchemaRelationshipFindingByName(context.Background(), "Finding")
				require.NoError(t, err)

				assert.Equal(t, extensionId, finding.SchemaExtensionId)
				assert.Equal(t, "Finding", finding.Name)
				assert.Equal(t, "Finding Display Name", finding.DisplayName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			extensionId := tt.setupData(t, testSuite.BHDatabase)

			var findingResponse model.SchemaRelationshipFinding
			// Wrap the call in a transaction
			err := testSuite.BHDatabase.Transaction(context.Background(), func(tx *database.BloodhoundDB) error {
				finding, err := tx.UpsertFinding(
					context.Background(),
					extensionId,
					tt.args.sourceKindName,
					tt.args.relationshipKindName,
					tt.args.environmentKind,
					tt.args.name,
					tt.args.displayName,
				)
				findingResponse = finding
				return err
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, findingResponse.ID, "Finding should have been created/updated")
			}

			if tt.assert != nil {
				tt.assert(t, testSuite.BHDatabase, extensionId)
			}
		})
	}
}
