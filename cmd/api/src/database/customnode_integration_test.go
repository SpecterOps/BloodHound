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
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCustomNodeKinds(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() IntegrationTestSuite
		input   model.CustomNodeKinds
		wantMap model.CustomNodeKindMap
		wantErr error
	}{
		{
			name: "success - create single kind",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			input: model.CustomNodeKinds{
				{
					KindName: "TestKind",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "coffee", Color: "#FFFFFF"},
					},
				},
			},
			wantMap: model.CustomNodeKindMap{"TestKind": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "coffee", Color: "#FFFFFF"}}},
			wantErr: nil,
		},
		{
			name: "success - create multiple kinds",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			input: model.CustomNodeKinds{
				{
					KindName: "TestKindA",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "house", Color: "#000000"},
					},
				},
				{
					KindName: "TestKindB",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star", Color: "#FF0000"},
					},
				},
			},
			wantMap: model.CustomNodeKindMap{"TestKindA": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "house", Color: "#000000"}}, "TestKindB": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star", Color: "#FF0000"}}},
			wantErr: nil,
		},

		{
			name: "success - create multiple kinds with some missing fields",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			input: model.CustomNodeKinds{
				{
					KindName: "TestKindA",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{Type: "font-awesome", Color: "#000000"},
					},
				},
				{
					KindName: "TestKindB",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star"},
					},
				},
			},
			wantMap: model.CustomNodeKindMap{"TestKindA": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Color: "#000000"}}, "TestKindB": model.CustomNodeKindConfig{Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star"}}},
			wantErr: nil,
		},
		{
			name: "fail - duplicate kind name",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)
				_, err := testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, model.CustomNodeKinds{
					{
						KindName: "DuplicateKind",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "coffee", Color: "#FFFFFF"},
						},
					},
				})
				require.NoError(t, err)
				return testSuite
			},
			input: model.CustomNodeKinds{
				{
					KindName: "DuplicateKind",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "coffee", Color: "#FFFFFF"},
					},
				},
			},
			wantErr: database.ErrDuplicateCustomNodeKindName,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			created, err := testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, testCase.input)
			if testCase.wantErr != nil {
				require.ErrorIs(t, err, testCase.wantErr)
			} else {
				require.NoError(t, err)
				require.Len(t, created, len(testCase.input))
				for _, kind := range created {
					assert.Equal(t, testCase.wantMap[kind.KindName], kind.Config)
				}
			}
		})
	}
}

func TestGetCustomNodeKinds(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() IntegrationTestSuite
		wantLen int
		wantErr error
	}{
		{
			name: "success - returns empty list when none exist",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name: "success - returns all created kinds",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)
				_, err := testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, model.CustomNodeKinds{
					{
						KindName: "KindOne",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "fire", Color: "#FF5733"},
						},
					},
					{
						KindName: "KindTwo",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star", Color: "#FFFF00"},
						},
					},
				})
				require.NoError(t, err)
				return testSuite
			},
			wantLen: 2,
			wantErr: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			kinds, err := testSuite.BHDatabase.GetCustomNodeKinds(testSuite.Context, nil)
			if testCase.wantErr != nil {
				assert.EqualError(t, err, testCase.wantErr.Error())
			} else {
				require.NoError(t, err)
				assert.Len(t, kinds, testCase.wantLen)
			}
		})
	}
}

func TestGetCustomNodeKindsMap(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() IntegrationTestSuite
		wantLen int
	}{
		{
			name: "success - retrieves nil when OG disabled",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)
				flag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, appcfg.FeatureOpenGraphSearch)
				require.NoError(t, err)
				flag.Enabled = false
				require.NoError(t, testSuite.BHDatabase.SetFlag(testSuite.Context, flag))

				_, err = testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, model.CustomNodeKinds{
					{
						KindName: "RetrievableKind",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "bell", Color: "#123456"},
						},
					},
				})
				require.NoError(t, err)
				return testSuite
			},
			wantLen: 0,
		},
		{
			name: "success - retrieves kind map when OG enabled",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)
				flag, err := testSuite.BHDatabase.GetFlagByKey(testSuite.Context, appcfg.FeatureOpenGraphSearch)
				require.NoError(t, err)
				flag.Enabled = true
				require.NoError(t, testSuite.BHDatabase.SetFlag(testSuite.Context, flag))

				_, err = testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, model.CustomNodeKinds{
					{
						KindName: "RetrievableKind",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "bell", Color: "#123456"},
						},
					},
				})
				require.NoError(t, err)
				return testSuite
			},
			wantLen: 1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			customNodeKindMap, err := testSuite.BHDatabase.GetCustomNodeKindsMap(testSuite.Context)
			require.NoError(t, err)
			assert.Len(t, customNodeKindMap, testCase.wantLen)
		})
	}
}

