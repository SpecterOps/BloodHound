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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertRemediation(t *testing.T) {
	type args struct {
		shortDescription, longDescription, shortRemediation, longRemediation string
	}
	tests := []struct {
		name          string
		setupData     func(t *testing.T, db *database.BloodhoundDB) int32 // Returns findingID
		args          args
		assert        func(t *testing.T, db *database.BloodhoundDB, findingId int32)
		expectedError string
	}{
		{
			name: "Success: Update existing remediation",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				env, err := db.CreateEnvironment(context.Background(), ext.ID, 1, 1)
				require.NoError(t, err)

				finding, err := db.CreateSchemaRelationshipFinding(context.Background(), ext.ID, 1, env.ID, "Finding", "Finding Display Name")
				require.NoError(t, err)

				_, err = db.CreateRemediation(context.Background(), finding.ID, "short", "long", "short rem", "long rem")
				require.NoError(t, err)

				return finding.ID
			},
			args: args{
				shortDescription: "updated short description",
				longDescription:  "updated long description",
				shortRemediation: "updated short remediation",
				longRemediation:  "updated long remediation",
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, findingId int32) {
				t.Helper()

				remediation, err := db.GetRemediationByFindingId(context.Background(), findingId)
				require.NoError(t, err)

				assert.Equal(t, findingId, remediation.FindingID)
				assert.Equal(t, "updated short description", remediation.ShortDescription)
				assert.Equal(t, "updated long description", remediation.LongDescription)
				assert.Equal(t, "updated short remediation", remediation.ShortRemediation)
				assert.Equal(t, "updated long remediation", remediation.LongRemediation)
			},
		},
		{
			name: "Success: Create remediation when none exists",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				env, err := db.CreateEnvironment(context.Background(), ext.ID, 1, 1)
				require.NoError(t, err)

				// Create Finding but do not create Remediation
				finding, err := db.CreateSchemaRelationshipFinding(context.Background(), ext.ID, 1, env.ID, "Finding2", "Finding 2 Display Name")
				require.NoError(t, err)

				return finding.ID
			},
			args: args{
				shortDescription: "new short description",
				longDescription:  "new long description",
				shortRemediation: "new short remediation",
				longRemediation:  "new long remediation",
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, findingId int32) {
				t.Helper()

				remediation, err := db.GetRemediationByFindingId(context.Background(), findingId)
				require.NoError(t, err)

				assert.Equal(t, findingId, remediation.FindingID)
				assert.Equal(t, "new short description", remediation.ShortDescription)
				assert.Equal(t, "new long description", remediation.LongDescription)
				assert.Equal(t, "new short remediation", remediation.ShortRemediation)
				assert.Equal(t, "new long remediation", remediation.LongRemediation)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			findingId := tt.setupData(t, testSuite.BHDatabase)

			// Wrap the call in a transaction
			err := testSuite.BHDatabase.Transaction(context.Background(), func(tx *database.BloodhoundDB) error {
				return tx.UpsertRemediation(
					context.Background(),
					findingId,
					tt.args.shortDescription,
					tt.args.longDescription,
					tt.args.shortRemediation,
					tt.args.longRemediation,
				)
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.assert != nil {
				tt.assert(t, testSuite.BHDatabase, findingId)
			}
		})
	}
}
