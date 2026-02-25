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

// the v7.3.0 migration initializes the kind table with Tag_Tier_Zero, so we're
// simply testing the kind exists
func TestGetKindByName(t *testing.T) {
	type args struct {
		name string
	}
	type want struct {
		err  error
		kind model.Kind
	}
	tests := []struct {
		name  string
		args  args
		setup func() IntegrationTestSuite
		want  want
	}{
		{
			name: "Success: Retrieves Kind Tag_Tier_Zero by name",
			args: args{
				name: "Tag_Tier_Zero",
			},
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			want: want{
				err: nil,
				kind: model.Kind{
					ID:   1,
					Name: "Tag_Tier_Zero",
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			kind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, testCase.args.name)
			if testCase.want.err != nil {
				assert.EqualError(t, err, testCase.want.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want.kind, kind)
			}
		})
	}
}

func TestGetKindsByIDs(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	type want struct {
		err  error
		kind model.Kind
	}

	tests := []struct {
		name  string
		setup func(*testing.T) model.Kind
		want  want
	}{
		{
			name: "fail - unknown kind",
			setup: func(t *testing.T) model.Kind {
				t.Helper()
				return model.Kind{
					ID: 2141,
				}
			},
			want: want{
				err: database.ErrNotFound,
			},
		},
		{
			name: "success - single kind",
			setup: func(t *testing.T) model.Kind {
				t.Helper()

				var kind model.Kind
				result := testSuite.DB.WithContext(testSuite.Context).Raw(`
					INSERT INTO kind (name)
					VALUES ('Test_Get_Kinds_By_IDs')
					RETURNING id, name;`).Scan(&kind)
				require.NoError(t, result.Error)
				return kind
			},
			want: want{
				kind: model.Kind{
					Name: "Test_Get_Kinds_By_IDs",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdKind := tt.setup(t)

			if kinds, err := testSuite.BHDatabase.GetKindsByIDs(testSuite.Context, createdKind.ID); tt.want.err != nil {
				assert.EqualError(t, err, tt.want.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, kinds, 1)
				assert.Equal(t, tt.want.kind.Name, kinds[0].Name)
				assert.Greater(t, kinds[0].ID, int32(0))
			}
		})
	}

	t.Run("success - multiple kinds", func(t *testing.T) {
		var kindA, kindB model.Kind

		result := testSuite.DB.WithContext(testSuite.Context).Raw(`
			INSERT INTO kind (name)
			VALUES ('Test_Multiple_Kind_A')
			RETURNING id, name;`).Scan(&kindA)
		require.NoError(t, result.Error)

		result = testSuite.DB.WithContext(testSuite.Context).Raw(`
			INSERT INTO kind (name)
			VALUES ('Test_Multiple_Kind_B')
			RETURNING id, name;`).Scan(&kindB)
		require.NoError(t, result.Error)

		kinds, err := testSuite.BHDatabase.GetKindsByIDs(testSuite.Context, kindA.ID, kindB.ID)
		require.NoError(t, err)
		require.Len(t, kinds, 2)

		kindNames := []string{kinds[0].Name, kinds[1].Name}
		assert.Contains(t, kindNames, "Test_Multiple_Kind_A")
		assert.Contains(t, kindNames, "Test_Multiple_Kind_B")
	})

	t.Run("success - duplicate IDs returns deduplicated results", func(t *testing.T) {
		var kind model.Kind

		result := testSuite.DB.WithContext(testSuite.Context).Raw(`
			INSERT INTO kind (name)
			VALUES ('Test_Dedupe_Kind')
			RETURNING id, name;`).Scan(&kind)
		require.NoError(t, result.Error)

		kinds, err := testSuite.BHDatabase.GetKindsByIDs(testSuite.Context, kind.ID, kind.ID)
		require.NoError(t, err)
		require.Len(t, kinds, 1)
		assert.Equal(t, "Test_Dedupe_Kind", kinds[0].Name)
	})
}