func TestGetCustomNodeKind(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IntegrationTestSuite
		kindName string
		wantErr  error
	}{
		{
			name: "fail - kind not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			kindName: "NonExistentKind",
			wantErr:  database.ErrNotFound,
		},
		{
			name: "success - retrieves kind by name",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)
				_, err := testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, model.CustomNodeKinds{
					{
						KindName: "RetrievableKind",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "bell", Color: "#123456"},
						},
					},
				})
				require.NoError(t, err)
				return testSuite
			},
			kindName: "RetrievableKind",
			wantErr:  nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			kind, err := testSuite.BHDatabase.GetCustomNodeKind(testSuite.Context, testCase.kindName)
			if testCase.wantErr != nil {
				require.ErrorIs(t, err, testCase.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.kindName, kind.KindName)
			}
		})
	}
}

func TestUpdateCustomNodeKind(t *testing.T) {
	var schemaNodeKindID int32
	tests := []struct {
		name    string
		setup   func() IntegrationTestSuite
		input   model.CustomNodeKind
		wantErr error
		want    struct {
			CustomNodeKind model.CustomNodeKind
			SchemaNodeKind model.GraphSchemaNodeKind
		}
	}{
		{
			name: "fail - kind not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			input: model.CustomNodeKind{
				KindName: "NonExistentKind",
				Config: model.CustomNodeKindConfig{
					Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "bell", Color: "#FFFFFF"},
				},
			},
			wantErr: database.ErrNotFound,
		},
		{
			name: "success - updates kind config",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)
				_, err := testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, model.CustomNodeKinds{
					{
						KindName: "UpdatableKind",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "coffee", Color: "#FFFFFF"},
						},
					},
				})
				require.NoError(t, err)
				return testSuite
			},
			input: model.CustomNodeKind{
				KindName: "UpdatableKind",
				Config: model.CustomNodeKindConfig{
					Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star", Color: "#000000"},
				},
			},
			want: struct {
				CustomNodeKind model.CustomNodeKind
				SchemaNodeKind model.GraphSchemaNodeKind
			}{
				CustomNodeKind: model.CustomNodeKind{
					KindName: "UpdatableKind",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star", Color: "#000000"},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "success - updates kind config and schema node kind icon config",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)
				createdExtension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "Test Extension",
					"v1.0.0", "test_namespace")
				require.NoError(t, err)
				schemaNodeKind, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, "UpdatableKind", createdExtension.ID, "Test Kind", "A test kind", true, "coffee", "#FFFFFF")
				require.NoError(t, err)
				schemaNodeKindID = schemaNodeKind.ID
				_, err = testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, model.CustomNodeKinds{
					{
						KindName:         "UpdatableKind",
						SchemaNodeKindId: &schemaNodeKindID,
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "coffee", Color: "#FFFFFF"},
						},
					},
				})
				require.NoError(t, err)

				return testSuite
			},
			input: model.CustomNodeKind{
				KindName:         "UpdatableKind",
				SchemaNodeKindId: &schemaNodeKindID,
				Config: model.CustomNodeKindConfig{
					Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star", Color: "#000000"},
				},
			},
			want: struct {
				CustomNodeKind model.CustomNodeKind
				SchemaNodeKind model.GraphSchemaNodeKind
			}{
				CustomNodeKind: model.CustomNodeKind{
					KindName: "UpdatableKind",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "star", Color: "#000000"},
					},
				},
				SchemaNodeKind: model.GraphSchemaNodeKind{
					IconColor: "#000000",
					Icon:      "star",
				},
			},
			wantErr: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			updated, err := testSuite.BHDatabase.UpdateCustomNodeKind(testSuite.Context, testCase.input)
			if testCase.wantErr != nil {
				require.ErrorIs(t, err, testCase.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want.CustomNodeKind.Config.Icon, updated.Config.Icon)
				if testCase.want.SchemaNodeKind.Icon != "" || testCase.want.SchemaNodeKind.IconColor != "" {
					require.NotNil(t, updated.SchemaNodeKindId)
					gotSchemaNodeKind, err := testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, *updated.SchemaNodeKindId)
					require.NoError(t, err)
					assert.Equal(t, testCase.want.SchemaNodeKind.Icon, gotSchemaNodeKind.Icon)
					assert.Equal(t, testCase.want.SchemaNodeKind.IconColor, gotSchemaNodeKind.IconColor)
				}
			}
		})
	}
}

func TestDeleteCustomNodeKind(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IntegrationTestSuite
		kindName string
		wantErr  error
	}{
		{
			name: "fail - kind not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			kindName: "NonExistentKind",
			wantErr:  database.ErrNotFound,
		},
		{
			name: "success - deletes existing kind",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)
				_, err := testSuite.BHDatabase.CreateCustomNodeKinds(testSuite.Context, model.CustomNodeKinds{
					{
						KindName: "DeletableKind",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{Type: "font-awesome", Name: "trash", Color: "#FF0000"},
						},
					},
				})
				require.NoError(t, err)
				return testSuite
			},
			kindName: "DeletableKind",
			wantErr:  nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			err := testSuite.BHDatabase.DeleteCustomNodeKind(testSuite.Context, testCase.kindName)
			if testCase.wantErr != nil {
				require.ErrorIs(t, err, testCase.wantErr)
			} else {
				require.NoError(t, err)
				_, getErr := testSuite.BHDatabase.GetCustomNodeKind(testSuite.Context, testCase.kindName)
				require.ErrorIs(t, getErr, database.ErrNotFound)
			}
		})
	}
}
