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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCustomNodeKinds(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() IntegrationTestSuite
		input   model.CustomNodeKinds
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

			kinds, err := testSuite.BHDatabase.GetCustomNodeKinds(testSuite.Context)
			if testCase.wantErr != nil {
				assert.EqualError(t, err, testCase.wantErr.Error())
			} else {
				require.NoError(t, err)
				assert.Len(t, kinds, testCase.wantLen)
			}
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
	tests := []struct {
		name    string
		setup   func() IntegrationTestSuite
		input   model.CustomNodeKind
		wantErr error
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
				assert.Equal(t, testCase.input.KindName, updated.KindName)
				assert.Equal(t, testCase.input.Config.Icon.Name, updated.Config.Icon.Name)
				assert.Equal(t, testCase.input.Config.Icon.Color, updated.Config.Icon.Color)
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
