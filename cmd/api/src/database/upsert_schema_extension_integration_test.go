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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertGraphSchemaExtension(t *testing.T) {
	type args struct {
		environments []database.EnvironmentInput
		findings     []database.FindingInput
	}
	tests := []struct {
		name          string
		setupData     func(t *testing.T, db *database.BloodhoundDB) int32 // Returns extensionId
		args          args
		assert        func(t *testing.T, db *database.BloodhoundDB, extensionId int32)
		expectedError string
	}{
		{
			name: "Success: Create new environments and findings with remediations",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = db.CreateEnvironment(context.Background(), ext.ID, int32(1), int32(1))
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environments: []database.EnvironmentInput{
					{
						EnvironmentKindName: "Tag_Tier_Zero",
						SourceKindName:      "Base",
						PrincipalKinds:      []string{"Tag_Owned", "Tag_Tier_Zero"},
					},
				},
				findings: []database.FindingInput{
					{
						Name:                 "Name 1",
						DisplayName:          "Display Name 1",
						RelationshipKindName: "Tag_Tier_Zero",
						EnvironmentKindName:  "Tag_Tier_Zero",
						SourceKindName:       "Base",
						RemediationInput: database.RemediationInput{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
					},
					{
						Name:                 "Name 2",
						DisplayName:          "Display Name 2",
						RelationshipKindName: "Tag_Tier_Zero",
						EnvironmentKindName:  "Tag_Tier_Zero",
						SourceKindName:       "Base",
						RemediationInput: database.RemediationInput{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
					},
				},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
				t.Helper()

				// Verify findings were created
				finding1, err := db.GetSchemaRelationshipFindingByName(context.Background(), "Name 1")
				require.NoError(t, err)
				assert.Equal(t, extensionId, finding1.SchemaExtensionId)
				assert.Equal(t, "Display Name 1", finding1.DisplayName)

				finding2, err := db.GetSchemaRelationshipFindingByName(context.Background(), "Name 2")
				require.NoError(t, err)
				assert.Equal(t, extensionId, finding2.SchemaExtensionId)
				assert.Equal(t, "Display Name 2", finding2.DisplayName)

				// Verify remediations were created
				remediation1, err := db.GetRemediationByFindingId(context.Background(), finding1.ID)
				require.NoError(t, err)
				assert.Equal(t, "Short Description", remediation1.ShortDescription)
				assert.Equal(t, "Long Description", remediation1.LongDescription)
				assert.Equal(t, "Short Remediation", remediation1.ShortRemediation)
				assert.Equal(t, "Long Remediation", remediation1.LongRemediation)

				remediation2, err := db.GetRemediationByFindingId(context.Background(), finding2.ID)
				require.NoError(t, err)
				assert.Equal(t, "Short Description", remediation2.ShortDescription)
				assert.Equal(t, "Long Description", remediation2.LongDescription)
				assert.Equal(t, "Short Remediation", remediation2.ShortRemediation)
				assert.Equal(t, "Long Remediation", remediation2.LongRemediation)
			},
		},
		{
			name: "Success: Update existing findings and remediations",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt2", "Test2", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				env, err := db.CreateEnvironment(context.Background(), ext.ID, 1, 1)
				require.NoError(t, err)

				// Create initial finding with remediation
				finding, err := db.CreateSchemaRelationshipFinding(context.Background(), ext.ID, 1, env.ID, "ExistingFinding", "Old Display Name")
				require.NoError(t, err)

				_, err = db.CreateRemediation(context.Background(), finding.ID, "old short", "old long", "old short rem", "old long rem")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environments: []database.EnvironmentInput{
					{
						EnvironmentKindName: "Tag_Tier_Zero",
						SourceKindName:      "Base",
						PrincipalKinds:      []string{"Tag_Tier_Zero"},
					},
				},
				findings: []database.FindingInput{
					{
						Name:                 "ExistingFinding",
						DisplayName:          "Updated Display Name",
						RelationshipKindName: "Tag_Tier_Zero",
						EnvironmentKindName:  "Tag_Tier_Zero",
						SourceKindName:       "Base",
						RemediationInput: database.RemediationInput{
							ShortDescription: "Updated Short Description",
							LongDescription:  "Updated Long Description",
							ShortRemediation: "Updated Short Remediation",
							LongRemediation:  "Updated Long Remediation",
						},
					},
				},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
				t.Helper()

				// Verify finding was updated (deleted and recreated)
				finding, err := db.GetSchemaRelationshipFindingByName(context.Background(), "ExistingFinding")
				require.NoError(t, err)
				assert.Equal(t, extensionId, finding.SchemaExtensionId)
				assert.Equal(t, "Updated Display Name", finding.DisplayName)

				// Verify remediation was updated
				remediation, err := db.GetRemediationByFindingId(context.Background(), finding.ID)
				require.NoError(t, err)
				assert.Equal(t, "Updated Short Description", remediation.ShortDescription)
				assert.Equal(t, "Updated Long Description", remediation.LongDescription)
				assert.Equal(t, "Updated Short Remediation", remediation.ShortRemediation)
				assert.Equal(t, "Updated Long Remediation", remediation.LongRemediation)
			},
		},
		{
			name: "Success: Empty environments and findings",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt3", "Test3", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				environments: []database.EnvironmentInput{},
				findings:     []database.FindingInput{},
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
				t.Helper()
				// Nothing to assert - just verify no error occurred
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			extensionId := tt.setupData(t, testSuite.BHDatabase)

			err := testSuite.BHDatabase.UpsertGraphSchemaExtension(
				context.Background(),
				extensionId,
				tt.args.environments,
				tt.args.findings,
			)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.assert != nil {
				tt.assert(t, testSuite.BHDatabase, extensionId)
			}
		})
	}
}
