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
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEnvironmentWithPrincipalKinds(t *testing.T) {
	t.Parallel()

	type args struct {
		environmentKindName string
		sourceKindName      string
		principalKinds      []string
	}
	type testData struct {
		name   string
		args   args
		setup  func(t *testing.T, db *database.BloodhoundDB) int32
		assert func(t *testing.T, db *database.BloodhoundDB, env model.SchemaEnvironment, err error, args args)
	}

	tt := []testData{
		{
			name: "Success: creates environment row and all principal kind rows",
			args: args{
				environmentKindName: "Tag_Tier_Zero",
				sourceKindName:      "Base",
				principalKinds:      []string{"Tag_Tier_Zero", "Tag_Owned"},
			},
			setup: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "CreateEnvWithPK", "Test", "v1.0.0", "test_ns")
				require.NoError(t, err)
				return ext.ID
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, env model.SchemaEnvironment, err error, args args) {
				t.Helper()
				require.NoError(t, err)
				assert.NotZero(t, env.ID)

				envKind, err := db.GetKindByName(context.Background(), args.environmentKindName)
				require.NoError(t, err)
				assert.Equal(t, envKind.ID, env.EnvironmentKindId)

				sourceKind, err := db.GetKindByName(context.Background(), args.sourceKindName)
				require.NoError(t, err)
				assert.Equal(t, sourceKind.ID, env.SourceKindId)

				principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), env.ID)
				require.NoError(t, err)
				require.Len(t, principalKinds, len(args.principalKinds))
				for i, pk := range principalKinds {
					expectedKind, err := db.GetKindByName(context.Background(), args.principalKinds[i])
					require.NoError(t, err)
					assert.Equal(t, expectedKind.ID, pk.PrincipalKind)
				}
			},
		},
		{
			name: "Error: invalid environment kind name returns error and no environment is created",
			args: args{
				environmentKindName: "DoesNotExist",
				sourceKindName:      "Base",
				principalKinds:      []string{},
			},
			setup: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "CreateEnvBadKind", "Test", "v1.0.0", "test_ns2")
				require.NoError(t, err)
				return ext.ID
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, env model.SchemaEnvironment, err error, args args) {
				t.Helper()
				require.Error(t, err)
				assert.Zero(t, env.ID)
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			extensionId := testCase.setup(t, testSuite.BHDatabase)
			input := model.EnvironmentInput{
				EnvironmentKindName: testCase.args.environmentKindName,
				SourceKindName:      testCase.args.sourceKindName,
				PrincipalKinds:      testCase.args.principalKinds,
			}

			env, err := testSuite.BHDatabase.CreateEnvironmentWithPrincipalKinds(testSuite.Context, extensionId, input)
			testCase.assert(t, testSuite.BHDatabase, env, err, testCase.args)
		})
	}
}

func TestUpdateEnvironmentWithPrincipalKinds(t *testing.T) {
	t.Parallel()

	type args struct {
		newSourceKindName string
		newPrincipalKinds []string
	}
	type testData struct {
		name   string
		args   args
		setup  func(t *testing.T, db *database.BloodhoundDB) (int32, model.SchemaEnvironment)
		assert func(t *testing.T, db *database.BloodhoundDB, updated model.SchemaEnvironment, err error, args args)
	}

	tt := []testData{
		{
			name: "Error: unknown principal kind name fails to build principal kinds",
			args: args{
				newSourceKindName: "Base",
				newPrincipalKinds: []string{"NonExistentKind"},
			},
			setup: func(t *testing.T, db *database.BloodhoundDB) (int32, model.SchemaEnvironment) {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "UpdateEnvBadPKExt", "Test", "v1.0.0", "test_ns4")
				require.NoError(t, err)
				env, err := db.CreateEnvironmentWithPrincipalKinds(context.Background(), ext.ID, model.EnvironmentInput{
					EnvironmentKindName: "Tag_Tier_Zero",
					SourceKindName:      "Base",
					PrincipalKinds:      []string{"Tag_Tier_Zero"},
				})
				require.NoError(t, err)
				return ext.ID, env
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, updated model.SchemaEnvironment, err error, args args) {
				t.Helper()
				require.ErrorContains(t, err, "error retrieving principal kind 'NonExistentKind'")
				assert.Zero(t, updated.ID)
			},
		},
		{
			name: "Error: duplicate principal kind names cause reconcile failure",
			args: args{
				newSourceKindName: "Base",
				newPrincipalKinds: []string{"Tag_Tier_Zero", "Tag_Tier_Zero"},
			},
			setup: func(t *testing.T, db *database.BloodhoundDB) (int32, model.SchemaEnvironment) {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "UpdateEnvDupePKExt", "Test", "v1.0.0", "test_ns5")
				require.NoError(t, err)
				env, err := db.CreateEnvironmentWithPrincipalKinds(context.Background(), ext.ID, model.EnvironmentInput{
					EnvironmentKindName: "Tag_Tier_Zero",
					SourceKindName:      "Base",
					PrincipalKinds:      []string{},
				})
				require.NoError(t, err)
				return ext.ID, env
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, updated model.SchemaEnvironment, err error, args args) {
				t.Helper()
				require.Error(t, err)
				assert.Zero(t, updated.ID)
			},
		},
		{
			name: "Success: adds new principal kind and removes stale principal kind",
			args: args{
				newSourceKindName: "Base",
				newPrincipalKinds: []string{"Tag_Tier_Zero"},
			},
			setup: func(t *testing.T, db *database.BloodhoundDB) (int32, model.SchemaEnvironment) {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "UpdateEnvPKExt", "Test", "v1.0.0", "test_ns3")
				require.NoError(t, err)
				input := model.EnvironmentInput{
					EnvironmentKindName: "Tag_Tier_Zero",
					SourceKindName:      "Base",
					PrincipalKinds:      []string{"Tag_Tier_Zero", "Tag_Owned"},
				}
				env, err := db.CreateEnvironmentWithPrincipalKinds(context.Background(), ext.ID, input)
				require.NoError(t, err)
				return ext.ID, env
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, updated model.SchemaEnvironment, err error, args args) {
				t.Helper()
				require.NoError(t, err)
				assert.NotZero(t, updated.ID)

				sourceKind, err := db.GetKindByName(context.Background(), args.newSourceKindName)
				require.NoError(t, err)
				assert.Equal(t, sourceKind.ID, updated.SourceKindId)

				principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), updated.ID)
				require.NoError(t, err)
				require.Len(t, principalKinds, len(args.newPrincipalKinds))
				for i, pk := range principalKinds {
					expectedKind, err := db.GetKindByName(context.Background(), args.newPrincipalKinds[i])
					require.NoError(t, err)
					assert.Equal(t, expectedKind.ID, pk.PrincipalKind)
				}
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			_, env := testCase.setup(t, testSuite.BHDatabase)
			input := model.EnvironmentInput{
				EnvironmentKindName: "Tag_Tier_Zero",
				SourceKindName:      testCase.args.newSourceKindName,
				PrincipalKinds:      testCase.args.newPrincipalKinds,
			}

			updated, err := testSuite.BHDatabase.UpdateEnvironmentWithPrincipalKinds(testSuite.Context, env, input)
			testCase.assert(t, testSuite.BHDatabase, updated, err, testCase.args)
		})
	}
}
