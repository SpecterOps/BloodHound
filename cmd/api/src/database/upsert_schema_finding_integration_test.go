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

func TestCreateFindingWithRemediation(t *testing.T) {
	t.Parallel()

	type args struct {
		name                 string
		displayName          string
		relationshipKindName string
		environmentKindName  string
		remediation          model.RemediationInput
	}
	type testData struct {
		name   string
		args   args
		setup  func(t *testing.T, db *database.BloodhoundDB) int32
		assert func(t *testing.T, db *database.BloodhoundDB, finding model.SchemaFinding, err error, args args)
	}

	tt := []testData{
		{
			name: "Success: creates finding and remediation rows with stable id",
			args: args{
				name:                 "TestFinding",
				displayName:          "Test Finding",
				relationshipKindName: "HasSession",
				environmentKindName:  "Tag_Tier_Zero",
				remediation: model.RemediationInput{
					ShortDescription: "short desc",
					LongDescription:  "long desc",
					ShortRemediation: "short rem",
					LongRemediation:  "long rem",
				},
			},
			setup: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "CreateFindingExt", "Test", "v1.0.0", "test_ns")
				require.NoError(t, err)
				envInput := model.EnvironmentInput{
					EnvironmentKindName: "Tag_Tier_Zero",
					SourceKindName:      "Base",
					PrincipalKinds:      []string{},
				}
				_, err = db.CreateEnvironmentWithPrincipalKinds(context.Background(), ext.ID, envInput)
				require.NoError(t, err)
				return ext.ID
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, finding model.SchemaFinding, err error, args args) {
				t.Helper()
				require.NoError(t, err)
				assert.NotZero(t, finding.ID)
				assert.Equal(t, args.name, finding.Name)
				assert.Equal(t, args.displayName, finding.DisplayName)
				assert.Equal(t, model.SchemaFindingTypeRelationship, finding.Type)
			},
		},
		{
			name: "Error: invalid relationship kind name returns error",
			args: args{
				name:                 "BadFinding",
				displayName:          "Bad Finding",
				relationshipKindName: "KindThatDoesNotExist",
				environmentKindName:  "Tag_Tier_Zero",
			},
			setup: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "CreateFindingBadKind", "Test", "v1.0.0", "test_ns2")
				require.NoError(t, err)
				return ext.ID
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, finding model.SchemaFinding, err error, args args) {
				t.Helper()
				require.Error(t, err)
				assert.Zero(t, finding.ID)
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			extensionId := testCase.setup(t, testSuite.BHDatabase)
			input := model.RelationshipFindingInput{
				Name:                 testCase.args.name,
				DisplayName:          testCase.args.displayName,
				RelationshipKindName: testCase.args.relationshipKindName,
				EnvironmentKindName:  testCase.args.environmentKindName,
				RemediationInput:     testCase.args.remediation,
			}

			finding, err := testSuite.BHDatabase.CreateFindingWithRemediation(testSuite.Context, extensionId, input)
			testCase.assert(t, testSuite.BHDatabase, finding, err, testCase.args)
		})
	}
}

func TestUpdateFindingWithRemediation(t *testing.T) {
	t.Parallel()

	type args struct {
		newDisplayName       string
		relationshipKindName string
		environmentKindName  string
		remediation          model.RemediationInput
	}
	type testData struct {
		name   string
		args   args
		setup  func(t *testing.T, db *database.BloodhoundDB) (int32, model.SchemaFinding)
		assert func(t *testing.T, db *database.BloodhoundDB, existing model.SchemaFinding, updated model.SchemaFinding, err error, args args)
	}

	tt := []testData{
		{
			name: "Success: updates display name and remediation, preserves id and name",
			args: args{
				newDisplayName:       "Updated Display",
				relationshipKindName: "HasSession",
				environmentKindName:  "Tag_Tier_Zero",
				remediation: model.RemediationInput{
					ShortDescription: "updated short",
					LongDescription:  "updated long",
					ShortRemediation: "updated rem short",
					LongRemediation:  "updated rem long",
				},
			},
			setup: func(t *testing.T, db *database.BloodhoundDB) (int32, model.SchemaFinding) {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "UpdateFindingExt", "Test", "v1.0.0", "test_ns3")
				require.NoError(t, err)
				envInput := model.EnvironmentInput{
					EnvironmentKindName: "Tag_Tier_Zero",
					SourceKindName:      "Base",
					PrincipalKinds:      []string{},
				}
				_, err = db.CreateEnvironmentWithPrincipalKinds(context.Background(), ext.ID, envInput)
				require.NoError(t, err)
				createInput := model.RelationshipFindingInput{
					Name:                 "UpdateableFinding",
					DisplayName:          "Original Display",
					RelationshipKindName: "HasSession",
					EnvironmentKindName:  "Tag_Tier_Zero",
				}
				finding, err := db.CreateFindingWithRemediation(context.Background(), ext.ID, createInput)
				require.NoError(t, err)
				return ext.ID, finding
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, existing model.SchemaFinding, updated model.SchemaFinding, err error, args args) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, existing.ID, updated.ID)
				assert.Equal(t, existing.Name, updated.Name)
				assert.Equal(t, args.newDisplayName, updated.DisplayName)
				assert.Equal(t, model.SchemaFindingTypeRelationship, updated.Type)
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			_, finding := testCase.setup(t, testSuite.BHDatabase)
			input := model.RelationshipFindingInput{
				Name:                 finding.Name,
				DisplayName:          testCase.args.newDisplayName,
				RelationshipKindName: testCase.args.relationshipKindName,
				EnvironmentKindName:  testCase.args.environmentKindName,
				RemediationInput:     testCase.args.remediation,
			}

			updated, err := testSuite.BHDatabase.UpdateFindingWithRemediation(testSuite.Context, finding, input)
			testCase.assert(t, testSuite.BHDatabase, finding, updated, err, testCase.args)
		})
	}
}
